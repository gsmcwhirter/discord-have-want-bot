package commands

import (
	"fmt"

	"github.com/gsmcwhirter/eso-discord/pkg/cmdhandler"
	"github.com/gsmcwhirter/eso-discord/pkg/parser"
	"github.com/gsmcwhirter/eso-discord/pkg/storage"
)

type dependencies interface {
	TrialAPI() storage.TrialAPI
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
	ch.SetHandler("create", CreateCommandHandler(deps, fmt.Sprintf("%screate", opts.CmdIndicator)))
	ch.SetHandler("signup", SignupCommandHandler(deps, fmt.Sprintf("%ssignup", opts.CmdIndicator)))
	ch.SetHandler("withdraw", WithdrawCommandHandler(deps, fmt.Sprintf("%swithdraw", opts.CmdIndicator)))
	ch.SetHandler("show", ShowCommandHandler(deps, fmt.Sprintf("%sshow", opts.CmdIndicator)))
	ch.SetHandler("list", ShowCommandHandler(deps, fmt.Sprintf("%slist", opts.CmdIndicator)))

	return ch
}
