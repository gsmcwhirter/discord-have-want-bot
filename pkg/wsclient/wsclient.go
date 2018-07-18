package wsclient

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/websocket"

	"github.com/gsmcwhirter/eso-discord/pkg/logging"
	"github.com/gsmcwhirter/eso-discord/pkg/util"
)

// MessageType TODOC
type MessageType int

// Websocket Message types
const (
	Text   MessageType = 1
	Binary             = 2
)

func (t MessageType) String() string {
	switch t {
	case Text:
		return "Text"
	case Binary:
		return "Binary"
	default:
		return fmt.Sprintf("(unknown: %d)", int(t))
	}
}

// WSMessage TODOC
type WSMessage struct {
	Ctx             context.Context
	MessageType     MessageType
	MessageContents []byte
}

func (m WSMessage) String() string {
	return fmt.Sprintf("WSMessage{Type=%v, Contents=%v}", m.MessageType, m.MessageContents)
}

// WSClient TODOC
type WSClient interface {
	SetGateway(string)
	SetHandler(MessageHandler)
	Connect(string) error
	Close()
	HandleRequests(interrupt chan os.Signal)
	SendMessage(msg WSMessage)
}

type dependencies interface {
	Logger() log.Logger
}

type wsClient struct {
	deps dependencies

	gatewayURL string
	dialer     *websocket.Dialer
	conn       *websocket.Conn
	handler    MessageHandler

	responses     chan WSMessage
	done          chan struct{}
	stopResponses chan struct{}

	controls   *sync.WaitGroup
	pool       *sync.WaitGroup
	poolTokens chan struct{}

	closeLock sync.Mutex
	isClosed  bool
}

// Options TODOC
type Options struct {
	GatewayURL            string
	Dialer                *websocket.Dialer
	MaxConcurrentHandlers int
}

// NewWSClient TODOC
func NewWSClient(deps dependencies, options Options) WSClient {
	c := &wsClient{
		deps:       deps,
		gatewayURL: options.GatewayURL,
	}

	if options.Dialer != nil {
		c.dialer = options.Dialer
	} else {
		c.dialer = websocket.DefaultDialer
	}

	c.done = make(chan struct{})
	c.stopResponses = make(chan struct{})

	c.pool = &sync.WaitGroup{}
	c.controls = &sync.WaitGroup{}
	if options.MaxConcurrentHandlers <= 0 {
		c.poolTokens = make(chan struct{}, 20)
		c.responses = make(chan WSMessage, 20)
	} else {
		c.poolTokens = make(chan struct{}, options.MaxConcurrentHandlers)
		c.responses = make(chan WSMessage, options.MaxConcurrentHandlers)
	}

	return c
}

func (c *wsClient) SetGateway(url string) {
	c.gatewayURL = url
}

func (c *wsClient) SetHandler(handler MessageHandler) {
	c.handler = handler
}

func (c *wsClient) Connect(token string) (err error) {
	ctx := util.NewRequestContext()
	logger := logging.WithContext(ctx, c.deps.Logger())

	dialHeader := http.Header{
		"Authorization": []string{fmt.Sprintf("Bot %s", token)},
	}

	var dialResp *http.Response

	_ = level.Debug(logger).Log(
		"message", "ws client dial start",
		"url", c.gatewayURL,
	)

	start := time.Now()
	c.conn, dialResp, err = c.dialer.Dial(c.gatewayURL, dialHeader)

	_ = level.Debug(logger).Log(
		"message", "ws client dial complete",
		"duration_ns", time.Since(start).Nanoseconds(),
		"status_code", dialResp.StatusCode,
	)

	if err != nil {
		return err
	}
	defer dialResp.Body.Close() // nolint: errcheck

	return
}

func (c *wsClient) ensureClosedChannels() {
	for range c.done {
	}

	select {
	case _, ok := <-c.done: // closed or something in it
		if ok {
			close(c.done)
		}
	default: // nothing in it but not closed
		close(c.done)
	}

	// TODO: responses? poolTokens?
}

func (c *wsClient) Close() {
	c.ensureClosedChannels()
	c.pool.Wait()
	c.controls.Wait()
	if c.conn != nil {
		c.conn.Close() // nolint: errcheck
	}
}

func (c *wsClient) HandleRequests(interrupt chan os.Signal) {
	_ = level.Debug(c.deps.Logger()).Log("message", "starting response handler")
	c.controls.Add(1)
	go c.handleResponses()

	_ = level.Debug(c.deps.Logger()).Log("message", "starting message reader")
	c.controls.Add(1)
	go c.readMessages(interrupt)

	_ = level.Info(c.deps.Logger()).Log("message", "connected and listening")

	c.controls.Wait()
	_ = level.Info(c.deps.Logger()).Log("message", "shutting down")
}

func (c *wsClient) gracefulClose() {
	c.closeLock.Lock()
	defer c.closeLock.Unlock()

	if c.isClosed {
		return
	}

	c.isClosed = true

	close(c.done)
	_ = c.conn.SetReadDeadline(time.Now())

	// Close the socket connection gracefully
	close(c.stopResponses)
}

