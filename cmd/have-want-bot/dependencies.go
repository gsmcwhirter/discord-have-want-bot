package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gsmcwhirter/discord-bot-lib/httpclient"
	"github.com/gsmcwhirter/discord-bot-lib/wsclient"
	"github.com/gsmcwhirter/eso-discord/pkg/cmdhandler"
	"golang.org/x/time/rate"

	"github.com/gsmcwhirter/eso-discord/cmd/have-want-bot/commands"
	configCommands "github.com/gsmcwhirter/eso-discord/pkg/commands"
	"github.com/gsmcwhirter/eso-discord/pkg/msghandler"
	"github.com/gsmcwhirter/eso-discord/pkg/storage"
)

type dependencies struct {
	logger             log.Logger
	db                 *bolt.DB
	userAPI            storage.UserAPI
	guildAPI           storage.GuildAPI
	httpClient         httpclient.HTTPClient
	wsClient           wsclient.WSClient
	cmdHandler         *cmdhandler.CommandHandler
	configHandler      *cmdhandler.CommandHandler
	messageRateLimiter *rate.Limiter
	connectRateLimiter *rate.Limiter
	msgHandlers        msghandler.Handlers
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

	d.httpClient = httpclient.NewHTTPClient(d)
	h := http.Header{}
	h.Add("User-Agent", fmt.Sprintf("DiscordBot (%s, %s)", conf.ClientURL, BuildVersion))
	h.Add("Authorization", fmt.Sprintf("Bot %s", conf.ClientToken))
	d.httpClient.SetHeaders(h)

	d.wsClient = wsclient.NewWSClient(d, wsclient.Options{MaxConcurrentHandlers: conf.NumWorkers})

	d.cmdHandler = commands.CommandHandler(d, conf.Version, commands.Options{})
	d.configHandler = configCommands.ConfigHandler(d, conf.Version, configCommands.Options{})

	d.connectRateLimiter = rate.NewLimiter(rate.Every(5*time.Second), 1)
	d.messageRateLimiter = rate.NewLimiter(rate.Every(60*time.Second), 120)

	d.msgHandlers = msghandler.NewHandlers(d, msghandler.Options{
		DefaultCommandIndicator: "!",
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

func (d *dependencies) Logger() log.Logger {
	return d.logger
}

func (d *dependencies) HTTPClient() httpclient.HTTPClient {
	return d.httpClient
}

func (d *dependencies) CommandHandler() *cmdhandler.CommandHandler {
	return d.cmdHandler
}

func (d *dependencies) ConfigHandler() *cmdhandler.CommandHandler {
	return d.configHandler
}

func (d *dependencies) MessageHandler() msghandler.Handlers {
	return d.msgHandlers
}

func (d *dependencies) WSClient() wsclient.WSClient {
	return d.wsClient
}

func (d *dependencies) GuildAPI() storage.GuildAPI {
	return d.guildAPI
}

func (d *dependencies) UserAPI() storage.UserAPI {
	return d.userAPI
}

func (d *dependencies) MessageRateLimiter() *rate.Limiter {
	return d.messageRateLimiter
}

func (d *dependencies) ConnectRateLimiter() *rate.Limiter {
	return d.connectRateLimiter
}
