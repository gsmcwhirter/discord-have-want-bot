package need

import (
	"github.com/gsmcwhirter/eso-discord/pkg/cmdhandler"
	"github.com/gsmcwhirter/eso-discord/pkg/parser"
)

func points(args []rune) (string, error) {
	return "", nil
}

func item(args []rune) (string, error) {
	return "", nil
}

// CommandHandler TODOC
func CommandHandler() cmdhandler.CommandHandler {
	p := parser.NewParser(parser.Options{
		CmdIndicator: ' ',
		KnownCommands: []string{
			"pts",
			"item",
		},
	})
	ch := cmdhandler.NewCommandHandler(p)
	ch.SetHandler("pts", cmdhandler.NewLineHandler(points))
	ch.SetHandler("item", cmdhandler.NewLineHandler(item))

	return ch
}
