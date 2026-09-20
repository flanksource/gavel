package merge

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/bulk"
	"github.com/flanksource/gavel/todos/types"
)

// mergeActor is recorded on the plan revision a merge saves.
const mergeActor = "merge"

// Run is the whole operation: order the selection, ask the model, validate the
// proposal, and write it. It returns the proposal alongside the result so a
// caller can render what was merged — a dry run returns exactly the same
// proposal, having written nothing.
func Run(ctx context.Context, targets []bulk.Target, opts Options) (bulk.Result, *Proposal, error) {
	todoList := make([]*types.TODO, 0, len(targets))
	for _, target := range targets {
		todoList = append(todoList, target.Todo)
	}
	survivor, retired, err := Targets(todoList, opts.Into)
	if err != nil {
		return bulk.Result{}, nil, err
	}
	if opts.Plans == nil {
		opts.Plans = Plan(providerOf(targets, survivor))
	}

	proposal, err := Propose(ctx, survivor, retired, opts)
	if err != nil {
		return bulk.Result{}, nil, err
	}
	selection, err := proposal.Validate(survivor, retired, hasFixture(todoList), hasPlan(ctx, opts, todoList), opts.WorkDir)
	if err != nil {
		return bulk.Result{}, nil, err
	}
	if opts.DryRun {
		return dryRunResult(targets, selection), proposal, nil
	}
	result := Apply(ctx, targets, selection, proposal, opts)
	return result, proposal, nil
}

// Apply writes the merge.
//
// The retirements land first. A failure halfway must never leave a survivor
// claiming work that is still open elsewhere: retire, then rewrite. Within one
// retirement the comment and the link come before the soft delete, so the
// rationale and the pointer to the survivor are recorded even if the delete
// fails.
//
// Per-item failures live in the Result and the error stays nil — see bulk.Result.
func Apply(ctx context.Context, targets []bulk.Target, selection *Selection, proposal *Proposal, opts Options) bulk.Result {
	result := bulk.Result{Action: "merge", Results: make([]bulk.ItemResult, 0, len(targets))}
	survivorRef := Ref(selection.Survivor)

	for _, todo := range selection.Retired {
		item := itemFor(targets, todo)
		if err := retire(ctx, providerOf(targets, todo), todo, selection.Survivor, proposal); err != nil {
			item.Error = err.Error()
		} else {
			item.Status = "merged into " + survivorRef
		}
		add(&result, item)
	}

	item := itemFor(targets, selection.Survivor)
	if err := rewrite(ctx, providerOf(targets, selection.Survivor), selection, proposal); err != nil {
		item.Error = err.Error()
	} else {
		item.Title = proposal.Title
		item.Status = fmt.Sprintf("merged %d todos in", len(selection.Retired)+1)
	}
	add(&result, item)

	for _, todo := range selection.Excluded {
		excluded := itemFor(targets, todo)
		excluded.Status = "excluded: not the same work"
		add(&result, excluded)
	}
	return result
}

// retire records why this TODO is part of the merged one, links it to the
// survivor, and soft-deletes it. Provider.Delete transitions the issue to
// cancelled and keeps its history — the TODO is recoverable, its comments and
// runs intact.
func retire(ctx context.Context, provider todos.Provider, todo, survivor *types.TODO, proposal *Proposal) error {
	comment := fmt.Sprintf("Merged into %s — %s.", Ref(survivor), proposal.Title)
	if rationale := strings.TrimSpace(proposal.Rationale); rationale != "" {
		comment += " " + rationale
	}
	if err := provider.Comment(ctx, todo, comment); err != nil {
		return fmt.Errorf("record the merge rationale on %s: %w", Ref(todo), err)
	}
	relationships, ok := provider.(todos.RelationshipProvider)
	if !ok {
		return fmt.Errorf("merging %s into %s needs issue links, which this TODO provider does not support",
			Ref(todo), Ref(survivor))
	}
	if _, err := relationships.Link(ctx, todo, Ref(survivor), types.RelationRelatedTo); err != nil {
		return fmt.Errorf("link %s to %s: %w", Ref(todo), Ref(survivor), err)
	}
	if err := provider.Delete(ctx, todo); err != nil {
		return fmt.Errorf("retire %s: %w", Ref(todo), err)
	}
	return nil
}

