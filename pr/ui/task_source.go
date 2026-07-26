package ui

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/flanksource/clicky/metrics"
	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/internal/taskhistory"
	"github.com/flanksource/gavel/procfile"
)

type supervisorTaskSource struct {
	history func(context.Context) ([]taskhistory.Record, error)
}

func newSupervisorTaskSource() *supervisorTaskSource {
	return &supervisorTaskSource{history: loadTaskHistory}
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
	archived, err := s.history(ctx)
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
	archived, err := s.history(ctx)
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

func loadTaskHistory(ctx context.Context) ([]taskhistory.Record, error) {
	now := time.Now().UTC()
	roots, err := configuredProjectRoots()
	if err != nil {
		return nil, err
	}
	spooled, err := taskhistory.LoadRoots(roots, now)
	if err != nil {
		return nil, err
	}
	db, err := database.Shared(ctx)
	if err != nil {
		return nil, fmt.Errorf("open task history database: %w", err)
	}
	if db.Disabled() {
		return spooled, nil
	}
	store, err := taskhistory.NewStore(db.Gorm())
	if err != nil {
		return nil, err
	}
	if err := store.Import(ctx, spooled); err != nil {
		return nil, err
	}
	if err := store.Prune(ctx, now); err != nil {
		return nil, err
	}
	return store.List(ctx, now)
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
