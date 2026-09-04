package ui

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/flanksource/gavel/procfile"
)

// marshalLean is the exact change-detection input handleProcStatusStream
// compares (json.Marshal of the lean projection). Two snapshots are
// indistinguishable to the stream when their marshalled lean forms are equal.
func marshalLean(t *testing.T, byKey map[string]procStatus) string {
	t.Helper()
	b, err := json.Marshal(leanProcStatus(byKey))
	if err != nil {
		t.Fatalf("marshal lean status: %v", err)
	}
	return string(b)
}

func runningProc(name string, cpu float64, rss uint64) procfile.ProcState {
	started := time.Date(2026, 6, 23, 10, 0, 0, 0, time.UTC)
	return procfile.ProcState{
		Name:       name,
		Command:    "echo hi",
		PID:        1234,
		Status:     procfile.StatusRunning,
		Started:    &started,
		Ports:      []int{3000},
		CPUPercent: cpu,
		MemoryRSS:  rss,
		OpenFiles:  12,
		Tree: []procfile.ProcNode{
			{PID: 1234, PPID: 1, Command: "echo hi", CPUPercent: cpu, MemoryRSS: rss},
		},
	}
}

// snapshots differing only in the live resource sample (cpu/mem and the
// per-node tree values) must be identical to the stream so it sends a ping
// rather than a re-render-firing data frame.
func TestLeanProcStatusIgnoresResourceChurn(t *testing.T) {
	low := map[string]procStatus{
		"gavel": {HasProcfile: true, Running: true, Processes: []procfile.ProcState{runningProc("web", 4.2, 100<<20)}},
	}
	high := map[string]procStatus{
		"gavel": {HasProcfile: true, Running: true, Processes: []procfile.ProcState{runningProc("web", 350.0, 4<<30)}},
	}

	if a, b := marshalLean(t, low), marshalLean(t, high); a != b {
		t.Errorf("lean status differs on cpu/mem churn alone:\n low=%s\nhigh=%s", a, b)
	}
}

// a real supervision transition (status, ports, restarts, …) must still produce
// a different frame so the dashboard updates.
func TestLeanProcStatusKeepsStableFields(t *testing.T) {
	base := runningProc("web", 4.2, 100<<20)

	stopped := base
	stopped.Status = procfile.StatusStopped

	reported := base
	reported.OpenFiles = 99

	cases := []struct {
		name string
		proc procfile.ProcState
	}{
		{"status change", stopped},
		{"open-files change", reported},
	}
	want := marshalLean(t, map[string]procStatus{"gavel": {HasProcfile: true, Running: true, Processes: []procfile.ProcState{base}}})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := marshalLean(t, map[string]procStatus{"gavel": {HasProcfile: true, Running: true, Processes: []procfile.ProcState{tc.proc}}})
			if got == want {
				t.Errorf("%s did not change the lean frame; both = %s", tc.name, got)
			}
		})
	}
}
