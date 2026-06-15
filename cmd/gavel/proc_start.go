package main

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/procfile"
)

type ProcStartOptions struct {
	Procfile string   `json:"procfile,omitempty" flag:"procfile" help:"Path to the Procfile (default: nearest Procfile up to the git root)"`
	Profile  string   `json:"profile,omitempty" flag:"profile" help:"Active profile; entries with 'profiles' auto-start only when it matches (default: .gavel.yaml procfile.profile)"`
	Names    []string `json:"-" args:"true"`
}

func (ProcStartOptions) Help() string {
	return `Start the Procfile's processes as a detached background daemon.

A supervisor process is spawned in its own session; it owns every process and
survives this command returning. Use ` + "`gavel proc status`" + ` to inspect it,
` + "`gavel proc logs -f`" + ` to follow output, and ` + "`gavel proc stop`" + ` to shut it
down. Refuses to start when a daemon is already running for this Procfile.

Pass process names to start only a subset:
  gavel proc start web worker`
}

func init() {
	cmd := clicky.AddNamedCommand("start", procCmd, ProcStartOptions{}, runProcStart)
	cmd.Use = "start [process...]"
}

func runProcStart(opts ProcStartOptions) (any, error) {
	workDir, err := getWorkingDir()
	if err != nil {
		return nil, err
	}
	report, err := procfile.Start(workDir, opts.Procfile, opts.Names, opts.Profile)
	if err != nil {
		return nil, err
	}
	renderProcReadiness(workDir, opts.Procfile, "Starting processes", report, opts.Names)
	return procfile.Status(workDir, opts.Procfile)
}
