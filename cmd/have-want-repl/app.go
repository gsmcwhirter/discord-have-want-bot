package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gsmcwhirter/discord-bot-lib/cmdhandler"
	"github.com/gsmcwhirter/discord-bot-lib/snowflake"
	"github.com/pkg/errors"

	"github.com/steven-ferrer/gonsole"

	"github.com/gsmcwhirter/discord-have-want-bot/pkg/commands"
)

type config struct {
	Database string `mapstructure:"database"`
	User     string `mapstructure:"user"`
	Guild    string `mapstructure:"guild"`
	Channel  string `mapstructure:"channel"`
}

func start(c config) error {
	fmt.Printf("%+v\n", c)

	deps, err := createDependencies(c)
	if err != nil {
		return err
	}
	defer deps.Close()

	ch, err := commands.CommandHandler(deps, fmt.Sprintf("%s (%s) (%s)", BuildVersion, BuildSHA, BuildDate), commands.Options{CmdIndicator: "!"})
	if err != nil {
		return err
	}

	uid, err := snowflake.FromString(c.User)
	if err != nil {
		return errors.Wrap(err, "could not parse user id")
	}

	gid, err := snowflake.FromString(c.Guild)
	if err != nil {
		return errors.Wrap(err, "could not parse guild id")
	}

	cid, err := snowflake.FromString(c.Channel)
	if err != nil {
		return errors.Wrap(err, "could not parse channel id")
	}

	baseMsg := cmdhandler.NewSimpleMessage(context.Background(), uid, gid, cid, 0, "")

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

		resp, err = ch.HandleLine(cmdhandler.NewWithContents(baseMsg, line))
		if err != nil {
			resp.IncludeError(err)
		}

		fmt.Println(resp.ToString())
	}

	return nil
}
