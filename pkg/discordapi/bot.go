package discordapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-kit/kit/log/level"

	"github.com/go-kit/kit/log"

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
	discordClient discordMessageHandler
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

	level.Debug(logger).Log(
		"response_body", body,
		"response_bytes", bodySize,
	)

	respData := jsonapi.GatewayResponse{}
	err = respData.UnmarshalJSON(body[:bodySize])
	if err != nil {
		return err
	}

	level.Debug(logger).Log(
		"gateway_url", respData.URL,
		"gateway_shards", respData.Shards,
	)

	d.discordClient = newDiscordMessageHandler()

	connectURL, err := url.Parse(respData.URL)
	if err != nil {
		return err
	}
	q := connectURL.Query()
	q.Add("v", "6")
	q.Add("encoding", "etf")
	connectURL.RawQuery = q.Encode()

	level.Info(logger).Log(
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

	fmt.Printf("To add to a guild, go to: https://discordapp.com/api/oauth2/authorize?client_id=%s&scope=bot&permissions=%d\n", d.config.ClientID, botPermissions)

	return nil
}

func (d *discordBot) Disconnect() error {
	d.deps.WSClient().Close()
	return nil
}

func (d *discordBot) Run() {
	done := make(chan struct{})
	defer close(done)

	go d.heartbeatHandler(done)
	d.deps.WSClient().HandleRequests(d.config.NumWorkers)

}

func (d *discordBot) heartbeatHandler(done chan struct{}) {
	for {
		if d.heartbeat == nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}

		select {
		case <-done:
			return
		case req := <-d.heartbeats:
			if req.interval > 0 {
				d.heartbeat.Stop()
				d.heartbeat = time.NewTicker(time.Duration(req.interval) * time.Millisecond)
			} else {
				ctx := req.ctx
				if ctx == nil {
					ctx = util.NewRequestContext()
				}
				d.deps.WSClient().SendMessage(d.discordClient.FormatHeartbeat(ctx, d.lastSequence))
			}
		case <-d.heartbeat.C:
			ctx := util.NewRequestContext()
			d.deps.WSClient().SendMessage(d.discordClient.FormatHeartbeat(ctx, d.lastSequence))
		}
	}
}
