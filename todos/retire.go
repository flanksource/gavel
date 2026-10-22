package todos

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos/types"
)

// Retirement is one TODO being closed and, when its work lives on elsewhere,
// the survivor it went to.
type Retirement struct {
	// Retired is the TODO being closed.
	Retired *types.TODO
	// Survivor is where the work lives on. It is never Retired, and nil when the
	// work is being dropped rather than handed on.
	Survivor *types.TODO
	// Reason leads the comment recorded on Retired ("Duplicate of"). It is
	// unused without a Survivor to name.
	Reason string
	// Relation links Retired to Survivor.
	Relation types.RelationKind
	// Rationale is why the TODO is being closed, recorded on it before the close.
	Rationale string
}

// Retire closes one TODO: its open children settled, a comment naming where the
// work went and why, a link to the survivor, then the soft delete.
//
// The comment and the link land before the delete, so a failed delete leaves a
// TODO that still says where its work went rather than a silently orphaned one.
// Nobody can be asked what to do with the retired TODO's open children, so they
// follow the work to the survivor, or become top-level TODOs when there is no
// open one to take them. That happens first, so a failure leaves the TODO open
// rather than closed over children nobody can see.
func Retire(ctx context.Context, provider Provider, retirement Retirement) error {
	retired, survivor := retirement.Retired, retirement.Survivor
	children, err := SettleChildren(ctx, provider, retired, ArchiveOptions{Survivor: survivor, Children: ChildrenDetach})
	if err != nil {
		return fmt.Errorf("retire %s: %w", RefOf(retired), err)
	}
	if note := retirementNote(retirement, children); note != "" {
		if err := provider.Comment(ctx, retired, CommentRequest{Body: note}); err != nil {
			return fmt.Errorf("record why %s was retired: %w", RefOf(retired), err)
		}
	}
	if survivor != nil {
		if err := linkRetired(ctx, provider, retirement); err != nil {
			return err
		}
	}
	if err := provider.Delete(ctx, retired); err != nil {
		return fmt.Errorf("retire %s: %w", RefOf(retired), err)
	}
	return nil
}

// linkRetired links the retired TODO to its survivor. An earlier run may
// already have linked the pair; the edge is the same, so it is not re-written.
func linkRetired(ctx context.Context, provider Provider, retirement Retirement) error {
	retired, survivor := retirement.Retired, retirement.Survivor
	relationships, ok := provider.(RelationshipProvider)
	if !ok {
		return fmt.Errorf("retiring %s into %s needs issue links, which the TODO provider does not support",
			RefOf(retired), RefOf(survivor))
	}
	existing, err := relationships.Links(ctx, retired)
	if err != nil {
		return fmt.Errorf("read links of %s: %w", RefOf(retired), err)
	}
	if HasLink(existing, retirement.Relation, survivor.ID) {
		return nil
	}
	if _, err := relationships.Link(ctx, retired, RefOf(survivor), retirement.Relation); err != nil {
		return fmt.Errorf("link %s to %s: %w", RefOf(retired), RefOf(survivor), err)
	}
	return nil
}

// retirementNote is the comment left on a retired TODO: where its work went,
// why, and where its open children went.
func retirementNote(retirement Retirement, children ChildrenOutcome) string {
	var parts []string
	if survivor := retirement.Survivor; survivor != nil {
		note := fmt.Sprintf("%s %s", retirement.Reason, RefOf(survivor))
		if title := strings.TrimSpace(survivor.Title); title != "" {
			note += " — " + title
		}
		parts = append(parts, note+".")
	}
	if rationale := strings.TrimSpace(retirement.Rationale); rationale != "" {
		parts = append(parts, rationale)
	}
	if sentence := children.Sentence(); sentence != "" {
		parts = append(parts, sentence)
	}
	return strings.Join(parts, "\n\n")
}

// HasLink reports whether links already carry relation to ref, by short or full id.
func HasLink(links []Link, relation types.RelationKind, ref string) bool {
	for _, link := range links {
		if link.Relation == relation && (strings.EqualFold(link.TargetShortID, ref) || strings.EqualFold(link.TargetID, ref)) {
			return true
		}
	}
	return false
}

// RefOf names a TODO in a message. It prefers the short id because that is
// what the CLI accepts and what a person can go and look up.
func RefOf(todo *types.TODO) string {
	if todo != nil && strings.TrimSpace(todo.ShortID) != "" {
		return todo.ShortID
	}
	return TODOReference(todo)
}

// SelfRefs is the set of references that name one TODO, lowercased.
func SelfRefs(todo *types.TODO) map[string]bool {
	refs := map[string]bool{}
	if todo == nil {
		return refs
	}
	for _, ref := range []string{todo.ID, todo.ShortID} {
		if ref = strings.TrimSpace(ref); ref != "" {
			refs[strings.ToLower(ref)] = true
		}
	}
	return refs
}
