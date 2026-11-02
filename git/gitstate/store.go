package gitstate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/status"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChangeChannel is the NOTIFY channel a repo's id is sent on whenever its rows
// change, so every gavel process sharing the database can push the change to
// its UI. The triggers of schema/171 send it, not this package.
const ChangeChannel = "gavel_git_change"

// ErrNotFound reports a repository that is not one of the store's org's.
var ErrNotFound = errors.New("git repo not found")

// Store reads and writes one org's rows of the git_* tables. Triggers on those
// tables (schema/171) advance the repo's generation and NOTIFY ChangeChannel in
// the same transaction whenever a write changes what a reader sees.
type Store struct {
	db  *gorm.DB
	org uuid.UUID
}

// NewStore returns the local org's store.
func NewStore(db *gorm.DB) *Store {
	return &Store{db: db, org: tenant.LocalOrgID}
}

// ForOrg returns a Store over org's git state, sharing the connection.
func (s *Store) ForOrg(org uuid.UUID) *Store {
	return &Store{db: s.db, org: org}
}

// inOrgRepo limits a repo_id column to the org's repositories. The returned
// condition takes the org ID as its one argument.
func inOrgRepo(column string) string {
	return column + ` IN (SELECT id FROM git_repos WHERE org_id = ?)`
}

// requireRepo fails with ErrNotFound unless repoID is one of the org's repos.
func (s *Store) requireRepo(db *gorm.DB, repoID uuid.UUID) error {
	var exists bool
	if err := db.Raw(`SELECT EXISTS(SELECT 1 FROM git_repos WHERE id = ? AND org_id = ?)`, repoID, s.org).Scan(&exists).Error; err != nil {
		return fmt.Errorf("look up git repo %s: %w", repoID, err)
	}
	if !exists {
		return fmt.Errorf("%w: %s", ErrNotFound, repoID)
	}
	return nil
}

// WorktreeRef is a worktree as a ref scan records it.
type WorktreeRef struct {
	gavelgit.Worktree
	LastCommitAt time.Time
}

// RefsUpdate is the ref-level state of a repository: its base, worktrees and
// branches. Upstream and UpstreamSHA are the primary checkout's upstream,
// empty when it has none.
type RefsUpdate struct {
	Fingerprint    string
	BaseBranch     string
	BaseSHA        string
	CurrentBranch  string
	BaseCheckedOut bool
	Upstream       string
	UpstreamSHA    string
	Worktrees      []WorktreeRef
	Branches       []gavelgit.BranchRef
}

// StatusUpdate is the uncommitted state of one worktree.
type StatusUpdate struct {
	Fingerprint string
	Changes     Changes
	TouchedAt   *time.Time
	Files       []status.FileStatus
}

// RangeKey names a range by the two commits it compares.
type RangeKey struct {
	Base, Head string
}

// Range is the cached comparison of Head against Base.
type Range struct {
	RangeKey
	gavelgit.RangeCompare
}

// EnsureRepo returns the id of the org's repository whose primary checkout is
// rootDir, creating its row on first sight.
func (s *Store) EnsureRepo(ctx context.Context, rootDir, commonDir string) (uuid.UUID, error) {
	var row struct{ ID uuid.UUID }
	err := s.db.WithContext(ctx).Raw(`
		INSERT INTO git_repos (org_id, user_id, root_dir, common_dir) VALUES (?, ?, ?, ?)
		ON CONFLICT (org_id, root_dir) DO UPDATE SET common_dir = EXCLUDED.common_dir
		RETURNING id`, s.org, tenant.FromContext(ctx).UserID, rootDir, commonDir).Scan(&row).Error
	if err != nil {
		return uuid.Nil, fmt.Errorf("register git repo %s: %w", rootDir, err)
	}
	return row.ID, nil
}

// errLocked reports that another session holds the repo's scan lock.
var errLocked = errors.New("git repo scan lock held by another session")

