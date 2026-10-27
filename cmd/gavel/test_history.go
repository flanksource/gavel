package main

import (
	"context"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/testrunner/history"
	"github.com/flanksource/gavel/workspace"
)

type testHistoryOptions struct {
	Paths []string `json:"paths,omitempty" args:"true"`
}

func (o testHistoryOptions) Help() api.Textable {
	return clicky.Text(`Show local test execution history from .gavel/run-*.json snapshots.

The history report aggregates completed gavel test runs and shows executable
test leaves grouped by package, file, and suite. Each test row includes
execution count, pass rate, min/avg/max duration, added date, last passed, and
last failed.

Optionally pass package or file paths to limit the report. Paths are matched
relative to --cwd.`)
}

func runTestHistory(ctx context.Context, opts testHistoryOptions) (any, error) {
	dirs, err := workspace.From(ctx)
	if err != nil {
		return nil, err
	}
	return history.Load(history.Options{
		WorkDir:   dirs.Invocation,
		StateRoot: dirs.StateRoot,
		Paths:     opts.Paths,
	})
}

func init() {
	cmd := clicky.AddNamedCommandWithContext("history", testCmd, testHistoryOptions{}, runTestHistory)
	cmd.Short = "Show test execution history from .gavel snapshots"
	cmd.Flags().SetInterspersed(true)
}
