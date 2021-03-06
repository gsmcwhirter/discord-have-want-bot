package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/websocket"
	"github.com/gsmcwhirter/discord-bot-lib/bot"
	"github.com/gsmcwhirter/discord-bot-lib/cmdhandler"
	"github.com/gsmcwhirter/discord-bot-lib/etfapi"
	"github.com/gsmcwhirter/discord-bot-lib/httpclient"
	"github.com/gsmcwhirter/discord-bot-lib/messagehandler"
	"github.com/gsmcwhirter/discord-bot-lib/wsclient"
	"golang.org/x/time/rate"

	"github.com/gsmcwhirter/discord-have-want-bot/pkg/commands"
	"github.com/gsmcwhirter/discord-have-want-bot/pkg/msghandler"
	"github.com/gsmcwhirter/discord-have-want-bot/pkg/storage"
)

type dependencies struct {
	logger log.Logger

	db       *bolt.DB
	userAPI  storage.UserAPI
	guildAPI storage.GuildAPI

	httpDoer   httpclient.Doer
	httpClient httpclient.HTTPClient
	wsDialer   wsclient.Dialer
	wsClient   wsclient.WSClient

	messageRateLimiter *rate.Limiter
	connectRateLimiter *rate.Limiter
	botSession         *etfapi.Session

	cmdHandler        *cmdhandler.CommandHandler
	configHandler     *cmdhandler.CommandHandler
	discordMsgHandler bot.DiscordMessageHandler
	msgHandlers       msghandler.Handlers
}

func createDependencies(conf config) (d *dependencies, err error) {
	d = &dependencies{}

	var logger log.Logger
	if conf.LogFormat == "json" {
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	} else {
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	}

	switch conf.LogLevel {
	case "debug":
		logger = level.NewFilter(logger, level.AllowDebug())
	case "info":
		logger = level.NewFilter(logger, level.AllowInfo())
	case "warn":
		logger = level.NewFilter(logger, level.AllowWarn())
	case "error":
		logger = level.NewFilter(logger, level.AllowError())
	default:
		logger = level.NewFilter(logger, level.AllowAll())
	}

	logger = log.With(logger, "timestamp", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	d.logger = logger

	d.db, err = bolt.Open(conf.Database, 0660, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return
	}

	d.userAPI, err = storage.NewBoltUserAPI(d.db)
	if err != nil {
		return
	}

	d.guildAPI, err = storage.NewBoltGuildAPI(d.db)
	if err != nil {
		return
	}

	d.httpDoer = &http.Client{}
	d.httpClient = httpclient.NewHTTPClient(d)
	h := http.Header{}
	h.Add("User-Agent", fmt.Sprintf("DiscordBot (%s, %s)", conf.ClientURL, BuildVersion))
	h.Add("Authorization", fmt.Sprintf("Bot %s", conf.ClientToken))
	d.httpClient.SetHeaders(h)

	d.wsDialer = wsclient.WrapDialer(websocket.DefaultDialer)
	d.wsClient = wsclient.NewWSClient(d, wsclient.Options{MaxConcurrentHandlers: conf.NumWorkers})

	d.discordMsgHandler = messagehandler.NewDiscordMessageHandler(d)
	d.botSession = etfapi.NewSession()

	d.cmdHandler, err = commands.CommandHandler(d, conf.Version, commands.Options{CmdIndicator: " "})
	if err != nil {
		return
	}
	d.configHandler, err = commands.ConfigHandler(d, conf.Version, commands.Options{CmdIndicator: " "})
	if err != nil {
		return
	}

	d.connectRateLimiter = rate.NewLimiter(rate.Every(5*time.Second), 1)
	d.messageRateLimiter = rate.NewLimiter(rate.Every(60*time.Second), 120)

	d.msgHandlers = msghandler.NewHandlers(d, msghandler.Options{
		DefaultCommandIndicator: "!",
		ErrorColor:              0xff0000,
		SuccessColor:            0x62aa00,
	})

	return
}

func (d *dependencies) Close() {
	if d.db != nil {
		d.db.Close() // nolint: errcheck
	}

	if d.wsClient != nil {
		d.wsClient.Close()
	}
}

func (d *dependencies) Logger() log.Logger                         { return d.logger }
func (d *dependencies) GuildAPI() storage.GuildAPI                 { return d.guildAPI }
func (d *dependencies) UserAPI() storage.UserAPI                   { return d.userAPI }
func (d *dependencies) HTTPDoer() httpclient.Doer                  { return d.httpDoer }
func (d *dependencies) HTTPClient() httpclient.HTTPClient          { return d.httpClient }
func (d *dependencies) WSDialer() wsclient.Dialer                  { return d.wsDialer }
func (d *dependencies) WSClient() wsclient.WSClient                { return d.wsClient }
func (d *dependencies) MessageRateLimiter() *rate.Limiter          { return d.messageRateLimiter }
func (d *dependencies) ConnectRateLimiter() *rate.Limiter          { return d.connectRateLimiter }
func (d *dependencies) BotSession() *etfapi.Session                { return d.botSession }
func (d *dependencies) CommandHandler() *cmdhandler.CommandHandler { return d.cmdHandler }
func (d *dependencies) ConfigHandler() *cmdhandler.CommandHandler  { return d.configHandler }
func (d *dependencies) MessageHandler() msghandler.Handlers        { return d.msgHandlers }
func (d *dependencies) DiscordMessageHandler() bot.DiscordMessageHandler {
	return d.discordMsgHandler
}
