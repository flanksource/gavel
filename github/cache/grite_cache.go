package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/flanksource/gavel/todos"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GriteIssueCache stores the projection of a grite issue plus its accumulated
// event history. It is a read cache: grite remains the source of truth (writes
// pass through the CLI) and this table is refreshed by an incremental
// `grite export --since`. The primary key is (Repo, IssueID) so each issue has
// at most one row per workspace.
type GriteIssueCache struct {
	Repo         string `gorm:"primaryKey;size:512"`
	IssueID      string `gorm:"primaryKey;size:64"`
	Title        string
	State        string `gorm:"size:16;index"`
	Labels       []byte // JSON-encoded []string
	EventsJSON   []byte // JSON-encoded []griteEvent, full accumulated history
	CreatedTS    int64
	UpdatedTS    int64
	CommentCount int
	SyncedAt     time.Time
}

// GriteSyncCursor tracks, per workspace, the max event timestamp already pulled
// into the cache (the `--since` argument for the next export) and when the last
// sync ran (the TTL clock the read path checks).
type GriteSyncCursor struct {
	Repo        string `gorm:"primaryKey;size:512"`
	LastEventTS int64
	SyncedAt    time.Time
}

// Enabled reports whether the store can serve as a grite read cache.
func (s *Store) Enabled() bool { return !s.Disabled() }

// LoadCursor returns the stored cursor for a repo. A missing row (or disabled
// store) yields a zero cursor and zero time, which the caller treats as "never
// synced".
func (s *Store) LoadCursor(ctx context.Context, repo string) (int64, time.Time, error) {
	if s.Disabled() {
		return 0, time.Time{}, nil
	}
	var cur GriteSyncCursor
	err := s.gorm().WithContext(ctx).Where("repo = ?", repo).First(&cur).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, time.Time{}, nil
	}
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("load grite sync cursor: %w", err)
	}
	return cur.LastEventTS, cur.SyncedAt, nil
}

// ListIssues returns every cached issue for a repo.
func (s *Store) ListIssues(ctx context.Context, repo string) ([]todos.CachedIssue, error) {
	if s.Disabled() {
		return nil, nil
	}
	var rows []GriteIssueCache
	if err := s.gorm().WithContext(ctx).Where("repo = ?", repo).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list grite issues: %w", err)
	}
	out := make([]todos.CachedIssue, 0, len(rows))
	for _, r := range rows {
		ci, err := r.toCached()
		if err != nil {
			return nil, err
		}
		out = append(out, ci)
	}
	return out, nil
}

// GetIssue returns a single cached issue, or (nil, nil) when absent.
func (s *Store) GetIssue(ctx context.Context, repo, issueID string) (*todos.CachedIssue, error) {
	if s.Disabled() {
		return nil, nil
	}
	var r GriteIssueCache
	err := s.gorm().WithContext(ctx).Where("repo = ? AND issue_id = ?", repo, issueID).First(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get grite issue %s: %w", issueID, err)
	}
	ci, err := r.toCached()
	if err != nil {
		return nil, err
	}
	return &ci, nil
}

// UpsertSync writes the issue snapshot and advances the cursor in one transaction.
// SyncedAt is stamped now on every call so the read-path TTL resets even when the
// export carried no new events.
func (s *Store) UpsertSync(ctx context.Context, repo string, issues []todos.CachedIssue, lastEventTS int64) error {
	if s.Disabled() {
		return nil
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	now := time.Now()
	return s.gorm().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, ci := range issues {
			labels, err := json.Marshal(ci.Labels)
			if err != nil {
				return fmt.Errorf("encode labels for %s: %w", ci.IssueID, err)
			}
			row := GriteIssueCache{
				Repo:         repo,
				IssueID:      ci.IssueID,
				Title:        ci.Title,
				State:        ci.State,
				Labels:       labels,
				EventsJSON:   ci.EventsJSON,
				CreatedTS:    ci.CreatedTS,
				UpdatedTS:    ci.UpdatedTS,
				CommentCount: ci.CommentCount,
				SyncedAt:     now,
			}
			if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error; err != nil {
				return fmt.Errorf("upsert grite issue %s: %w", ci.IssueID, err)
			}
		}
		cursor := GriteSyncCursor{Repo: repo, LastEventTS: lastEventTS, SyncedAt: now}
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&cursor).Error; err != nil {
			return fmt.Errorf("upsert grite sync cursor: %w", err)
		}
		return nil
	})
}

func (r GriteIssueCache) toCached() (todos.CachedIssue, error) {
	var labels []string
	if len(r.Labels) > 0 {
		if err := json.Unmarshal(r.Labels, &labels); err != nil {
			return todos.CachedIssue{}, fmt.Errorf("decode labels for %s: %w", r.IssueID, err)
		}
	}
	return todos.CachedIssue{
		Repo:         r.Repo,
		IssueID:      r.IssueID,
		Title:        r.Title,
		State:        r.State,
		Labels:       labels,
		CreatedTS:    r.CreatedTS,
		UpdatedTS:    r.UpdatedTS,
		CommentCount: r.CommentCount,
		EventsJSON:   r.EventsJSON,
	}, nil
}
