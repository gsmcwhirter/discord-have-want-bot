package discordapi

import (
	"fmt"

	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/etfapi"
	"github.com/gsmcwhirter/eso-discord/pkg/wsclient"
)

type discordMessageHandler struct {
}

// NewDiscordMessageHandler TODOC
func NewDiscordMessageHandler() wsclient.MessageHandler {
	c := discordMessageHandler{}
	return c
}

func (c discordMessageHandler) FormatHeartbeat(lastSeq *int) wsclient.WSMessage {
	p := etfapi.Payload{
		OpCode: 1,
		Data:   []byte{},
	}
	fmt.Println(p)
	return wsclient.WSMessage{}
}

func (c discordMessageHandler) HandleRequest(req wsclient.WSMessage) wsclient.WSMessage {
	fmt.Printf("%+v\n", req)
	fmt.Println(string(req.MessageContents))
	p, err := etfapi.Unmarshal(req.MessageContents)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("payload %+v\n", p)
	return wsclient.WSMessage{}
}
