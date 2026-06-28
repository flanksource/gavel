package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TestRunCache stores the per-run summary of a .gavel/run-*.json snapshot plus
// the absolute path to the file on disk. It is a read cache populated by the PR
// UI's background scanner: the run-*.json files remain the source of truth, and
// the detail endpoint reads the file at Path on demand — the snapshot body is
// never stored here. Primary key is (WorkspaceDir, RunID) so each run has at
// most one row per workspace.
type TestRunCache struct {
	WorkspaceDir string `gorm:"primaryKey;size:512"`
	RunID        string `gorm:"primaryKey;size:255"` // run-*.json file stem
	Path         string `gorm:"size:1024"`           // absolute path to the run-*.json
	Kind         string `gorm:"size:16;index"`       // test | lint | test+lint
	Repo         string `gorm:"size:255"`
	SHA          string `gorm:"size:64"`
	StartedTS    int64  `gorm:"index"` // unix nanos, the scan watermark
	EndedTS      int64

	Passed  int
	Failed  int
	Skipped int
	Warned  int
	Total   int

	LintViolations int
	LintLinters    int

	Frameworks []byte // JSON-encoded []string
	SyncedAt   time.Time
}

// TestRunCursor records, per workspace, the highest run start timestamp already
// scanned into the cache (the `since` watermark for the next scan) and when the
// last scan ran.
type TestRunCursor struct {
	WorkspaceDir string `gorm:"primaryKey;size:512"`
	LastRunTS    int64
	SyncedAt     time.Time
}

// LoadTestRunCursor returns the stored cursor for a workspace. A missing row
// (or disabled store) yields a zero watermark and zero time — "never scanned".
func (s *Store) LoadTestRunCursor(ctx context.Context, dir string) (int64, time.Time, error) {
	if s.Disabled() {
		return 0, time.Time{}, nil
	}
	var cur TestRunCursor
	err := s.gorm().WithContext(ctx).Where("workspace_dir = ?", dir).First(&cur).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, time.Time{}, nil
	}
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("load test run cursor: %w", err)
	}
	return cur.LastRunTS, cur.SyncedAt, nil
}

// ListTestRuns returns every cached run across all workspaces, newest-first.
func (s *Store) ListTestRuns(ctx context.Context) ([]TestRunCache, error) {
	if s.Disabled() {
		return nil, nil
	}
	var rows []TestRunCache
	if err := s.gorm().WithContext(ctx).Order("started_ts DESC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list test runs: %w", err)
	}
	return rows, nil
}

// ListTestRunsForDir returns the cached runs for a single workspace, newest-first.
func (s *Store) ListTestRunsForDir(ctx context.Context, dir string) ([]TestRunCache, error) {
	if s.Disabled() {
		return nil, nil
	}
	var rows []TestRunCache
	err := s.gorm().WithContext(ctx).Where("workspace_dir = ?", dir).Order("started_ts DESC").Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list test runs for %s: %w", dir, err)
	}
	return rows, nil
}

// GetTestRunPath returns the on-disk path of a cached run, or "" when absent.
func (s *Store) GetTestRunPath(ctx context.Context, dir, runID string) (string, error) {
	if s.Disabled() {
		return "", nil
	}
	var row TestRunCache
	err := s.gorm().WithContext(ctx).
		Where("workspace_dir = ? AND run_id = ?", dir, runID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get test run path %s/%s: %w", dir, runID, err)
	}
	return row.Path, nil
}

// UpsertTestRuns writes the run summaries and advances the workspace cursor in
// one transaction. SyncedAt is stamped now on every call so the read-path TTL
// resets even when a scan found no new runs (rows empty). lastRunTS is the new
// high-watermark the caller computed (max of the prior cursor and any new runs).
func (s *Store) UpsertTestRuns(ctx context.Context, dir string, rows []TestRunCache, lastRunTS int64) error {
	if s.Disabled() {
		return nil
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	now := time.Now()
	return s.gorm().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, row := range rows {
			row.WorkspaceDir = dir
			row.SyncedAt = now
			if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error; err != nil {
				return fmt.Errorf("upsert test run %s: %w", row.RunID, err)
			}
		}
		cursor := TestRunCursor{WorkspaceDir: dir, LastRunTS: lastRunTS, SyncedAt: now}
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&cursor).Error; err != nil {
			return fmt.Errorf("upsert test run cursor: %w", err)
		}
		return nil
	})
}
