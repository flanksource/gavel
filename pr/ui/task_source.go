package ui

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/flanksource/clicky/metrics"
	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/internal/taskhistory"
	"github.com/flanksource/gavel/procfile"
)

// supervisorTaskSource serves archived runs to clicky's task manager. The
// archive is only ever written by the spool sweep (importTaskHistory), so the
// run listing is cached until the sweep invalidates it rather than for a
// wall-clock TTL: a dashboard stream polling twice a second reads memory, and
// the database is asked once per sweep. Snapshots are read per run on demand;
// the listing never carries them because captured output dwarfs run metadata.
type supervisorTaskSource struct {
	runs     func(context.Context) ([]taskhistory.RunSummary, error)
	snapshot func(context.Context, string) ([]clickytask.TaskSnapshot, error)

	mu     sync.Mutex
	cached []taskhistory.RunSummary
	valid  bool

	// Which supervisors are live, memoized separately from the listing: it is
	// discovered by sweeping every configured project, not by the spool sweep,
	// so it has no invalidating writer to hang off. See supervisorRoots.
	rootsMu     sync.Mutex
	rootsCached []string
	rootsAt     time.Time
}

// supervisorRootsTTL bounds how stale the live-supervisor set may be. A start or
// stop through the dashboard invalidates it outright; the TTL only covers
// supervisors started elsewhere, e.g. `gavel proc start` in a terminal.
//
// It has to be a TTL and not just invalidation because clicky's runs stream
// polls every 500ms and its per-task stream every 200ms, per subscriber, and
// each poll asks this source for the live roots first.
const supervisorRootsTTL = 2 * time.Second

func newSupervisorTaskSource() *supervisorTaskSource {
	return &supervisorTaskSource{runs: loadArchivedRuns, snapshot: loadArchivedSnapshot}
}

// invalidate drops the cached listing so the next read reloads it.
func (s *supervisorTaskSource) invalidate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cached, s.valid = nil, false
}

// invalidateRoots drops the memoized live-supervisor set so the next read sees a
// supervisor this process just started or stopped.
func (s *supervisorTaskSource) invalidateRoots() {
	if s == nil {
		return
	}
	s.rootsMu.Lock()
	defer s.rootsMu.Unlock()
	s.rootsCached, s.rootsAt = nil, time.Time{}
}

// supervisorRoots memoizes runningSupervisorRoots for supervisorRootsTTL.
//
// Every RunSource method needs the live roots before it can do anything, and
// discovering them costs one procfile.Status per configured project — each of
// which re-reads the Procfile, the state file and, when no supervisor is
// running, deep-merges up to three .gavel.yaml files. Uncached at the stream
// cadence that sweep was ~27% of this process's CPU.
func (s *supervisorTaskSource) supervisorRoots() ([]string, error) {
	s.rootsMu.Lock()
	defer s.rootsMu.Unlock()
	if !s.rootsAt.IsZero() && time.Since(s.rootsAt) < supervisorRootsTTL {
		return s.rootsCached, nil
	}
	roots, err := runningSupervisorRoots()
	if err != nil {
		return nil, err
	}
	s.rootsCached, s.rootsAt = roots, time.Now()
	return roots, nil
}

// archivedRuns serves the listing from cache. taskhistory.RunSummary values
// are only ever replaced wholesale, so callers may read the slice.
func (s *supervisorTaskSource) archivedRuns(ctx context.Context) ([]taskhistory.RunSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.valid {
		runs, err := s.runs(ctx)
		if err != nil {
			return nil, err
		}
		s.cached, s.valid = runs, true
	}
	return s.cached, nil
}

