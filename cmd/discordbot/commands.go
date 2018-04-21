package main

import (
	"github.com/gsmcwhirter/eso-discord/cmd/discordbot/char"
	"github.com/gsmcwhirter/eso-discord/cmd/discordbot/got"
	"github.com/gsmcwhirter/eso-discord/cmd/discordbot/need"
	"github.com/gsmcwhirter/eso-discord/pkg/cmdhandler"
	"github.com/gsmcwhirter/eso-discord/pkg/parser"
)

func items(args []rune) (string, error) {
	return "", nil
}

// CommandHandler TODOC
func CommandHandler() cmdhandler.CommandHandler {
	p := parser.NewParser(parser.Options{
		CmdIndicator: '!',
		KnownCommands: []string{
			"char",
			"items",
			"need",
			"got",
		},
	})
	ch := cmdhandler.NewCommandHandler(p)
	ch.SetHandler("char", char.CommandHandler())
	ch.SetHandler("need", need.CommandHandler())
	ch.SetHandler("got", got.CommandHandler())
	ch.SetHandler("items", cmdhandler.NewLineHandler(items))

	return ch
}
