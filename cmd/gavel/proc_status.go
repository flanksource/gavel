package main

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/procfile"
)

type ProcStatusOptions struct {
	Procfile string `json:"procfile,omitempty" flag:"procfile" help:"Path to the Procfile (default: nearest Procfile up to the git root)"`
}

func (ProcStatusOptions) Help() string {
	return `Show the status of each process: state, pid, uptime and restart count.

When a daemon is running the data is live (queried over the control socket);
otherwise every configured process is shown as stopped.`
}

func init() {
	clicky.AddNamedCommand("status", procCmd, ProcStatusOptions{}, runProcStatus)
}

func runProcStatus(opts ProcStatusOptions) (any, error) {
	workDir, err := getWorkingDir()
	if err != nil {
		return nil, err
	}
	return procfile.Status(workDir, opts.Procfile)
}
