package main

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/procfile"
)

type ProcRunOptions struct {
	Procfile string   `json:"procfile,omitempty" flag:"procfile" help:"Path to the Procfile (default: nearest Procfile up to the git root)"`
	Detached bool     `json:"-" flag:"detached" help:"Internal: supervise without multiplexing output (used by 'proc start')"`
	Names    []string `json:"-" args:"true"`
}

func (ProcRunOptions) Help() string {
	return `Run the Procfile's processes in the foreground (foreman-style).

Each process's output is streamed to the terminal with a coloured "name |"
prefix and also captured to .gavel/proc/<name>.log. Press Ctrl-C to stop every
process and exit.

Pass process names to run only a subset:
  gavel proc run web worker`
}

func init() {
	cmd := clicky.AddNamedCommand("run", procCmd, ProcRunOptions{}, runProcRun)
	cmd.Use = "run [process...]"
	_ = cmd.Flags().MarkHidden("detached")
}

func runProcRun(opts ProcRunOptions) (any, error) {
	workDir, err := getWorkingDir()
	if err != nil {
		return nil, err
	}
	return nil, procfile.Run(workDir, opts.Procfile, opts.Names, !opts.Detached)
}
