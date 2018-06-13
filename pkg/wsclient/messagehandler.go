package wsclient

// MessageHandler TODOC
type MessageHandler interface {
	HandleRequest(WSMessage, chan WSMessage)
}

type messageHandler struct {
	handler func(WSMessage, chan WSMessage)
}

// NewMessageHandler TODOC
func NewMessageHandler(h func(WSMessage, chan WSMessage)) MessageHandler {
	return messageHandler{handler: h}
}

func (mh messageHandler) HandleRequest(req WSMessage, resp chan WSMessage) {
	mh.handler(req, resp)
}
