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
// worker), or reaches a terminal state. readyTimeout bounds a process that never
// leaves "starting".
const (
	procReadyPoll    = 300 * time.Millisecond
	procReadyTimeout = 30 * time.Second
	procPortGrace    = 4 * time.Second
)

// renderProcReadiness renders one live clicky task per tracked process, polling
// each to readiness and updating its label as it moves starting → running →
// listening on :PORT (or crashed/exited). names selects the subset to track;
// empty tracks every process in report. It blocks until every tracked process
// settles so the caller can re-fetch a final, accurate status afterwards.
func renderProcReadiness(workDir, pf, groupName string, report *procfile.StatusReport, names []string) {
	tracked := trackedProcNames(report.Processes, names)
	if len(tracked) == 0 {
		return
	}
	group := clicky.StartGroup[procfile.ProcState](groupName)
	for _, name := range tracked {
		name := name
		group.Add(name, func(ctx commonsCtx.Context, t *task.Task) (procfile.ProcState, error) {
			return awaitProcReady(ctx, t, workDir, pf, name)
		})
	}
	group.WaitFor()
}

// awaitProcReady polls a single process's status until it settles, keeping the
// task label in sync. A crash fails the task; a clean early exit warns; a
// running process (with or without a port) succeeds.
func awaitProcReady(ctx commonsCtx.Context, t *task.Task, workDir, pf, name string) (procfile.ProcState, error) {
	deadline := time.Now().Add(procReadyTimeout)
	var runningSince time.Time
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

		switch ps.Status {
		case procfile.StatusCrashed:
			return ps, fmt.Errorf("%s crashed (exit %s)", name, exitCodeStr(ps.ExitCode))
		case procfile.StatusExited:
			t.Warning()
			return ps, nil
		case procfile.StatusRunning:
			if len(ps.Ports) > 0 {
				return ps, nil
			}
			if runningSince.IsZero() {
				runningSince = time.Now()
			} else if time.Since(runningSince) >= procPortGrace {
				return ps, nil
			}
		default:
			runningSince = time.Time{}
		}

		if time.Now().After(deadline) {
			t.Warning()
			return ps, nil
		}
		select {
		case <-ctx.Done():
			return ps, ctx.Err()
		case <-time.After(procReadyPoll):
		}
	}
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
// names when non-empty (preserving its order), otherwise every process in procs.
func trackedProcNames(procs []procfile.ProcState, names []string) []string {
	if len(names) > 0 {
		return names
	}
	out := make([]string, 0, len(procs))
	for _, p := range procs {
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
