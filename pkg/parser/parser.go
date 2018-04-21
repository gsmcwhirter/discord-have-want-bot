package parser

import "errors"

// Parser TODOC
type Parser interface {
	ParseCommand(line []rune) (cmd string, rest []rune, err error)
	KnownCommand(cmd string) bool
}

// ErrNotACommand TODOC
var ErrNotACommand = errors.New("not a command")

// ErrUnknownCommand TODOC
var ErrUnknownCommand = errors.New("unknown command")

type parser struct {
	CmdIndicator  rune
	knownCommands map[string]bool
}

// Options TODOC
type Options struct {
	CmdIndicator  rune
	KnownCommands []string
}

// NewParser TODOC
func NewParser(opts Options) Parser {
	p := parser{
		CmdIndicator:  opts.CmdIndicator,
		knownCommands: map[string]bool{},
	}

	for _, cmd := range opts.KnownCommands {
		p.knownCommands[cmd] = true
	}
	return p
}

// KnownCommand TODOC
func (p parser) KnownCommand(cmd string) bool {
	known, _ := p.knownCommands[cmd]
	return known
}

func (p parser) ParseCommand(line []rune) (cmd string, rest []rune, err error) {
	if len(line) == 0 {
		err = ErrNotACommand
		return
	}

	if line[0] != p.CmdIndicator {
		err = ErrNotACommand
		return
	}

	for i := range line {
		if line[i] == ' ' {
			cmd = string(line[1:i])
			rest = line[i:]

			if known, _ := p.knownCommands[cmd]; !known {
				err = ErrUnknownCommand
			}
			return
		}
	}

	cmd = string(line[1:])
	rest = line[0:0]
	if known, _ := p.knownCommands[cmd]; !known {
		err = ErrUnknownCommand
	}
	return
}
