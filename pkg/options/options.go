package options

import (
	"fmt"

	"github.com/tj/kingpin"
)

const usageTemplate = `{{define "FormatCommand"}}\
{{if .FlagSummary}} {{.FlagSummary}}{{end}}\
{{range .Args}} {{if not .Required}}[{{end}}<{{.Name}}>{{if .Value|IsCumulative}}...{{end}}{{if not .Required}}]{{end}}{{end}}\
{{end}}\

{{define "FormatCommands"}}\
{{range .FlattenedCommands}}\
{{if not .Hidden}}\
    {{printf "%-20s %s" .FullCommand .Help}}
{{end}}\
{{end}}\
{{end}}\

{{define "FormatUsage"}}\
{{template "FormatCommand" .}}{{if .Commands}} <command> [<args> ...]{{end}}

  {{"Description:" | bold}}
{{if .Help}}
{{.Help|Wrap 4}}\
{{end}}\
{{end}}

{{if .Context.SelectedCommand}}\
  {{"Usage:" | bold}}
    {{.App.Name}} {{.Context.SelectedCommand}}{{template "FormatUsage" .Context.SelectedCommand}}
{{else}}\
  {{"Usage:" | bold}}
    {{.App.Name}}{{template "FormatUsage" .App}}
{{end}}\
{{if .Context.Flags}}\
  {{"Flags:" | bold}}
{{.Context.Flags|FlagsToTwoColumns|FormatTwoColumns}}
{{end}}\
{{if .Context.Args}}\
  {{"Args:" | bold}}
{{.Context.Args|ArgsToTwoColumns|FormatTwoColumns}}
{{end}}\
{{if .Context.SelectedCommand}}\
{{if len .Context.SelectedCommand.Commands}}\
  {{"Subcommands:" | bold}}
{{template "FormatCommands" .Context.SelectedCommand}}
{{end}}\
{{else if .App.Commands}}\
  {{"Commands:" | bold}}
{{template "FormatCommands" .App}}
{{end}}\
{{define "Examples"}}\
{{if .}}\
  {{"Examples:" | bold}}
  {{range .}}
    {{.Help}}
    $ {{.Usage}}
  {{end}}
{{end}}\
{{end}}\
{{if .Context.SelectedCommand}}\
{{template "Examples" .Context.SelectedCommand.Examples}}\
{{else if .App.Examples}}\
{{template "Examples" .App.Examples}}\
{{end}}\
`

// OptionParser creates a new, standard option parser layout
func OptionParser(appName, appAuthor, buildVersion, buildSHA, buildDate, helpStr string) *kingpin.Application {
	CLI := kingpin.New(appName, helpStr)
	CLI.UsageTemplate(usageTemplate)
	CLI.Version(fmt.Sprintf("%s (%s)", buildVersion, buildSHA))
	CLI.Author(fmt.Sprintf("%s (built on %s, sha %s)", appAuthor, buildDate, buildSHA))

	return CLI
}