func (s *supervisorTaskSource) Runs(ctx context.Context, filter clickytask.RunFilter) ([]clickytask.RunMeta, error) {
	var runs []clickytask.RunMeta
	roots, err := s.supervisorRoots()
	if err != nil {
		return nil, err
	}
	for _, root := range roots {
		remote, err := procfile.TaskRuns(root, filter)
		if err != nil {
			return nil, fmt.Errorf("list supervisor tasks for %s: %w", root, err)
		}
		runs = append(runs, remote...)
	}
	seen := make(map[string]struct{}, len(runs))
	for _, run := range runs {
		seen[run.ID] = struct{}{}
	}
	archived, err := s.archivedRuns(ctx)
	if err != nil {
		return nil, err
	}
	for _, record := range archived {
		if _, ok := seen[record.Run.ID]; ok || !filter.Matches(record.Run) {
			continue
		}
		runs = append(runs, record.Run)
	}
	sort.SliceStable(runs, func(i, j int) bool { return runs[i].StartedAt > runs[j].StartedAt })
	return runs, nil
}

func (s *supervisorTaskSource) Snapshot(ctx context.Context, id string) ([]clickytask.TaskSnapshot, error) {
	roots, err := s.supervisorRoots()
	if err != nil {
		return nil, err
	}
	for _, root := range roots {
		snapshots, err := procfile.TaskSnapshot(root, id)
		if err != nil {
			return nil, fmt.Errorf("load supervisor task %s from %s: %w", id, root, err)
		}
		if len(snapshots) > 0 {
			return snapshots, nil
		}
	}
	archived, err := s.archivedRuns(ctx)
	if err != nil {
		return nil, err
	}
	for _, record := range archived {
		if record.Run.ID == id {
			return s.snapshot(ctx, id)
		}
	}
	return nil, nil
}

// loadArchivedRuns and loadArchivedSnapshot are pure reads: folding the
// on-disk spool into the database is the sweep's job. Reading from here would
// make every dashboard poll rewrite the whole archive.
func loadArchivedRuns(ctx context.Context) ([]taskhistory.RunSummary, error) {
	now := time.Now().UTC()
	db, err := database.Shared(ctx)
	if err != nil {
		return nil, fmt.Errorf("open task history database: %w", err)
	}
	if db.Disabled() {
		records, err := loadSpool(now)
		if err != nil {
			return nil, err
		}
		runs := make([]taskhistory.RunSummary, 0, len(records))
		for _, record := range records {
			runs = append(runs, taskhistory.RunSummary{Run: record.Run, ArchivedAt: record.ArchivedAt})
		}
		return runs, nil
	}
	store, err := taskhistory.NewStore(db.Gorm())
	if err != nil {
		return nil, err
	}
	return store.ListRuns(ctx, now)
}

func loadArchivedSnapshot(ctx context.Context, id string) ([]clickytask.TaskSnapshot, error) {
	db, err := database.Shared(ctx)
	if err != nil {
		return nil, fmt.Errorf("open task history database: %w", err)
	}
	if db.Disabled() {
		records, err := loadSpool(time.Now().UTC())
		if err != nil {
			return nil, err
		}
		for _, record := range records {
			if record.Run.ID == id {
				return record.Snapshots, nil
			}
		}
		return nil, nil
	}
	store, err := taskhistory.NewStore(db.Gorm())
	if err != nil {
		return nil, err
	}
	return store.Snapshot(ctx, id)
}

func loadSpool(now time.Time) ([]taskhistory.Record, error) {
	roots, err := configuredProjectRoots()
	if err != nil {
		return nil, err
	}
	return taskhistory.LoadRoots(roots, taskhistory.LoadOptions{Now: now})
}

// taskHistorySweepInterval is how often spool files written by other processes
// (detached supervisors, CLI test runs) are folded into the database. Runs this
// server archives itself nudge the sweep, so the ticker only has to cover
// out-of-process writers.
const taskHistorySweepInterval = time.Minute

// nudgeTaskHistoryImport asks the sweep to run now because this process just
// archived a run. Never blocks: a full channel already means a sweep is queued.
func (s *Server) nudgeTaskHistoryImport() {
	select {
	case s.taskHistoryImport <- struct{}{}:
	default:
	}
}

