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

// procSample is one supervisor resource measurement. Every field here moves on
// each sampling tick, so none of them may reach the stream — modelling them
// together keeps the fixture honest about what a real sample carries.
type procSample struct {
	CPU       float64
	RSS       uint64
	VMS       uint64
	SampledAt time.Time
}

func runningProc(name string, sample procSample) procfile.ProcState {
	started := time.Date(2026, 6, 23, 10, 0, 0, 0, time.UTC)
	sampledAt := sample.SampledAt
	return procfile.ProcState{
		Name:       name,
		Command:    "echo hi",
		PID:        1234,
		Status:     procfile.StatusRunning,
		Started:    &started,
		Ports:      []int{3000},
		CPUPercent: sample.CPU,
		MemoryRSS:  sample.RSS,
		MemoryVMS:  sample.VMS,
		OpenFiles:  12,
		SampledAt:  &sampledAt,
		PeakCPU:    sample.CPU,
		PeakRSS:    sample.RSS,
		PeakVMS:    sample.VMS,
		Tree: []procfile.ProcNode{
			{PID: 1234, PPID: 1, Command: "echo hi", CPUPercent: sample.CPU, MemoryRSS: sample.RSS},
		},
	}
}

// snapshots differing only in the live resource sample (cpu/mem, the sample
// timestamp, the running peaks and the per-node tree values) must be identical
// to the stream so it sends a ping rather than a re-render-firing data frame.
//
// The sample timestamp is the field that matters most in practice: the
// supervisor stamps a new SampledAt on every tick even when nothing else moved,
// so leaking it makes the change-detection in handleProcStatusStream fire on
// every cadence forever.
func TestLeanProcStatusIgnoresResourceChurn(t *testing.T) {
	base := time.Date(2026, 6, 23, 10, 0, 0, 0, time.UTC)
	low := map[string]procStatus{
		"gavel": {HasProcfile: true, Running: true, Processes: []procfile.ProcState{
			runningProc("web", procSample{CPU: 4.2, RSS: 100 << 20, VMS: 400 << 20, SampledAt: base}),
		}},
	}
	high := map[string]procStatus{
		"gavel": {HasProcfile: true, Running: true, Processes: []procfile.ProcState{
			runningProc("web", procSample{CPU: 350.0, RSS: 4 << 30, VMS: 8 << 30, SampledAt: base.Add(3 * time.Second)}),
		}},
	}

	if a, b := marshalLean(t, low), marshalLean(t, high); a != b {
		t.Errorf("lean status differs on resource churn alone:\n low=%s\nhigh=%s", a, b)
	}
}

// a real supervision transition (status, ports, restarts, …) must still produce
// a different frame so the dashboard updates.
func TestLeanProcStatusKeepsStableFields(t *testing.T) {
	base := runningProc("web", procSample{
		CPU:       4.2,
		RSS:       100 << 20,
		VMS:       400 << 20,
		SampledAt: time.Date(2026, 6, 23, 10, 0, 0, 0, time.UTC),
	})

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
