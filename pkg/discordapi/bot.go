package discordapi

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/jsonapi"
	"github.com/gsmcwhirter/eso-discord/pkg/util"
	"github.com/gsmcwhirter/eso-discord/pkg/wsclient"
)

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
	// AuthURL      string
	// TokenURL     string
	// RedirectURL  string
	// GatewayHost string
	// GatewayPort string
	NumWorkers int
}

type discordBot struct {
	config     BotConfig
	httpClient *http.Client
	wsClient   wsclient.WSClient
}

// NewDiscordBot TODOC
func NewDiscordBot(conf BotConfig) DiscordBot {
	return &discordBot{
		config:     conf,
		httpClient: &http.Client{},
	}
}

func (d *discordBot) HTTPGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bot %s", d.config.BotToken))
	return d.httpClient.Do(req)
}

func (d *discordBot) AuthenticateAndConnect() error {
	// ctx := context.Background()
	// conf := &oauth2.Config{
	// 	ClientID:     d.config.ClientID,
	// 	ClientSecret: d.config.ClientSecret,
	// 	Scopes:       []string{"bot"},
	// 	Endpoint: oauth2.Endpoint{
	// 		AuthURL:  d.config.AuthURL,
	// 		TokenURL: d.config.TokenURL,
	// 	},
	// 	RedirectURL: d.config.RedirectURL,
	// }

	// client := &http.Client{}
	// authUrl := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	// resp, err := client.Get(authUrl)
	// if err != nil {
	// 	return errors.Wrap(err, "could not authenticate")
	// }

	// if resp.StatusCode != 301 && resp.StatusCode != 302 {
	// 	return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	// }

	// loc := resp.Header.Get("Location")
	// redirectUrl, err := url.Parse(loc)
	// if err != nil {
	// 	return errors.Wrap(err, "could not parse oauth2 redirect url")
	// }

	// code := redirectUrl.Query().Get("code")
	// if code == "" {
	// 	return errors.New("could not find code in redirect url")
	// }

	// token, err := conf.Exchange(ctx, code)
	// if err != nil {
	// 	return errors.Wrap(err, "could not get a valid token")
	// }

	// d.token = token

	resp, err := d.HTTPGet(fmt.Sprintf("%s/gateway/bot", d.config.APIURL))
	if err != nil {
		return err
	}
	defer util.CheckDefer(resp.Body.Close)

	fmt.Printf("%+v\n", resp)

	bodySize, body, err := util.ReadBody(resp.Body, 200)
	if err != nil {
		return err
	}

	respData := jsonapi.GatewayResponse{}
	err = respData.UnmarshalJSON(body[:bodySize])
	if err != nil {
		return err
	}

	fmt.Printf("gateway %+v\n", respData)

	discordClient := NewDiscordMessageHandler()

	connectURL, err := url.Parse(respData.URL)
	if err != nil {
		return err
	}
	q := connectURL.Query()
	q.Add("v", "6")
	q.Add("encoding", "etf")
	connectURL.RawQuery = q.Encode()

	fmt.Printf("Connecting to %s\n", connectURL.String())

	wsClient := wsclient.NewWSClient(connectURL.String())
	wsClient.SetHandler(discordClient)
	d.wsClient = wsClient

	err = d.wsClient.Connect(d.config.BotToken)
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
	d.wsClient.Close()
	return nil
}

func (d *discordBot) Run() {
	d.wsClient.HandleRequests(d.config.NumWorkers)
}
