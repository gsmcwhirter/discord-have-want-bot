package discordapi

import (
	"fmt"

	"github.com/go-kit/kit/log/level"

	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/constants"
	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/etfapi"
	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/etfapi/payloads"
	"github.com/gsmcwhirter/eso-discord/pkg/logging"
	"github.com/gsmcwhirter/eso-discord/pkg/wsclient"
)

type discordMessageHandlerFunc func(*etfapi.Payload, wsclient.WSMessage, chan<- wsclient.WSMessage, <-chan struct{})

type discordMessageHandler struct {
	bot               *discordBot
	deps              dependencies
	heartbeatReconfig chan hbReconfig
	wsCodeDispatch    map[constants.ResponseCode]wsclient.MessageHandlerFunc
	opCodeDispatch    map[constants.OpCode]discordMessageHandlerFunc
}

// NewDiscordMessageHandler TODOC
func newDiscordMessageHandler(bot *discordBot, deps dependencies, heartbeats chan hbReconfig) *discordMessageHandler {
	c := discordMessageHandler{
		bot:               bot,
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
		constants.Heartbeat:       c.handleHeartbeat,
		constants.HeartbeatAck:    noop,
		constants.InvalidSession:  nil,
		constants.InvalidSequence: nil,
		constants.Reconnect:       nil,
		constants.Dispatch:        c.handleDispatch,
	}

	return &c
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

	_ = level.Debug(logger).Log("message", "processing server message", "ws_msg", fmt.Sprintf("%v", []byte(req.MessageContents)))

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

func (c *discordMessageHandler) handleError(req wsclient.WSMessage, resp chan<- wsclient.WSMessage, done <-chan struct{}) {
	select {
	case <-done:
		return
	default:
	}

	logger := logging.WithContext(req.Ctx, c.deps.Logger())
	_ = level.Error(logger).Log("message", "error code received from websocket", "ws_msg", req)
}

func (c *discordMessageHandler) handleHello(p *etfapi.Payload, req wsclient.WSMessage, resp chan<- wsclient.WSMessage, done <-chan struct{}) {
	select {
	case <-done:
		return
	default:
	}

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
	ip := payloads.IdentifyPayload{
		Token: c.bot.config.BotToken,
		Properties: payloads.IdentifyPayloadProperties{
			OS:      "linux",
			Browser: "eso-have-want-bot#0286",
			Device:  "eso-have-want-bot#0286",
		},
		LargeThreshold: 250,
		Shard: payloads.IdentifyPayloadShard{
			ID:    0,
			MaxID: 0,
		},
		Presence: payloads.IdentifyPayloadPresence{
			Game: payloads.IdentifyPayloadGame{
				Name: "List Manager 2018",
				Type: 0,
			},
			Status: "online",
			Since:  0,
			AFK:    false,
		},
	}

	m, err := payloads.ETFPayloadToMessage(req.Ctx, ip)
	if err != nil {
		_ = level.Error(logger).Log("message", "error generating identify payload", "err", err)
	} else {
		_ = level.Debug(logger).Log("message", "sending response to channel", "message", m, "msg_len", len(m.MessageContents))
		resp <- m
	}
}

func (c *discordMessageHandler) handleHeartbeat(p *etfapi.Payload, req wsclient.WSMessage, resp chan<- wsclient.WSMessage, done <-chan struct{}) {
	select {
	case <-done:
		return
	default:
	}

	logger := logging.WithContext(req.Ctx, c.deps.Logger())
	_ = level.Info(logger).Log("message", "requesting manual heartbeat")
	c.heartbeatReconfig <- hbReconfig{
		ctx: req.Ctx,
	}
	_ = level.Debug(logger).Log("message", "manual heartbeat done")
}

func (c *discordMessageHandler) handleDispatch(p *etfapi.Payload, req wsclient.WSMessage, resp chan<- wsclient.WSMessage, done <-chan struct{}) {
	// TODO
	fmt.Println(p.PrettyString("", false))
}

func noop(p *etfapi.Payload, req wsclient.WSMessage, resp chan<- wsclient.WSMessage, done <-chan struct{}) {
}
