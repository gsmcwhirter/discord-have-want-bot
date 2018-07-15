package discordapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-kit/kit/log/level"

	"github.com/go-kit/kit/log"

	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/etfapi/payloads"
	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/jsonapi"
	"github.com/gsmcwhirter/eso-discord/pkg/httpclient"
	"github.com/gsmcwhirter/eso-discord/pkg/logging"
	"github.com/gsmcwhirter/eso-discord/pkg/util"
	"github.com/gsmcwhirter/eso-discord/pkg/wsclient"
)

type dependencies interface {
	Logger() log.Logger
	HTTPClient() httpclient.HTTPClient
	WSClient() wsclient.WSClient
}

// DiscordBot TODOC
type DiscordBot interface {
	AuthenticateAndConnect() error
	Disconnect() error
	Run()
}

// BotConfig TODOC
type BotConfig struct {
	ClientID     string
	ClientSecret string
	BotToken     string
	APIURL       string
	NumWorkers   int
}

type hbReconfig struct {
	ctx      context.Context
	interval int
}

type discordBot struct {
	config        BotConfig
	discordClient *discordMessageHandler
	deps          dependencies
	httpHeaders   http.Header

	heartbeat    *time.Ticker
	heartbeats   chan hbReconfig
	lastSequence *int
}

// NewDiscordBot TODOC
func NewDiscordBot(deps dependencies, conf BotConfig) DiscordBot {
	d := discordBot{
		config:      conf,
		deps:        deps,
		httpHeaders: http.Header{},

		heartbeats: make(chan hbReconfig),
	}

	d.httpHeaders.Add("Authorization", fmt.Sprintf("Bot %s", d.config.BotToken))

	return &d
}

func (d *discordBot) AuthenticateAndConnect() error {
	ctx := util.NewRequestContext()
	logger := logging.WithContext(ctx, d.deps.Logger())

	resp, err := d.deps.HTTPClient().Get(ctx, fmt.Sprintf("%s/gateway/bot", d.config.APIURL), &d.httpHeaders)
	if err != nil {
		return err
	}
	defer util.CheckDefer(resp.Body.Close)

	bodySize, body, err := util.ReadBody(resp.Body, 200)
	if err != nil {
		return err
	}

	_ = level.Debug(logger).Log(
		"response_body", body,
		"response_bytes", bodySize,
	)

	respData := jsonapi.GatewayResponse{}
	err = respData.UnmarshalJSON(body[:bodySize])
	if err != nil {
		return err
	}

	_ = level.Debug(logger).Log(
		"gateway_url", respData.URL,
		"gateway_shards", respData.Shards,
	)

	d.discordClient = newDiscordMessageHandler(d, d.deps, d.heartbeats)

	connectURL, err := url.Parse(respData.URL)
	if err != nil {
		return err
	}
	q := connectURL.Query()
	q.Add("v", "6")
	q.Add("encoding", "etf")
	connectURL.RawQuery = q.Encode()

	_ = level.Info(logger).Log(
		"message", "connecting to gateway",
		"gateway_url", connectURL.String(),
	)

	d.deps.WSClient().SetGateway(connectURL.String())
	d.deps.WSClient().SetHandler(d.discordClient)

	err = d.deps.WSClient().Connect(d.config.BotToken)
	if err != nil {
		return err
	}

	// See https://discordapp.com/developers/docs/topics/permissions#permissions-bitwise-permission-flags
	botPermissions := 0x00000040 // add reactions
	botPermissions |= 0x00000400 // view channel (including read messages)
	botPermissions |= 0x00000800 // send messages

	fmt.Printf("\nTo add to a guild, go to: https://discordapp.com/api/oauth2/authorize?client_id=%s&scope=bot&permissions=%d\n\n", d.config.ClientID, botPermissions)

	return nil
}

func (d *discordBot) Disconnect() error {
	d.deps.WSClient().Close()
	return nil
}

func (d *discordBot) Run() {
	_ = level.Debug(d.deps.Logger()).Log("message", "setting bot signal watcher")
	interrupt := make(chan os.Signal, 2)
	defer func() {
		for range interrupt {
		}
	}()
	defer close(interrupt)
	signal.Notify(interrupt, os.Interrupt)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go d.heartbeatHandler(&wg, interrupt)
	d.deps.WSClient().HandleRequests(interrupt)
	wg.Wait()
}

func (d *discordBot) heartbeatHandler(wg *sync.WaitGroup, done chan os.Signal) {
	defer wg.Done()

	_ = level.Debug(d.deps.Logger()).Log("message", "waiting for heartbeat config")

	// wait for init
	if d.heartbeat == nil {
		select {
		case sig := <-done:
			_ = level.Info(d.deps.Logger()).Log("message", "heartbeat loop stopping before config")
			done <- sig
			return
		case req := <-d.heartbeats:
			if req.interval > 0 {
				d.heartbeat = time.NewTicker(time.Duration(req.interval) * time.Millisecond)
				_ = level.Info(d.deps.Logger()).Log("message", "starting heartbeat loop", "interval", req.interval)
			}
		}

	}

	// in the groove
	for {
		select {
		case sig := <-done: // quit
			_ = level.Info(d.deps.Logger()).Log("message", "heartbeat quitting at request")
			done <- sig
			d.heartbeat.Stop()
			return

		case req := <-d.heartbeats: // reconfigure
			if req.interval > 0 {
				_ = level.Info(d.deps.Logger()).Log("message", "reconfiguring heartbeat loop", "interval", req.interval)
				d.heartbeat.Stop()
				d.heartbeat = time.NewTicker(time.Duration(req.interval) * time.Millisecond)
			} else {
				ctx := req.ctx
				if ctx == nil {
					ctx = util.NewRequestContext()
				}

				_ = level.Debug(logging.WithContext(ctx, d.deps.Logger())).Log("message", "manual heartbeat requested")

				m, err := payloads.ETFPayloadToMessage(ctx, payloads.HeartbeatPayload{
					Sequence: d.lastSequence,
				})
				if err != nil {
					_ = level.Error(logging.WithContext(ctx, d.deps.Logger())).Log("message", "error formatting heartbeat", "err", err)
				} else {
					d.deps.WSClient().SendMessage(m)
				}
			}

		case <-d.heartbeat.C: // tick
			_ = level.Debug(d.deps.Logger()).Log("message", "bum-bum")
			ctx := util.NewRequestContext()

			m, err := payloads.ETFPayloadToMessage(ctx, payloads.HeartbeatPayload{
				Sequence: d.lastSequence,
			})
			if err != nil {
				_ = level.Error(logging.WithContext(ctx, d.deps.Logger())).Log("message", "error formatting heartbeat", "err", err)
			} else {
				d.deps.WSClient().SendMessage(m)
			}
		}
	}
}
