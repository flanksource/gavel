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

// historyCacheTTL collapses the archived-run reads issued by a single dashboard
// render. Runs() and Snapshot() each need the full archive, and the task stream
// re-asks per subscriber, so an uncached read multiplies one render into many
// full-table scans.
const historyCacheTTL = 2 * time.Second

type supervisorTaskSource struct {
	history func(context.Context) ([]taskhistory.Record, error)

	mu       sync.Mutex
	cached   []taskhistory.Record
	cachedAt time.Time
}

func newSupervisorTaskSource() *supervisorTaskSource {
	return &supervisorTaskSource{history: loadTaskHistory}
}

// archivedRuns serves the archive from a short-lived cache. taskhistory.Record
// values are only ever replaced wholesale, so callers may read the slice.
func (s *supervisorTaskSource) archivedRuns(ctx context.Context) ([]taskhistory.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cachedAt.IsZero() || time.Since(s.cachedAt) >= historyCacheTTL {
		records, err := s.history(ctx)
		if err != nil {
			return nil, err
		}
		s.cached, s.cachedAt = records, time.Now()
	}
	return s.cached, nil
}

func (s *supervisorTaskSource) Runs(ctx context.Context, filter clickytask.RunFilter) ([]clickytask.RunMeta, error) {
	var runs []clickytask.RunMeta
	roots, err := runningSupervisorRoots()
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
	roots, err := runningSupervisorRoots()
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
			return record.Snapshots, nil
		}
	}
	return nil, nil
}

// loadTaskHistory reads the archived runs. It is a pure read: folding the
// on-disk spool into the database is importTaskHistory's job, driven by the
// archive write path and the periodic sweep. Importing from here would make
// every dashboard poll rewrite the whole archive.
func loadTaskHistory(ctx context.Context) ([]taskhistory.Record, error) {
	now := time.Now().UTC()
	db, err := database.Shared(ctx)
	if err != nil {
		return nil, fmt.Errorf("open task history database: %w", err)
	}
	if db.Disabled() {
		roots, err := configuredProjectRoots()
		if err != nil {
			return nil, err
		}
		return taskhistory.LoadRoots(roots, now)
	}
	store, err := taskhistory.NewStore(db.Gorm())
	if err != nil {
		return nil, err
	}
	return store.List(ctx, now)
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
	t := time.NewTicker(taskHistorySweepInterval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
		case <-s.taskHistoryImport:
		}
		if err := importTaskHistory(context.Background()); err != nil {
			logger.Errorf("import task history: %v", err)
		}
	}
}

// importTaskHistory folds every retained spool record into the database mirror
// and drops expired rows. This is the archive's only write path: call it when a
// run is archived and from the periodic sweep, never from a request handler.
func importTaskHistory(ctx context.Context) error {
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
	spooled, err := taskhistory.LoadRoots(roots, now)
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
	return store.Prune(ctx, now)
}

func (s *supervisorTaskSource) Control(_ context.Context, id string, action clickytask.ControlAction) error {
	roots, err := runningSupervisorRoots()
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
	roots, err := runningSupervisorRoots()
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
	roots, err := runningSupervisorRoots()
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
