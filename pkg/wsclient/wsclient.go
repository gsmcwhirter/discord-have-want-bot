package wsclient

import (
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type MessageHandler interface {
}

type WSClient struct {
	url       url.URL
	heartbeat *time.Ticker
	conn      *websocket.Conn
	handler   MessageHandler

	readErr
}

func NewWSClient(host, path string, handler MessageHandler) WSClient {
	c := WSClient{
		url:     url.URL{Scheme: "ws", Host: host, Path: path},
		handler: handler,
	}

	return c
}

func (c *WSClient) Connect() (err error) {
	// nil here is a request header
	c.conn, _, err = websocket.DefaultDialer.Dial(c.url.String(), nil)
	return
}

func (c *WSClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *WSClient) Communicate() {

}
