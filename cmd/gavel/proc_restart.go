package main

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/task"
	commonsCtx "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/procfile"
)

type ProcRestartOptions struct {
	Procfile string   `json:"procfile,omitempty" flag:"procfile" help:"Path to the Procfile (default: nearest Procfile up to the git root)"`
	Profile  string   `json:"profile,omitempty" flag:"profile" help:"Active profile when starting a daemon that isn't running (default: .gavel.yaml procfile.profile)"`
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

	report, err := restartProcs(workDir, opts.Procfile, opts.Names, opts.Profile)
	if err != nil {
		return nil, err
	}
	renderProcReadiness(workDir, opts.Procfile, "Restarting processes", report, opts.Names)
	return procfile.Status(workDir, opts.Procfile)
}

// restartProcs performs the stop+restart, surfacing the stop as its own live
// task when a daemon is already running (so the user sees "Stopping…" before the
// per-process readiness view). When nothing is running, restart is a plain start
// and no stop task is shown.
func restartProcs(workDir, pf string, names []string, profile string) (*procfile.StatusReport, error) {
	st, err := procfile.Status(workDir, pf)
	if err != nil {
		return nil, err
	}
	if !st.Running {
		return procfile.Restart(workDir, pf, names, profile)
	}

	// The task returns an empty string (not the report) so clicky renders a
	// single "✓ Stopped processes" line rather than pretty-printing the whole
	// status table inline — the readiness view and final status do that.
	var report *procfile.StatusReport
	t := clicky.StartTask[string]("Stopping processes",
		func(_ commonsCtx.Context, _ *task.Task) (string, error) {
			r, err := procfile.Restart(workDir, pf, names, profile)
			report = r
			return "", err
		})
	if _, err := t.GetResult(); err != nil {
		t.Failed()
		return nil, err
	}
	t.SetName("Stopped processes")
	t.Success()
	return report, nil
}
