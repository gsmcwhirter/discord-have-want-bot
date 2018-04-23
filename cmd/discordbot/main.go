package main

import (
	"fmt"
	"os"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/gsmcwhirter/eso-discord/pkg/commands"
	"github.com/gsmcwhirter/eso-discord/pkg/discordapi"
	"github.com/gsmcwhirter/eso-discord/pkg/options"
	"github.com/gsmcwhirter/eso-discord/pkg/storage"
)

// build time variables
var (
	AppName      string
	AppAuthor    string
	BuildDate    string
	BuildVersion string
	BuildSHA     string
)

func main() {
	code, err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", AppName, err) // nolint: gas
	}

	os.Exit(code)
}

func run() (int, error) {
	helpStr := ""
	cli := options.OptionParser(AppName, AppAuthor, BuildVersion, BuildSHA, BuildDate, helpStr)
	cid := cli.Flag("client-id", "The discord bot client id").Default("437746796207538176").String()
	cs := cli.Flag("client-secret", "The discord bot client secret").Required().String()
	tok := cli.Flag("client-token", "The discord bot client token").Required().String()
	dbFile := cli.Flag("db", "The bolt database file").Short('d').Required().String()
	numWorkers := cli.Flag("num-workers", "The number of worker goroutines to run").Short('n').Default("10").Int()

	_, err := cli.Parse(os.Args[1:])
	if err != nil {
		return -1, err
	}

	db, err := bolt.Open(*dbFile, 0660, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return -1, err
	}
	defer db.Close() // nolint: errcheck

	userAPI, err := storage.NewBoltUserAPI(db)
	if err != nil {
		return -1, err
	}

	_ = commands.CommandHandler(userAPI, fmt.Sprintf("%s (%s) (%s)", BuildVersion, BuildSHA, BuildDate))
	botConfig := discordapi.BotConfig{
		ClientID:     *cid,
		ClientSecret: *cs,
		BotToken:     *tok,
		APIURL:       "https://discordapp.com/api",
		NumWorkers:   *numWorkers,
	}
	bot := discordapi.NewDiscordBot(botConfig)
	err = bot.AuthenticateAndConnect()
	if err != nil {
		return -1, err
	}
	defer bot.Disconnect() // nolint: errcheck

	bot.Run()

	return 0, nil
}
