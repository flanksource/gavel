package main

import (
	"fmt"
	"os"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/commons/duration"
	"github.com/flanksource/gavel/service"
)

type SystemStartOptions struct {
	Wait        bool              `flag:"wait" help:"Poll until the UI port accepts connections (or the process dies) before returning" default:"true"`
	WaitTimeout duration.Duration `flag:"wait-timeout" help:"How long to wait for readiness before giving up" default:"30s"`
}

func (SystemStartOptions) Help() string {
	return `Start the PR UI in the background as a detached process.

Runs ` + "`gavel pr list --all --ui --menu-bar`" + ` — the --menu-bar flag shows the
status indicator on macOS and is a no-op on Linux.

If an existing gavel pr-ui daemon is running (tracked via the pidfile at
~/.config/gavel/pr-ui.pid) it is terminated first so the new instance can
bind the UI port. The logfile at ~/.config/gavel/pr-ui.log is truncated on
each start so you only see output from the current run.

Readiness (--wait, default true):
  The command polls every 500ms until the UI port (read from
  ~/.config/gavel/pr-ui.port, defaulting to 9092) accepts TCP connections
  or --wait-timeout elapses. If the daemon crashes during startup the full
  logfile is printed so you can see what went wrong. Pass --wait=false to
  return immediately after spawning the process.`
}

func init() {
	clicky.AddNamedCommand("start", systemCmd, SystemStartOptions{}, runSystemStart)
}

// startReport is what system start returns when it completes — captures the
// pid and readiness outcome so Pretty() can render a colored summary.
type startReport struct {
	PID      int
	PidFile  string
	LogFile  string
	Waited   bool
	Outcome  service.Readiness
	Timeout  time.Duration
	Duration time.Duration
}

func runSystemStart(opts SystemStartOptions) (any, error) {
	pid, err := service.Start(service.StartOptions{})
	if err != nil {
		return nil, err
	}
	pidFile, _ := service.PidFile()
	logPath, _ := service.LogFile()
	timeout := time.Duration(opts.WaitTimeout)
	report := startReport{
		PID:     pid,
		PidFile: pidFile,
		LogFile: logPath,
		Waited:  opts.Wait,
		Outcome: service.ReadinessReady,
		Timeout: timeout,
	}

	if !opts.Wait {
		return report, nil
	}

	start := time.Now()
	outcome, err := service.WaitForReady(timeout)
	report.Duration = time.Since(start)
	report.Outcome = outcome
	if err != nil {
		return report, fmt.Errorf("wait for ready: %w", err)
	}

	if outcome != service.ReadinessReady {
		dumpStartupLog(logPath)
		return report, fmt.Errorf("daemon did not become ready: %s", outcome)
	}
	return report, nil
}

// dumpStartupLog writes the whole logfile to stderr so operators can see why
// the daemon crashed or failed to bind. We intentionally read the full file
// (not TailLog) because the log was just truncated on this start and a
// failing daemon rarely produces more than a few KB.
func dumpStartupLog(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n--- unable to read log %s: %v ---\n", path, err)
		return
	}
	fmt.Fprintf(os.Stderr, "\n--- log %s ---\n", path)
	_, _ = os.Stderr.Write(b)
	if len(b) > 0 && b[len(b)-1] != '\n' {
		fmt.Fprintln(os.Stderr)
	}
	fmt.Fprintln(os.Stderr, "--- end log ---")
}

func (r startReport) Pretty() api.Text {
	t := api.Text{}
	switch {
	case !r.Waited:
		t = t.Add(icons.Success).Space().
			Append("Started gavel pr UI", "text-green-600").
			Append(" (pid ").Append(r.PID).Append(")").NewLine()
	case r.Outcome == service.ReadinessReady:
		t = t.Add(icons.Success).Space().
			Append("Ready", "text-green-600").
			Append(" (pid ").Append(r.PID).
			Append(", ").Append(r.Duration.Round(time.Millisecond).String()).
			Append(")").NewLine()
	case r.Outcome == service.ReadinessCrashed:
		t = t.Add(icons.Error).Space().
			Append("Crashed during startup", "error").NewLine()
	case r.Outcome == service.ReadinessTimedOut:
		t = t.Add(icons.Warning).Space().
			Append("Timed out waiting for UI port after ", "warning").
			Append(r.Timeout.String()).NewLine()
	}
	t = t.Append(kv("pidfile", r.PidFile)).NewLine()
	t = t.Append(kv("logfile", r.LogFile)).NewLine()
	t = t.Append(kv("stop", "")).Append("gavel system stop", "font-bold").NewLine()
	return t
}
