package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	gavelgit "github.com/flanksource/gavel/git"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CommitStatCache stores the aggregated git diff footprint of a todo's commits
// (those carrying its Gavel-Issue-Id trailer), keyed by (Repo, IssueID). It is a
// pure read cache derived from `git log`: the working tree stays the source of
// truth and SaveCommitStats refreshes the rows. Recomputing the whole workspace
// is one git pass, so the cache spares the todo-list endpoint that pass on every
// poll.
type CommitStatCache struct {
	Repo     string `gorm:"primaryKey;size:512"`
	IssueID  string `gorm:"primaryKey;size:128"`
	Commits  int
	Files    int
	Adds     int
	Dels     int
	SyncedAt time.Time
}

// CommitStatCursor records, per repo, when the commit-stat rows were last
// recomputed — the TTL clock the read path checks before triggering a fresh
// `git log` pass.
type CommitStatCursor struct {
	Repo     string `gorm:"primaryKey;size:512"`
	SyncedAt time.Time
}

// CommitStats returns the cached per-issue diff stats for a repo along with the
// time they were last recomputed (zero when never synced or the store is
// disabled). Reads are lock-free thanks to postgres MVCC.
func (s *Store) CommitStats(ctx context.Context, repo string) (map[string]gavelgit.DiffStat, time.Time, error) {
	if s.Disabled() {
		return map[string]gavelgit.DiffStat{}, time.Time{}, nil
	}
	var cursor CommitStatCursor
	err := s.gorm().WithContext(ctx).Where("repo = ?", repo).First(&cursor).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, time.Time{}, fmt.Errorf("load commit-stat cursor: %w", err)
	}
	var rows []CommitStatCache
	if err := s.gorm().WithContext(ctx).Where("repo = ?", repo).Find(&rows).Error; err != nil {
		return nil, time.Time{}, fmt.Errorf("list commit stats: %w", err)
	}
	stats := make(map[string]gavelgit.DiffStat, len(rows))
	for _, r := range rows {
		stats[r.IssueID] = gavelgit.DiffStat{Commits: r.Commits, Files: r.Files, Adds: r.Adds, Dels: r.Dels}
	}
	return stats, cursor.SyncedAt, nil
}

// SaveCommitStats replaces a repo's commit-stat rows with the freshly computed
// map and advances its cursor, all in one transaction. Issue rows no longer in
// the map are deleted so a rewritten/dropped history never leaves stale counts.
func (s *Store) SaveCommitStats(ctx context.Context, repo string, stats map[string]gavelgit.DiffStat) error {
	if s.Disabled() {
		return nil
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	now := time.Now()
	return s.gorm().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("repo = ?", repo).Delete(&CommitStatCache{}).Error; err != nil {
			return fmt.Errorf("clear commit stats for %s: %w", repo, err)
		}
		for issueID, st := range stats {
			row := CommitStatCache{
				Repo:     repo,
				IssueID:  issueID,
				Commits:  st.Commits,
				Files:    st.Files,
				Adds:     st.Adds,
				Dels:     st.Dels,
				SyncedAt: now,
			}
			if err := tx.Create(&row).Error; err != nil {
				return fmt.Errorf("write commit stat %s: %w", issueID, err)
			}
		}
		cursor := CommitStatCursor{Repo: repo, SyncedAt: now}
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&cursor).Error; err != nil {
			return fmt.Errorf("upsert commit-stat cursor: %w", err)
		}
		return nil
	})
}
