package discordapi

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log/level"

	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/constants"
	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/etfapi"
	"github.com/gsmcwhirter/eso-discord/pkg/logging"
	"github.com/gsmcwhirter/eso-discord/pkg/wsclient"
)

type discordMessageHandler struct {
	deps           dependencies
	wsCodeDispatch map[constants.ResponseCode]wsclient.MessageHandler
	opCodeDispatch map[constants.OpCode]wsclient.MessageHandler
}

// NewDiscordMessageHandler TODOC
func newDiscordMessageHandler(deps dependencies) discordMessageHandler {
	c := discordMessageHandler{
		deps: deps,
	}

	c.wsCodeDispatch = map[constants.ResponseCode]wsclient.MessageHandler{
		constants.UnknownError:         nil,
		constants.UnknownOpcode:        nil,
		constants.DecodeError:          nil,
		constants.NotAuthenticated:     nil,
		constants.AuthenticationFailed: nil,
		constants.AlreadyAuthenticated: nil,
		constants.InvalidSequence:      nil,
		constants.RateLimited:          nil,
		constants.SessionTimeout:       nil,
		constants.InvalidShard:         nil,
		constants.ShardingRequired:     nil,
	}

	c.opCodeDispatch = map[constants.OpCode]wsclient.MessageHandler{
		constants.Hello:           nil,
		constants.Heartbeat:       nil,
		constants.HeartbeatAck:    nil,
		constants.InvalidSession:  nil,
		constants.InvalidSequence: nil,
		constants.Reconnect:       nil,
		constants.Dispatch:        nil,
	}

	return c
}

func (c discordMessageHandler) FormatHeartbeat(ctx context.Context, lastSeq *int) wsclient.WSMessage {
	p := etfapi.Payload{
		OpCode: constants.Heartbeat,
	}
	fmt.Println(p)
	return wsclient.WSMessage{
		Ctx: ctx,
	}
}

func (c discordMessageHandler) HandleRequest(req wsclient.WSMessage, resp chan wsclient.WSMessage) {
	logger := logging.WithContext(req.Ctx, c.deps.Logger())

	errHandler, ok := c.wsCodeDispatch[constants.ResponseCode(req.MessageType)]
	if ok {
		level.Debug(logger).Log("message", "sending request to a websocket error handler")
		errHandler.HandleRequest(req, resp)
		return
	}

	level.Debug(logger).Log("message", "handling message start")

	p, err := etfapi.Unmarshal(req.MessageContents)
	if err != nil {
		level.Error(logger).Log("message", "error unmarshaling payload", "error", err)
		return
	}

	level.Debug(logger).Log("message", "received payload", "payload", p)

	opHandler, ok := c.opCodeDispatch[p.OpCode]
	if !ok {
		level.Error(logger).Log("message", "unrecognized OpCode", "op_code", p.OpCode)
		return
	}

	opHandler.HandleRequest(req, resp)

	// resp <- wsclient.WSMessage{
	// 	Ctx: req.Ctx,
	// }
}

func (c discordMessageHandler) handleHello(req wsclient.WSMessage, resp chan wsclient.WSMessage) {
	// set heartbeat stuff
	// send identify
}
