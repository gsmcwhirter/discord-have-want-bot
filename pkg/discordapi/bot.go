package discordapi

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"

	"github.com/gsmcwhirter/eso-discord/pkg/cmdhandler"
	"github.com/gsmcwhirter/eso-discord/pkg/commands"
	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/etfapi/payloads"
	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/jsonapi"
	"github.com/gsmcwhirter/eso-discord/pkg/httpclient"
	"github.com/gsmcwhirter/eso-discord/pkg/logging"
	"github.com/gsmcwhirter/eso-discord/pkg/snowflake"
	"github.com/gsmcwhirter/eso-discord/pkg/storage"
	"github.com/gsmcwhirter/eso-discord/pkg/util"
	"github.com/gsmcwhirter/eso-discord/pkg/wsclient"
)

// ErrResponse TODOC
var ErrResponse = errors.New("error response")

type dependencies interface {
	Logger() log.Logger
	HTTPClient() httpclient.HTTPClient
	WSClient() wsclient.WSClient
	UserAPI() storage.UserAPI
}

// DiscordBot TODOC
type DiscordBot interface {
	AuthenticateAndConnect() error
	Disconnect() error
	Run()
}

// BotConfig TODOC
type BotConfig struct {
	ClientID                string
	ClientSecret            string
	BotToken                string
	APIURL                  string
	Version                 string
	NumWorkers              int
	DefaultCommandIndicator rune
}

type hbReconfig struct {
	ctx      context.Context
	interval int
}

type discordBot struct {
	config         BotConfig
	deps           dependencies
	messageHandler *discordMessageHandler
	commandHandler *cmdhandler.CommandHandler

	session               Session
	connectionRateLimiter *rate.Limiter
	rateLimiter           *rate.Limiter

	heartbeat  *time.Ticker
	heartbeats chan hbReconfig

	seqLock      sync.Mutex
	lastSequence int
}

// NewDiscordBot TODOC
func NewDiscordBot(deps dependencies, conf BotConfig) DiscordBot {
	d := discordBot{
		config:         conf,
		deps:           deps,
		commandHandler: commands.CommandHandler(deps, conf.Version, commands.Options{}),

		session:               NewSession(),
		connectionRateLimiter: rate.NewLimiter(rate.Every(5*time.Second), 1),
		rateLimiter:           rate.NewLimiter(rate.Every(60*time.Second), 120),

		heartbeats: make(chan hbReconfig),

		lastSequence: -1,
	}

	return &d
}

func (d *discordBot) AuthenticateAndConnect() error {
	ctx := util.NewRequestContext()
	logger := logging.WithContext(ctx, d.deps.Logger())

	err := d.connectionRateLimiter.Wait(ctx)
	if err != nil {
		return err
	}

	resp, body, err := d.deps.HTTPClient().GetBody(ctx, fmt.Sprintf("%s/gateway/bot", d.config.APIURL), nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Wrap(ErrResponse, "non-200 response")
	}

	_ = level.Debug(logger).Log(
		"response_body", body,
		"response_bytes", len(body),
	)

	respData := jsonapi.GatewayResponse{}
	err = respData.UnmarshalJSON(body)
	if err != nil {
		return err
	}

	_ = level.Debug(logger).Log(
		"gateway_url", respData.URL,
		"gateway_shards", respData.Shards,
	)

	_ = level.Info(logger).Log("message", "acquired gateway url")

	d.messageHandler = newDiscordMessageHandler(d, d.heartbeats)

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
	d.deps.WSClient().SetHandler(d.messageHandler)

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

func (d *discordBot) SendMessage(ctx context.Context, cid snowflake.Snowflake, m jsonapi.Message) (resp *http.Response, body []byte, err error) {
	logger := logging.WithContext(ctx, d.deps.Logger())

	_ = level.Info(logger).Log("message", "sending message to channel")

	b, err := m.MarshalJSON()
	if err != nil {
		return
	}
	r := bytes.NewReader(b)

	err = d.rateLimiter.Wait(ctx)
	if err != nil {
		return nil, nil, err
	}

	header := http.Header{}
	header.Add("Content-Type", "application/json")
	resp, body, err = d.deps.HTTPClient().PostBody(ctx, fmt.Sprintf("%s/channels/%d/messages", d.config.APIURL, cid), &header, r)
	if err != nil {
		err = errors.Wrap(err, "could not complete the message send")
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = errors.Wrap(ErrResponse, "non-200 response")
	}

	return
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

func (d *discordBot) LastSequence() int {
	d.seqLock.Lock()
	defer d.seqLock.Unlock()

	return d.lastSequence
}

func (d *discordBot) UpdateSequence(seq int) bool {
	d.seqLock.Lock()
	defer d.seqLock.Unlock()

	if seq < d.lastSequence {
		return false
	}
	d.lastSequence = seq
	return true
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
					Sequence: d.LastSequence(),
				})
				if err != nil {
					_ = level.Error(logging.WithContext(ctx, d.deps.Logger())).Log("message", "error formatting heartbeat", "err", err)
					return
				}

				err = d.rateLimiter.Wait(m.Ctx)
				if err != nil {
					_ = level.Error(logging.WithContext(ctx, d.deps.Logger())).Log("message", "error rate limiting", "err", err)
					return
				}
				d.deps.WSClient().SendMessage(m)
			}

		case <-d.heartbeat.C: // tick
			_ = level.Debug(d.deps.Logger()).Log("message", "bum-bum")
			ctx := util.NewRequestContext()

			m, err := payloads.ETFPayloadToMessage(ctx, payloads.HeartbeatPayload{
				Sequence: d.lastSequence,
			})
			if err != nil {
				_ = level.Error(logging.WithContext(ctx, d.deps.Logger())).Log("message", "error formatting heartbeat", "err", err)
				return
			}

			err = d.rateLimiter.Wait(m.Ctx)
			if err != nil {
				_ = level.Error(logging.WithContext(ctx, d.deps.Logger())).Log("message", "error rate limiting", "err", err)
				return
			}
			d.deps.WSClient().SendMessage(m)
		}
	}
}
