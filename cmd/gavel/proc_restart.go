package main

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/procfile"
)

type ProcRestartOptions struct {
	Procfile string   `json:"procfile,omitempty" flag:"procfile" help:"Path to the Procfile (default: nearest Procfile up to the git root)"`
	Names    []string `json:"-" args:"true"`
}

func (ProcRestartOptions) Help() string {
	return `Restart the running processes.

With no arguments every process is restarted on the running supervisor. Pass
names to restart only those:
  gavel proc restart web

When no daemon is running, restart starts one (equivalent to ` + "`gavel proc start`" + `).`
}

func init() {
	cmd := clicky.AddNamedCommand("restart", procCmd, ProcRestartOptions{}, runProcRestart)
	cmd.Use = "restart [process...]"
}

func runProcRestart(opts ProcRestartOptions) (any, error) {
	workDir, err := getWorkingDir()
	if err != nil {
		return nil, err
	}
	return procfile.Restart(workDir, opts.Procfile, opts.Names)
}
