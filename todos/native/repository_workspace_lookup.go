package native

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// WorkspaceLookup is a read-only snapshot of the workspaces a set of paths and
// repository keys resolve to, taken in one query by LookupWorkspaces. It answers
// the same two questions GetWorkspaceByPath and GetWorkspaceByRepoKey do — with
// the same normalization and the same not-found errors — so a caller resolving
// many projects at once pays one round trip instead of one per key.
//
// It only knows the keys it was built for: asking about any other key is a
// caller bug, and is reported as one rather than as a workspace that does not
// exist.
type WorkspaceLookup struct {
	paths     map[string]*Workspace
	repoKeys  map[string]*Workspace
	requested map[string]bool
}

type workspaceLookupRow struct {
	ID          uuid.UUID
	RepoKey     string
	RootPath    string
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	MatchedBy   string
	MatchedKey  string
}

// LookupWorkspaces resolves every path (current or retained) and repository key
// to its workspace in a single query.
func (r *Repository) LookupWorkspaces(ctx context.Context, paths, repoKeys []string) (*WorkspaceLookup, error) {
	lookup := &WorkspaceLookup{
		paths: map[string]*Workspace{}, repoKeys: map[string]*Workspace{}, requested: map[string]bool{},
	}
	normalizedPaths := lookup.request("path:", paths, normalizeWorkspacePath)
	normalizedKeys := lookup.request("repo:", repoKeys, normalizeToken)
	if len(lookup.requested) == 0 {
		return lookup, nil
	}
	var rows []workspaceLookupRow
	// IN with an empty list expands to IN (NULL), which matches nothing, so
	// either arm may be empty.
	err := r.db.WithContext(ctx).Raw(`
		SELECT `+workspaceColumns+`, 'path' AS matched_by, matched.path AS matched_key
		FROM `+workspaceFrom+`
		JOIN todo_workspace_paths AS matched ON matched.workspace_id = workspace.id
		WHERE matched.path IN ?
		UNION ALL
		SELECT `+workspaceColumns+`, 'repo' AS matched_by, workspace.repo_key AS matched_key
		FROM `+workspaceFrom+`
		WHERE workspace.repo_key IN ?`,
		normalizedPaths, normalizedKeys,
	).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("look up workspaces by %d paths and %d repository keys: %w", len(normalizedPaths), len(normalizedKeys), err)
	}
	for _, row := range rows {
		workspace := workspaceRecord{
			ID: row.ID, RepoKey: row.RepoKey, RootPath: row.RootPath, DisplayName: row.DisplayName,
			CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}.workspace()
		switch row.MatchedBy {
		case "path":
			lookup.paths[row.MatchedKey] = workspace
		case "repo":
			lookup.repoKeys[row.MatchedKey] = workspace
		default:
			return nil, fmt.Errorf("workspace lookup matched workspace %s by unknown key kind %q", row.ID, row.MatchedBy)
		}
	}
	return lookup, nil
}

// request normalizes keys the way the single-key lookups do, records them as
// asked for under kind, and returns the non-empty ones.
func (l *WorkspaceLookup) request(kind string, keys []string, normalize func(string) string) []string {
	normalized := make([]string, 0, len(keys))
	for _, key := range keys {
		if key = normalize(key); key != "" {
			normalized = append(normalized, key)
			l.requested[kind+key] = true
		}
	}
	return normalized
}

// GetWorkspaceByPath mirrors Repository.GetWorkspaceByPath over the snapshot.
func (l *WorkspaceLookup) GetWorkspaceByPath(_ context.Context, path string) (*Workspace, error) {
	path = normalizeWorkspacePath(path)
	if path == "" {
		return nil, fmt.Errorf("%w: workspace path is required", ErrInvalidInput)
	}
	return l.get("path:"+path, l.paths[path], fmt.Sprintf("workspace path %q", path))
}

// GetWorkspaceByRepoKey mirrors Repository.GetWorkspaceByRepoKey over the
// snapshot.
func (l *WorkspaceLookup) GetWorkspaceByRepoKey(_ context.Context, repoKey string) (*Workspace, error) {
	repoKey = normalizeToken(repoKey)
	if repoKey == "" {
		return nil, fmt.Errorf("%w: workspace repo key is required", ErrInvalidInput)
	}
	return l.get("repo:"+repoKey, l.repoKeys[repoKey], fmt.Sprintf("workspace repo key %q", repoKey))
}

func (l *WorkspaceLookup) get(key string, workspace *Workspace, subject string) (*Workspace, error) {
	if !l.requested[key] {
		return nil, fmt.Errorf("%w: %s was not part of this workspace lookup", ErrInvalidInput, subject)
	}
	if workspace == nil {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, subject)
	}
	copied := *workspace
	return &copied, nil
}
