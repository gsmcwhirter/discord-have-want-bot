package commands

import (
	"fmt"

	"github.com/gsmcwhirter/eso-discord/pkg/cmdhandler"
	"github.com/gsmcwhirter/eso-discord/pkg/parser"
	"github.com/gsmcwhirter/eso-discord/pkg/storage"
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

func (r rootCommands) version(user, guildw, args string) (string, error) {
	return r.versionStr, nil
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
