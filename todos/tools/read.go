package tools

import (
	"context"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

func (t *toolset) readTools() []api.ToolDefinition {
	return []api.ToolDefinition{
		t.def("todo_search", "Search the open TODOs of this workspace by title or body. An empty query lists the whole open backlog.",
			api.ToolPolicyAllow, object(nil, map[string]any{"query": str("Words to match, case-insensitively")}), t.search),
		t.def("todo_get", "Read one TODO in full: description, acceptance criteria, verification fixture, plan, labels, status and parent.",
			api.ToolPolicyAllow, object(nil, map[string]any{"ref": refProp()}), t.get),
	}
}

type searchInput struct {
	Query string `json:"query"`
}

func (t *toolset) search(ctx context.Context, input map[string]any) (any, error) {
	var in searchInput
	if err := decode(input, &in); err != nil {
		return nil, err
	}
	list, err := t.opts.Provider.List(ctx, todos.DiscoveryFilters{})
	if err != nil {
		return nil, err
	}
	query := strings.ToLower(strings.TrimSpace(in.Query))
	found := []map[string]any{}
	for _, todo := range list {
		if todo.ID == t.opts.Subject.ID || closed(todo) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(todo.Title+"\n"+todo.MarkdownBody), query) {
			continue
		}
		found = append(found, map[string]any{
			"id": todo.ShortID, "title": todo.Title, "status": todo.Status, "labels": todo.Labels, "parentId": todo.ParentID,
		})
	}
	return found, nil
}

type getInput struct {
	Ref string `json:"ref"`
}

func (t *toolset) get(ctx context.Context, input map[string]any) (any, error) {
	var in getInput
	if err := decode(input, &in); err != nil {
		return nil, err
	}
	todo, err := t.lookup(ctx, in.Ref)
	if err != nil {
		return nil, err
	}
	plan := ""
	if plans, ok := t.opts.Provider.(todos.PlanContentProvider); ok {
		if plan, err = plans.PlanMarkdown(ctx, todo, types.ModePlan); err != nil {
			return nil, err
		}
	}
	return map[string]any{
		"id": todo.ShortID, "title": todo.Title, "body": todo.MarkdownBody, "verification": todo.VerificationMarkdown,
		"criteria": todo.AcceptanceCriteria, "plan": plan, "labels": todo.Labels, "status": todo.Status,
		"priority": todo.Priority, "parentId": todo.ParentID, "pr": todo.PR,
	}, nil
}
