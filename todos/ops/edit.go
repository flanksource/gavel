// Package ops holds the TODO write operations that are not per-item loops: the
// validation and splitting a create or an edit needs before it touches a
// provider.
//
// It is a package rather than CLI-private code because the clicky entity has to
// call exactly this logic. An `edit` that reached the dashboard through a
// second, hand-written path would be the drift the entity exists to remove —
// and it was: the CLI validated statuses the API did not.
package ops

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/labels"
	"github.com/flanksource/gavel/todos/types"
)

// EditFlags is the already-resolved (file references expanded) input to an
// edit. Content pointers are nil when the caller did not name that field, which
// is what distinguishes "leave it alone" from "set it to empty".
type EditFlags struct {
	Title        *string
	Body         *string
	Plan         *string
	Verification *string
	Status       string
	Priority     string
	// Labels is nil when no labels were named. ClearLabels is --clear-labels;
	// the two are mutually exclusive because "replace with nothing" and "replace
	// with this list" cannot both be meant.
	Labels      *[]string
	ClearLabels bool
}

// EditChanges is an edit split by the provider call that applies it. They are
// separate writes — content, plan revision, state — and the order matters,
// because the first one refreshes the optimistic-lock version the rest reuse.
type EditChanges struct {
	Content todos.EditRequest
	Plan    *string
	State   todos.StateUpdate
}

// BuildEdit splits edit flags into content, plan, and state updates. It rejects
// statuses storage will not persist so callers see a failure rather than a
// silently declined write.
func BuildEdit(flags EditFlags) (EditChanges, error) {
	var changes EditChanges

	if flags.Title != nil {
		title := strings.TrimSpace(*flags.Title)
		if title == "" {
			return changes, fmt.Errorf("--title cannot be empty")
		}
		changes.Content.Title = &title
	}
	if flags.Body != nil {
		changes.Content.Body = flags.Body
	}
	if flags.Plan != nil {
		plan := strings.TrimSpace(*flags.Plan)
		if plan == "" {
			return changes, fmt.Errorf("--plan cannot be empty")
		}
		changes.Plan = &plan
	}
	if flags.Verification != nil {
		changes.Content.Verification = flags.Verification
	}
	if flags.ClearLabels && flags.Labels != nil {
		return changes, fmt.Errorf("--clear-labels cannot be combined with --label")
	}
	if flags.ClearLabels {
		changes.Content.Labels = &[]string{}
	} else if flags.Labels != nil {
		labelSet := make([]string, 0, len(*flags.Labels))
		for _, label := range *flags.Labels {
			if label = labels.Normalize(label); label != "" {
				labelSet = append(labelSet, label)
			}
		}
		changes.Content.Labels = &labelSet
	}
	if raw := strings.TrimSpace(flags.Status); raw != "" {
		status := types.Status(raw)
		if err := types.ValidateAssignableStatus(status); err != nil {
			return changes, err
		}
		changes.State.Status = &status
	}
	if raw := strings.TrimSpace(flags.Priority); raw != "" {
		priority := types.Priority(raw)
		if err := types.ValidatePriority(priority); err != nil {
			return changes, err
		}
		changes.State.Priority = &priority
	}

	if changes.Content.IsEmpty() && changes.Plan == nil && changes.State.Status == nil && changes.State.Priority == nil {
		return changes, fmt.Errorf("nothing to edit: provide --title, --body, --plan, --verification, --status, --priority, --label, and/or --clear-labels")
	}
	return changes, nil
}

// ApplyEdit performs the three writes in the order the optimistic lock
// requires and returns the TODO the last of them left behind.
func ApplyEdit(ctx context.Context, provider todos.Provider, todo *types.TODO, changes EditChanges) (*types.TODO, error) {
	var planRevisions todos.PlanRevisionProvider
	if changes.Plan != nil {
		var ok bool
		planRevisions, ok = provider.(todos.PlanRevisionProvider)
		if !ok {
			return nil, fmt.Errorf("TODO provider does not support plan revisions")
		}
	}
	// Content first: Edit refreshes the TODO's optimistic-lock version, which
	// the subsequent plan and state updates then reuse.
	if !changes.Content.IsEmpty() {
		if err := provider.Edit(ctx, todo, changes.Content); err != nil {
			return nil, err
		}
	}
	if changes.Plan != nil {
		revised, err := planRevisions.SavePlanRevision(ctx, todo, *changes.Plan, "")
		if err != nil {
			return nil, err
		}
		if revised == nil {
			return nil, fmt.Errorf("plan revision provider returned no TODO")
		}
		todo = revised
	}
	if changes.State.Status != nil || changes.State.Priority != nil {
		if err := provider.UpdateState(ctx, todo, changes.State); err != nil {
			return nil, err
		}
	}
	return todo, nil
}

// ParsePriority validates a severity, defaulting to medium when unset.
func ParsePriority(raw string) (types.Priority, error) {
	priority := types.Priority(strings.TrimSpace(raw))
	if priority == "" {
		return types.PriorityMedium, nil
	}
	switch priority {
	case types.PriorityHigh, types.PriorityMedium, types.PriorityLow:
		return priority, nil
	default:
		return "", fmt.Errorf("invalid --priority %q: expected high, medium, or low", raw)
	}
}

// CreateLifecycle is the starting lifecycle position of a new TODO. `approved`
// is not a status — it is pending with the plan already blessed — so it is
// parsed into the two facts storage actually holds.
type CreateLifecycle struct {
	Status       types.Status
	PlanApproved bool
}

// ParseLifecycle resolves the --status of a create.
func ParseLifecycle(raw string) (CreateLifecycle, error) {
	if strings.EqualFold(strings.TrimSpace(raw), "approved") {
		return CreateLifecycle{Status: types.StatusPending, PlanApproved: true}, nil
	}
	status := types.Status(strings.TrimSpace(raw))
	if status == "" {
		return CreateLifecycle{Status: types.StatusPending}, nil
	}
	if !types.IsKnownStatus(status) {
		known := make([]string, 0, len(types.KnownStatuses())+1)
		for _, candidate := range types.KnownStatuses() {
			known = append(known, string(candidate))
		}
		known = append(known, "approved")
		return CreateLifecycle{}, fmt.Errorf("invalid --status %q: expected %s", raw, strings.Join(known, ", "))
	}
	return CreateLifecycle{Status: status}, nil
}

// ValidateCreatePlan refuses an approved plan that does not exist.
func ValidateCreatePlan(plan string, lifecycle CreateLifecycle) error {
	if lifecycle.PlanApproved && strings.TrimSpace(plan) == "" {
		return fmt.Errorf("--status approved requires --plan")
	}
	return nil
}
