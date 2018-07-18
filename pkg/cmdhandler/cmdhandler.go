package cmdhandler

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/gsmcwhirter/eso-discord/pkg/parser"
)

// ErrMissingHandler TODOC
var ErrMissingHandler = errors.New("missing handler for command")

// Options TODOC
type Options struct {
	Placeholder string
	PreCommand  string
}

// CommandHandler TODOC
type CommandHandler struct {
	parser      parser.Parser
	commands    map[string]LineHandler
	helpCmd     []rune
	placeholder string
	preCommand  string
}

// NewCommandHandler TODOC
func NewCommandHandler(parser parser.Parser, opts Options) *CommandHandler {
	ch := CommandHandler{
		commands:   map[string]LineHandler{},
		preCommand: opts.PreCommand,
	}
	ch.SetParser(parser)

	if opts.Placeholder != "" {
		ch.placeholder = opts.Placeholder
	} else {
		ch.placeholder = "command"
	}

	ch.SetHandler("", NewLineHandler(ch.showHelp))
	ch.SetHandler("help", NewLineHandler(ch.showHelp))
	return &ch
}

// CommandIndicator TODOC
func (ch *CommandHandler) CommandIndicator() rune {
	return ch.parser.LeadChar()
}

// SetParser sets the parser for the command handler
func (ch *CommandHandler) SetParser(p parser.Parser) {
	ch.parser = p
	ch.calculateHelpCmd()
}

func (ch *CommandHandler) calculateHelpCmd() {
	ch.helpCmd = append([]rune{ch.parser.LeadChar()}, []rune("help")...)
}

func (ch *CommandHandler) showHelp(user string, line []rune) (string, error) {
	var helpStr string
	if ch.preCommand != "" {
		helpStr = fmt.Sprintf("Usage: %s [%s]\n\n", ch.preCommand, ch.placeholder)
	}
	leadChar := string(ch.parser.LeadChar())
	helpStr += fmt.Sprintf("Available %ss:\n", ch.placeholder)
	for cmd := range ch.commands {
		if cmd != "" {
			helpStr += fmt.Sprintf("  %s%s\n", leadChar, cmd)
		}
	}
	return helpStr, nil
}

// SetHandler TODOC
func (ch *CommandHandler) SetHandler(cmd string, handler LineHandler) {
	ch.commands[cmd] = handler
}

// HandleLine TODOC
func (ch *CommandHandler) HandleLine(user string, line []rune) (string, error) {
	cmd, rest, err := ch.parser.ParseCommand(line)

	subHandler, cmdExists := ch.commands[cmd]
	if err == parser.ErrNotACommand && cmd != "" && !cmdExists {
		fmt.Println(cmd)
		return "", err
	}

	if err == parser.ErrUnknownCommand {
		cmd2, rest, err := ch.parser.ParseCommand(ch.helpCmd)

		subHandler, cmdExists := ch.commands[cmd2]
		if !cmdExists {
			return "", ErrMissingHandler
		}

		if err != nil {
			return fmt.Sprintf("Unknown command '%s'", cmd2), err
		}

		return subHandler.HandleLine(user, rest)

	}

	if err != nil && err != parser.ErrNotACommand {
		return "", err
	}

	if !cmdExists {
		return "", ErrMissingHandler
	}

	return subHandler.HandleLine(user, rest)
}
