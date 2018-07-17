package main

import (
	"fmt"
	"os"

	toml "github.com/pelletier/go-toml"

	"github.com/gsmcwhirter/eso-discord/pkg/commands"
	"github.com/gsmcwhirter/eso-discord/pkg/options"
	"github.com/steven-ferrer/gonsole"
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

type config struct {
	Database string `toml:"database"`
	User     string `toml:"__not_there__"`
}

func setup() (config, error) {
	conf := config{}
	var err error

	helpStr := ""
	cli := options.OptionParser(AppName, AppAuthor, BuildVersion, BuildSHA, BuildDate, helpStr)
	cfile := cli.Flag("config", "The config file to use").Default("./config.toml").String()
	user := cli.Flag("user", "The discord user to impersonate").Short('u').Required().String()
	dbFile := cli.Flag("db", "The bolt database file").String()

	_, err = cli.Parse(os.Args[1:])
	if err != nil {
		return conf, err
	}

	tomlConf, err := toml.LoadFile(*cfile)
	if err != nil {
		fmt.Printf("Could not load config file %s: %s\n", *cfile, err)
	}

	err = tomlConf.Unmarshal(&conf)
	if err != nil {
		fmt.Printf("Could not load config file settings: %s\n", err)
	}

	if user != nil && *user != "" {
		conf.User = *user
	}

	if dbFile != nil && *dbFile != "" {
		conf.Database = *dbFile
	}

	return conf, err
}

func run() (int, error) {

	conf, err := setup()
	if err != nil {
		return -1, err
	}

	fmt.Printf("%+v\n", conf)

	deps, err := createDependencies(conf)
	if err != nil {
		return -1, err
	}
	defer deps.Close()

	ch := commands.CommandHandler(deps, fmt.Sprintf("%s (%s) (%s)", BuildVersion, BuildSHA, BuildDate), commands.Options{CmdIndicator: '!'})

	scanner := gonsole.NewReader(os.Stdin)
	var line string
	var resp string
	for {
		fmt.Print("> ")
		line, err = scanner.Line()

		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		if line == "" || line == "!q" {
			break
		}

		resp, err = ch.HandleLine(conf.User, []rune(line))
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			fmt.Println(resp)
		}
	}

	return 0, nil
}
