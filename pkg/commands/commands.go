package commands

import (
	"github.com/gsmcwhirter/eso-discord/pkg/cmdhandler"
	"github.com/gsmcwhirter/eso-discord/pkg/commands/char"
	"github.com/gsmcwhirter/eso-discord/pkg/commands/got"
	"github.com/gsmcwhirter/eso-discord/pkg/commands/need"
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

// CommandHandler TODOC
func CommandHandler(deps dependencies, versionStr string) cmdhandler.CommandHandler {
	p := parser.NewParser(parser.Options{
		CmdIndicator: '!',
		KnownCommands: []string{
			"help",
			"version",
			"char",
			"items",
			"need",
			"got",
		},
	})
	rh := rootCommands{
		versionStr: versionStr,
	}
	ch := cmdhandler.NewCommandHandler(p)
	ch.SetHandler("version", cmdhandler.NewLineHandler(rh.version))
	ch.SetHandler("char", char.CommandHandler(deps))
	ch.SetHandler("need", need.CommandHandler(deps))
	ch.SetHandler("got", got.CommandHandler(deps))

	return ch
}
