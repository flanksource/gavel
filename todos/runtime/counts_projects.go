package runtime

import (
	"context"
	"errors"

	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// workspaceFinder is what resolveWorkspace needs to find a project's
// workspace: the repository for one project, a native.WorkspaceLookup snapshot
// for many.
type workspaceFinder interface {
	GetWorkspaceByPath(ctx context.Context, path string) (*native.Workspace, error)
	GetWorkspaceByRepoKey(ctx context.Context, repoKey string) (*native.Workspace, error)
}

// ProjectStatusCounts is one configured project's share of CountProjects: its
// TODO counts by derived status, or why they could not be read.
type ProjectStatusCounts struct {
	Counts map[types.Status]int
	Err    error
}

// CountProjects counts every configured project's TODOs by derived status —
// exactly what each project's own CountByStatus reports — in a fixed number of
// queries however many projects there are.
//
// Unlike Open it is read-only with respect to workspaces: a project whose
// workspace does not exist yet is not initialized, and reports empty counts,
// which is what counting the workspace Open would have created reports. A
// workspace whose registration drifted from the project (display name, primary
// root, repository key) is counted as it stands and reconciled on the next Open.
//
// It returns an error only when nothing could be counted (no database, a schema
// older than this binary, or a failed shared query). A project that cannot be
// resolved — invalid options, or a path and repository that name different
// workspaces — reports that on its own entry.
func CountProjects(ctx context.Context, projects []WorkspaceOptions) ([]ProjectStatusCounts, error) {
	db, err := database.Require(ctx, "native TODO storage")
	if err != nil {
		return nil, err
	}
	if err := requireVerificationColumn(ctx, db); err != nil {
		return nil, err
	}
	return countProjects(ctx, db, projects)
}

// countProjects is CountProjects over an already opened pool: one workspace
// lookup, then countWorkspaces over every workspace the projects resolve to.
func countProjects(ctx context.Context, db *gorm.DB, projects []WorkspaceOptions) ([]ProjectStatusCounts, error) {
	global, err := NewGlobal(db)
	if err != nil {
		return nil, err
	}
	out := make([]ProjectStatusCounts, len(projects))
	normalized := make([]WorkspaceOptions, len(projects))
	var paths, repoKeys []string
	for i, options := range projects {
		if normalized[i], out[i].Err = normalizeWorkspaceOptions(options); out[i].Err == nil {
			paths = append(paths, normalized[i].RootPath)
			repoKeys = append(repoKeys, normalized[i].Repositories...)
		}
	}
	lookup, err := global.repository.LookupWorkspaces(ctx, paths, repoKeys)
	if err != nil {
		return nil, err
	}
	providers := map[uuid.UUID]*Provider{}
	resolved := make([]uuid.UUID, len(projects))
	for i, options := range normalized {
		if out[i].Err != nil {
			continue
		}
		workspace, _, err := resolveWorkspace(ctx, lookup, options)
		switch {
		case errors.Is(err, native.ErrNotFound):
			out[i].Counts = map[types.Status]int{}
		case err != nil:
			out[i].Err = err
		default:
			resolved[i] = workspace.ID
			if providers[workspace.ID] == nil {
				providers[workspace.ID] = global.withWorkspace(workspace, options.RootPath)
			}
		}
	}
	counts, err := countWorkspaces(ctx, providers)
	if err != nil {
		return nil, err
	}
	for i, id := range resolved {
		if id != uuid.Nil {
			out[i].Counts = counts[id]
		}
	}
	return out, nil
}
