package snapshots

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	testui "github.com/flanksource/gavel/testrunner/ui"
)

// RunningName is the in-flight pointer file written periodically while a
// `gavel test` run is active. It is removed on clean completion and left
// behind on interrupt / timeout / crash so external observers can see what
// was running when things went wrong.
const RunningName = "running.json"

// PerRunPrefix is the stem of the per-run timestamped snapshot. The full
// filename includes a UTC timestamp, e.g. run-2026-04-27T10-43-42Z.json.
const PerRunPrefix = "run-"

// PerRunTimestampLayout is filesystem-safe (no colons) but still
// chronologically sortable.
const PerRunTimestampLayout = "2006-01-02T15-04-05Z"

// Recorder writes periodic snapshots of an in-flight `gavel test` run to
// .gavel/running.json. A 3-second ticker drives writes; ticks where nothing
// has changed since the last write are no-ops.
//
// Lifecycle:
//
//	r := snapshots.NewRecorder(workDir)
//	r.Start()
//	r.Update(snap)   // call whenever the live snapshot changes
//	...
//	r.Stop(success)  // success=true removes running.json; false keeps it
type Recorder struct {
	workDir  string
	interval time.Duration

	mu      sync.Mutex
	pending *testui.Snapshot
	dirty   bool

	stop   chan struct{}
	doneCh chan struct{}
}

// NewRecorder creates a Recorder that will write to workDir/.gavel/running.json
// every 3 seconds. The Recorder is inert until Start() is called.
func NewRecorder(workDir string) *Recorder {
	return &Recorder{
		workDir:  workDir,
		interval: 3 * time.Second,
		stop:     make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Update records the latest snapshot to be written on the next tick. Cheap
// (single mutex lock + pointer assignment); safe to call from any goroutine.
// A nil snapshot is ignored.
func (r *Recorder) Update(snap *testui.Snapshot) {
	if r == nil || snap == nil {
		return
	}
	r.mu.Lock()
	r.pending = snap
	r.dirty = true
	r.mu.Unlock()
}

// Start launches the background ticker. Calling Start more than once is a
// no-op after the first call.
func (r *Recorder) Start() {
	if r == nil || r.workDir == "" {
		return
	}
	go r.loop()
}

func (r *Recorder) loop() {
	defer close(r.doneCh)
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-r.stop:
			return
		case <-ticker.C:
			r.flush()
		}
	}
}

// flush writes running.json if Update has been called since the last flush.
// Errors are non-fatal: the in-flight run continues even when the disk is
// full or read-only.
func (r *Recorder) flush() {
	r.mu.Lock()
	if !r.dirty || r.pending == nil {
		r.mu.Unlock()
		return
	}
	snap := r.pending
	r.dirty = false
	r.mu.Unlock()

	if err := writeRunningJSON(r.workDir, snap); err != nil {
		// Logged at the call site only on the first failure; flushes don't
		// have a logger handy and shouldn't spam every 3s.
		_ = err
	}
}

// Stop ends the periodic writes. If success is true, running.json is removed
// (the run finished cleanly and the data lives in last.json + the per-run
// timestamped file). If success is false (interrupt, timeout, panic),
// running.json is left in place as a forensic record of what was in flight.
//
// Stop is idempotent and safe to call from defer.
func (r *Recorder) Stop(success bool) {
	if r == nil {
		return
	}
	select {
	case <-r.stop:
		return
	default:
		close(r.stop)
	}
	<-r.doneCh

	if success {
		path := filepath.Join(r.workDir, Dir, RunningName)
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			_ = err
		}
		return
	}
	// Best-effort final flush so the file on disk reflects the latest state
	// at the moment of failure.
	r.mu.Lock()
	snap := r.pending
	r.dirty = false
	r.mu.Unlock()
	if snap != nil {
		_ = writeRunningJSON(r.workDir, snap)
	}
}

func writeRunningJSON(workDir string, snap *testui.Snapshot) error {
	if workDir == "" {
		return errors.New("snapshots: workDir is required")
	}
	dir := filepath.Join(workDir, Dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}
	return writeJSON(filepath.Join(dir, RunningName), snap)
}

// SavePerRun writes a per-run timestamped snapshot to .gavel/run-<ts>.json
// alongside the existing sha-keyed snapshot from Save(). The timestamp uses
// the run's start time (UTC) so reruns triggered in the same git state still
// land in distinct files. Returns the absolute path written.
//
// Unlike Save(), this does not write or update last.json / <branch>.json
// pointers — those are owned by Save() and reflect the latest sha-keyed
// snapshot, not the per-run history.
func SavePerRun(workDir string, snap *testui.Snapshot, started time.Time) (string, error) {
	if snap == nil {
		return "", errors.New("snapshots.SavePerRun: snapshot is nil")
	}
	if workDir == "" {
		return "", errors.New("snapshots.SavePerRun: workDir is required")
	}
	if started.IsZero() {
		started = time.Now().UTC()
	}
	dir := filepath.Join(workDir, Dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}
	name := PerRunPrefix + started.UTC().Format(PerRunTimestampLayout) + ".json"
	path := filepath.Join(dir, name)
	if err := writeJSON(path, snap); err != nil {
		return "", err
	}
	return path, nil
}
