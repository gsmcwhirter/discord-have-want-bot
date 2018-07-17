package cmdhandler

import (
	"errors"
	"fmt"

	"github.com/gsmcwhirter/eso-discord/pkg/parser"
)

// ErrMissingHandler TODOC
var ErrMissingHandler = errors.New("missing handler for command")

// CommandHandler TODOC
type CommandHandler struct {
	parser   parser.Parser
	commands map[string]LineHandler
	helpCmd  []rune
}

// NewCommandHandler TODOC
func NewCommandHandler(parser parser.Parser) *CommandHandler {
	ch := CommandHandler{
		commands: map[string]LineHandler{},
	}
	ch.SetParser(parser)

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
	leadChar := string(ch.parser.LeadChar())
	helpStr := "Available Commands:\n"
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

	if cmd == "" && err == parser.ErrUnknownCommand {
		cmd, rest, err = ch.parser.ParseCommand(ch.helpCmd)
	}

	if err == parser.ErrUnknownCommand {
		return fmt.Sprintf("Unknown command '%s'", cmd), nil
	}

	if err != nil {
		return "", err
	}

	subHandler, ok := ch.commands[cmd]
	if !ok {
		return "", ErrMissingHandler
	}

	return subHandler.HandleLine(user, rest)
}
