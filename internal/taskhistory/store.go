package taskhistory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

type storedRecord struct {
	Run        string    `gorm:"column:run"`
	Snapshots  string    `gorm:"column:snapshots"`
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
	result := db.Exec(`
		INSERT INTO task_run_history (id, started_at, run, snapshots, archived_at, expires_at)
		VALUES (?, ?, CAST(? AS jsonb), CAST(? AS jsonb), ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			started_at = EXCLUDED.started_at,
			run = EXCLUDED.run,
			snapshots = EXCLUDED.snapshots,
			archived_at = EXCLUDED.archived_at,
			expires_at = EXCLUDED.expires_at`,
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

func (s *Store) List(ctx context.Context, now time.Time) ([]Record, error) {
	var stored []storedRecord
	if err := s.db.WithContext(ctx).Raw(`
		SELECT CAST(run AS text) AS run, CAST(snapshots AS text) AS snapshots, archived_at
		FROM task_run_history
		WHERE expires_at > ?
		ORDER BY started_at DESC`, now).Scan(&stored).Error; err != nil {
		return nil, fmt.Errorf("list task history database: %w", err)
	}
	records := make([]Record, 0, len(stored))
	for _, row := range stored {
		record, err := row.record()
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (row storedRecord) record() (Record, error) {
	record := Record{ArchivedAt: row.ArchivedAt.UTC()}
	if err := json.Unmarshal([]byte(row.Run), &record.Run); err != nil {
		return Record{}, fmt.Errorf("parse stored task run: %w", err)
	}
	if err := json.Unmarshal([]byte(row.Snapshots), &record.Snapshots); err != nil {
		return Record{}, fmt.Errorf("parse stored task snapshots %q: %w", record.Run.ID, err)
	}
	if err := validateRecord(record); err != nil {
		return Record{}, fmt.Errorf("parse stored task history: %w", err)
	}
	return record, nil
}
