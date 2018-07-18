package discordapi

import (
	"fmt"

	"github.com/go-kit/kit/log/level"

	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/constants"
	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/etfapi"
	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/etfapi/payloads"
	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/jsonapi"
	"github.com/gsmcwhirter/eso-discord/pkg/logging"
	"github.com/gsmcwhirter/eso-discord/pkg/wsclient"
)

type discordMessageHandlerFunc func(*etfapi.Payload, wsclient.WSMessage, chan<- wsclient.WSMessage, <-chan struct{})

type discordMessageHandler struct {
	bot               *discordBot
	heartbeatReconfig chan hbReconfig
	opCodeDispatch    map[constants.OpCode]discordMessageHandlerFunc
	eventDispatch     map[string]discordMessageHandlerFunc
}

func noop(p *etfapi.Payload, req wsclient.WSMessage, resp chan<- wsclient.WSMessage, done <-chan struct{}) {
}

// NewDiscordMessageHandler TODOC
func newDiscordMessageHandler(bot *discordBot, heartbeats chan hbReconfig) *discordMessageHandler {
	c := discordMessageHandler{
		bot:               bot,
		heartbeatReconfig: heartbeats,
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

	c.eventDispatch = map[string]discordMessageHandlerFunc{
		"READY":          c.handleReady,
		"GUILD_CREATE":   c.handleGuildCreate,
		"MESSAGE_CREATE": c.handleMessage,
	}

	return &c
}

func (c *discordMessageHandler) HandleRequest(req wsclient.WSMessage, resp chan<- wsclient.WSMessage, done <-chan struct{}) {
	logger := logging.WithContext(req.Ctx, c.bot.deps.Logger())
	_ = level.Debug(logger).Log("message", "discordapi dispatching request")

	select {
	case <-done:
		_ = level.Debug(logger).Log("message", "discordapi already done. skipping request")
		return
	default:
	}

	_ = level.Debug(logger).Log("message", "processing server message", "ws_msg", fmt.Sprintf("%v", req.MessageContents))

	p, err := etfapi.Unmarshal(req.MessageContents)
	if err != nil {
		_ = level.Error(logger).Log("message", "error unmarshaling payload", "error", err)
		return
	}

	if p.SeqNum != nil {
		c.bot.UpdateSequence(*p.SeqNum)
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

	logger := logging.WithContext(req.Ctx, c.bot.deps.Logger())
	_ = level.Error(logger).Log("message", "error code received from websocket", "ws_msg", req)
}

func (c *discordMessageHandler) handleHello(p *etfapi.Payload, req wsclient.WSMessage, resp chan<- wsclient.WSMessage, done <-chan struct{}) {
	select {
	case <-done:
		return
	default:
	}

	logger := logging.WithContext(req.Ctx, c.bot.deps.Logger())
	rawInterval, ok := p.Data["heartbeat_interval"]

	if ok {
		// set heartbeat stuff
		interval, err := rawInterval.ToInt()
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
	var m wsclient.WSMessage
	var err error

	sessID := c.bot.session.ID()
	if sessID != "" {
		rp := payloads.ResumePayload{
			Token:     c.bot.config.BotToken,
			SessionID: sessID,
			SeqNum:    c.bot.LastSequence(),
		}

		m, err = payloads.ETFPayloadToMessage(req.Ctx, rp)
	} else {
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

		m, err = payloads.ETFPayloadToMessage(req.Ctx, ip)
	}

	if err != nil {
		_ = level.Error(logger).Log("message", "error generating identify payload", "err", err)
		return
	}

	err = c.bot.rateLimiter.Wait(req.Ctx)
	if err != nil {
		_ = level.Error(logger).Log("message", "error ratelimiting", "err", err)
		return
	}

	_ = level.Debug(logger).Log("message", "sending response to channel", "message", m, "msg_len", len(m.MessageContents))
	resp <- m
}

func (c *discordMessageHandler) handleHeartbeat(p *etfapi.Payload, req wsclient.WSMessage, resp chan<- wsclient.WSMessage, done <-chan struct{}) {
	select {
	case <-done:
		return
	default:
	}

	logger := logging.WithContext(req.Ctx, c.bot.deps.Logger())
	_ = level.Info(logger).Log("message", "requesting manual heartbeat")
	c.heartbeatReconfig <- hbReconfig{
		ctx: req.Ctx,
	}
	_ = level.Debug(logger).Log("message", "manual heartbeat done")
}

func (c *discordMessageHandler) handleDispatch(p *etfapi.Payload, req wsclient.WSMessage, resp chan<- wsclient.WSMessage, done <-chan struct{}) {
	select {
	case <-done:
		return
	default:
	}

	logger := logging.WithContext(req.Ctx, c.bot.deps.Logger())
	eventHandler, ok := c.eventDispatch[p.EventName]
	if ok {
		_ = level.Debug(logger).Log("message", "processing event", "event_name", p.EventName)
		eventHandler(p, req, resp, done)
	}
}

func (c *discordMessageHandler) handleReady(p *etfapi.Payload, req wsclient.WSMessage, resp chan<- wsclient.WSMessage, done <-chan struct{}) {
	select {
	case <-done:
		return
	default:
	}

	logger := logging.WithContext(req.Ctx, c.bot.deps.Logger())

	err := c.bot.session.UpdateFromReady(p)
	if err != nil {
		_ = level.Error(logger).Log("message", "error setting up session", "err", err)
		return
	}
}

func (c *discordMessageHandler) handleGuildCreate(p *etfapi.Payload, req wsclient.WSMessage, resp chan<- wsclient.WSMessage, done <-chan struct{}) {
	select {
	case <-done:
		return
	default:
	}

	logger := logging.WithContext(req.Ctx, c.bot.deps.Logger())
	err := c.bot.session.UpsertGuildFromElementMap(p.Data)
	if err != nil {
		_ = level.Error(logger).Log("message", "error processing guild create", "err", err)
		return
	}
}

func (c *discordMessageHandler) handleMessage(p *etfapi.Payload, req wsclient.WSMessage, resp chan<- wsclient.WSMessage, done <-chan struct{}) {
	select {
	case <-done:
		return
	default:
	}

	logger := logging.WithContext(req.Ctx, c.bot.deps.Logger())

	m, err := etfapi.MessageFromElementMap(p.Data)
	if err != nil {
		_ = level.Error(logger).Log("message", "error inflating message", "err", err)
		return
	}

	if m.MessageType() != etfapi.DefaultMessage {
		_ = level.Debug(logger).Log("message", "message was not a default type")
		return
	}

	content := m.ContentRunes()
	if len(content) == 0 {
		_ = level.Debug(logger).Log("message", "message contents empty")
		return
	}

	// TODO: guild-specific command indicator
	if content[0] != c.bot.config.DefaultCommandIndicator {
		_ = level.Debug(logger).Log("message", "not a command")
		return
	}

	content[0] = c.bot.commandHandler.CommandIndicator()
	respStr, err := c.bot.commandHandler.HandleLine(m.AuthorIDString(), content)
	if err != nil {
		_ = level.Error(logger).Log("message", "error handling command", "contents", string(content[1:]), "err", err)
		respStr += fmt.Sprintf("\nError: %v\n", err)
	}

	err = c.bot.rateLimiter.Wait(req.Ctx)
	if err != nil {
		_ = level.Error(logger).Log("message", "error generating identify payload", "err", err)
		return
	}

	msg := jsonapi.Message{
		Content: fmt.Sprintf("%s\n\n%s\n", m.AuthorIDString(), respStr),
	}

	sendResp, body, err := c.bot.SendMessage(req.Ctx, m.ChannelID(), msg)
	if err != nil {
		_ = level.Error(logger).Log("message", "could not send message", "err", err, "resp_body", string(body), "status_code", sendResp.StatusCode)
		return
	}

	_ = level.Info(logger).Log("message", "successfully sent message to channel", "channel_id", m.ChannelID().ToString())

	return
}
