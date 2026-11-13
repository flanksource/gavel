package runtime

import (
	"context"
	"fmt"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

// ListOrg lists the todos of every workspace in the provider's organization whose
// root path is not in excludeRootPaths, as the context's user sees them. It reads
// the database only: nothing is reconciled and no workspace is prepared on disk.
// Each todo's CWD is its workspace's root path.
func (p *Provider) ListOrg(ctx context.Context, filters todos.DiscoveryFilters, excludeRootPaths []string) (types.TODOS, error) {
	workspaces, err := p.repository.ListOrgWorkspaces(ctx, excludeRootPaths)
	if err != nil {
		return nil, err
	}
	if len(workspaces) == 0 {
		return types.TODOS{}, nil
	}
	providers := make([]*Provider, 0, len(workspaces))
	ids := make([]uuid.UUID, 0, len(workspaces))
	for index := range workspaces {
		provider := p.withWorkspace(&workspaces[index], workspaces[index].RootPath)
		provider.readOnly = true
		providers = append(providers, provider)
		ids = append(ids, workspaces[index].ID)
	}
	issues, err := p.repository.ListWorkspaceIssues(ctx, ids)
	if err != nil {
		return nil, err
	}
	failures, err := prepareWorkspaceLists(ctx, providers, ids, issues)
	if err != nil {
		return nil, err
	}
	byWorkspace := map[uuid.UUID][]native.Issue{}
	for _, issue := range issues {
		byWorkspace[issue.WorkspaceID] = append(byWorkspace[issue.WorkspaceID], issue)
	}
	result := types.TODOS{}
	for _, provider := range providers {
		id := provider.workspace.ID
		if failure := failures[id]; failure != nil {
			return nil, fmt.Errorf("read the phase runs of workspace %s: %w", provider.workspace.RootPath, failure)
		}
		items, err := provider.listIssues(ctx, filters, byWorkspace[id])
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	}
	result.Sort()
	return result, nil
}