// Locked runs fn in a transaction holding the advisory scan lock named key,
// so two gavel processes sharing the database never run one scan at once.
// Unless wait is set it reports false, without running fn, when another
// session holds the lock.
func (s *Store) Locked(ctx context.Context, key string, wait bool, fn func(tx *Store) error) (bool, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if wait {
			if err := tx.Exec(`SELECT pg_advisory_xact_lock(hashtextextended(?, 0))`, key).Error; err != nil {
				return fmt.Errorf("lock %s: %w", key, err)
			}
			return fn(&Store{db: tx, org: s.org})
		}
		var acquired bool
		if err := tx.Raw(`SELECT pg_try_advisory_xact_lock(hashtextextended(?, 0))`, key).Scan(&acquired).Error; err != nil {
			return fmt.Errorf("lock %s: %w", key, err)
		}
		if !acquired {
			return errLocked
		}
		return fn(&Store{db: tx, org: s.org})
	})
	if errors.Is(err, errLocked) {
		return false, nil
	}
	return err == nil, err
}

// MarkRefsUnchanged records a ref scan whose fingerprint equals the one the
// repo's last good scan stored, and reports whether it did. Like
// MarkStatusUnchanged it writes only the UNLOGGED git_repo_overview row
// (schema/171); false means the refs changed, the last scan failed, or the
// overview has no row, and the caller applies the scan instead.
func (s *Store) MarkRefsUnchanged(ctx context.Context, repoID uuid.UUID, fingerprint string) (bool, error) {
	result := s.db.WithContext(ctx).Exec(`
		UPDATE git_repo_overview SET refs_scanned_at = now()
		WHERE repo_id = ? AND refs_fingerprint = ? AND `+inOrgRepo("repo_id"), repoID, fingerprint, s.org)
	if result.Error != nil {
		return false, fmt.Errorf("mark refs of git repo %s scanned: %w", repoID, result.Error)
	}
	return result.RowsAffected == 1, nil
}

// ApplyRefs replaces the repo's ref-level rows with update: worktrees missing
// from it are marked removed, branches missing from it are deleted.
func (s *Store) ApplyRefs(ctx context.Context, repoID uuid.UUID, update RefsUpdate) error {
	db := s.db.WithContext(ctx)
	if err := s.requireRepo(db, repoID); err != nil {
		return err
	}
	if err := db.Exec(`
		UPDATE git_repos SET refs_fingerprint = ?, refs_scanned_at = now(), base_branch = ?, base_sha = ?,
			current_branch = ?, base_checked_out = ?, upstream_branch = NULLIF(?, ''), upstream_sha = NULLIF(?, ''),
			error = NULL, error_at = NULL
		WHERE id = ?`,
		update.Fingerprint, update.BaseBranch, update.BaseSHA, update.CurrentBranch, update.BaseCheckedOut,
		update.Upstream, update.UpstreamSHA, repoID).Error; err != nil {
		return fmt.Errorf("update refs of git repo %s: %w", repoID, err)
	}
	paths := make([]string, 0, len(update.Worktrees))
	for i, wt := range update.Worktrees {
		paths = append(paths, wt.Path)
		var lastCommitAt *time.Time
		if !wt.LastCommitAt.IsZero() {
			lastCommitAt = &wt.LastCommitAt
		}
		if err := db.Exec(`
			INSERT INTO git_worktrees (repo_id, path, branch, head_sha, is_primary, detached, prunable, ordinal, last_commit_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (repo_id, path) DO UPDATE SET branch = EXCLUDED.branch, head_sha = EXCLUDED.head_sha,
				is_primary = EXCLUDED.is_primary, detached = EXCLUDED.detached, prunable = EXCLUDED.prunable,
				ordinal = EXCLUDED.ordinal, last_commit_at = EXCLUDED.last_commit_at, removed_at = NULL`,
			repoID, wt.Path, wt.Branch, wt.Head, wt.Primary, wt.Detached, wt.Prunable, i, lastCommitAt).Error; err != nil {
			return fmt.Errorf("upsert worktree %s: %w", wt.Path, err)
		}
	}
	if err := db.Exec(`UPDATE git_worktrees SET removed_at = now() WHERE repo_id = ? AND removed_at IS NULL AND NOT (path IN ?)`,
		repoID, paths).Error; err != nil {
		return fmt.Errorf("mark removed worktrees of git repo %s: %w", repoID, err)
	}
	if err := db.Exec(`DELETE FROM git_branches WHERE repo_id = ?`, repoID).Error; err != nil {
		return fmt.Errorf("clear branches of git repo %s: %w", repoID, err)
	}
	for _, branch := range update.Branches {
		var worktree *string
		if branch.Worktree != "" {
			worktree = &branch.Worktree
		}
		if err := db.Exec(`INSERT INTO git_branches (repo_id, name, head_sha, worktree_path, last_commit_at) VALUES (?, ?, ?, ?, ?)`,
			repoID, branch.Name, branch.Head, worktree, branch.LastCommitAt).Error; err != nil {
			return fmt.Errorf("insert branch %s: %w", branch.Name, err)
		}
	}
	return nil
}

