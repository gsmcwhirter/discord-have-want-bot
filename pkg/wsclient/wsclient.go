package wsclient

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/websocket"

	"github.com/gsmcwhirter/eso-discord/pkg/logging"
	"github.com/gsmcwhirter/eso-discord/pkg/util"
)

// MessageHandler TODOC
type MessageHandler interface {
	HandleRequest(req WSMessage) WSMessage
}

// WSMessage TODOC
type WSMessage struct {
	Ctx             context.Context
	MessageType     int
	MessageContents []byte
}

// WSClient TODOC
type WSClient interface {
	SetGateway(string)
	SetHandler(MessageHandler)
	Connect(string) error
	Close()
	HandleRequests(int)
	SendMessage(WSMessage)
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

	requests  chan WSMessage
	responses chan WSMessage

	interrupts chan struct{}
	closeAcks  chan struct{}
}

// Options TODOC
type Options struct {
	GatewayURL string
	Dialer     *websocket.Dialer
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

	level.Debug(logger).Log(
		"message", "ws client dial start",
		"url", c.gatewayURL,
	)

	start := time.Now()
	c.conn, dialResp, err = c.dialer.Dial(c.gatewayURL, dialHeader)

	level.Debug(logger).Log(
		"message", "ws client dial complete",
		"duration_ns", time.Since(start).Nanoseconds(),
		"status_code", dialResp.StatusCode,
	)

	if err != nil {
		return err
	}
	defer dialResp.Body.Close()

	return
}

func (c *wsClient) Close() {
	if c.conn != nil {
		defer util.CheckDefer(c.conn.Close)
	}
}

func (c *wsClient) createChannels() {
	c.requests = make(chan WSMessage)
	c.responses = make(chan WSMessage)
	c.interrupts = make(chan struct{})
	c.closeAcks = make(chan struct{})
}

func (c *wsClient) closeChannels() {
	close(c.requests)
	close(c.responses)
	close(c.interrupts)
	close(c.closeAcks)
}

func (c *wsClient) HandleRequests(handlers int) {
	level.Debug(c.deps.Logger()).Log("message", "creating channels")
	c.createChannels()
	defer c.closeChannels()

	level.Debug(c.deps.Logger()).Log("message", "starting request handlers", "num_handlers", handlers)
	for i := 0; i < handlers; i++ {
		go c.requestHandler()
	}

	level.Debug(c.deps.Logger()).Log("message", "setting signal watcher")
	interrupt := make(chan os.Signal, 1)
	defer close(interrupt)
	signal.Notify(interrupt, os.Interrupt)

	wsDone := make(chan struct{})
	defer close(wsDone)

	level.Debug(c.deps.Logger()).Log("message", "starting response handler")
	go c.handleResponses(handlers, wsDone, interrupt)
	level.Debug(c.deps.Logger()).Log("message", "starting message reader")
	go c.readMessages(wsDone, interrupt)

	level.Info(c.deps.Logger()).Log("message", "connected and listening")
	<-wsDone // block until one of the inner goroutines is done
	level.Info(c.deps.Logger()).Log("message", "shutting down")
	select { // wait up to 5 more seconds for the other one to finish
	case <-wsDone:
	case <-time.After(5 * time.Second):
	}
}

func (c wsClient) SendMessage(msg WSMessage) {
	logger := logging.WithContext(msg.Ctx, c.deps.Logger())
	level.Debug(logger).Log(
		"message", "adding message to response queue",
		"ws_msg_type", msg.MessageType,
		"ws_msg_len", len(msg.MessageContents))

	c.responses <- msg
}

func (c wsClient) readMessages(done chan struct{}, interrupt chan os.Signal) {
	logger := c.deps.Logger()
	for {
		select {
		case sig := <-interrupt:
			level.Info(logger).Log("message", "readMessages received interrupt -- shutting down")
			interrupt <- sig // allows other things to respond also
			done <- struct{}{}
			return
		default:

			msgType, msg, err := c.conn.ReadMessage()
			if err != nil {
				c.deps.Logger().Log(
					"message", "read error",
					"error", err,
				)
				break
			}

			ctx := util.NewRequestContext()
			wsMsg := WSMessage{Ctx: ctx, MessageType: msgType, MessageContents: msg}
			level.Debug(logging.WithContext(ctx, logger)).Log(
				"message", "received message",
				"ws_msg_type", msgType,
				"ws_msg_len", len(msg),
			)
			c.requests <- wsMsg
		}
	}
}

func (c wsClient) handleResponses(activeHandlers int, done chan struct{}, interrupt chan os.Signal) { //nolint: gocyclo
	logger := c.deps.Logger()
	for {
		select {
		case sig := <-interrupt: // time to stop
			level.Info(logger).Log("message", "handleResponses received interrupt -- shutting down")
			interrupt <- sig // allows other things to respond also

			level.Debug(logger).Log("message", "killing request handlers")
			for i := 0; i < activeHandlers; i++ { // close all the workers
				c.interrupts <- struct{}{}
			}

			level.Debug(logger).Log("message", "waiting for request handlers to die", "num_alive", activeHandlers)
			for activeHandlers > 0 { // wait for all the workers to close
				timeout := false

				select {
				case <-c.closeAcks:
					activeHandlers--
				case <-time.After(5 * time.Second):
					timeout = true
				}

				if timeout {
					break
				}
			}
			level.Debug(logger).Log("message", "request handlers dead", "num_alive", activeHandlers)

			// Close the socket connection gracefully
			level.Debug(logger).Log("message", "gracefully closing the socket")
			err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				level.Error(logger).Log("message", "Unable to write websocket close message", "error", err)
			}
			done <- struct{}{}
			return

		case <-c.closeAcks: // worker closed before time -- restart it
			level.Warn(logger).Log("message", "Unexpected request handler close ACK -- restarting")
			go c.requestHandler()
		case resp := <-c.responses: // handle pending responses
			level.Debug(logging.WithContext(resp.Ctx, logger)).Log(
				"message", "starting sending message",
				"ws_msg_type", resp.MessageType,
				"ws_msg_len", len(resp.MessageContents),
			)

			start := time.Now()
			err := c.conn.WriteMessage(resp.MessageType, resp.MessageContents)

			level.Debug(logging.WithContext(resp.Ctx, logger)).Log(
				"message", "done sending message",
				"elapsed_ns", time.Since(start).Nanoseconds(),
			)

			if err != nil {
				level.Error(logging.WithContext(resp.Ctx, logger)).Log(
					"message", "error sending message",
					"error", err,
				)
			}
		}
	}
}

func (c wsClient) requestHandler() {
	logger := c.deps.Logger()
	for {
		select {
		case <-c.interrupts:
			level.Info(logger).Log("message", "requestHandler received interrupt -- shutting down")
			c.closeAcks <- struct{}{}
			return
		case req := <-c.requests:
			level.Debug(logging.WithContext(req.Ctx, logger)).Log(
				"message", "requestHandler passing message to the registered handler",
			)
			c.responses <- c.handler.HandleRequest(req)
		}
	}
}