func (s *Server) taskHistoryImportLoop() {
	sweep := &spoolSweep{imported: map[string]time.Time{}}
	t := time.NewTicker(taskHistorySweepInterval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
		case <-s.taskHistoryImport:
		}
		if err := sweep.run(context.Background()); err != nil {
			logger.Errorf("import task history: %v", err)
		}
		s.taskSource.invalidate()
	}
}

// spoolSweep folds retained spool records into the database mirror and drops
// expired rows. It is the archive's only write path: it runs when this process
// archives a run and on the periodic tick, never from a request handler.
//
// imported remembers every spool file the previous pass saw, keyed by path with
// the file's modification time, so a pass over an unchanged spool parses
// nothing and offers the database nothing. Files that disappear fall out of
// the map because each pass rebuilds it from what it was offered.
type spoolSweep struct {
	imported map[string]time.Time
}

func (s *spoolSweep) run(ctx context.Context) error {
	db, err := database.Shared(ctx)
	if err != nil {
		return fmt.Errorf("open task history database: %w", err)
	}
	if db.Disabled() {
		return nil
	}
	roots, err := configuredProjectRoots()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	offered := map[string]time.Time{}
	spooled, err := taskhistory.LoadRoots(roots, taskhistory.LoadOptions{
		Now: now,
		Skip: func(path string, modTime time.Time) bool {
			offered[path] = modTime
			return s.imported[path].Equal(modTime)
		},
	})
	if err != nil {
		return err
	}
	store, err := taskhistory.NewStore(db.Gorm())
	if err != nil {
		return err
	}
	if err := store.Import(ctx, spooled); err != nil {
		return err
	}
	s.imported = offered
	return store.Prune(ctx, now)
}

func (s *supervisorTaskSource) Control(_ context.Context, id string, action clickytask.ControlAction) error {
	roots, err := s.supervisorRoots()
	if err != nil {
		return err
	}
	for _, root := range roots {
		snapshots, err := procfile.TaskSnapshot(root, id)
		if err != nil {
			return fmt.Errorf("locate supervisor task %s in %s: %w", id, root, err)
		}
		if len(snapshots) == 0 {
			continue
		}
		return procfile.TaskControl(root, id, action)
	}
	return fmt.Errorf("run %q not found", id)
}

func (s *supervisorTaskSource) ControlTask(_ context.Context, runID, taskID string, action clickytask.ControlAction) error {
	roots, err := s.supervisorRoots()
	if err != nil {
		return err
	}
	for _, root := range roots {
		snapshots, err := procfile.TaskSnapshot(root, runID)
		if err != nil {
			return fmt.Errorf("locate supervisor task %s in %s: %w", runID, root, err)
		}
		if len(snapshots) == 0 {
			continue
		}
		return procfile.TaskControlTask(root, runID, taskID, action)
	}
	return fmt.Errorf("run %q not found", runID)
}

func (s *supervisorTaskSource) QueryMetric(_ context.Context, request metrics.QueryRequest) ([]metrics.Point, error) {
	roots, err := s.supervisorRoots()
	if err != nil {
		return nil, err
	}
	for _, root := range roots {
		points, err := procfile.TaskMetrics(root, request)
		if err != nil {
			return nil, fmt.Errorf("query supervisor metric %s from %s: %w", request.ID, root, err)
		}
		if len(points) > 0 {
			return points, nil
		}
	}
	return nil, nil
}

func runningSupervisorRoots() ([]string, error) {
	configured, err := configuredProjectRoots()
	if err != nil {
		return nil, err
	}
	var roots []string
	for _, root := range configured {
		report, err := procfile.Status(root, "")
		if errors.Is(err, procfile.ErrProcfileNotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read supervisor status for %s: %w", root, err)
		}
		if !report.Running {
			continue
		}
		roots = append(roots, root)
	}
	return roots, nil
}

func configuredProjectRoots() ([]string, error) {
	projects, err := LoadProjects()
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	var roots []string
	for _, project := range projects {
		root := project.ResolvedDir()
		if root == "" {
			continue
		}
		if _, ok := seen[root]; ok {
			continue
		}
		seen[root] = struct{}{}
		roots = append(roots, root)
	}
	return roots, nil
}
