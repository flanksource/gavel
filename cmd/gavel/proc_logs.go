package main

import (
	"os"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/procfile"
)

type ProcLogsOptions struct {
	Procfile string   `json:"procfile,omitempty" flag:"procfile" help:"Path to the Procfile (default: nearest Procfile up to the git root)"`
	Lines    int      `json:"lines,omitempty" flag:"lines" short:"n" help:"Number of trailing log lines to show per process" default:"50"`
	Follow   bool     `json:"follow,omitempty" flag:"follow" short:"f" help:"Stream new output until interrupted"`
	Names    []string `json:"-" args:"true"`
}

func (ProcLogsOptions) Help() string {
	return `Tail the per-process logs under .gavel/proc/<name>.log.

With no arguments every process's log is shown; pass names to select a subset.
Use -f/--follow to stream new output until interrupted:
  gavel proc logs -f web worker`
}

func init() {
	cmd := clicky.AddNamedCommand("logs", procCmd, ProcLogsOptions{}, runProcLogs)
	cmd.Use = "logs [process...]"
}

func runProcLogs(opts ProcLogsOptions) (any, error) {
	workDir, err := getWorkingDir()
	if err != nil {
		return nil, err
	}
	return nil, procfile.Logs(workDir, opts.Procfile, opts.Names, opts.Lines, opts.Follow, os.Stdout)
}
