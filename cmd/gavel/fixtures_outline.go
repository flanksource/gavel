package main

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/fixtures"
	"github.com/spf13/cobra"
)

type fixturesOutlineOptions struct {
	Paths []string `json:"paths,omitempty" args:"true"`
}

func (o fixturesOutlineOptions) Help() api.Textable {
	return clicky.Text(`Statically outline fixture markdown files without running them.

Renders the parsed fixture file, section, table, and test-step tree with one row
per fixture. Rows include fixture kind, source location, origin kind, and kind
counts. This command only parses fixture files; it does not run build commands,
daemons, exec fixtures, test/lint runner steps, skip commands, or AI checks.`)
}

func runFixturesOutline(opts fixturesOutlineOptions) (any, error) {
	wd, err := getWorkingDir()
	if err != nil {
		return nil, err
	}
	return fixtures.Outline(fixtures.OutlineOptions{
		Paths:   opts.Paths,
		WorkDir: wd,
	})
}

func init() {
	cmd := clicky.AddNamedCommand("outline", fixturesCmd, fixturesOutlineOptions{}, runFixturesOutline)
	cmd.Short = "Static outline of fixture markdown without running fixtures"
	cmd.Args = cobra.MinimumNArgs(1)
	cmd.Flags().SetInterspersed(true)
}
