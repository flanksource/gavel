package bulk

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/labels"
	"github.com/flanksource/gavel/todos/types"
)

// StatusFlags are the parameters of the `status` bulk action.
type StatusFlags struct {
	To      string `flag:"to" help:"Status to assign" enum:"draft,pending,verified,completed,skipped" required:"true"`
	Comment string `flag:"comment" help:"Comment recorded on each TODO alongside the change"`
}

func (StatusFlags) ClickyActionFlags() {}

// SetStatus assigns one status to every selected TODO. Only assignable statuses
// are accepted: the rest are projected from the last run, so writing one would
// claim something the run history contradicts.
func SetStatus(flags StatusFlags) (ItemFunc, error) {
	status := types.Status(strings.ToLower(strings.TrimSpace(flags.To)))
	if err := types.ValidateAssignableStatus(status); err != nil {
		return nil, err
	}
	comment := strings.TrimSpace(flags.Comment)
	return func(ctx context.Context, provider todos.Provider, todo *types.TODO) (ItemResult, error) {
		if err := provider.UpdateState(ctx, todo, todos.StateUpdate{Status: &status}); err != nil {
			return ItemResult{}, err
		}
		if comment != "" {
			if err := provider.Comment(ctx, todo, comment); err != nil {
				return ItemResult{}, err
			}
		}
		return ItemResult{Status: string(status)}, nil
	}, nil
}

// PriorityFlags are the parameters of the `priority` bulk action.
type PriorityFlags struct {
	To string `flag:"to" help:"Severity to assign" enum:"high,medium,low" required:"true"`
}

func (PriorityFlags) ClickyActionFlags() {}

func SetPriority(flags PriorityFlags) (ItemFunc, error) {
	priority := types.Priority(strings.ToLower(strings.TrimSpace(flags.To)))
	if err := types.ValidatePriority(priority); err != nil {
		return nil, err
	}
	return func(ctx context.Context, provider todos.Provider, todo *types.TODO) (ItemResult, error) {
		if err := provider.UpdateState(ctx, todo, todos.StateUpdate{Priority: &priority}); err != nil {
			return ItemResult{}, err
		}
		return ItemResult{Status: string(priority)}, nil
	}, nil
}

// LabelFlags are the parameters of the `labels` bulk action.
type LabelFlags struct {
	Add    []string `flag:"add" help:"Labels to add"`
	Remove []string `flag:"remove" help:"Labels to remove"`
}

func (LabelFlags) ClickyActionFlags() {}

// EditLabels adds and removes labels per TODO.
//
// It is a read-modify-write per item rather than one write, because the
// provider's Labels edit replaces the whole set — see labels.Apply, which is the
// same merge triage performs on a single TODO.
func EditLabels(flags LabelFlags) (ItemFunc, error) {
	add := labels.Split(flags.Add)
	remove := labels.Split(flags.Remove)
	if len(add) == 0 && len(remove) == 0 {
		return nil, fmt.Errorf("at least one of --add or --remove is required")
	}
	for _, label := range add {
		if labels.Contains(remove, label) {
			return nil, fmt.Errorf("label %q is both added and removed", label)
		}
	}
	return func(ctx context.Context, provider todos.Provider, todo *types.TODO) (ItemResult, error) {
		next := labels.Apply(todo.Labels, add, remove)
		if err := provider.Edit(ctx, todo, todos.EditRequest{Labels: &next}); err != nil {
			return ItemResult{}, err
		}
		return ItemResult{Status: strings.Join(next, ", ")}, nil
	}, nil
}

// CommentFlags are the parameters of the `comment` bulk action.
type CommentFlags struct {
	Body string `flag:"body" help:"Comment to append to each TODO" required:"true"`
}

func (CommentFlags) ClickyActionFlags() {}

func AddComment(flags CommentFlags) (ItemFunc, error) {
	body := strings.TrimSpace(flags.Body)
	if body == "" {
		return nil, fmt.Errorf("comment body is required")
	}
	return func(ctx context.Context, provider todos.Provider, todo *types.TODO) (ItemResult, error) {
		if err := provider.Comment(ctx, todo, body); err != nil {
			return ItemResult{}, err
		}
		return ItemResult{}, nil
	}, nil
}

// DeleteFlags are the parameters of the `delete` bulk action.
type DeleteFlags struct {
	// Confirm is required because a filter-mode delete acts on a selection the
	// caller never enumerated. A UI can gate this behind its own prompt; a
	// script has to say it out loud.
	Confirm bool `flag:"confirm" help:"Required: confirm the deletion" required:"true"`
}

func (DeleteFlags) ClickyActionFlags() {}

func Delete(flags DeleteFlags) (ItemFunc, error) {
	if !flags.Confirm {
		return nil, fmt.Errorf("refusing to delete without --confirm")
	}
	return func(ctx context.Context, provider todos.Provider, todo *types.TODO) (ItemResult, error) {
		if err := provider.Delete(ctx, todo); err != nil {
			return ItemResult{}, err
		}
		return ItemResult{Status: "deleted"}, nil
	}, nil
}
