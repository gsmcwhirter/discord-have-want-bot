package msghandler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gsmcwhirter/discord-bot-lib/bot"
	"github.com/gsmcwhirter/discord-bot-lib/cmdhandler"
	"github.com/gsmcwhirter/discord-bot-lib/etfapi"
	"github.com/gsmcwhirter/discord-bot-lib/logging"
	"github.com/gsmcwhirter/discord-bot-lib/snowflake"
	"github.com/gsmcwhirter/discord-bot-lib/wsclient"
	"github.com/gsmcwhirter/go-util/parser"
	"golang.org/x/time/rate"

	"github.com/gsmcwhirter/discord-have-want-bot/pkg/storage"
)

var errUnauthorized = errors.New("unauthorized")

// ErrNoResponse is the error that should be returned by a command handler
// if the bot should not provide a response
var ErrNoResponse = errors.New("no response")

type dependencies interface {
	Logger() log.Logger
	GuildAPI() storage.GuildAPI
	CommandHandler() *cmdhandler.CommandHandler
	ConfigHandler() *cmdhandler.CommandHandler
	MessageRateLimiter() *rate.Limiter
	BotSession() *etfapi.Session
}

// Handlers is the dependency interface for a set of message handlers that connect themselves
// to a discord bot
type Handlers interface {
	ConnectToBot(bot.DiscordBot)
}

type handlers struct {
	bot                     bot.DiscordBot
	deps                    dependencies
	defaultCommandIndicator string
	successColor            int
	errorColor              int
}

// Options is how to set response colors etc. when creating a Handlers
type Options struct {
	DefaultCommandIndicator string
	SuccessColor            int
	ErrorColor              int
}

// NewHandlers creates a new Handlers object
func NewHandlers(deps dependencies, opts Options) Handlers {
	h := handlers{
		deps: deps,
		defaultCommandIndicator: opts.DefaultCommandIndicator,
		successColor:            opts.SuccessColor,
		errorColor:              opts.ErrorColor,
	}

	return &h
}

func (h *handlers) ConnectToBot(bot bot.DiscordBot) {
	h.bot = bot

	bot.AddMessageHandler("MESSAGE_CREATE", h.handleMessage)
}

func (h *handlers) channelGuild(cid snowflake.Snowflake) (gid snowflake.Snowflake) {
	gid, _ = h.deps.BotSession().GuildOfChannel(cid)
	return
}

func (h *handlers) guildCommandIndicator(gid snowflake.Snowflake) string {
	if gid == 0 {
		return h.defaultCommandIndicator
	}

	s, err := storage.GetSettings(h.deps.GuildAPI(), gid)
	if err != nil {
		return h.defaultCommandIndicator
	}

	if s.ControlSequence == "" {
		return h.defaultCommandIndicator
	}

	return s.ControlSequence
}

func (h *handlers) attemptConfigAndAdminHandlers(msg cmdhandler.Message, req wsclient.WSMessage, cmdIndicator string, content string, m etfapi.Message, gid snowflake.Snowflake) (resp cmdhandler.Response, err error) {
	// TODO: check auth
	logger := logging.WithMessage(msg, h.deps.Logger())

	if !h.deps.BotSession().IsGuildAdmin(gid, m.AuthorID()) {
		_ = level.Debug(logging.WithContext(req.Ctx, h.deps.Logger())).Log("message", "non-admin trying to config", "author_id", m.AuthorID().ToString(), "guild_id", gid.ToString())

		err = errUnauthorized
		return
	}

	_ = level.Debug(logger).Log("message", "admin trying to config")
	cmdContent := h.deps.ConfigHandler().CommandIndicator() + strings.TrimPrefix(content, cmdIndicator)
	resp, err = h.deps.ConfigHandler().HandleMessage(cmdhandler.NewWithContents(msg, cmdContent))
	return
}

func (h *handlers) handleMessage(p *etfapi.Payload, req wsclient.WSMessage, respChan chan<- wsclient.WSMessage) {
	if h.bot == nil {
		return
	}

	select {
	case <-req.Ctx.Done():
		return
	default:
	}

	logger := logging.WithContext(req.Ctx, h.deps.Logger())

	m, err := etfapi.MessageFromElementMap(p.Data)
	if err != nil {
		_ = level.Error(logger).Log("message", "error inflating message", "err", err)
		return
	}

	if m.MessageType() != etfapi.DefaultMessage {
		_ = level.Info(logger).Log("message", "message was not a default type")
		return
	}

	content := m.ContentString()
	if len(content) == 0 {
		_ = level.Info(logger).Log("message", "message contents empty")
		return
	}

	gid := h.channelGuild(m.ChannelID())
	cmdIndicator := h.guildCommandIndicator(gid)

	if !strings.HasPrefix(content, cmdIndicator) {
		_ = level.Info(logger).Log("message", "not a command")
		return
	}

	msg := cmdhandler.NewSimpleMessage(req.Ctx, m.AuthorID(), gid, m.ChannelID(), m.ID(), "")
	logger = logging.WithMessage(msg, h.deps.Logger())
	resp, err := h.attemptConfigAndAdminHandlers(msg, req, cmdIndicator, content, m, gid)

	if err != nil && (err == errUnauthorized || err == parser.ErrUnknownCommand) {
		_ = level.Debug(logger).Log("message", "admin not successful; processing as real message")
		cmdContent := h.deps.CommandHandler().CommandIndicator() + strings.TrimPrefix(content, cmdIndicator)
		resp, err = h.deps.CommandHandler().HandleMessage(cmdhandler.NewWithContents(msg, cmdContent))
	}

	if err == ErrNoResponse {
		return
	}

	if err != nil {
		_ = level.Error(logger).Log("message", "error handling command", "contents", content, "err", err)
		resp.IncludeError(err)
	}

	if resp.HasErrors() {
		resp.SetColor(h.errorColor)
	} else {
		resp.SetColor(h.successColor)
	}

	err = h.deps.MessageRateLimiter().Wait(req.Ctx)
	if err != nil {
		_ = level.Error(logger).Log("message", "error waiting for ratelimiting", "err", err)
		return
	}

	_ = level.Info(logger).Log("message", "sending message", "resp", fmt.Sprintf("%+v", resp))

	sendTo := resp.Channel()
	if sendTo == 0 {
		sendTo = m.ChannelID()
	}

	sendResp, body, err := h.bot.SendMessage(req.Ctx, sendTo, resp.ToMessage())
	if err != nil {
		_ = level.Error(logger).Log("message", "could not send message", "err", err, "resp_body", string(body), "status_code", sendResp.StatusCode)
		return
	}

	_ = level.Info(logger).Log("message", "successfully sent message to channel", "channel_id", sendTo.ToString())
}
