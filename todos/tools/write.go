package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/labels"
	"github.com/flanksource/gavel/todos/ops"
	"github.com/flanksource/gavel/todos/types"
)

func (t *toolset) todoTools() []api.ToolDefinition {
	defs := t.readTools()
	defs = append(defs, t.writeTools()...)
	return append(defs, t.closeTools()...)
}

func (t *toolset) writeTools() []api.ToolDefinition {
	allow := api.ToolPolicyAllow
	return []api.ToolDefinition{
		t.def("todo_update", "Rewrite a TODO's title, body, verification fixture, priority or status. Send only the fields that change.", allow,
			object(nil, map[string]any{
				"ref": refProp(), "title": str("Replacement title"), "body": str("Replacement description markdown"),
				"verification": str("Replacement ## Verification fixture markdown, without the heading"),
				"priority":     enum("Severity", string(types.PriorityHigh), string(types.PriorityMedium), string(types.PriorityLow)),
				"status":       enum("Directly assignable status", assignableStatuses()...),
			}), t.update),
		t.def("todo_labels", "Add labels from the workspace taxonomy to a TODO, or remove labels that no longer describe it.", allow,
			object(nil, map[string]any{"ref": refProp(), "add": strs("Labels to add, from the taxonomy only"), "remove": strs("Labels to remove")}), t.labels),
		t.def("todo_link", "Link a TODO to another TODO. Closes nothing.", allow,
			object([]string{"target", "relation"}, map[string]any{
				"ref": refProp(), "target": str("Short id of the other TODO"),
				"relation": enum("How the TODO relates to the target", string(types.RelationDuplicates), string(types.RelationRelatedTo),
					string(types.RelationDependsOn), string(types.RelationBlocks)),
			}), t.link),
		t.def("todo_link_pr", "Link a pull request whose diff or description addresses a TODO, as a related PR. The TODO's own PR is not changed.", allow,
			object([]string{"url"}, map[string]any{"ref": refProp(), "url": str("Full pull request URL")}), t.linkPR),
		t.def("todo_set_parent", "Place a TODO under a top-level parent TODO.", allow,
			object([]string{"parent"}, map[string]any{"ref": refProp(), "parent": str("Short id of the top-level parent TODO")}), t.setParent),
		t.def("todo_comment", "Record a comment on a TODO's history.", allow,
			object([]string{"body"}, map[string]any{"ref": refProp(), "body": str("Comment markdown")}), t.comment),
		t.def("todo_plan", "Save the complete implementation plan of the TODO this plan run is for.", allow,
			object([]string{"content"}, map[string]any{"content": str("The full plan markdown")}), t.plan),
	}
}

func assignableStatuses() []string {
	return []string{string(types.StatusDraft), string(types.StatusPending), string(types.StatusVerified),
		string(types.StatusCompleted), string(types.StatusSkipped)}
}

type updateInput struct {
	Ref          string  `json:"ref"`
	Title        *string `json:"title"`
	Body         *string `json:"body"`
	Verification *string `json:"verification"`
	Priority     string  `json:"priority"`
	Status       string  `json:"status"`
}

func (t *toolset) update(ctx context.Context, input map[string]any) (any, error) {
	var in updateInput
	if err := decode(input, &in); err != nil {
		return nil, err
	}
	todo, err := t.resolve(ctx, in.Ref)
	if err != nil {
		return nil, err
	}
	changes, err := ops.BuildEdit(ops.EditFlags{
		Title: in.Title, Body: in.Body, Verification: in.Verification, Priority: in.Priority, Status: in.Status,
	})
	if err != nil {
		return nil, err
	}
	if in.Verification != nil {
		if _, err := todos.ParseVerificationMarkdown(todos.VerificationMarkdownOptions{
			Name: "verification", Markdown: *in.Verification, SourceDir: t.opts.WorkDir,
		}); err != nil {
			return nil, fmt.Errorf("verification fixture for %s does not parse: %w", todos.RefOf(todo), err)
		}
	}
	if _, err := ops.ApplyEdit(ctx, t.opts.Provider, todo, changes); err != nil {
		return nil, fmt.Errorf("update %s: %w", todos.RefOf(todo), err)
	}
	t.opts.Run.record("todo_update", todo)
	return fmt.Sprintf("updated %s", todos.RefOf(todo)), nil
}

