package main

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/procfile"
)

type ProcStopOptions struct {
	Procfile string   `json:"procfile,omitempty" flag:"procfile" help:"Path to the Procfile (default: nearest Procfile up to the git root)"`
	Names    []string `json:"-" args:"true"`
}

func (ProcStopOptions) Help() string {
	return `Stop the running processes.

With no arguments the whole daemon is stopped (every process is terminated and
the supervisor exits). Pass process names to stop only those, leaving the
supervisor running:
  gavel proc stop worker

Stopping does nothing when no daemon is running.`
}

func init() {
	cmd := clicky.AddNamedCommand("stop", procCmd, ProcStopOptions{}, runProcStop)
	cmd.Use = "stop [process...]"
}

func runProcStop(opts ProcStopOptions) (any, error) {
	workDir, err := getWorkingDir()
	if err != nil {
		return nil, err
	}
	return procfile.Stop(workDir, opts.Procfile, opts.Names)
}
