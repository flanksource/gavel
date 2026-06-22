package main

import (
	"github.com/flanksource/clicky"
)

type ProcStopOptions struct {
	Procfile string   `json:"procfile,omitempty" flag:"procfile" help:"Path to the Procfile (default: nearest Procfile up to the git root)"`
	Follow   bool     `json:"follow,omitempty" flag:"follow" short:"f" help:"Keep streaming process logs until interrupted (default: stream until they stop)"`
	Names    []string `json:"-" args:"true"`
}

func (ProcStopOptions) Help() string {
	return `Stop the running processes.

With no arguments the whole daemon is stopped (every process is terminated and
the supervisor exits). Pass process names to stop only those, leaving the
supervisor running:
  gavel proc stop worker

Process logs are streamed while the processes shut down; -f/--follow keeps
streaming until interrupted. Stopping does nothing when no daemon is running.`
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
	return stopAndStream(workDir, opts.Procfile, opts.Names, opts.Follow)
}
