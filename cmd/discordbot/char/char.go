package char

import (
	"github.com/gsmcwhirter/eso-discord/pkg/cmdhandler"
	"github.com/gsmcwhirter/eso-discord/pkg/parser"
)

func show(args []rune) (string, error) {
	return "", nil
}

func create(args []rune) (string, error) {
	return "", nil
}

func delete(args []rune) (string, error) {
	return "", nil
}

// CommandHandler TODOC
func CommandHandler() cmdhandler.CommandHandler {
	p := parser.NewParser(parser.Options{
		CmdIndicator: ' ',
		KnownCommands: []string{
			"create",
			"delete",
			"show",
		},
	})
	ch := cmdhandler.NewCommandHandler(p)
	ch.SetHandler("show", cmdhandler.NewLineHandler(show))
	ch.SetHandler("create", cmdhandler.NewLineHandler(create))
	ch.SetHandler("create", cmdhandler.NewLineHandler(delete))

	return ch
}
