package commands

import (
	"fmt"

	"github.com/gsmcwhirter/eso-discord/pkg/cmdhandler"
	"github.com/gsmcwhirter/eso-discord/pkg/parser"
	"github.com/gsmcwhirter/eso-discord/pkg/storage"
)

type dependencies interface {
	GuildAPI() storage.GuildAPI
}

// Options TODOC
type Options struct {
	CmdIndicator string
}

// ConfigHandler TODOC
func ConfigHandler(deps dependencies, versionStr string, opts Options) *cmdhandler.CommandHandler {
	p := parser.NewParser(parser.Options{
		CmdIndicator: opts.CmdIndicator,
	})

	ch := cmdhandler.NewCommandHandler(p, cmdhandler.Options{})
	ch.SetHandler("config", ConfigCommandHandler(deps, fmt.Sprintf("%sconfig", opts.CmdIndicator)))

	return ch
}
