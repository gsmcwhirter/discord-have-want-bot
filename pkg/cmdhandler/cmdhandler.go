package cmdhandler

import (
	"errors"
	"fmt"

	"github.com/gsmcwhirter/eso-discord/pkg/parser"
)

// ErrMissingHandler TODOC
var ErrMissingHandler = errors.New("missing handler for command")

// LineHandler TODOC
type LineHandler interface {
	HandleLine(line []rune) (string, error)
}

type lineHandlerFunc struct {
	handler func(line []rune) (string, error)
}

// NewLineHandler TODOC
func NewLineHandler(f func([]rune) (string, error)) LineHandler {
	return lineHandlerFunc{handler: f}
}

func (lh lineHandlerFunc) HandleLine(line []rune) (string, error) {
	return lh.handler(line)
}

// CommandHandler TODOC
type CommandHandler struct {
	Parser   parser.Parser
	commands map[string]LineHandler
}

// NewCommandHandler TODOC
func NewCommandHandler(parser parser.Parser) CommandHandler {
	ch := CommandHandler{
		Parser:   parser,
		commands: map[string]LineHandler{},
	}
	return ch
}

// SetHandler TODOC
func (ch *CommandHandler) SetHandler(cmd string, handler LineHandler) {
	ch.commands[cmd] = handler
}

// HandleLine TODOC
func (ch CommandHandler) HandleLine(line []rune) (string, error) {
	cmd, rest, err := ch.Parser.ParseCommand(line)
	if err == parser.ErrUnknownCommand {
		return fmt.Sprintf("Unknown command '%s'", cmd), nil
	} else if err != nil {
		return "", err
	}

	subHandler, ok := ch.commands[cmd]
	if !ok {
		return "", ErrMissingHandler
	}

	return subHandler.HandleLine(rest)
}
