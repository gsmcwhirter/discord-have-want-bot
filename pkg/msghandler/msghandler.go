package msghandler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gsmcwhirter/discord-bot-lib/cmdhandler"
	"github.com/gsmcwhirter/discord-bot-lib/discordapi"
	"github.com/gsmcwhirter/discord-bot-lib/discordapi/etfapi"
	"github.com/gsmcwhirter/discord-bot-lib/logging"
	"github.com/gsmcwhirter/discord-bot-lib/snowflake"
	"github.com/gsmcwhirter/discord-bot-lib/wsclient"
	"github.com/gsmcwhirter/go-util/parser"
	"golang.org/x/time/rate"

	"github.com/gsmcwhirter/discord-have-want-bot/pkg/storage"
)

var errUnauthorized = errors.New("unauthorized")

type dependencies interface {
	Logger() log.Logger
	GuildAPI() storage.GuildAPI
	CommandHandler() *cmdhandler.CommandHandler
	ConfigHandler() *cmdhandler.CommandHandler
	MessageRateLimiter() *rate.Limiter
}

// Handlers TODOC
type Handlers interface {
	ConnectToBot(discordapi.DiscordBot)
}

type handlers struct {
	bot                     discordapi.DiscordBot
	deps                    dependencies
	defaultCommandIndicator string
}

// Options TODOC
type Options struct {
	DefaultCommandIndicator string
}

// NewHandlers TODOC
func NewHandlers(deps dependencies, opts Options) Handlers {
	h := handlers{
		deps: deps,
		defaultCommandIndicator: opts.DefaultCommandIndicator,
	}

	return &h
}

func (h *handlers) ConnectToBot(bot discordapi.DiscordBot) {
	h.bot = bot

	bot.AddMessageHandler("MESSAGE_CREATE", h.handleMessage)
}

func (h *handlers) channelGuild(cid snowflake.Snowflake) (gid snowflake.Snowflake) {
	gid, _ = h.bot.GuildOfChannel(cid)
	return
}

func (h *handlers) guildCommandIndicator(gid snowflake.Snowflake) string {
	if gid == 0 {
		return h.defaultCommandIndicator
	}

	s, err := GetSettings(h.deps.GuildAPI(), gid)
	if err != nil {
		return h.defaultCommandIndicator
	}

	if s.ControlSequence == "" {
		return h.defaultCommandIndicator
	}

	return s.ControlSequence
}

func (h *handlers) attemptConfigHandler(req wsclient.WSMessage, cmdIndicator string, content string, m etfapi.Message, gid snowflake.Snowflake) (resp cmdhandler.Response, err error) {
	// TODO: check auth
	if !h.bot.IsGuildAdmin(gid, m.AuthorID()) {
		_ = level.Debug(logging.WithContext(req.Ctx, h.deps.Logger())).Log("message", "non-admin trying to config", "author_id", m.AuthorID().ToString(), "guild_id", gid.ToString())

		err = errUnauthorized
		return
	}

	_ = level.Debug(logging.WithContext(req.Ctx, h.deps.Logger())).Log("message", "admin trying to config", "author_id", m.AuthorID().ToString(), "guild_id", gid.ToString())
	cmdContent := h.deps.ConfigHandler().CommandIndicator() + strings.TrimPrefix(content, cmdIndicator)
	resp, err = h.deps.ConfigHandler().HandleLine(m.AuthorIDString(), gid.ToString(), cmdContent)
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
		_ = level.Debug(logger).Log("message", "message was not a default type")
		return
	}

	content := m.ContentString()
	if len(content) == 0 {
		_ = level.Debug(logger).Log("message", "message contents empty")
		return
	}

	gid := h.channelGuild(m.ChannelID())
	cmdIndicator := h.guildCommandIndicator(gid)

	if !strings.HasPrefix(content, cmdIndicator) {
		_ = level.Debug(logger).Log("message", "not a command")
		return
	}

	_ = level.Info(logger).Log("message", "attempting to handle command")
	resp, err := h.attemptConfigHandler(req, cmdIndicator, content, m, gid)
	if err != nil && (err == errUnauthorized || err == parser.ErrUnknownCommand) {
		cmdContent := h.deps.CommandHandler().CommandIndicator() + strings.TrimPrefix(content, cmdIndicator)
		resp, err = h.deps.CommandHandler().HandleLine(m.AuthorIDString(), gid.ToString(), cmdContent)
	}

	if err != nil {
		_ = level.Error(logger).Log("message", "error handling command", "contents", content, "err", err)
		resp.IncludeError(err)
		resp.SetColor(0xff0000)
	} else {
		resp.SetColor(0x62aa00)
	}

	err = h.deps.MessageRateLimiter().Wait(req.Ctx)
	if err != nil {
		_ = level.Error(logger).Log("message", "error waiting for ratelimiting", "err", err)
		return
	}

	_ = level.Debug(logger).Log("message", "sending message", "marshaler", fmt.Sprintf("%+v", resp.ToMessage()), "resp", fmt.Sprintf("%+v", resp))

	sendResp, body, err := h.bot.SendMessage(req.Ctx, m.ChannelID(), resp.ToMessage())
	if err != nil {
		_ = level.Error(logger).Log("message", "could not send message", "err", err, "resp_body", string(body), "status_code", sendResp.StatusCode)
		return
	}

	_ = level.Info(logger).Log("message", "successfully sent message to channel", "channel_id", m.ChannelID().ToString())

	return
}
