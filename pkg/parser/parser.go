package parser

import "errors"

// Parser TODOC
type Parser interface {
	ParseCommand(line []rune) (cmd string, rest []rune, err error)
	KnownCommand(cmd string) bool
	LeadChar() rune
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

// LeadChar returns the character that identifies commands
func (p parser) LeadChar() rune {
	return p.CmdIndicator
}

func (p parser) ParseCommand(line []rune) (cmd string, rest []rune, err error) {
	if len(line) == 0 {
		if p.CmdIndicator == ' ' && p.KnownCommand("") {
			return "", line, nil
		}

		err = ErrNotACommand
		return
	}

	if line[0] != p.CmdIndicator {
		err = ErrNotACommand
		return
	}

	for i := range line {
		if i == 0 {
			continue
		}

		if line[i] == ' ' {
			cmd = string(line[1:i])
			rest = line[i:]

			if known, _ := p.knownCommands[cmd]; known {
				return
			}
		}
	}

	cmd = string(line[1:])
	rest = line[0:0]
	if known, _ := p.knownCommands[cmd]; !known {
		err = ErrUnknownCommand
	}
	return
}

var digits = map[rune]bool{
	'0': true,
	'1': true,
	'2': true,
	'3': true,
	'4': true,
	'5': true,
	'6': true,
	'7': true,
	'8': true,
	'9': true,
}

func MaybeCount(line []rune) (l []rune, c []rune) {
	l = line

	for i := len(line) - 1; i >= 0; i-- {
		_, isDigit := digits[line[i]]
		if !isDigit {
			if line[i] == 'x' {
				l = line[:i]
				c = line[i+1:]
			} else {
				l = line[:i+1]
				c = line[i+1:]
			}
			return
		}
	}

	return
}
