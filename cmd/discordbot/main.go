package main

import (
	"fmt"
	"os"

	"github.com/gsmcwhirter/eso-discord/pkg/commands"
	"github.com/gsmcwhirter/eso-discord/pkg/discordapi"
	"github.com/gsmcwhirter/eso-discord/pkg/options"
	"github.com/pelletier/go-toml"
)

// build time variables
var (
	AppName      string
	AppAuthor    string
	BuildDate    string
	BuildVersion string
	BuildSHA     string
)

type config struct {
	DiscordAPI   string `toml:"discord_api"`
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
	ClientToken  string `toml:"client_token"`
	Database     string `toml:"database"`
	NumWorkers   int    `toml:"num_workers"`
}

func main() {
	code, err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", AppName, err) // nolint: gas
	}

	os.Exit(code)
}

func setup() (conf config, err error) {
	helpStr := ""
	cli := options.OptionParser(AppName, AppAuthor, BuildVersion, BuildSHA, BuildDate, helpStr)
	cfile := cli.Flag("config", "The config file to use").Default("./config.toml").String()
	cid := cli.Flag("client-id", "The discord bot client id").String()
	cs := cli.Flag("client-secret", "The discord bot client secret").String()
	tok := cli.Flag("client-token", "The discord bot client token").String()
	dbFile := cli.Flag("database", "The bolt database file").String()
	numWorkers := cli.Flag("num-workers", "The number of worker goroutines to run").Int()

	_, err = cli.Parse(os.Args[1:])
	if err != nil {
		return
	}

	tomlConf, err := toml.LoadFile(*cfile)
	if err != nil {
		fmt.Printf("Could not load config file %s: %s\n", *cfile, err)
	}

	err = tomlConf.Unmarshal(&conf)
	if err != nil {
		fmt.Printf("Could not load config gile settings: %s\n", err)
	}

	if cid != nil && *cid != "" {
		conf.ClientID = *cid
	}

	if cs != nil && *cs != "" {
		conf.ClientSecret = *cs
	}

	if tok != nil && *tok != "" {
		conf.ClientToken = *tok
	}

	if dbFile != nil && *dbFile != "" {
		conf.Database = *dbFile
	}

	if numWorkers != nil && *numWorkers > 0 {
		conf.NumWorkers = *numWorkers
	}

	return
}

func run() (int, error) {
	config, err := setup()
	if err != nil {
		return -1, err
	}

	deps, err := createDependencies(config)
	if err != nil {
		return -1, err
	}
	defer deps.Close()

	_ = commands.CommandHandler(deps, fmt.Sprintf("%s (%s) (%s)", BuildVersion, BuildSHA, BuildDate))
	botConfig := discordapi.BotConfig{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		BotToken:     config.ClientToken,
		APIURL:       config.DiscordAPI,
		NumWorkers:   config.NumWorkers,
	}
	bot := discordapi.NewDiscordBot(deps, botConfig)
	err = bot.AuthenticateAndConnect()
	if err != nil {
		return -1, err
	}
	defer bot.Disconnect() // nolint: errcheck

	bot.Run()

	return 0, nil
}
