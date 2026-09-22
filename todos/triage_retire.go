package todos

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos/types"
)

// triageRetirement is one TODO being closed and the survivor its work went to.
// Every verdict that closes a TODO reduces to a list of these, which is what lets
// them share one write path.
type triageRetirement struct {
	// Retired is the TODO being closed: a folded TODO for merge-into, the TODO
	// being triaged for duplicate-of and retire.
	Retired *types.TODO
	// Survivor is where the work lives on. It is never Retired, and it is nil for
	// a retire — the work is being dropped, not handed on.
	Survivor *types.TODO
	// Reason leads the comment recorded on Retired. It is unused without a
	// Survivor to name: a retire's rationale is the verdict's own comment.
	Reason string
}

// resolveTriageRetirements resolves every TODO a verdict closes, plus its
// survivor, before any write happens.
//
// Retirement is the one thing triage does that cannot be walked back, so each
// target is proved to exist, to be open, and not to be its own survivor while the
// whole verdict can still be rejected as a unit.
func resolveTriageRetirements(ctx context.Context, provider Provider, todo *types.TODO, env *types.TriageEnvelope) ([]triageRetirement, error) {
	switch env.Verdict {
	case types.VerdictMergeInto:
		return resolveFolds(ctx, provider, todo, env)
	case types.VerdictRetire:
		// Nothing to resolve: the work is being dropped, so there is no survivor to
		// prove exists. The rationale the verdict requires is already the comment
		// ApplyTriage writes before the close.
		return []triageRetirement{{Retired: todo}}, nil
	case types.VerdictDuplicateOf:
		survivor, err := resolveTriageTarget(ctx, provider, todo, env.DuplicateOf)
		if err != nil {
			return nil, err
		}
		// The survivor may be closed: "this duplicates work already finished" is a
		// normal verdict. Only the TODO being retired has to be open.
		return []triageRetirement{{Retired: todo, Survivor: survivor, Reason: "Duplicate of"}}, nil
	default:
		return nil, nil
	}
}

func resolveFolds(ctx context.Context, provider Provider, todo *types.TODO, env *types.TriageEnvelope) ([]triageRetirement, error) {
	refs := env.RetirementTargets()
	out := make([]triageRetirement, 0, len(refs))
	seen := map[string]bool{}
	for _, ref := range refs {
		folded, err := resolveTriageTarget(ctx, provider, todo, ref)
		if err != nil {
			return nil, err
		}
		if seen[folded.ID] {
			return nil, fmt.Errorf("triage of %s names %q twice among the TODOs it folds in", triageRef(todo), ref)
		}
		// A closed TODO is either finished work or one already folded elsewhere —
		// Delete soft-deletes to cancelled, which projects to completed. Folding it
		// again would re-close it and bury the first rationale under a second one.
		// This is also the guard against two TODOs in one batch each absorbing the
		// other: whichever verdict lands second is refused.
		if folded.Status == types.StatusCompleted {
			return nil, fmt.Errorf("triage of %s folds in %s, which is already closed", triageRef(todo), triageRef(folded))
		}
		seen[folded.ID] = true
		out = append(out, triageRetirement{Retired: folded, Survivor: todo, Reason: "Merged into"})
	}
	return out, nil
}

// resolveTriageTarget turns one agent-supplied reference into a TODO. An
// unresolvable reference is reported rather than skipped: a silently dropped fold
// is how two TODOs stay duplicated.
func resolveTriageTarget(ctx context.Context, provider Provider, todo *types.TODO, ref string) (*types.TODO, error) {
	ref = strings.TrimSpace(ref)
	if selfRefs(todo)[strings.ToLower(ref)] {
		return nil, fmt.Errorf("triage of %s names itself (%q) as the surviving TODO", triageRef(todo), ref)
	}
	target, err := provider.Get(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("triage of %s names %q, which could not be resolved: %w", triageRef(todo), ref, err)
	}
	if target == nil {
		return nil, fmt.Errorf("triage of %s names %q, which matched no TODO", triageRef(todo), ref)
	}
	if target.ID == todo.ID {
		return nil, fmt.Errorf("triage of %s names itself (%q) as the surviving TODO", triageRef(todo), ref)
	}
	return target, nil
}

// applyTriageRetirements closes what the verdict decided.
func applyTriageRetirements(ctx context.Context, provider Provider, retiring []triageRetirement, rationale string) error {
	for _, retirement := range retiring {
		if err := retireInto(ctx, provider, retirement, rationale); err != nil {
			return err
		}
	}
	return nil
}

// retireInto closes one TODO into its survivor: a comment naming where the work
// went, a related_to link to it, then the soft delete.
//
// The order is `todos/merge/apply.go` retire()'s, and for the same reason — the
// comment and the link must land before the delete, so a failed delete leaves a
// TODO that still says where its work went rather than a silently orphaned one.
// related_to is the only honest relation: depends_on means "blocked until", which
// is not what a duplicate is.
//
// A retirement with no survivor is just the delete: there is nowhere to point at,
// and ApplyTriage has already recorded the rationale that verdict requires.
func retireInto(ctx context.Context, provider Provider, retirement triageRetirement, rationale string) error {
	retired, survivor := retirement.Retired, retirement.Survivor
	if survivor == nil {
		if err := provider.Delete(ctx, retired); err != nil {
			return fmt.Errorf("retire %s: %w", triageRef(retired), err)
		}
		return nil
	}
	note := fmt.Sprintf("%s %s", retirement.Reason, triageRef(survivor))
	if title := strings.TrimSpace(survivor.Title); title != "" {
		note += " — " + title
	}
	note += "."
	if rationale = strings.TrimSpace(rationale); rationale != "" {
		note += "\n\n" + rationale
	}
	if err := provider.Comment(ctx, retired, note); err != nil {
		return fmt.Errorf("record why %s was retired: %w", triageRef(retired), err)
	}

	relationships, ok := provider.(RelationshipProvider)
	if !ok {
		return fmt.Errorf("triage retired %s into %s but the TODO provider does not support links",
			triageRef(retired), triageRef(survivor))
	}
	if _, err := relationships.Link(ctx, retired, triageRef(survivor), types.RelationRelatedTo); err != nil {
		return fmt.Errorf("link %s to %s: %w", triageRef(retired), triageRef(survivor), err)
	}

	if err := provider.Delete(ctx, retired); err != nil {
		return fmt.Errorf("retire %s: %w", triageRef(retired), err)
	}
	return nil
}

// retirementPriority is the priority a survivor keeps after absorbing others:
// never lower than the most urgent TODO folded into it. Merging work does not
// make it less urgent, and `gavel todos merge` applies the same rule through the
// same helper.
func retirementPriority(current types.Priority, retiring []triageRetirement) types.Priority {
	priorities := []types.Priority{current}
	for _, retirement := range retiring {
		priorities = append(priorities, retirement.Retired.Priority)
	}
	return types.HighestPriority(priorities...)
}

// selfRefs is the set of references that name one TODO, lowercased. Title is
// included because an agent asked for a short id will sometimes answer with the
// title it was reading.
func selfRefs(todo *types.TODO) map[string]bool {
	refs := map[string]bool{}
	if todo == nil {
		return refs
	}
	for _, ref := range []string{todo.ID, todo.ShortID, todo.Title} {
		if ref = strings.TrimSpace(ref); ref != "" {
			refs[strings.ToLower(ref)] = true
		}
	}
	return refs
}