// rewrite replaces the survivor's content with the merged content: the body and
// the fixture in one edit, then the plan, the priority, and the summary comment.
// Each write is guarded on the issue version and refreshes it, so they cannot be
// batched or reordered.
func rewrite(ctx context.Context, provider todos.Provider, selection *Selection, proposal *Proposal) error {
	survivor := selection.Survivor
	title := strings.TrimSpace(proposal.Title)
	body := strings.TrimSpace(proposal.Body)
	labels := mergedLabels(selection, proposal)
	edit := todos.EditRequest{Title: &title, Body: &body, Labels: &labels}
	if fixture := strings.TrimSpace(proposal.Verification); fixture != "" {
		edit.Verification = &fixture
	}
	if err := provider.Edit(ctx, survivor, edit); err != nil {
		return fmt.Errorf("apply the merged content to %s: %w", Ref(survivor), err)
	}

	if plan := strings.TrimSpace(proposal.Plan); plan != "" {
		revisions, ok := provider.(todos.PlanRevisionProvider)
		if !ok {
			return fmt.Errorf("the merge produced a plan for %s but this TODO provider does not store plans", Ref(survivor))
		}
		// Saved unapproved on purpose: the merge changed the scope the previous
		// approval was given for, so the plan goes back through review.
		updated, err := revisions.SavePlanRevision(ctx, survivor, plan, mergeActor)
		if err != nil {
			return fmt.Errorf("save the merged plan on %s: %w", Ref(survivor), err)
		}
		if updated != nil {
			*survivor = *updated
		}
	}

	priority := mergedPriority(selection, proposal)
	if priority != "" && priority != survivor.Priority {
		if err := provider.UpdateState(ctx, survivor, todos.StateUpdate{Status: nil, Priority: &priority}); err != nil {
			return fmt.Errorf("apply the merged severity to %s: %w", Ref(survivor), err)
		}
	}

	comment := strings.TrimSpace(proposal.Summary) + "\n\nMerged in: " + strings.Join(refs(selection.Retired), ", ")
	if err := provider.Comment(ctx, survivor, comment); err != nil {
		return fmt.Errorf("record the merge summary on %s: %w", Ref(survivor), err)
	}
	return nil
}

// mergedLabels is the model's set when it proposed one, otherwise the union of
// the sources' labels. A merged TODO that silently lost the labels its sources
// were filtered by would disappear from the boards that tracked it.
func mergedLabels(selection *Selection, proposal *Proposal) []string {
	if len(proposal.Labels) > 0 {
		return normalize(proposal.Labels)
	}
	var all []string
	for _, todo := range append([]*types.TODO{selection.Survivor}, selection.Retired...) {
		all = append(all, todo.Labels...)
	}
	return normalize(all)
}

// mergedPriority is the model's severity, else the highest among the sources:
// merging never lowers how urgent the work was.
func mergedPriority(selection *Selection, proposal *Proposal) types.Priority {
	if raw := strings.TrimSpace(proposal.Priority); raw != "" {
		return types.Priority(raw)
	}
	rank := map[types.Priority]int{types.PriorityLow: 1, types.PriorityMedium: 2, types.PriorityHigh: 3}
	highest := selection.Survivor.Priority
	for _, todo := range selection.Retired {
		if rank[todo.Priority] > rank[highest] {
			highest = todo.Priority
		}
	}
	return highest
}

func dryRunResult(targets []bulk.Target, selection *Selection) bulk.Result {
	result := bulk.Result{Action: "merge", Results: make([]bulk.ItemResult, 0, len(targets))}
	survivor := itemFor(targets, selection.Survivor)
	survivor.Status = "dry-run: would be merged into"
	add(&result, survivor)
	for _, todo := range selection.Retired {
		item := itemFor(targets, todo)
		item.Status = "dry-run: would be retired into " + Ref(selection.Survivor)
		add(&result, item)
	}
	for _, todo := range selection.Excluded {
		item := itemFor(targets, todo)
		item.Status = "excluded: not the same work"
		add(&result, item)
	}
	return result
}

func add(result *bulk.Result, item bulk.ItemResult) {
	if item.Error != "" {
		result.Failed++
	} else {
		result.Applied++
	}
	result.Results = append(result.Results, item)
}

func itemFor(targets []bulk.Target, todo *types.TODO) bulk.ItemResult {
	item := bulk.ItemResult{Ref: Ref(todo)}
	if todo != nil {
		item.Title = todo.Title
		item.Dir = todo.CWD
	}
	for _, target := range targets {
		if target.Todo == todo && strings.TrimSpace(target.Ref) != "" {
			item.Ref = target.Ref
		}
	}
	return item
}

func providerOf(targets []bulk.Target, todo *types.TODO) todos.Provider {
	for _, target := range targets {
		if target.Todo == todo {
			return target.Provider
		}
	}
	return nil
}

func hasFixture(todoList []*types.TODO) bool {
	return fixtureSource(todoList) != nil
}

func hasPlan(ctx context.Context, opts Options, todoList []*types.TODO) bool {
	if opts.Plans == nil {
		return false
	}
	for _, todo := range todoList {
		plan, err := opts.Plans(ctx, todo)
		if err == nil && strings.TrimSpace(plan) != "" {
			return true
		}
	}
	return false
}

func refs(todoList []*types.TODO) []string {
	out := make([]string, 0, len(todoList))
	for _, todo := range todoList {
		out = append(out, Ref(todo))
	}
	return out
}

func normalize(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		duplicate := false
		for _, existing := range out {
			if strings.EqualFold(existing, value) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			out = append(out, value)
		}
	}
	return out
}
