package main

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/procfile"
)

type ProcListOptions struct {
	Procfile string `json:"procfile,omitempty" flag:"procfile" help:"Path to the Procfile (default: nearest Procfile up to the git root)"`
}

func (ProcListOptions) Help() string {
	return `List the processes defined in the Procfile and their commands, annotated
with each process's current status.`
}

func init() {
	clicky.AddNamedCommand("list", procCmd, ProcListOptions{}, runProcList)
}

func runProcList(opts ProcListOptions) (any, error) {
	workDir, err := getWorkingDir()
	if err != nil {
		return nil, err
	}
	return procfile.List(workDir, opts.Procfile)
}
