package main

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/testrunner/history"
)

type testHistoryOptions struct {
	Paths []string `json:"paths,omitempty" args:"true"`
}

func (o testHistoryOptions) Help() string {
	return `Show local test execution history from .gavel/run-*.json snapshots.

The history report aggregates completed gavel test runs and shows executable
test leaves grouped by package, file, and suite. Each test row includes
execution count, pass rate, min/avg/max duration, added date, last passed, and
last failed.

Optionally pass package or file paths to limit the report. Paths are matched
relative to --cwd.`
}

func runTestHistory(opts testHistoryOptions) (any, error) {
	workDir, err := getWorkingDir()
	if err != nil {
		return nil, err
	}
	return history.Load(history.Options{
		WorkDir: workDir,
		Paths:   opts.Paths,
	})
}

func init() {
	cmd := clicky.AddNamedCommand("history", testCmd, testHistoryOptions{}, runTestHistory)
	cmd.Short = "Show test execution history from .gavel snapshots"
	cmd.Flags().SetInterspersed(true)
}
