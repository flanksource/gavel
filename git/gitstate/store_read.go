package gitstate

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/status"
	"github.com/google/uuid"
)

type repoRow struct {
	ID             uuid.UUID
	RootDir        string
	BaseBranch     *string
	BaseSHA        *string
	CurrentBranch  *string
	BaseCheckedOut bool
	RefsScannedAt  *time.Time
	Generation     int64
	Error          *string
}

type worktreeRow struct {
	RepoID          uuid.UUID
	Path            string
	Branch          *string
	HeadSHA         *string
	IsPrimary       bool
	Detached        bool
	Prunable        bool
	LastCommitAt    *time.Time
	Staged          int
	Unstaged        int
	Both            int
	Untracked       int
	Conflict        int
	Adds            int
	Dels            int
	TouchedAt       *time.Time
	StatusScannedAt *time.Time
	StatusError     *string
}

type branchRow struct {
	RepoID        uuid.UUID
	Name          string
	HeadSHA       string
	WorktreePath  *string
	LastCommitAt  *time.Time
	Ahead, Behind int
	Commits       int
	Files         int
	Adds, Dels    int
}

// Load assembles the State of each repo in repoIDs. A repo that was never
// scanned successfully has no entry.
func (s *Store) Load(ctx context.Context, repoIDs []uuid.UUID) (map[uuid.UUID]State, error) {
	states := map[uuid.UUID]State{}
	if len(repoIDs) == 0 {
		return states, nil
	}
	db := s.db.WithContext(ctx)
	var repos []repoRow
	if err := db.Raw(`
		SELECT id, root_dir, base_branch, base_sha, current_branch, base_checked_out, refs_scanned_at, generation, error
		FROM git_repos WHERE id IN ? AND (refs_scanned_at IS NOT NULL OR error IS NOT NULL)`, repoIDs).Scan(&repos).Error; err != nil {
		return nil, fmt.Errorf("load git repos: %w", err)
	}
	var worktrees []worktreeRow
	if err := db.Raw(`
		SELECT repo_id, path, branch, head_sha, is_primary, detached, prunable, last_commit_at,
			staged, unstaged, "both", untracked, conflict, adds, dels, touched_at, status_scanned_at, status_error
		FROM git_worktrees WHERE repo_id IN ? AND removed_at IS NULL ORDER BY repo_id, ordinal`, repoIDs).Scan(&worktrees).Error; err != nil {
		return nil, fmt.Errorf("load git worktrees: %w", err)
	}
	var branches []branchRow
	if err := db.Raw(`
		SELECT b.repo_id, b.name, b.head_sha, b.worktree_path, b.last_commit_at,
			r.ahead, r.behind, r.commits, r.files, r.adds, r.dels
		FROM git_branches b
		JOIN git_repos g ON g.id = b.repo_id
		JOIN git_range_stats r ON r.repo_id = b.repo_id AND r.base_sha = g.base_sha AND r.head_sha = b.head_sha
		WHERE b.repo_id IN ? AND b.name <> g.base_branch AND r.ahead > 0
		ORDER BY b.repo_id, b.name`, repoIDs).Scan(&branches).Error; err != nil {
		return nil, fmt.Errorf("load git branches: %w", err)
	}

	for _, repo := range repos {
		states[repo.ID] = State{
			Base: deref(repo.BaseBranch), BaseSHA: deref(repo.BaseSHA),
			CurrentBranch: deref(repo.CurrentBranch), BaseCheckedOut: repo.BaseCheckedOut,
			Worktrees: []Worktree{}, Branches: []gavelgit.BranchInfo{},
			ComputedAt: utc(repo.RefsScannedAt), Generation: repo.Generation, Error: deref(repo.Error),
		}
	}
	ahead := map[uuid.UUID]map[string]int{}
	for _, b := range branches {
		state, ok := states[b.RepoID]
		if !ok {
			continue
		}
		state.Branches = append(state.Branches, gavelgit.BranchInfo{
			Name: b.Name, Head: b.HeadSHA, Ahead: b.Ahead, Behind: b.Behind, Worktree: deref(b.WorktreePath),
			Diff:         gavelgit.DiffStat{Commits: b.Commits, Files: b.Files, Adds: b.Adds, Dels: b.Dels},
			LastCommitAt: utc(b.LastCommitAt),
		})
		states[b.RepoID] = state
		if ahead[b.RepoID] == nil {
			ahead[b.RepoID] = map[string]int{}
		}
		ahead[b.RepoID][b.Name] = b.Ahead
	}
	for _, w := range worktrees {
		state, ok := states[w.RepoID]
		if !ok {
			continue
		}
		wt := Worktree{
			Worktree: gavelgit.Worktree{
				Path: w.Path, Branch: deref(w.Branch), Head: deref(w.HeadSHA),
				Primary: w.IsPrimary, Detached: w.Detached, Prunable: w.Prunable,
			},
			Changes: Changes{
				Staged: w.Staged, Unstaged: w.Unstaged, Both: w.Both, Untracked: w.Untracked,
				Conflict: w.Conflict, Adds: w.Adds, Dels: w.Dels,
			},
			Ahead: ahead[w.RepoID][deref(w.Branch)], LastCommitAt: utc(w.LastCommitAt),
			StatusError: deref(w.StatusError),
		}
		if w.TouchedAt != nil {
			touched := w.TouchedAt.UTC()
			wt.TouchedAt = &touched
		}
		if w.StatusScannedAt != nil {
			scanned := w.StatusScannedAt.UTC()
			wt.StatusScannedAt = &scanned
		}
		state.Worktrees = append(state.Worktrees, wt)
		states[w.RepoID] = state
	}
	return states, nil
}

