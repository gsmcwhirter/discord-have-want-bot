package options

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// CommandOptions TODOC
type CommandOptions struct {
	ShortHelp    string
	LongHelp     string
	Example      string
	PosArgsUsage string
	Deprecated   string
	Args         cobra.PositionalArgs

	Aliases      []string
	Hidden       bool
	SilenceUsage bool
}

// NoArgs TODOC
var NoArgs = cobra.NoArgs

// ExactArgs TODOC
var ExactArgs = cobra.ExactArgs

// MinimumNArgs TODOC
var MinimumNArgs = cobra.MinimumNArgs

// MaximumNArgs TODOC
var MaximumNArgs = cobra.MaximumNArgs

// RangeArgs TODOC
var RangeArgs = cobra.RangeArgs

// Command TODOC
type Command struct {
	*cobra.Command
}

// NewCLI TODOC
func NewCLI(appName, buildVersion, buildSHA, buildDate string, opts CommandOptions) *Command {
	c := NewCommand(appName, opts)
	c.Version = fmt.Sprintf("%s (%s, %s)", buildVersion, buildDate, buildSHA)
	return c
}

// NewCommand TODOC
func NewCommand(cmdName string, opts CommandOptions) *Command {
	var use string
	if opts.PosArgsUsage != "" {
		use = fmt.Sprintf("%s %s", cmdName, opts.PosArgsUsage)
	} else {
		use = cmdName
	}

	return &Command{
		&cobra.Command{
			Use:     use,
			Short:   opts.ShortHelp,
			Long:    opts.LongHelp,
			Example: opts.Example,

			Deprecated:   opts.Deprecated,
			SilenceUsage: opts.SilenceUsage,

			Args: opts.Args,

			Aliases: opts.Aliases,
			Hidden:  opts.Hidden,

			RunE: func(cmd *cobra.Command, args []string) error {
				return cmd.Help()
			},
		},
	}
}

// AddExamples TODOC
func (c *Command) AddExamples(descCmds ...string) {
	b := strings.Builder{}
	_, _ = b.WriteString(c.Example)
	for i := 0; i < len(descCmds)/2*2; i += 2 {
		_, _ = b.WriteString(fmt.Sprintf(`
  %s:
	$ %s
`, descCmds[i], descCmds[i+1]))
	}
	c.Example = b.String()
}

// SetRunFunc TODOC
func (c *Command) SetRunFunc(run func(cmd *Command, args []string) error) {
	c.RunE = func(ccmd *cobra.Command, args []string) error {
		return run(&Command{ccmd}, args)
	}
}

// AddSubCommands TODOC
func (c *Command) AddSubCommands(cmds ...*Command) {
	ccmds := make([]*cobra.Command, 0, len(cmds))
	for _, cmd := range cmds {
		ccmds = append(ccmds, cmd.Command)
	}
	c.AddCommand(ccmds...)
}
