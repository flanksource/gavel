package triagenew

import (
	"context"
	"fmt"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"strings"
)

func (r *Review) search(ctx context.Context, input map[string]any) (any, error) {
	query, _ := input["query"].(string)
	list, err := r.Provider.List(ctx, todos.DiscoveryFilters{})
	if err != nil {
		return nil, err
	}
	var found []map[string]any
	for _, todo := range list {
		if todo.ID == r.SourceID || todo.Status == types.StatusCompleted || todo.Status == types.StatusSkipped || todo.Status == types.StatusVerified {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(todo.Title+"\n"+todo.MarkdownBody), strings.ToLower(query)) {
			continue
		}
		found = append(found, map[string]any{"id": todo.ID, "title": todo.Title, "status": todo.Status, "labels": todo.Labels, "parentId": todo.ParentID})
	}
	return found, nil
}

func (r *Review) get(ctx context.Context, input map[string]any) (any, error) {
	ref, ok := input["ref"].(string)
	if !ok || strings.TrimSpace(ref) == "" {
		return nil, fmt.Errorf("ref is required")
	}
	todo, err := r.Provider.Get(ctx, ref)
	if err != nil {
		return nil, err
	}
	plan := ""
	if plans, ok := r.Provider.(todos.PlanContentProvider); ok {
		plan, err = plans.PlanMarkdown(ctx, todo, types.ModePlan)
		if err != nil {
			return nil, err
		}
	}
	return map[string]any{"id": todo.ID, "version": todo.Version, "title": todo.Title, "body": todo.MarkdownBody, "verification": todo.VerificationMarkdown, "criteria": todo.AcceptanceCriteria, "plan": plan, "labels": todo.Labels, "status": todo.Status, "parentId": todo.ParentID}, nil
}
