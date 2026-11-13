// Serves GET /api/todos?exclude=<filter> without a dir: one flat list of the todos of
// every workspace in the organization except those the excluded filters hold.
package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"slices"

	"github.com/flanksource/gavel/filters"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// orgTodoLister lists every workspace's todos from the database.
type orgTodoLister interface {
	ListOrg(ctx context.Context, filters todos.DiscoveryFilters, excludeRootPaths []string) (types.TODOS, error)
}

// handleOrgTodosList reads only the database: no workspace is prepared and no git range
// is compared, so a workspace's state on disk cannot fail the list. Each item carries
// its workspace's dir, and the diff footprint workspaceData has prepared for it.
func (s *Server) handleOrgTodosList(w http.ResponseWriter, r *http.Request) {
	excluded, err := s.excludedValues(r, filters.KindWorkspace)
	if err != nil {
		writeTodoError(w, filterErrorStatus(err), err)
		return
	}
	todoFilters, err := todoFiltersFromRequest(r)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	resp, err := s.orgTodoList(r.Context(), todoFilters, slices.Sorted(maps.Keys(excluded)))
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

func (s *Server) orgTodoList(ctx context.Context, todoFilters todos.DiscoveryFilters, excluded []string) (todoListResponse, error) {
	roots, err := s.excludedWorkspaceRoots(excluded)
	if err != nil {
		return todoListResponse{}, err
	}
	global, err := s.globalTodoProvider(ctx)
	if err != nil {
		return todoListResponse{}, err
	}
	lister, ok := global.(orgTodoLister)
	if !ok {
		return todoListResponse{}, fmt.Errorf("todo provider %T cannot list an organization's todos", global)
	}
	items, err := lister.ListOrg(ctx, todoFilters, roots)
	if err != nil {
		return todoListResponse{}, err
	}
	diffs := s.preparedDiffStats()
	resp := todoListResponse{Counts: summarizeTodos(items), Items: make([]todoSummary, 0, len(items))}
	for _, item := range items {
		sum := summarizeTodo(item, false)
		if sum.Dir, err = s.workspaceDir(item.CWD); err != nil {
			return todoListResponse{}, err
		}
		sum.Diff = diffStatFor(diffs[item.WorkspaceID], item.ID)
		resp.Items = append(resp.Items, sum)
	}
	return resp, nil
}

// excludedWorkspaceRoots turns workspace filter values into the workspace root paths
// the database keys workspaces by: a hosted repo:<id> becomes its checkout directory.
func (s *Server) excludedWorkspaceRoots(values []string) ([]string, error) {
	if s.workspaces == nil {
		return values, nil
	}
	roots := make([]string, 0, len(values))
	for _, value := range values {
		id, ok := s.workspaces.repoID(s.orgID(), value)
		if !ok {
			return nil, fmt.Errorf("excluded workspace %q %w: name it repo:<id>", value, errHostedWorkspace)
		}
		roots = append(roots, s.workspaces.rootPath(s.orgID(), id))
	}
	return roots, nil
}

// workspaceDir names the workspace at root as the projects list does: repo:<id> for a
// hosted repository's checkout, the root itself otherwise.
func (s *Server) workspaceDir(root string) (string, error) {
	if s.workspaces == nil {
		return root, nil
	}
	id, ok := s.workspaces.repoID(s.orgID(), root)
	if !ok {
		return "", fmt.Errorf("todo workspace root %q %w", root, errHostedWorkspace)
	}
	return RepoWorkspace(id), nil
}

// preparedDiffStats is each workspace's prepared diff footprints keyed by workspace id.
// A workspace whose run ranges workspaceData hasn't compared yet has none: its items
// list without a diff until the background preparation stores one.
func (s *Server) preparedDiffStats() map[string]map[string]gavelgit.DiffStat {
	stats := map[string]map[string]gavelgit.DiffStat{}
	if s.workspaceData == nil {
		return stats
	}
	s.workspaceData.mu.RLock()
	defer s.workspaceData.mu.RUnlock()
	for id, workspace := range s.workspaceData.diffs {
		stats[id.String()] = workspace
	}
	return stats
}
