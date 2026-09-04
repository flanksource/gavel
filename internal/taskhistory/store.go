package taskhistory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	clickytask "github.com/flanksource/clicky/task"
	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

// RunSummary is the listing projection of an archived run: everything a run
// listing needs and nothing a run detail needs. Snapshots stay in the database
// until Snapshot asks for one run, because the archive's snapshot payload is
// two orders of magnitude larger than its run metadata.
type RunSummary struct {
	Run        clickytask.RunMeta
	ArchivedAt time.Time
}

type storedRun struct {
	Run        string    `gorm:"column:run"`
	ArchivedAt time.Time `gorm:"column:archived_at"`
}

func NewStore(db *gorm.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("task history store requires a database")
	}
	return &Store{db: db}, nil
}

func (s *Store) Import(ctx context.Context, records []Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, record := range records {
			if err := importRecord(tx, record); err != nil {
				return err
			}
		}
		return nil
	})
}

func importRecord(db *gorm.DB, record Record) error {
	if err := validateRecord(record); err != nil {
		return err
	}
	startedAt, err := time.Parse(time.RFC3339Nano, record.Run.StartedAt)
	if err != nil {
		return fmt.Errorf("parse task run %q startedAt: %w", record.Run.ID, err)
	}
	run, err := json.Marshal(record.Run)
	if err != nil {
		return fmt.Errorf("marshal task run %q: %w", record.Run.ID, err)
	}
	snapshots, err := json.Marshal(record.Snapshots)
	if err != nil {
		return fmt.Errorf("marshal task snapshots %q: %w", record.Run.ID, err)
	}
	// The DO UPDATE is guarded because archived runs are immutable and the spool
	// sweep re-offers every retained record on each pass. Without the guard an
	// unchanged re-import still writes a new tuple, so a 40-row table accumulates
	// dead tuples (and WAL) in proportion to sweep frequency rather than content.
	result := db.Exec(`
		INSERT INTO task_run_history (id, started_at, run, snapshots, archived_at, expires_at)
		VALUES (?, ?, CAST(? AS jsonb), CAST(? AS jsonb), ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			started_at = EXCLUDED.started_at,
			run = EXCLUDED.run,
			snapshots = EXCLUDED.snapshots,
			archived_at = EXCLUDED.archived_at,
			expires_at = EXCLUDED.expires_at
		WHERE task_run_history.started_at IS DISTINCT FROM EXCLUDED.started_at
			OR task_run_history.archived_at IS DISTINCT FROM EXCLUDED.archived_at
			OR task_run_history.expires_at IS DISTINCT FROM EXCLUDED.expires_at
			OR task_run_history.run IS DISTINCT FROM EXCLUDED.run
			OR task_run_history.snapshots IS DISTINCT FROM EXCLUDED.snapshots`,
		record.Run.ID, startedAt, string(run), string(snapshots), record.ArchivedAt, record.ArchivedAt.Add(Retention))
	if result.Error != nil {
		return fmt.Errorf("import task history %q: %w", record.Run.ID, result.Error)
	}
	return nil
}

func (s *Store) Prune(ctx context.Context, now time.Time) error {
	if err := s.db.WithContext(ctx).Exec("DELETE FROM task_run_history WHERE expires_at <= ?", now).Error; err != nil {
		return fmt.Errorf("prune task history database: %w", err)
	}
	return nil
}

// ListRuns returns every retained run's metadata, newest first. It never reads
// the snapshots column: with hundreds of runs that column is tens of megabytes
// of captured output, and sorting rows that wide spills the sort to disk.
func (s *Store) ListRuns(ctx context.Context, now time.Time) ([]RunSummary, error) {
	var stored []storedRun
	if err := s.db.WithContext(ctx).Raw(`
		SELECT CAST(run AS text) AS run, archived_at
		FROM task_run_history
		WHERE expires_at > ?
		ORDER BY started_at DESC`, now).Scan(&stored).Error; err != nil {
		return nil, fmt.Errorf("list task history database: %w", err)
	}
	runs := make([]RunSummary, 0, len(stored))
	for _, row := range stored {
		summary := RunSummary{ArchivedAt: row.ArchivedAt.UTC()}
		if err := json.Unmarshal([]byte(row.Run), &summary.Run); err != nil {
			return nil, fmt.Errorf("parse stored task run: %w", err)
		}
		if err := validateRunID(summary.Run.ID); err != nil {
			return nil, fmt.Errorf("parse stored task run: %w", err)
		}
		runs = append(runs, summary)
	}
	return runs, nil
}

// Snapshot returns one archived run's task snapshots by primary key. A run
// that was never archived yields nil: that is the ordinary case for a run
// still owned by a live process, not an error.
func (s *Store) Snapshot(ctx context.Context, id string) ([]clickytask.TaskSnapshot, error) {
	if err := validateRunID(id); err != nil {
		return nil, err
	}
	var stored []string
	if err := s.db.WithContext(ctx).Raw(`
		SELECT CAST(snapshots AS text)
		FROM task_run_history
		WHERE id = ?`, id).Scan(&stored).Error; err != nil {
		return nil, fmt.Errorf("load task history snapshot %q: %w", id, err)
	}
	if len(stored) == 0 {
		return nil, nil
	}
	var snapshots []clickytask.TaskSnapshot
	if err := json.Unmarshal([]byte(stored[0]), &snapshots); err != nil {
		return nil, fmt.Errorf("parse stored task snapshots %q: %w", id, err)
	}
	if len(snapshots) == 0 {
		return nil, fmt.Errorf("stored task history %q has no snapshots", id)
	}
	return snapshots, nil
}
