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

// RootCommands holds the commands at the root level
type rootCommands struct {
	versionStr string
}

func (r rootCommands) version(user string, args []rune) (string, error) {
	return r.versionStr, nil
}

// Options TODOC
type Options struct {
	CmdIndicator rune
}

// CommandHandler TODOC
func CommandHandler(deps dependencies, versionStr string, opts Options) *cmdhandler.CommandHandler {
	p := parser.NewParser(parser.Options{
		CmdIndicator: opts.CmdIndicator,
		KnownCommands: []string{
			"help",
			"version",
			"char",
			"list",
			"need",
			"got",
		},
	})
	rh := rootCommands{
		versionStr: versionStr,
	}
	ciString := string([]rune{opts.CmdIndicator})
	ch := cmdhandler.NewCommandHandler(p, cmdhandler.Options{})
	ch.SetHandler("version", cmdhandler.NewLineHandler(rh.version))
	ch.SetHandler("char", CharCommandHandler(deps, fmt.Sprintf("%schar", ciString)))
	ch.SetHandler("need", NeedCommandHandler(deps, fmt.Sprintf("%sneed", ciString)))
	ch.SetHandler("got", GotCommandHandler(deps, fmt.Sprintf("%sgot", ciString)))
	ch.SetHandler("list", ListCommandHandler(deps, fmt.Sprintf("%slist", ciString)))

	return ch
}
