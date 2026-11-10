package runtime

import (
	"context"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type WorkspaceList struct {
	Items types.TODOS
	Err   error
}

// ListWorkspaces evaluates the joined overview once for the requested workspace set.
func ListWorkspaces(ctx context.Context, providers []*Provider) (map[uuid.UUID]WorkspaceList, error) {
	result := map[uuid.UUID]WorkspaceList{}
	if len(providers) == 0 {
		return result, nil
	}
	byID := map[uuid.UUID]*Provider{}
	var ids []uuid.UUID
	for _, provider := range providers {
		id := provider.workspace.ID
		if byID[id] != nil {
			continue
		}
		byID[id] = provider
		ids = append(ids, id)
	}
	if !providers[0].readOnly {
		if err := reconcileWorkspaceAnswers(ctx, byID); err != nil {
			return nil, err
		}
	}
	issues, err := providers[0].repository.ListWorkspaceIssues(ctx, ids)
	if err != nil {
		return nil, err
	}
	byWorkspace := map[uuid.UUID][]native.Issue{}
	for _, issue := range issues {
		byWorkspace[issue.WorkspaceID] = append(byWorkspace[issue.WorkspaceID], issue)
	}
	preparation, err := prepareWorkspaceLists(ctx, providers, ids, issues)
	if err != nil {
		return nil, err
	}
	var group errgroup.Group
	group.SetLimit(4)
	lists := make([]WorkspaceList, len(ids))
	for index, id := range ids {
		index, id := index, id
		group.Go(func() error {
			if preparation[id] != nil {
				lists[index] = WorkspaceList{Err: preparation[id]}
				return nil
			}
			items, readErr := byID[id].listIssues(ctx, todos.DiscoveryFilters{}, byWorkspace[id])
			lists[index] = WorkspaceList{Items: items, Err: readErr}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	for index, id := range ids {
		result[id] = lists[index]
	}
	return result, nil
}
