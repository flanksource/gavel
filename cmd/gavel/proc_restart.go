package main

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/procfile"
)

type ProcRestartOptions struct {
	Procfile string   `json:"procfile,omitempty" flag:"procfile" help:"Path to the Procfile (default: nearest Procfile up to the git root)"`
	Profile  string   `json:"profile,omitempty" flag:"profile" help:"Active profile when starting a daemon that isn't running (default: .gavel.yaml procfile.profile)"`
	Follow   bool     `json:"follow,omitempty" flag:"follow" short:"f" help:"Keep streaming process logs until interrupted (default: stream until they settle)"`
	Names    []string `json:"-" args:"true"`
}

func (ProcRestartOptions) Help() string {
	return `Restart the running processes.

With no arguments every process is restarted on the running supervisor. Pass
names to restart only those:
  gavel proc restart web

When no daemon is running, restart starts one (equivalent to ` + "`gavel proc start`" + `).

Process logs are streamed while the restarted processes come back up; -f/--follow
keeps streaming until interrupted. Exits non-zero if a process fails to start.`
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
	report, err := procfile.Restart(workDir, opts.Procfile, opts.Names, opts.Profile)
	if err != nil {
		return nil, err
	}
	if err := streamUntilReady(workDir, opts.Procfile, report, opts.Names, opts.Follow); err != nil {
		return nil, err
	}
	return procfile.Status(workDir, opts.Procfile)
}