type labelsInput struct {
	Ref    string   `json:"ref"`
	Add    []string `json:"add"`
	Remove []string `json:"remove"`
}

func (t *toolset) labels(ctx context.Context, input map[string]any) (any, error) {
	var in labelsInput
	if err := decode(input, &in); err != nil {
		return nil, err
	}
	if err := t.checkLabelDelta(in.Add, in.Remove); err != nil {
		return nil, err
	}
	todo, err := t.resolve(ctx, in.Ref)
	if err != nil {
		return nil, err
	}
	next := labels.Apply(todo.Labels, in.Add, in.Remove)
	if labels.Equal(next, todo.Labels) {
		return fmt.Sprintf("%s already carries those labels", todos.RefOf(todo)), nil
	}
	if err := t.opts.Provider.Edit(ctx, todo, todos.EditRequest{Labels: &next}); err != nil {
		return nil, fmt.Errorf("label %s: %w", todos.RefOf(todo), err)
	}
	todo.Labels = next
	t.opts.Run.record("todo_labels", todo)
	return map[string]any{"id": todos.RefOf(todo), "labels": next}, nil
}

// checkLabelDelta holds a label delta to the taxonomy. Removing a namespaced
// label such as `source:todo` is refused: it records where the TODO came from,
// not what it is about.
func (t *toolset) checkLabelDelta(add, remove []string) error {
	if len(add) == 0 && len(remove) == 0 {
		return fmt.Errorf("todo_labels needs labels to add or remove")
	}
	for _, label := range append(append([]string(nil), add...), remove...) {
		if strings.TrimSpace(label) == "" {
			return fmt.Errorf("todo_labels was given a blank label")
		}
	}
	for _, label := range add {
		if labels.Contains(remove, label) {
			return fmt.Errorf("label %q is both added and removed", label)
		}
		if !labels.Contains(t.opts.Labels, label) {
			return fmt.Errorf("label %q is not in the workspace taxonomy; allowed labels: %s", label, strings.Join(t.opts.Labels, ", "))
		}
	}
	for _, label := range remove {
		if labels.Key(label) != "" {
			return fmt.Errorf("label %q is namespaced and records provenance; it cannot be removed", label)
		}
	}
	return nil
}

type linkInput struct {
	Ref      string `json:"ref"`
	Target   string `json:"target"`
	Relation string `json:"relation"`
}

func (t *toolset) link(ctx context.Context, input map[string]any) (any, error) {
	var in linkInput
	if err := decode(input, &in); err != nil {
		return nil, err
	}
	relation := types.RelationKind(in.Relation)
	switch relation {
	case types.RelationDuplicates, types.RelationRelatedTo, types.RelationDependsOn, types.RelationBlocks:
	default:
		return nil, fmt.Errorf("relation %q is not one of duplicates, related_to, depends_on, blocks", in.Relation)
	}
	todo, err := t.resolve(ctx, in.Ref)
	if err != nil {
		return nil, err
	}
	target, err := t.lookup(ctx, in.Target)
	if err != nil {
		return nil, err
	}
	if target.ID == todo.ID {
		return nil, fmt.Errorf("a TODO cannot be linked to itself")
	}
	relationships, ok := t.opts.Provider.(todos.RelationshipProvider)
	if !ok {
		return nil, fmt.Errorf("the TODO provider does not support links")
	}
	existing, err := relationships.Links(ctx, todo)
	if err != nil {
		return nil, fmt.Errorf("read links of %s: %w", todos.RefOf(todo), err)
	}
	if todos.HasLink(existing, relation, target.ID) {
		return fmt.Sprintf("%s already %s %s", todos.RefOf(todo), relation, todos.RefOf(target)), nil
	}
	if _, err := relationships.Link(ctx, todo, todos.RefOf(target), relation); err != nil {
		return nil, fmt.Errorf("link %s to %s: %w", todos.RefOf(todo), todos.RefOf(target), err)
	}
	t.opts.Run.record("todo_link", todo)
	return fmt.Sprintf("linked %s %s %s", todos.RefOf(todo), relation, todos.RefOf(target)), nil
}

type linkPRInput struct {
	Ref string `json:"ref"`
	URL string `json:"url"`
}

