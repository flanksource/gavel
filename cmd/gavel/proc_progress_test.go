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
		tr := &procTracker{}
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
			tr := &procTracker{}
			if outcome, _ := tr.observe(procState("job", status, nil, nil), base); outcome != outcomeWarn {
				t.Fatalf("%s outcome = %d, want outcomeWarn", status, outcome)
			}
		}
	})

	t.Run("running with a detected port is ready", func(t *testing.T) {
		tr := &procTracker{}
		if outcome, _ := tr.observe(procState("web", procfile.StatusRunning, []int{3000}, nil), base); outcome != outcomeReady {
			t.Fatalf("outcome = %d, want outcomeReady", outcome)
		}
	})

	t.Run("a portless worker is ready only after the grace window", func(t *testing.T) {
		tr := &procTracker{}
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
		tr := &procTracker{}
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

	t.Run("running but slow to bind past the deadline is still ready", func(t *testing.T) {
		tr := &procTracker{}
		if outcome, _ := tr.observe(procState("web", procfile.StatusRunning, nil, nil), base); outcome != outcomePending {
			t.Fatalf("first observe = %d, want outcomePending", outcome)
		}
		if outcome, _ := tr.observe(procState("web", procfile.StatusRunning, nil, nil), deadline.Add(time.Second)); outcome != outcomeReady {
			t.Fatalf("past deadline = %d, want outcomeReady", outcome)
		}
	})

	t.Run("compiling gets the longer budget", func(t *testing.T) {
		tr := &procTracker{}
		compiling := procState("web", procfile.StatusCompiling, nil, nil)
		if outcome, _ := tr.observe(compiling, base); outcome != outcomePending {
			t.Fatalf("first observe = %d, want outcomePending", outcome)
		}
		if outcome, _ := tr.observe(compiling, base.Add(procCompileTimeout)); outcome != outcomePending {
			t.Fatalf("at %s = %d, want outcomePending", procCompileTimeout, outcome)
		}
		outcome, err := tr.observe(compiling, base.Add(procCompileTimeout+time.Second))
		if outcome != outcomeFailed {
			t.Fatalf("past %s = %d, want outcomeFailed", procCompileTimeout, outcome)
		}
		if err == nil || !strings.Contains(err.Error(), procCompileTimeout.String()) {
			t.Fatalf("err = %v, want one naming the %s compile budget", err, procCompileTimeout)
		}
	})

	t.Run("leaving compiling restarts the stage clock", func(t *testing.T) {
		tr := &procTracker{}
		compiledFor := procCompileTimeout - time.Minute
		if outcome, _ := tr.observe(procState("web", procfile.StatusCompiling, nil, nil), base); outcome != outcomePending {
			t.Fatalf("compiling = %d, want outcomePending", outcome)
		}
		// The long compile must not eat into the following stage's budget.
		if outcome, _ := tr.observe(procState("web", procfile.StatusStarting, nil, nil), base.Add(compiledFor)); outcome != outcomePending {
			t.Fatalf("starting after a %s compile = %d, want outcomePending", compiledFor, outcome)
		}
		if outcome, _ := tr.observe(procState("web", procfile.StatusStarting, nil, nil), base.Add(compiledFor+procReadyTimeout+time.Second)); outcome != outcomeFailed {
			t.Fatalf("starting past its own budget = %d, want outcomeFailed", outcome)
		}
	})
}
