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

// Options TODOC
type Options struct {
	CmdIndicator string
}

// RootCommands holds the commands at the root level
type rootCommands struct {
	versionStr string
}

func (c *rootCommands) version(user, guildw, args string) (cmdhandler.Response, error) {
	r := &cmdhandler.SimpleEmbedResponse{
		To:          user,
		Description: c.versionStr,
	}

	return r, nil
}

// CommandHandler TODOC
func CommandHandler(deps dependencies, versionStr string, opts Options) *cmdhandler.CommandHandler {
	p := parser.NewParser(parser.Options{
		CmdIndicator: opts.CmdIndicator,
	})
	rh := rootCommands{
		versionStr: versionStr,
	}

	ch := cmdhandler.NewCommandHandler(p, cmdhandler.Options{})
	ch.SetHandler("version", cmdhandler.NewLineHandler(rh.version))
	ch.SetHandler("char", CharCommandHandler(deps, fmt.Sprintf("%schar", opts.CmdIndicator)))
	ch.SetHandler("need", NeedCommandHandler(deps, fmt.Sprintf("%sneed", opts.CmdIndicator)))
	ch.SetHandler("got", GotCommandHandler(deps, fmt.Sprintf("%sgot", opts.CmdIndicator)))
	ch.SetHandler("list", ListCommandHandler(deps, fmt.Sprintf("%slist", opts.CmdIndicator)))

	return ch
}

type configDependencies interface {
	GuildAPI() storage.GuildAPI
}

// ConfigHandler TODOC
func ConfigHandler(deps configDependencies, versionStr string, opts Options) *cmdhandler.CommandHandler {
	p := parser.NewParser(parser.Options{
		CmdIndicator: opts.CmdIndicator,
	})

	ch := cmdhandler.NewCommandHandler(p, cmdhandler.Options{})
	ch.SetHandler("config", ConfigCommandHandler(deps, fmt.Sprintf("%sconfig", opts.CmdIndicator)))

	return ch
}
