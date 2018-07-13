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

type discordMessageHandlerFunc func(*etfapi.Payload, wsclient.WSMessage, chan<- wsclient.WSMessage, <-chan struct{})

type discordMessageHandler struct {
	deps              dependencies
	heartbeatReconfig chan hbReconfig
	wsCodeDispatch    map[constants.ResponseCode]wsclient.MessageHandlerFunc
	opCodeDispatch    map[constants.OpCode]discordMessageHandlerFunc
}

// NewDiscordMessageHandler TODOC
func newDiscordMessageHandler(deps dependencies, heartbeats chan hbReconfig) *discordMessageHandler {
	c := discordMessageHandler{
		deps:              deps,
		heartbeatReconfig: heartbeats,
	}

	c.wsCodeDispatch = map[constants.ResponseCode]wsclient.MessageHandlerFunc{
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

	c.opCodeDispatch = map[constants.OpCode]discordMessageHandlerFunc{
		constants.Hello:           c.handleHello,
		constants.Heartbeat:       nil,
		constants.HeartbeatAck:    nil,
		constants.InvalidSession:  nil,
		constants.InvalidSequence: nil,
		constants.Reconnect:       nil,
		constants.Dispatch:        nil,
	}

	return &c
}

func (c *discordMessageHandler) FormatHeartbeat(ctx context.Context, lastSeq *int) wsclient.WSMessage {
	p := etfapi.Payload{
		OpCode: constants.Heartbeat,
	}
	fmt.Println(p)
	return wsclient.WSMessage{
		Ctx: ctx,
	}
}

func (c *discordMessageHandler) HandleRequest(req wsclient.WSMessage, resp chan<- wsclient.WSMessage, done <-chan struct{}) {
	logger := logging.WithContext(req.Ctx, c.deps.Logger())
	_ = level.Debug(logger).Log("message", "discordapi dispatching request")

	select {
	case <-done:
		_ = level.Debug(logger).Log("message", "discordapi already done. skipping request")
		return
	default:
	}

	errHandler, ok := c.wsCodeDispatch[constants.ResponseCode(req.MessageType)]
	if ok {
		_ = level.Debug(logger).Log("message", "sending request to a websocket error handler")
		errHandler(req, resp, done)
		return
	}

	_ = level.Debug(logger).Log("message", "handling message start")

	p, err := etfapi.Unmarshal(req.MessageContents)
	if err != nil {
		_ = level.Error(logger).Log("message", "error unmarshaling payload", "error", err)
		return
	}

	_ = level.Debug(logger).Log("message", "received payload", "payload", p)

	opHandler, ok := c.opCodeDispatch[p.OpCode]
	if !ok {
		_ = level.Error(logger).Log("message", "unrecognized OpCode", "op_code", p.OpCode)
		return
	}

	if opHandler == nil {
		_ = level.Error(logger).Log("message", "no handler for OpCode", "op_code", p.OpCode)
		return
	}

	opHandler(p, req, resp, done)
}

func (c *discordMessageHandler) handleHello(p *etfapi.Payload, req wsclient.WSMessage, resp chan<- wsclient.WSMessage, done <-chan struct{}) {
	logger := logging.WithContext(req.Ctx, c.deps.Logger())
	rawInterval, ok := p.Data["heartbeat_interval"]

	if ok {
		// set heartbeat stuff
		var interval int
		err := rawInterval.Unmarshal(&interval)
		if err != nil {
			_ = level.Error(logger).Log("message", "error handling hello heartbeat config", "err", err)
			return
		}

		_ = level.Info(logger).Log("message", "configuring heartbeat", "interval", interval)
		c.heartbeatReconfig <- hbReconfig{
			ctx:      req.Ctx,
			interval: interval,
		}
		_ = level.Debug(logger).Log("message", "configuring heartbeat done")

	}

	// send identify
}
