package main

import (
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/procfile"
)

func procState(name, status string, ports []int, exit *int) procfile.ProcState {
	return procfile.ProcState{Name: name, Status: status, Ports: ports, ExitCode: exit}
}

// TestProcTrackerObserve drives the readiness classifier through each terminal
// verdict with deterministic timestamps (no real processes).
func TestProcTrackerObserve(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	deadline := base.Add(procReadyTimeout)
	exit1 := 1

	t.Run("a crash is a start failure", func(t *testing.T) {
		tr := &procTracker{deadline: deadline}
		outcome, err := tr.observe(procState("web", procfile.StatusCrashed, nil, &exit1), base)
		if outcome != outcomeFailed {
			t.Fatalf("outcome = %d, want outcomeFailed", outcome)
		}
		if err == nil || !strings.Contains(err.Error(), "crashed") {
			t.Fatalf("err = %v, want one mentioning crashed", err)
		}
	})

	t.Run("a clean early exit warns", func(t *testing.T) {
		for _, status := range []string{procfile.StatusExited, procfile.StatusStopped} {
			tr := &procTracker{deadline: deadline}
			if outcome, _ := tr.observe(procState("job", status, nil, nil), base); outcome != outcomeWarn {
				t.Fatalf("%s outcome = %d, want outcomeWarn", status, outcome)
			}
		}
	})

	t.Run("running with a detected port is ready", func(t *testing.T) {
		tr := &procTracker{deadline: deadline}
		if outcome, _ := tr.observe(procState("web", procfile.StatusRunning, []int{3000}, nil), base); outcome != outcomeReady {
			t.Fatalf("outcome = %d, want outcomeReady", outcome)
		}
	})

	t.Run("a portless worker is ready only after the grace window", func(t *testing.T) {
		tr := &procTracker{deadline: deadline}
		if outcome, _ := tr.observe(procState("worker", procfile.StatusRunning, nil, nil), base); outcome != outcomePending {
			t.Fatalf("first observe = %d, want outcomePending", outcome)
		}
		if outcome, _ := tr.observe(procState("worker", procfile.StatusRunning, nil, nil), base.Add(procPortGrace-time.Millisecond)); outcome != outcomePending {
			t.Fatalf("within grace = %d, want outcomePending", outcome)
		}
		if outcome, _ := tr.observe(procState("worker", procfile.StatusRunning, nil, nil), base.Add(procPortGrace)); outcome != outcomeReady {
			t.Fatalf("after grace = %d, want outcomeReady", outcome)
		}
	})

	t.Run("still starting at the deadline is a failure", func(t *testing.T) {
		tr := &procTracker{deadline: deadline}
		if outcome, _ := tr.observe(procState("web", procfile.StatusStarting, nil, nil), base); outcome != outcomePending {
			t.Fatalf("before deadline = %d, want outcomePending", outcome)
		}
		outcome, err := tr.observe(procState("web", procfile.StatusStarting, nil, nil), deadline.Add(time.Second))
		if outcome != outcomeFailed {
			t.Fatalf("after deadline = %d, want outcomeFailed", outcome)
		}
		if err == nil || !strings.Contains(err.Error(), "did not become ready") {
			t.Fatalf("err = %v, want one mentioning did not become ready", err)
		}
	})

	t.Run("running but slow to bind at the deadline is still ready", func(t *testing.T) {
		tr := &procTracker{deadline: deadline}
		if outcome, _ := tr.observe(procState("web", procfile.StatusRunning, nil, nil), deadline.Add(time.Second)); outcome != outcomeReady {
			t.Fatalf("outcome = %d, want outcomeReady", outcome)
		}
	})
}
