package snapshots

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/gavel/blob"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/traces"
)

// RunningName is the in-flight pointer file written periodically while a
// `gavel test` run is active. It is removed on clean completion and left
// behind on interrupt / timeout / crash so external observers can see what
// was running when things went wrong.
const RunningName = "running.json"

// PerRunPrefix is the stem of the per-run timestamped snapshot. The full
// filename includes a UTC timestamp, e.g. run-2026-04-27T10-43-42Z.json.
const PerRunPrefix = traces.RunPrefix

// PerRunTimestampLayout is filesystem-safe (no colons) but still
// chronologically sortable.
const PerRunTimestampLayout = traces.RunTimestampLayout

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

	mu         sync.Mutex
	pending    *testui.Snapshot
	dirty      bool
	err        error
	startOnce  sync.Once
	stopOnce   sync.Once
	finishOnce sync.Once

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
	if r == nil {
		return
	}
	r.startOnce.Do(func() { go r.loop() })
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
// Failed writes remain dirty for retry and are returned by Stop.
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
		r.mu.Lock()
		r.dirty = true
		if r.err == nil {
			r.err = err
		}
		r.mu.Unlock()
	}
}

// Stop ends the periodic writes. If success is true, running.json is removed
// (the run finished cleanly and the data lives in last.json + the per-run
// timestamped file). If success is false (interrupt, timeout, panic),
// running.json is left in place as a forensic record of what was in flight.
//
// Stop is idempotent and safe to call from defer.
func (r *Recorder) Stop(success bool) error {
	if r == nil {
		return nil
	}
	r.Start()
	r.stopOnce.Do(func() { close(r.stop) })
	<-r.doneCh
	r.finishOnce.Do(func() {
		r.flush()
		r.mu.Lock()
		defer r.mu.Unlock()
		if success && r.err == nil {
			path := filepath.Join(r.workDir, Dir, RunningName)
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				r.err = fmt.Errorf("remove %s: %w", path, err)
			}
		}
	})
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.err
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

// SavePerRun writes a per-run timestamped snapshot, run-<ts>.json, to workDir's snapshot store
// alongside the sha-keyed snapshot from Save(). The timestamp uses the run's start time (UTC) so
// reruns triggered in the same git state still land under distinct keys. It returns where the
// snapshot was written (see Location).
//
// label distinguishes runs that start within the same second — several fixture
// verification steps in one document, for instance — and is appended to the
// stem as run-<ts>-<label>.json. Pass "" for a whole-command run. ListRuns
// reads the start time from the snapshot metadata, so a labelled stem sorts
// identically to a bare one.
//
// Unlike Save(), this does not write or update last.json / <branch>.json
// pointers — those are owned by Save() and reflect the latest sha-keyed
// snapshot, not the per-run history.
func SavePerRun(workDir string, snap *testui.Snapshot, started time.Time, label string) (string, error) {
	if snap == nil {
		return "", errors.New("snapshots.SavePerRun: snapshot is nil")
	}
	if workDir == "" {
		return "", errors.New("snapshots.SavePerRun: workDir is required")
	}
	if started.IsZero() {
		started = time.Now().UTC()
	}
	store, err := OpenStore(workDir)
	if err != nil {
		return "", err
	}
	stem := traces.RunID(started)
	if label = slugifyLabel(label); label != "" {
		stem += "-" + label
	}
	key, err := uniqueRunKey(store, stem)
	if err != nil {
		return "", err
	}
	if err := putJSON(store, key, snap); err != nil {
		return "", err
	}
	return Location(workDir, key), nil
}

// uniqueRunKey returns <stem>.json, suffixing -2, -3, … when that key is taken. The per-run
// history is append-only: a second run in the same second must not silently overwrite the first.
func uniqueRunKey(store blob.Store, stem string) (string, error) {
	for i := 1; i < 1000; i++ {
		name := stem
		if i > 1 {
			name = fmt.Sprintf("%s-%d", stem, i)
		}
		key := name + ".json"
		exists, err := store.Exists(context.Background(), key)
		if err != nil {
			return "", err
		}
		if !exists {
			return key, nil
		}
	}
	return "", fmt.Errorf("snapshots.SavePerRun: no free name for %s", stem)
}

// slugifyLabel reduces a label to lowercase alphanumerics and dashes so it is
// safe in a filename and still readable in a directory listing.
func slugifyLabel(label string) string {
	var b strings.Builder
	lastDash := true // suppresses a leading dash
	for _, r := range strings.ToLower(strings.TrimSpace(label)) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastDash = false
		case !lastDash:
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