// Generations returns every repo's generation.
func (s *Store) Generations(ctx context.Context) (map[uuid.UUID]int64, error) {
	var rows []struct {
		ID         uuid.UUID
		Generation int64
	}
	if err := s.db.WithContext(ctx).Raw(`SELECT id, generation FROM git_repos`).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("load git repo generations: %w", err)
	}
	generations := make(map[uuid.UUID]int64, len(rows))
	for _, row := range rows {
		generations[row.ID] = row.Generation
	}
	return generations, nil
}

// WorktreeFiles returns the uncommitted files the last status scan of the
// worktree recorded, and whether it was ever scanned.
func (s *Store) WorktreeFiles(ctx context.Context, repoID uuid.UUID, path string) ([]status.FileStatus, bool, error) {
	var row struct {
		Files           string
		StatusScannedAt *time.Time
	}
	result := s.db.WithContext(ctx).Raw(`SELECT files::text AS files, status_scanned_at FROM git_worktrees WHERE repo_id = ? AND path = ? AND removed_at IS NULL`,
		repoID, path).Scan(&row)
	if result.Error != nil {
		return nil, false, fmt.Errorf("load files of worktree %s: %w", path, result.Error)
	}
	if result.RowsAffected == 0 || row.StatusScannedAt == nil {
		return nil, false, nil
	}
	var files []status.FileStatus
	if err := json.Unmarshal([]byte(row.Files), &files); err != nil {
		return nil, false, fmt.Errorf("decode files of worktree %s: %w", path, err)
	}
	return files, true, nil
}

// BranchHead returns the tip of the local branch name as the last ref scan
// recorded it, and whether the repo has such a branch.
func (s *Store) BranchHead(ctx context.Context, repoID uuid.UUID, name string) (string, bool, error) {
	var heads []string
	if err := s.db.WithContext(ctx).Raw(`SELECT head_sha FROM git_branches WHERE repo_id = ? AND name = ?`,
		repoID, name).Scan(&heads).Error; err != nil {
		return "", false, fmt.Errorf("load head of branch %s: %w", name, err)
	}
	if len(heads) == 0 {
		return "", false, nil
	}
	return heads[0], true, nil
}

