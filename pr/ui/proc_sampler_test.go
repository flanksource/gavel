package ui

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/flanksource/gavel/procfile"
)

func sampleWith(status string) map[string]procStatus {
	return map[string]procStatus{
		"gavel": {HasProcfile: true, Running: true, Processes: []procfile.ProcState{
			{Name: "web", Status: status},
		}},
	}
}

// Every open dashboard tab holds a proc-status stream, and each tick used to run
// the full LoadProjects + projectStatus scan on its own — a measured ~1.7s of
// filesystem and supervisor work, multiplied by the number of tabs. Concurrent
// callers inside one TTL window must collapse onto a single scan.
func TestProcSamplerSharesOneScanAcrossConcurrentCallers(t *testing.T) {
	var scans int32
	sampler := &procSampler{ttl: time.Minute, sample: func() (map[string]procStatus, error) {
		atomic.AddInt32(&scans, 1)
		time.Sleep(20 * time.Millisecond) // a scan is slow; overlap the callers
		return sampleWith(procfile.StatusRunning), nil
	}}

	const callers = 8
	var wg sync.WaitGroup
	wg.Add(callers)
	for i := 0; i < callers; i++ {
		go func() {
			defer wg.Done()
			if _, err := sampler.get(); err != nil {
				t.Errorf("get: %v", err)
			}
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&scans); got != 1 {
		t.Errorf("want %d concurrent callers to share 1 scan, got %d scans", callers, got)
	}
}

// A cached value must not outlive its TTL, or a start/restart would never reach
// the dashboard.
func TestProcSamplerRecomputesAfterTTL(t *testing.T) {
	var scans int32
	sampler := &procSampler{ttl: 10 * time.Millisecond, sample: func() (map[string]procStatus, error) {
		atomic.AddInt32(&scans, 1)
		return sampleWith(procfile.StatusRunning), nil
	}}

	if _, err := sampler.get(); err != nil {
		t.Fatalf("first get: %v", err)
	}
	if _, err := sampler.get(); err != nil {
		t.Fatalf("cached get: %v", err)
	}
	if got := atomic.LoadInt32(&scans); got != 1 {
		t.Fatalf("want the second get served from cache, got %d scans", got)
	}

	time.Sleep(20 * time.Millisecond)
	if _, err := sampler.get(); err != nil {
		t.Fatalf("post-TTL get: %v", err)
	}
	if got := atomic.LoadInt32(&scans); got != 2 {
		t.Errorf("want a rescan once the TTL lapsed, got %d scans", got)
	}
}

// Callers must each get their own copy: handleProcStatusStream marshals what it
// receives, and a shared map handed to several streams would race.
func TestProcSamplerReturnsIndependentCopies(t *testing.T) {
	sampler := &procSampler{ttl: time.Minute, sample: func() (map[string]procStatus, error) {
		return sampleWith(procfile.StatusRunning), nil
	}}

	first, err := sampler.get()
	if err != nil {
		t.Fatalf("first get: %v", err)
	}
	first["gavel"] = procStatus{Error: "mutated by one caller"}

	second, err := sampler.get()
	if err != nil {
		t.Fatalf("second get: %v", err)
	}
	if second["gavel"].Error != "" {
		t.Errorf("one caller's mutation leaked into another's snapshot: %+v", second["gavel"])
	}
}

// A failing scan must surface, not serve a stale or empty map as if it were
// current state.
func TestProcSamplerPropagatesScanError(t *testing.T) {
	want := errors.New("load projects: permission denied")
	sampler := &procSampler{ttl: time.Minute, sample: func() (map[string]procStatus, error) {
		return nil, want
	}}

	if _, err := sampler.get(); !errors.Is(err, want) {
		t.Errorf("want the scan error surfaced, got %v", err)
	}
}
