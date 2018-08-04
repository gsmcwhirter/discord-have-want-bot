package commands

import (
	"github.com/gsmcwhirter/eso-discord/pkg/cmdhandler"
	"github.com/gsmcwhirter/eso-discord/pkg/parser"
)

// ListCommandHandler TODOC
func ListCommandHandler(deps dependencies, preCommand string) *cmdhandler.CommandHandler {
	p := parser.NewParser(parser.Options{
		CmdIndicator: " ",
	})
	// cc := charCommands{
	// 	preCommand: preCommand,
	// 	deps:       deps,
	// }
	ch := cmdhandler.NewCommandHandler(p, cmdhandler.Options{
		PreCommand:  preCommand,
		Placeholder: "action",
	})

	// TODO

	return ch
}