func (c *wsClient) doReads(readerDone chan<- struct{}) {
	defer c.controls.Done()
	defer level.Info(c.deps.Logger()).Log("message", "websocket reader done") //nolint: errcheck
	defer close(readerDone)

	for {
		msgType, msg, err := c.conn.ReadMessage()
		if err != nil {
			_ = level.Error(c.deps.Logger()).Log(
				"message", "read error",
				"error", err,
				"ws_msg_type", msgType,
				"ws_content", msg,
			)
			return
		}

		ctx := util.NewRequestContext()
		mT := MessageType(msgType)
		mC := make([]byte, len(msg))
		copy(mC, msg)

		wsMsg := WSMessage{Ctx: ctx, MessageType: mT, MessageContents: mC}
		_ = level.Debug(logging.WithContext(ctx, c.deps.Logger())).Log(
			"message", "received message",
			"ws_msg_type", mT,
			"ws_msg_len", len(mC),
		)

		_ = level.Debug(logging.WithContext(ctx, c.deps.Logger())).Log(
			"message", "waiting for worker token",
		)
		c.poolTokens <- struct{}{}
		_ = level.Debug(logging.WithContext(ctx, c.deps.Logger())).Log(
			"message", "worker token acquired",
		)
		c.pool.Add(1)
		go c.handleRequest(wsMsg)
	}
}

func (c *wsClient) readMessages(interrupt chan os.Signal) {
	defer c.controls.Done()
	defer level.Info(c.deps.Logger()).Log("message", "readMessages shutdown complete") //nolint: errcheck

	readerDone := make(chan struct{})

	c.controls.Add(1)
	go c.doReads(readerDone)

	select {
	case <-readerDone:
		_ = level.Info(c.deps.Logger()).Log("message", "readMessages doReads error -- shutting down")
		interrupt <- os.Interrupt
		c.gracefulClose()
		return
	case sig := <-interrupt:
		_ = level.Info(c.deps.Logger()).Log("message", "readMessages received interrupt -- shutting down")
		interrupt <- sig // allows other things to respond also
		c.gracefulClose()
		return
	case <-c.done:
		_ = level.Info(c.deps.Logger()).Log("message", "readMessages received done -- shutting down")
		c.gracefulClose()
		return
	}
}

func (c *wsClient) handleRequest(req WSMessage) {
	defer c.pool.Done()

	logger := logging.WithContext(req.Ctx, c.deps.Logger())

	defer func() {
		<-c.poolTokens
		_ = level.Debug(logger).Log("message", "released worker token")
	}()

	select {
	case <-c.done:
		_ = level.Info(logger).Log("message", "handleRequest received interrupt -- shutting down")
		return
	default:
		_ = level.Debug(logger).Log("message", "handleRequest dispatching request")
		c.handler.HandleRequest(req, c.responses, c.done)
	}
}

func (c *wsClient) processResponse(resp WSMessage) {
	logger := logging.WithContext(resp.Ctx, c.deps.Logger())

	_ = level.Debug(logger).Log(
		"message", "starting sending message",
		"ws_msg_type", resp.MessageType,
		"ws_msg_len", len(resp.MessageContents),
	)

	start := time.Now()
	err := c.conn.WriteMessage(int(resp.MessageType), resp.MessageContents)

	_ = level.Debug(logging.WithContext(resp.Ctx, c.deps.Logger())).Log(
		"message", "done sending message",
		"elapsed_ns", time.Since(start).Nanoseconds(),
	)

	if err != nil {
		_ = level.Error(logger).Log(
			"message", "error sending message",
			"error", err,
		)
	}
}

func (c *wsClient) handleResponses() { //nolint: gocyclo
	defer c.controls.Done()
	defer level.Info(c.deps.Logger()).Log("message", "handleResponses shutdown complete") //nolint: errcheck

	for {
		select {
		case <-c.stopResponses: // time to stop
			_ = level.Info(c.deps.Logger()).Log("message", "handleResponses received interrupt -- shutting down")

			defer func() {
				_ = level.Debug(c.deps.Logger()).Log("message", "gracefully closing the socket")
				err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					_ = level.Error(c.deps.Logger()).Log("message", "Unable to write websocket close message", "error", err)
					return
				}
				_ = level.Debug(c.deps.Logger()).Log("message", "close message sent")
			}()

			// drain the remaining response queue
			for {
				select {
				case resp, ok := <-c.responses:
					if !ok {
						close(c.responses)
						return
					}
					c.processResponse(resp)
				case <-time.After(5 * time.Second):
					return
				}
			}

		case resp := <-c.responses: // handle pending responses
			c.processResponse(resp)
		}
	}
}

func (c *wsClient) SendMessage(msg WSMessage) {
	logger := logging.WithContext(msg.Ctx, c.deps.Logger())
	_ = level.Debug(logger).Log(
		"message", "adding message to response queue",
		"ws_msg_type", msg.MessageType,
		"ws_msg_len", len(msg.MessageContents),
	)

	c.responses <- msg
}