func (t *toolset) linkPR(ctx context.Context, input map[string]any) (any, error) {
	var in linkPRInput
	if err := decode(input, &in); err != nil {
		return nil, err
	}
	ref, err := types.PRRefFromURL(in.URL)
	if err != nil {
		return nil, err
	}
	todo, err := t.resolve(ctx, in.Ref)
	if err != nil {
		return nil, err
	}
	links, ok := t.opts.Provider.(todos.PRLinkProvider)
	if !ok {
		return nil, fmt.Errorf("the TODO provider does not support PR links")
	}
	existing, err := links.PRLinks(ctx, todo)
	if err != nil {
		return nil, fmt.Errorf("read PR links of %s: %w", todos.RefOf(todo), err)
	}
	for _, link := range existing {
		if link.SamePR(ref) {
			return fmt.Sprintf("%s already links %s", todos.RefOf(todo), ref.URL), nil
		}
	}
	request := todos.PRLinkRequest{URL: ref.URL, Relation: types.PRRelationRelated, Source: types.PRLinkSourceTriage}
	if _, err := links.LinkPR(ctx, todo, request); err != nil {
		return nil, fmt.Errorf("link %s to %s: %w", todos.RefOf(todo), ref.URL, err)
	}
	t.opts.Run.record("todo_link_pr", todo)
	return fmt.Sprintf("linked %s to %s", todos.RefOf(todo), ref.URL), nil
}

type setParentInput struct {
	Ref    string `json:"ref"`
	Parent string `json:"parent"`
}

func (t *toolset) setParent(ctx context.Context, input map[string]any) (any, error) {
	var in setParentInput
	if err := decode(input, &in); err != nil {
		return nil, err
	}
	todo, err := t.resolve(ctx, in.Ref)
	if err != nil {
		return nil, err
	}
	parent, err := t.resolve(ctx, in.Parent)
	if err != nil {
		return nil, err
	}
	if parent.ID == todo.ID || parent.ParentID != "" {
		return nil, fmt.Errorf("parent %s must be another top-level TODO", todos.RefOf(parent))
	}
	parents, ok := t.opts.Provider.(todos.ParentProvider)
	if !ok {
		return nil, fmt.Errorf("the TODO provider does not support parents")
	}
	if err := parents.SetParent(ctx, todo, todos.RefOf(parent)); err != nil {
		return nil, fmt.Errorf("place %s under %s: %w", todos.RefOf(todo), todos.RefOf(parent), err)
	}
	t.opts.Run.record("todo_set_parent", todo)
	return fmt.Sprintf("%s is now a child of %s", todos.RefOf(todo), todos.RefOf(parent)), nil
}

type commentInput struct {
	Ref  string `json:"ref"`
	Body string `json:"body"`
}

func (t *toolset) comment(ctx context.Context, input map[string]any) (any, error) {
	var in commentInput
	if err := decode(input, &in); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Body) == "" {
		return nil, fmt.Errorf("todo_comment needs a body")
	}
	todo, err := t.resolve(ctx, in.Ref)
	if err != nil {
		return nil, err
	}
	if err := t.opts.Provider.Comment(ctx, todo, todos.CommentRequest{Body: in.Body}); err != nil {
		return nil, fmt.Errorf("comment on %s: %w", todos.RefOf(todo), err)
	}
	t.opts.Run.record("todo_comment", todo)
	return fmt.Sprintf("commented on %s", todos.RefOf(todo)), nil
}

type planInput struct {
	Content string `json:"content"`
}

// plan records the run's plan. The attempt persists it as a plan revision when
// the run finishes, the way a native plan-mode file is persisted.
func (t *toolset) plan(_ context.Context, input map[string]any) (any, error) {
	var in planInput
	if err := decode(input, &in); err != nil {
		return nil, err
	}
	if !t.opts.PlanStep {
		return nil, fmt.Errorf("todo_plan saves the plan of a plan-mode step; this step implements")
	}
	if strings.TrimSpace(in.Content) == "" {
		return nil, fmt.Errorf("todo_plan needs the plan content")
	}
	t.opts.Run.mu.Lock()
	t.opts.Run.plan = in.Content
	t.opts.Run.mu.Unlock()
	t.opts.Run.record("todo_plan", t.opts.Subject)
	return "plan recorded; it is saved for review when this run finishes", nil
}
