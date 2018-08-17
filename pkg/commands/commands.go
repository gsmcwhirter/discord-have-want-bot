package commands

import (
	"fmt"

	"github.com/gsmcwhirter/discord-bot-lib/cmdhandler"
	"github.com/gsmcwhirter/discord-have-want-bot/pkg/storage"
	"github.com/gsmcwhirter/go-util/parser"
)

type dependencies interface {
	UserAPI() storage.UserAPI
}

// Options enables setting the command indicator string for a CommandHandler
type Options struct {
	CmdIndicator string
}

// CommandHandler creates a new command handler for !char, !need, !got, and !list
func CommandHandler(deps dependencies, versionStr string, opts Options) (*cmdhandler.CommandHandler, error) {
	p := parser.NewParser(parser.Options{
		CmdIndicator: opts.CmdIndicator,
	})

	ch, err := cmdhandler.NewCommandHandler(p, cmdhandler.Options{})
	if err != nil {
		return nil, err
	}

	cch, err := CharCommandHandler(deps, fmt.Sprintf("%schar", opts.CmdIndicator))
	if err != nil {
		return nil, err
	}
	ch.SetHandler("char", cch)

	nch, err := NeedCommandHandler(deps, fmt.Sprintf("%sneed", opts.CmdIndicator))
	if err != nil {
		return nil, err
	}
	ch.SetHandler("need", nch)

	gch, err := GotCommandHandler(deps, fmt.Sprintf("%sgot", opts.CmdIndicator))
	if err != nil {
		return nil, err
	}
	ch.SetHandler("got", gch)

	lch, err := ListCommandHandler(deps, fmt.Sprintf("%slist", opts.CmdIndicator))
	if err != nil {
		return nil, err
	}
	ch.SetHandler("list", lch)

	return ch, nil
}

type configDependencies interface {
	GuildAPI() storage.GuildAPI
}

// ConfigHandler creates a new command handler for !config-hw
func ConfigHandler(deps configDependencies, versionStr string, opts Options) (*cmdhandler.CommandHandler, error) {
	p := parser.NewParser(parser.Options{
		CmdIndicator: opts.CmdIndicator,
	})

	ch, err := cmdhandler.NewCommandHandler(p, cmdhandler.Options{
		NoHelpOnUnknownCommands: true,
	})
	if err != nil {
		return nil, err
	}

	cch, err := ConfigCommandHandler(deps, versionStr, fmt.Sprintf("%sconfig", opts.CmdIndicator))
	if err != nil {
		return nil, err
	}
	ch.SetHandler("config-hw", cch)
	// disable help for config
	ch.SetHandler("help", cmdhandler.NewMessageHandler(func(msg cmdhandler.Message) (cmdhandler.Response, error) {
		r := &cmdhandler.SimpleEmbedResponse{}
		return r, parser.ErrUnknownCommand
	}))

	return ch, nil
}
