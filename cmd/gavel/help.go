package main

import (
	"github.com/flanksource/clicky"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const globalHelpHint = "Use gavel --help to show arguments available to all commands."

const gavelUsage = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if .Aliases}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if or .IsAvailableCommand (eq .Name "help")}}
  {{rpad .Name .NamePadding}} {{.Short}}{{end}}{{end}}{{end}}{{range gavelHelpFlagGroups .}}

{{.Title}}:
{{.Flags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}{{if .HasParent}}

` + globalHelpHint + `{{end}}
`

func setGavelUsage(cmd *cobra.Command) {
	cobra.AddTemplateFunc("gavelHelpFlagGroups", helpFlagGroups)
	cmd.SetUsageTemplate(gavelUsage)
}

func helpFlagGroups(cmd *cobra.Command) []clicky.FlagGroup {
	groups := clicky.CommandFlagGroups(cmd)
	root := cmd.Root().PersistentFlags()
	showRootFlags := !cmd.HasParent()
	visible := groups[:0]
	for _, group := range groups {
		flags := pflag.NewFlagSet(group.Title, pflag.ContinueOnError)
		group.Flags.VisitAll(func(flag *pflag.Flag) {
			isGlobal := root.Lookup(flag.Name) == flag
			if isGlobal == showRootFlags || (showRootFlags && flag.Name == "help") {
				flags.AddFlag(flag)
			}
		})
		if flags.HasAvailableFlags() {
			title := group.Title
			if !showRootFlags && title == "Global flags" {
				title = "Shared command flags"
			}
			visible = append(visible, clicky.FlagGroup{Title: title, Flags: flags})
		}
	}
	return visible
}