// MissingRanges returns the keys with no cached comparison yet.
func (s *Store) MissingRanges(ctx context.Context, repoID uuid.UUID, keys []RangeKey) ([]RangeKey, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	var present []RangeKey
	heads := make([]string, 0, len(keys))
	for _, key := range keys {
		heads = append(heads, key.Head)
	}
	err := s.db.WithContext(ctx).Raw(`SELECT base_sha AS base, head_sha AS head FROM git_range_stats WHERE repo_id = ? AND head_sha IN ? AND `+inOrgRepo("repo_id"),
		repoID, heads, s.org).Scan(&present).Error
	if err != nil {
		return nil, fmt.Errorf("read cached ranges of git repo %s: %w", repoID, err)
	}
	have := make(map[RangeKey]bool, len(present))
	for _, key := range present {
		have[key] = true
	}
	var missing []RangeKey
	for _, key := range keys {
		if !have[key] {
			have[key] = true
			missing = append(missing, key)
		}
	}
	return missing, nil
}

// SaveRanges stores computed comparisons. A range is a pure function of its
// two commits, so a row that already exists is left as is.
func (s *Store) SaveRanges(ctx context.Context, repoID uuid.UUID, ranges []Range) error {
	db := s.db.WithContext(ctx)
	if err := s.requireRepo(db, repoID); err != nil {
		return err
	}
	for _, r := range ranges {
		if err := db.Exec(`
			INSERT INTO git_range_stats (repo_id, base_sha, head_sha, merge_base, ahead, behind, commits, files, adds, dels)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (repo_id, base_sha, head_sha) DO NOTHING`,
			repoID, r.Base, r.Head, r.MergeBase, r.Ahead, r.Behind, r.Diff.Commits, r.Diff.Files, r.Diff.Adds, r.Diff.Dels).Error; err != nil {
			return fmt.Errorf("save range %s..%s: %w", r.Base, r.Head, err)
		}
	}
	return nil
}

// MarkStatusUnchanged records a status scan whose fingerprint equals the one
// the worktree's last good scan stored, and reports whether it did. It writes
// only the UNLOGGED git_worktree_overview row (schema/170); false means the
// status changed, the last scan failed, or the overview has no row for the
// worktree, and the caller applies a full scan instead.
func (s *Store) MarkStatusUnchanged(ctx context.Context, repoID uuid.UUID, path, fingerprint string) (bool, error) {
	result := s.db.WithContext(ctx).Exec(`
		UPDATE git_worktree_overview SET status_scanned_at = now()
		WHERE repo_id = ? AND path = ? AND status_fingerprint = ? AND `+inOrgRepo("repo_id"), repoID, path, fingerprint, s.org)
	if result.Error != nil {
		return false, fmt.Errorf("mark worktree %s scanned: %w", path, result.Error)
	}
	return result.RowsAffected == 1, nil
}