// Ranges returns the cached comparisons among keys; a key never computed has
// no entry.
func (s *Store) Ranges(ctx context.Context, repoID uuid.UUID, keys []RangeKey) (map[RangeKey]Range, error) {
	ranges := make(map[RangeKey]Range, len(keys))
	if len(keys) == 0 {
		return ranges, nil
	}
	heads := make([]string, 0, len(keys))
	wanted := make(map[RangeKey]bool, len(keys))
	for _, key := range keys {
		heads = append(heads, key.Head)
		wanted[key] = true
	}
	var rows []struct {
		BaseSHA, HeadSHA, MergeBase               string
		Ahead, Behind, Commits, Files, Adds, Dels int
	}
	if err := s.db.WithContext(ctx).Raw(`
		SELECT base_sha, head_sha, merge_base, ahead, behind, commits, files, adds, dels FROM git_range_stats
		WHERE repo_id = ? AND head_sha IN ?`, repoID, heads).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("load cached ranges of git repo %s: %w", repoID, err)
	}
	for _, row := range rows {
		key := RangeKey{Base: row.BaseSHA, Head: row.HeadSHA}
		if !wanted[key] {
			continue
		}
		ranges[key] = Range{RangeKey: key, RangeCompare: gavelgit.RangeCompare{
			MergeBase: row.MergeBase, Ahead: row.Ahead, Behind: row.Behind,
			Diff: gavelgit.DiffStat{Commits: row.Commits, Files: row.Files, Adds: row.Adds, Dels: row.Dels},
		}}
	}
	return ranges, nil
}

// Range returns the cached comparison of head against base, nil when it was
// never computed.
func (s *Store) Range(ctx context.Context, repoID uuid.UUID, key RangeKey) (*Range, error) {
	var row struct {
		MergeBase                                 string
		Ahead, Behind, Commits, Files, Adds, Dels int
	}
	result := s.db.WithContext(ctx).Raw(`
		SELECT merge_base, ahead, behind, commits, files, adds, dels FROM git_range_stats
		WHERE repo_id = ? AND base_sha = ? AND head_sha = ?`, repoID, key.Base, key.Head).Scan(&row)
	if result.Error != nil {
		return nil, fmt.Errorf("load range %s..%s: %w", key.Base, key.Head, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &Range{RangeKey: key, RangeCompare: gavelgit.RangeCompare{
		MergeBase: row.MergeBase, Ahead: row.Ahead, Behind: row.Behind,
		Diff: gavelgit.DiffStat{Commits: row.Commits, Files: row.Files, Adds: row.Adds, Dels: row.Dels},
	}}, nil
}

// RangeFiles returns the cached file list of a range, and whether one was
// stored.
func (s *Store) RangeFiles(ctx context.Context, repoID uuid.UUID, key RangeKey) ([]gavelgit.CommitFile, bool, error) {
	var blob *string
	if err := s.db.WithContext(ctx).Raw(`
		SELECT file_list::text FROM git_range_stats WHERE repo_id = ? AND base_sha = ? AND head_sha = ?`,
		repoID, key.Base, key.Head).Scan(&blob).Error; err != nil {
		return nil, false, fmt.Errorf("load files of range %s..%s: %w", key.Base, key.Head, err)
	}
	if blob == nil {
		return nil, false, nil
	}
	var files []gavelgit.CommitFile
	if err := json.Unmarshal([]byte(*blob), &files); err != nil {
		return nil, false, fmt.Errorf("decode files of range %s..%s: %w", key.Base, key.Head, err)
	}
	return files, true, nil
}

// SaveRangeFiles stores the file list of a range whose stats are cached.
func (s *Store) SaveRangeFiles(ctx context.Context, repoID uuid.UUID, key RangeKey, files []gavelgit.CommitFile) error {
	blob, err := json.Marshal(files)
	if err != nil {
		return fmt.Errorf("encode files of range %s..%s: %w", key.Base, key.Head, err)
	}
	result := s.db.WithContext(ctx).Exec(`
		UPDATE git_range_stats SET file_list = ?::jsonb WHERE repo_id = ? AND base_sha = ? AND head_sha = ? AND file_list IS NULL`,
		string(blob), repoID, key.Base, key.Head)
	if result.Error != nil {
		return fmt.Errorf("save files of range %s..%s: %w", key.Base, key.Head, result.Error)
	}
	return nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func utc(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.UTC()
}
