package wsclient

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gsmcwhirter/eso-discord/pkg/util"
)

// MessageHandler TODOC
type MessageHandler interface {
	HandleRequest(req WSMessage) WSMessage
	FormatHeartbeat(*int) WSMessage
}

// WSMessage TODOC
type WSMessage struct {
	MessageType     int
	MessageContents []byte
}

// WSClient TODOC
type WSClient interface {
	SetHandler(MessageHandler)
	Connect(string) error
	Close()
	HandleRequests(int)
}

type hbReconfig struct {
	interval int
}

type wsClient struct {
	url       string
	dialer    *websocket.Dialer
	heartbeat *time.Ticker
	conn      *websocket.Conn
	handler   MessageHandler

	heartbeats          chan hbReconfig
	heartbeatInterrupts chan struct{}
	heartbeatCloseAcks  chan struct{}

	requests  chan WSMessage
	responses chan WSMessage

	interrupts   chan struct{}
	closeAcks    chan struct{}
	lastSequence *int
}

// NewWSClient TODOC
func NewWSClient(hostURL string) WSClient {
	c := &wsClient{
		dialer: websocket.DefaultDialer,
		url:    hostURL,
	}

	return c
}

func (c *wsClient) SetHandler(handler MessageHandler) {
	c.handler = handler
}

func (c *wsClient) Connect(token string) (err error) {
	// nil here is a request header
	dialHeader := http.Header{
		"Authorization": []string{fmt.Sprintf("Bot %s", token)},
	}
	var dialResp *http.Response
	c.conn, dialResp, err = c.dialer.Dial(c.url, dialHeader)
	fmt.Printf("%+v\n", dialResp)
	_, body, err := util.ReadBody(dialResp.Body, 200)
	if err != nil {
		return err
	}
	fmt.Printf("body: %s\n", string(body))

	c.heartbeat = time.NewTicker(14000 * time.Millisecond)

	return
}

func (c *wsClient) Close() {
	if c.conn != nil {
		defer util.CheckDefer(c.conn.Close)
	}
}

func (c *wsClient) createChannels() {
	c.heartbeats = make(chan hbReconfig)
	c.heartbeatInterrupts = make(chan struct{})
	c.heartbeatCloseAcks = make(chan struct{})
	c.requests = make(chan WSMessage)
	c.responses = make(chan WSMessage)
	c.interrupts = make(chan struct{})
	c.closeAcks = make(chan struct{})
}

func (c *wsClient) closeChannels() {
	close(c.heartbeats)
	close(c.heartbeatInterrupts)
	close(c.heartbeatCloseAcks)
	close(c.requests)
	close(c.responses)
	close(c.interrupts)
	close(c.closeAcks)
}

func (c *wsClient) HandleRequests(handlers int) {
	// TODO: create heartbeat timer?
	defer c.heartbeat.Stop()
	c.createChannels()
	defer c.closeChannels()

	go c.heartbeatHandler()
	for i := 0; i < handlers; i++ {
		go c.requestHandler()
	}

	interrupt := make(chan os.Signal, 1)
	defer close(interrupt)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})
	defer close(done)

	go c.handleResponses(handlers, done, interrupt)
	go c.readMessages(done, interrupt)

	<-done   // block until one of the inner goroutines is done
	select { // wait up to 5 more seconds for the other one to finish
	case <-done:
	case <-time.After(5 * time.Second):
	}
}

func (c wsClient) readMessages(done chan struct{}, interrupt chan os.Signal) {
	for {
		select {
		case sig := <-interrupt:
			interrupt <- sig // allows other things to respond also
			done <- struct{}{}
			return
		default:

			msgType, msg, err := c.conn.ReadMessage()
			if err != nil {
				fmt.Printf("read error: %s\n", err)
				// TODO: real error handling here -- but we're depending on the read error to indicate
				break
			}

			// TODO: differentiate heartbeats from normal requests

			c.requests <- WSMessage{MessageType: msgType, MessageContents: msg}
		}
	}
}

func (c wsClient) handleResponses(activeHandlers int, done chan struct{}, interrupt chan os.Signal) { //nolint: gocyclo
	for {
		select {
		case sig := <-interrupt: // time to stop
			interrupt <- sig                      // allows other things to respond also
			c.heartbeatInterrupts <- struct{}{}   // close the heartbeat
			for i := 0; i < activeHandlers; i++ { // close all the workers
				c.interrupts <- struct{}{}
			}

			select { // wait for the heartbeat to close
			case <-c.heartbeatCloseAcks:
			case <-time.After(5 * time.Second):
			}

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

			// Close the socket connection gracefully
			err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to write websocket close message: %s\n", err) // nolint: errcheck,gas
			}
			done <- struct{}{}
			return

		case <-c.closeAcks: // worker closed before time -- restart it
			go c.requestHandler()
		case <-c.heartbeatCloseAcks: // heartbeater close before time -- restart it
			go c.heartbeatHandler()
		case resp := <-c.responses: // handle pending responses
			err := c.conn.WriteMessage(resp.MessageType, resp.MessageContents)
			if err != nil {
				// TODO: error handling
				fmt.Println(err)
			}
		}
	}
}

func (c *wsClient) heartbeatHandler() {
	for {
		select {
		case <-c.heartbeatInterrupts:
			c.heartbeatCloseAcks <- struct{}{}
			return
		case req := <-c.heartbeats:
			if req.interval > 0 {
				c.heartbeat.Stop()
				c.heartbeat = time.NewTicker(time.Duration(req.interval) * time.Millisecond)
			} else {
				c.responses <- c.handler.FormatHeartbeat(c.lastSequence)
			}
		case <-c.heartbeat.C:
			c.responses <- c.handler.FormatHeartbeat(c.lastSequence)
		}
	}
}

func (c wsClient) requestHandler() {
	for {
		select {
		case <-c.interrupts:
			c.closeAcks <- struct{}{}
			return
		case req := <-c.requests:
			c.responses <- c.handler.HandleRequest(req)
		}
	}
}