// EnsureOverview rebuilds git_worktree_overview and git_repo_overview when
// crash recovery left them empty, announcing every repo so listeners reload,
// and reports whether it did. The rebuild covers every org's repos.
func (s *Store) EnsureOverview(ctx context.Context) (bool, error) {
	var rebuilt bool
	if err := s.db.WithContext(ctx).Raw(`SELECT public.gavel_git_overview_ensure()`).Scan(&rebuilt).Error; err != nil {
		return false, fmt.Errorf("ensure git overviews: %w", err)
	}
	return rebuilt, nil
}

// ApplyStatus replaces the worktree's uncommitted state.
func (s *Store) ApplyStatus(ctx context.Context, repoID uuid.UUID, path string, update StatusUpdate) error {
	files := update.Files
	if files == nil {
		files = []status.FileStatus{}
	}
	blob, err := json.Marshal(files)
	if err != nil {
		return fmt.Errorf("encode files of worktree %s: %w", path, err)
	}
	// Lock the repo row first: ApplyRefs locks it (and, through the schema/171
	// triggers, the overview row) before its worktree rows, so a status write
	// taking them the other way round deadlocks with the refs scan a commit
	// triggers alongside it. Callers run this inside Locked's transaction.
	if err := s.lockRepo(ctx, repoID); err != nil {
		return err
	}
	c := update.Changes
	result := s.db.WithContext(ctx).Exec(`
		UPDATE git_worktrees SET staged = ?, unstaged = ?, "both" = ?, untracked = ?, conflict = ?, adds = ?, dels = ?,
			touched_at = ?, files = ?::jsonb, status_fingerprint = ?, status_scanned_at = now(), status_error = NULL
		WHERE repo_id = ? AND path = ?`,
		c.Staged, c.Unstaged, c.Both, c.Untracked, c.Conflict, c.Adds, c.Dels,
		update.TouchedAt, string(blob), update.Fingerprint, repoID, path)
	if result.Error != nil {
		return fmt.Errorf("update status of worktree %s: %w", path, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update status of worktree %s: not a recorded worktree of git repo %s", path, repoID)
	}
	return nil
}

// RecordStatusError records a failed status scan of the worktree, keeping
// its last good state.
func (s *Store) RecordStatusError(ctx context.Context, repoID uuid.UUID, path string, scanErr error) error {
	// Repo row before worktree row, as in ApplyStatus, held until the update
	// commits.
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := (&Store{db: tx, org: s.org}).lockRepo(ctx, repoID); err != nil {
			return err
		}
		if err := tx.Exec(`UPDATE git_worktrees SET status_error = ?, status_scanned_at = now() WHERE repo_id = ? AND path = ?`,
			scanErr.Error(), repoID, path).Error; err != nil {
			return fmt.Errorf("record status error of worktree %s: %w", path, err)
		}
		return nil
	})
}

// lockRepo takes the repo row lock ApplyRefs' update takes, without writing
// the row, so status writes order their locks the same way. It fails with
// ErrNotFound unless repoID is one of the org's repos.
func (s *Store) lockRepo(ctx context.Context, repoID uuid.UUID) error {
	if err := s.requireRepo(s.db.WithContext(ctx), repoID); err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Exec(`SELECT 1 FROM git_repos WHERE id = ? FOR NO KEY UPDATE`, repoID).Error; err != nil {
		return fmt.Errorf("lock git repo %s: %w", repoID, err)
	}
	return nil
}

// RecordError records a failed ref scan of the repo, keeping its last good
// rows.
func (s *Store) RecordError(ctx context.Context, repoID uuid.UUID, scanErr error) error {
	if err := s.requireRepo(s.db.WithContext(ctx), repoID); err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Exec(`UPDATE git_repos SET error = ?, error_at = now() WHERE id = ?`, scanErr.Error(), repoID).Error; err != nil {
		return fmt.Errorf("record error of git repo %s: %w", repoID, err)
	}
	return nil
}
