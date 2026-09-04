package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/task"
	commonsCtx "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/procfile"
)

// Cadence for the start/restart readiness view. A process is "ready" once it is
// Running with a detected port, or has run portGrace without binding one (a
// worker), or reaches a terminal state. The stage timeouts bound a process that
// never leaves a transitional stage.
const (
	procReadyPoll = 300 * time.Millisecond
	// procCompileTimeout budgets the "compiling" stage, where a cold `go run`
	// build or a bundler's first pass legitimately runs for many minutes.
	procCompileTimeout = 900 * time.Second
	// procReadyTimeout budgets every other stage.
	procReadyTimeout = 180 * time.Second
	procPortGrace    = 4 * time.Second
)

// stageTimeout is how long a process may sit in one stage before its start is
// declared failed. The clock is per stage, so time spent compiling never eats
// into the budget of the stage that follows it.
func stageTimeout(status string) time.Duration {
	if status == procfile.StatusCompiling {
		return procCompileTimeout
	}
	return procReadyTimeout
}

// renderProcReadiness renders one live clicky task per tracked process, polling
// each to readiness and updating its label as it moves starting → running →
// listening on :PORT (or crashed/exited). names selects the subset to track;
// empty tracks every process in report. It blocks until every tracked process
// settles, returning a non-nil error when any failed to start (a crash or a
// process that never reached running) so the caller can exit non-zero.
func renderProcReadiness(workDir, pf, groupName string, report *procfile.StatusReport, names []string) error {
	tracked := trackedProcNames(report.Processes, names)
	if len(tracked) == 0 {
		return nil
	}
	group := clicky.StartGroup[procfile.ProcState](groupName)
	for _, name := range tracked {
		name := name
		group.Add(name, func(ctx commonsCtx.Context, t *task.Task) (procfile.ProcState, error) {
			return awaitProcReady(ctx, t, workDir, pf, name)
		})
	}
	group.WaitFor()
	_, err := group.GetResults()
	return err
}

// awaitProcReady polls a single process's status until it settles, keeping the
// task label in sync. A crash or a never-running process fails the task; a clean
// early exit warns; a running process (with or without a port) succeeds.
func awaitProcReady(ctx commonsCtx.Context, t *task.Task, workDir, pf, name string) (procfile.ProcState, error) {
	tr := &procTracker{}
	var last procfile.ProcState
	for {
		rep, err := procfile.Status(workDir, pf)
		if err != nil {
			return last, err
		}
		ps, ok := findProcState(rep.Processes, name)
		if !ok {
			return last, fmt.Errorf("process %q disappeared from status", name)
		}
		last = ps
		t.SetName(procReadyLabel(ps))

		switch outcome, ferr := tr.observe(ps, time.Now()); outcome {
		case outcomeReady:
			return ps, nil
		case outcomeWarn:
			t.Warning()
			return ps, nil
		case outcomeFailed:
			return ps, ferr
		}
		select {
		case <-ctx.Done():
			return ps, ctx.Err()
		case <-time.After(procReadyPoll):
		}
	}
}

// readyOutcome is the verdict of folding one status sample into a procTracker.
type readyOutcome int

const (
	outcomePending readyOutcome = iota // not settled yet
	outcomeReady                       // settled: running (ready) — a successful start
	outcomeWarn                        // settled: a clean early exit — soft, not a failure
	outcomeFailed                      // settled: crashed or never reached running
)

// procTracker classifies one process's progress toward "ready" across repeated
// status samples, owning the per-process timing (stage clock, port grace) so the
// caller only feeds it fresh ProcStates. A process is ready once it is Running
// with a detected port, or has run procPortGrace without binding one (a worker).
// A crash, or outlasting its stage's timeout without leaving it, is a start
// failure.
type procTracker struct {
	stage        string
	stageSince   time.Time
	runningSince time.Time
}

// observe folds one status sample into the tracker. now is passed in so tests
// can drive the timing deterministically. The returned error is non-nil only for
// outcomeFailed and describes why the start failed.
func (pt *procTracker) observe(ps procfile.ProcState, now time.Time) (readyOutcome, error) {
	if pt.stageSince.IsZero() || ps.Status != pt.stage {
		pt.stage = ps.Status
		pt.stageSince = now
	}

	switch ps.Status {
	case procfile.StatusCrashed:
		return outcomeFailed, fmt.Errorf("%s crashed (exit %s)", ps.Name, exitCodeStr(ps.ExitCode))
	case procfile.StatusExited, procfile.StatusStopped:
		return outcomeWarn, nil
	case procfile.StatusRunning:
		if len(ps.Ports) > 0 {
			return outcomeReady, nil
		}
		if pt.runningSince.IsZero() {
			pt.runningSince = now
		} else if now.Sub(pt.runningSince) >= procPortGrace {
			return outcomeReady, nil
		}
	default:
		pt.runningSince = time.Time{}
	}

	// A Running process is never failed here: it settles via the port grace
	// above, well inside its stage timeout. Only one still stuck in a
	// transitional stage runs out of budget.
	if timeout := stageTimeout(ps.Status); now.Sub(pt.stageSince) > timeout {
		return outcomeFailed, fmt.Errorf("%s did not become ready within %s (status %q)", ps.Name, timeout, ps.Status)
	}
	return outcomePending, nil
}

// procReadyLabel is the live task label for a process in a given state.
func procReadyLabel(ps procfile.ProcState) string {
	switch ps.Status {
	case procfile.StatusStarting:
		return ps.Name + ": starting"
	case procfile.StatusRestarting:
		return ps.Name + ": restarting"
	case procfile.StatusRunning:
		if len(ps.Ports) > 0 {
			return fmt.Sprintf("%s: listening on %s", ps.Name, procfile.PortsLabel(ps.Ports))
		}
		return fmt.Sprintf("%s: running (pid %d)", ps.Name, ps.PID)
	case procfile.StatusCrashed:
		return fmt.Sprintf("%s: crashed (exit %s)", ps.Name, exitCodeStr(ps.ExitCode))
	case procfile.StatusExited:
		return fmt.Sprintf("%s: exited (%s)", ps.Name, exitCodeStr(ps.ExitCode))
	default:
		return ps.Name + ": " + ps.Status
	}
}

// trackedProcNames returns the process names to follow during a start/restart:
// names when non-empty (preserving its order), otherwise every process that is
// actually starting. Registered-but-stopped entries (default:false or an
// inactive profile) are skipped so the readiness view doesn't wait on processes
// that were never meant to start.
func trackedProcNames(procs []procfile.ProcState, names []string) []string {
	if len(names) > 0 {
		return names
	}
	out := make([]string, 0, len(procs))
	for _, p := range procs {
		if p.Status == procfile.StatusStopped {
			continue
		}
		out = append(out, p.Name)
	}
	return out
}

func findProcState(procs []procfile.ProcState, name string) (procfile.ProcState, bool) {
	for _, p := range procs {
		if p.Name == name {
			return p, true
		}
	}
	return procfile.ProcState{}, false
}

func exitCodeStr(code *int) string {
	if code == nil {
		return "?"
	}
	return strconv.Itoa(*code)
}
