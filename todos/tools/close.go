package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// closeTools close TODOs, which no later run walks back, so each call is ask:
// the run's broker puts it to a person before the handler runs.
func (t *toolset) closeTools() []api.ToolDefinition {
	ask := api.ToolPolicyAsk
	return []api.ToolDefinition{
		t.def("todo_merge", "Fold source TODOs into a surviving TODO: the survivor is rewritten with the combined content, and every source is linked to it and closed. Requires approval.", ask,
			object([]string{"sources", "body"}, map[string]any{
				"into":         str("Short id of the surviving TODO; omit for the TODO this run is for"),
				"sources":      strs("Short ids of the TODOs folded in and closed"),
				"title":        str("Title of the merged TODO; omit to keep the survivor's"),
				"body":         str("Combined description carrying every acceptance criterion and attachment of every TODO merged"),
				"verification": str("Combined ## Verification fixture, required when any merged TODO has one"),
				"plan":         str("Combined plan, required when any merged TODO has one"),
				"rationale":    str("Why these TODOs are one piece of work"),
			}), t.merge),
		t.def("todo_close", "Close a TODO that is obsolete, or that duplicates another TODO entirely. Requires approval.", ask,
			object([]string{"reason"}, map[string]any{
				"ref":         refProp(),
				"reason":      str("Why the TODO is being closed, recorded on it"),
				"duplicateOf": str("Short id of the TODO that already covers this work; it is linked as the survivor"),
			}), t.close),
	}
}

type mergeInput struct {
	Into         string   `json:"into"`
	Sources      []string `json:"sources"`
	Title        string   `json:"title"`
	Body         string   `json:"body"`
	Verification string   `json:"verification"`
	Plan         string   `json:"plan"`
	Rationale    string   `json:"rationale"`
}

func (t *toolset) merge(ctx context.Context, input map[string]any) (any, error) {
	var in mergeInput
	if err := decode(input, &in); err != nil {
		return nil, err
	}
	merger, ok := t.opts.Provider.(todos.MergeProvider)
	if !ok {
		return nil, fmt.Errorf("the TODO provider does not support merges")
	}
	survivor, err := t.resolve(ctx, in.Into)
	if err != nil {
		return nil, err
	}
	retired, err := t.mergeSources(ctx, survivor, in.Sources)
	if err != nil {
		return nil, err
	}
	request := todos.MergeRequest{
		Survivor: survivor, Retired: retired, Title: strings.TrimSpace(in.Title), Body: in.Body,
		Verification: in.Verification, Plan: in.Plan, Rationale: in.Rationale, Preview: t.opts.Preview,
	}
	if request.Title == "" {
		request.Title = survivor.Title
	}
	if err := merger.MergeInto(ctx, request); err != nil {
		return nil, fmt.Errorf("merge into %s: %w", todos.RefOf(survivor), err)
	}
	refs := make([]string, len(retired))
	for i, todo := range retired {
		refs[i] = todos.RefOf(todo)
	}
	if t.opts.Preview {
		return fmt.Sprintf("preview: would merge %s into %s; nothing was written", strings.Join(refs, ", "), todos.RefOf(survivor)), nil
	}
	t.opts.Run.record("todo_merge", survivor)
	for _, todo := range retired {
		t.opts.Run.record("todo_merge", todo)
	}
	return fmt.Sprintf("merged %s into %s", strings.Join(refs, ", "), todos.RefOf(survivor)), nil
}

// mergeSources resolves every TODO a merge closes before anything is written,
// so a merge naming a missing, closed or repeated TODO is refused as a whole.
func (t *toolset) mergeSources(ctx context.Context, survivor *types.TODO, refs []string) ([]*types.TODO, error) {
	if len(refs) == 0 {
		return nil, fmt.Errorf("todo_merge needs at least one source TODO")
	}
	seen := map[string]bool{survivor.ID: true}
	retired := make([]*types.TODO, 0, len(refs))
	for _, ref := range refs {
		todo, err := t.resolve(ctx, ref)
		if err != nil {
			return nil, err
		}
		if seen[todo.ID] {
			return nil, fmt.Errorf("todo_merge names %s twice, or as both survivor and source", todos.RefOf(todo))
		}
		seen[todo.ID] = true
		retired = append(retired, todo)
	}
	return retired, nil
}

type closeInput struct {
	Ref         string `json:"ref"`
	Reason      string `json:"reason"`
	DuplicateOf string `json:"duplicateOf"`
}

func (t *toolset) close(ctx context.Context, input map[string]any) (any, error) {
	var in closeInput
	if err := decode(input, &in); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Reason) == "" {
		return nil, fmt.Errorf("todo_close needs a reason")
	}
	todo, err := t.resolve(ctx, in.Ref)
	if err != nil {
		return nil, err
	}
	retirement := todos.Retirement{Retired: todo, Rationale: in.Reason}
	if strings.TrimSpace(in.DuplicateOf) != "" {
		if retirement.Survivor, err = t.duplicateSurvivor(ctx, todo, in.DuplicateOf); err != nil {
			return nil, err
		}
		retirement.Reason, retirement.Relation = "Duplicate of", types.RelationDuplicates
	}
	if t.opts.Preview {
		return closePreview(retirement), nil
	}
	if err := t.carryPR(ctx, retirement); err != nil {
		return nil, err
	}
	if err := todos.Retire(ctx, t.opts.Provider, retirement); err != nil {
		return nil, err
	}
	t.opts.Run.record("todo_close", todo)
	return fmt.Sprintf("closed %s", todos.RefOf(todo)), nil
}

// duplicateSurvivor is the TODO a duplicate is closed into. It may itself be
// closed — "this duplicates finished work" is a normal finding — but it must
// be another TODO, and not one linked to a different PR.
func (t *toolset) duplicateSurvivor(ctx context.Context, todo *types.TODO, ref string) (*types.TODO, error) {
	survivor, err := t.lookup(ctx, ref)
	if err != nil {
		return nil, err
	}
	if survivor.ID == todo.ID {
		return nil, fmt.Errorf("a TODO cannot duplicate itself")
	}
	if _, err := types.MergedPR([]*types.TODO{survivor, todo}); err != nil {
		return nil, err
	}
	return survivor, nil
}

// carryPR keeps a closed duplicate's PR link on its survivor when the survivor
// has none, so closing it does not drop the PR the work is in.
func (t *toolset) carryPR(ctx context.Context, retirement todos.Retirement) error {
	survivor := retirement.Survivor
	if survivor == nil || survivor.PR != nil || retirement.Retired.PR == nil {
		return nil
	}
	pr := retirement.Retired.PR
	if err := t.opts.Provider.Edit(ctx, survivor, todos.EditRequest{PR: &pr}); err != nil {
		return fmt.Errorf("carry %s's PR to %s: %w", todos.RefOf(retirement.Retired), todos.RefOf(survivor), err)
	}
	t.opts.Run.record("todo_close", survivor)
	return nil
}

func closePreview(retirement todos.Retirement) string {
	line := fmt.Sprintf("preview: would close %s", todos.RefOf(retirement.Retired))
	if retirement.Survivor != nil {
		line += " as a duplicate of " + todos.RefOf(retirement.Survivor)
	}
	return line + "; nothing was written"
}
