package main

import (
	"fmt"
	"os"

	"github.com/gsmcwhirter/discord-bot-lib/cmdhandler"

	"github.com/steven-ferrer/gonsole"

	"github.com/gsmcwhirter/discord-have-want-bot/pkg/commands"
)

type config struct {
	Database  string `mapstructure:"database"`
	User      string `mapstructure:"user"`
	TestThing string `mapstructure:"test_thing"`
}

func start(c config) error {
	fmt.Printf("%+v\n", c)

	deps, err := createDependencies(c)
	if err != nil {
		return err
	}
	defer deps.Close()

	ch := commands.CommandHandler(deps, fmt.Sprintf("%s (%s) (%s)", BuildVersion, BuildSHA, BuildDate), commands.Options{CmdIndicator: "!"})

	scanner := gonsole.NewReader(os.Stdin)
	var line string
	var resp cmdhandler.Response
	for {
		fmt.Print("> ")
		line, err = scanner.Line()

		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		if line == "" || line == "!q" {
			break
		}

		resp, err = ch.HandleLine(c.User, "", line)
		if err != nil {
			resp.IncludeError(err)
		}

		fmt.Println(resp.ToString())
	}

	return nil
}
