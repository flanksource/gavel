package ui

import (
	"context"
	"encoding/json"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/snapshots"
)

// TestRunSyncer scans each registered workspace's .gavel/run-*.json files on an
// interval (and on demand via Notify) and upserts a per-run summary + the file
// path into the DB cache. The run-*.json files are the source of truth; this is
// a historical read cache. Modeled on DetailSyncer: a buffered
// notify channel coalesces bursts, and a ticker keeps the cache warm without a
// request. The source is local files (cheap, no rate limit) so the loop just
// rescans every workspace each tick — ListRuns' `since` watermark keeps each
// rescan to only the runs added since the last sync.
type TestRunSyncer struct {
	srv      *Server
	interval time.Duration
	notify   chan struct{}
}

func NewTestRunSyncer(srv *Server) *TestRunSyncer {
	return &TestRunSyncer{
		srv:      srv,
		interval: 30 * time.Second,
		notify:   make(chan struct{}, 1),
	}
}

// Notify requests an out-of-cycle scan. Non-blocking: a buffered channel of 1
// coalesces rapid calls into a single pending sync.
func (trs *TestRunSyncer) Notify() {
	select {
	case trs.notify <- struct{}{}:
	default:
	}
}

func (trs *TestRunSyncer) Start(ctx context.Context) {
	go trs.loop(ctx)
}

func (trs *TestRunSyncer) loop(ctx context.Context) {
	trs.syncAll(ctx) // populate before the first tick so the tab isn't empty
	ticker := time.NewTicker(trs.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-trs.notify:
		case <-ticker.C:
		}
		trs.syncAll(ctx)
	}
}

func (trs *TestRunSyncer) syncAll(ctx context.Context) {
	store := cache.Shared()
	if store.Disabled() {
		// No DB configured — the read path scans .gavel directly instead.
		return
	}
	changed := false
	projects, err := LoadProjects()
	if err != nil {
		logger.Errorf("sync test runs: %v", err)
		return
	}
	for _, p := range projects {
		dir := p.ResolvedDir()
		if dir == "" {
			continue
		}
		if trs.syncWorkspace(ctx, store, dir) {
			changed = true
		}
	}
	if changed {
		trs.srv.notify()
	}
}

// syncWorkspace scans one workspace and upserts any runs newer than its cursor.
// Returns true when new runs were written.
func (trs *TestRunSyncer) syncWorkspace(ctx context.Context, store *cache.Store, dir string) bool {
	sinceNanos, _, err := store.LoadTestRunCursor(ctx, dir)
	if err != nil {
		logger.Warnf("test run cursor %s: %v", dir, err)
		return false
	}
	since := time.Time{}
	if sinceNanos > 0 {
		since = time.Unix(0, sinceNanos).UTC()
	}

	runs, err := snapshots.ListRuns(dir, since)
	if err != nil {
		logger.Warnf("scan test runs %s: %v", dir, err)
		return false
	}

	rows := make([]cache.TestRunCache, 0, len(runs)+1)
	highWater := sinceNanos
	for _, r := range runs {
		rows = append(rows, testRunRow(dir, r))
		if ts := nanos(r.Started); ts > highWater {
			highWater = ts
		}
	}
	// The `last` pointer is a different snapshot family from the run-*.json
	// sweep above and is what the Tests tab renders for every project on every
	// request, so it is summarized here too. Without it that view decodes the
	// full snapshot per project per request — 61% of this server's CPU with the
	// tab open.
	if row, ok := lastPointerRow(ctx, store, dir); ok {
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		// Nothing new, but stamp the cursor so SyncedAt still advances.
		if err := store.UpsertTestRuns(ctx, dir, nil, highWater); err != nil {
			logger.Warnf("refresh test run cursor %s: %v", dir, err)
		}
		return false
	}

	if err := store.UpsertTestRuns(ctx, dir, rows, highWater); err != nil {
		logger.Warnf("upsert test runs %s: %v", dir, err)
		return false
	}
	return true
}

// lastPointerRow summarizes the snapshot .gavel/last.json names, unless the
// cache already holds it. Run snapshots are immutable, so an existing row is
// never stale and re-decoding would defeat the point.
func lastPointerRow(ctx context.Context, store *cache.Store, dir string) (cache.TestRunCache, bool) {
	pointer, err := snapshots.LoadPointer(dir, snapshots.PointerLast)
	if err != nil || pointer == nil {
		return cache.TestRunCache{}, false
	}
	runID := pointerRunID(pointer)
	if runID == "" {
		return cache.TestRunCache{}, false
	}
	if existing, err := store.GetTestRun(ctx, dir, runID); err == nil && existing != nil {
		return cache.TestRunCache{}, false
	}
	run, err := snapshots.LastRun(dir)
	if err != nil || run == nil {
		return cache.TestRunCache{}, false
	}
	row := testRunRow(dir, *run)
	// LastRun reports the pointer's own name as the run id; key the row by the
	// file it resolved to so the reader finds it from the pointer alone.
	row.RunID = runID
	return row, true
}

func testRunRow(dir string, r snapshots.RunInfo) cache.TestRunCache {
	frameworks, _ := json.Marshal(r.Frameworks)
	return cache.TestRunCache{
		WorkspaceDir:   dir,
		RunID:          r.RunID,
		Path:           r.Path,
		Kind:           r.Kind,
		Repo:           r.Repo,
		SHA:            r.SHA,
		StartedTS:      nanos(r.Started),
		EndedTS:        nanos(r.Ended),
		Passed:         r.Passed,
		Failed:         r.Failed,
		Skipped:        r.Skipped,
		Warned:         r.Warned,
		Total:          r.Total,
		LintViolations: r.LintViolations,
		LintLinters:    r.LintLinters,
		Frameworks:     frameworks,
	}
}

// nanos returns the unix-nanosecond timestamp, or 0 for the zero time (so an
// unset Ended doesn't serialize as a huge negative watermark).
func nanos(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixNano()
}
