package todos

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos/types"
)

// applyTriageLinks records the TODOs the agent found: related as related_to and
// duplicates as duplicates. Neither closes anything. depends_on means "blocked
// until", which is a claim triage is not in a position to make.
//
// A link to a TODO that cannot be resolved is reported rather than skipped — a
// silently dropped duplicate link is how two TODOs stay duplicated. A link the
// TODO already carries is skipped: re-triaging a TODO re-finds the same
// neighbours, and the storage refuses a second identical edge.
func applyTriageLinks(ctx context.Context, provider Provider, todo *types.TODO, env *types.TriageEnvelope) error {
	related, duplicates := triageLinkTargets(todo, env.Related), triageLinkTargets(todo, env.Duplicates)
	if len(related) == 0 && len(duplicates) == 0 {
		return nil
	}
	relationships, ok := provider.(RelationshipProvider)
	if !ok {
		return fmt.Errorf("triage linked %s to %s but the TODO provider does not support links",
			triageRef(todo), strings.Join(append(related, duplicates...), ", "))
	}
	existing, err := relationships.Links(ctx, todo)
	if err != nil {
		return fmt.Errorf("read links of %s: %w", triageRef(todo), err)
	}
	for _, batch := range []struct {
		relation types.RelationKind
		targets  []string
	}{{types.RelationRelatedTo, related}, {types.RelationDuplicates, duplicates}} {
		for _, target := range batch.targets {
			if hasLink(existing, batch.relation, target) {
				continue
			}
			if _, err := relationships.Link(ctx, todo, target, batch.relation); err != nil {
				return fmt.Errorf("link %s to %s as %s: %w", triageRef(todo), target, batch.relation, err)
			}
		}
	}
	return nil
}

func hasLink(links []Link, relation types.RelationKind, ref string) bool {
	for _, link := range links {
		if link.Relation == relation && (strings.EqualFold(link.TargetShortID, ref) || strings.EqualFold(link.TargetID, ref)) {
			return true
		}
	}
	return false
}

// triageLinkTargets deduplicates refs case-insensitively and drops the TODO
// itself. An agent asked to name a related TODO will sometimes name the one it is
// looking at; linking a TODO to itself is noise, not an error worth failing the
// whole triage over.
//
// duplicateOf and merges are deliberately absent: those edges are written by the
// retirement path, which links the TODO being closed to its survivor, and the
// envelope refuses a duplicates entry that repeats them.
func triageLinkTargets(todo *types.TODO, refs []string) []string {
	self := selfRefs(todo)
	seen := map[string]bool{}
	var targets []string
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		key := strings.ToLower(ref)
		if ref == "" || seen[key] || self[key] {
			continue
		}
		seen[key] = true
		targets = append(targets, ref)
	}
	return targets
}

// applyTriagePRLinks links the PRs the agent found as related, triage-sourced
// links. It never touches todo.PR: the primary PR is an explicit link a human or
// the PR-feedback flow made (prompt rule 6), and triage only adds to the list
// around it. A PR the TODO already links — under any relation — is skipped.
func applyTriagePRLinks(ctx context.Context, provider Provider, todo *types.TODO, env *types.TriageEnvelope) error {
	if len(env.PRs) == 0 {
		return nil
	}
	links, ok := provider.(PRLinkProvider)
	if !ok {
		return fmt.Errorf("triage linked %s to PRs %s but the TODO provider does not support PR links",
			triageRef(todo), strings.Join(env.PRs, ", "))
	}
	existing, err := links.PRLinks(ctx, todo)
	if err != nil {
		return fmt.Errorf("read PR links of %s: %w", triageRef(todo), err)
	}
	linked := make([]types.PRRef, 0, len(existing)+len(env.PRs))
	for _, link := range existing {
		linked = append(linked, link.PRRef)
	}
	for _, raw := range env.PRs {
		ref, err := types.PRRefFromURL(raw)
		if err != nil {
			return fmt.Errorf("triage PR link for %s: %w", triageRef(todo), err)
		}
		if containsPR(linked, ref) {
			continue
		}
		request := PRLinkRequest{URL: ref.URL, Relation: types.PRRelationRelated, Source: types.PRLinkSourceTriage}
		if _, err := links.LinkPR(ctx, todo, request); err != nil {
			return fmt.Errorf("link %s to %s: %w", triageRef(todo), ref.URL, err)
		}
		linked = append(linked, ref)
	}
	return nil
}

func containsPR(refs []types.PRRef, ref types.PRRef) bool {
	for _, candidate := range refs {
		if candidate.SamePR(ref) {
			return true
		}
	}
	return false
}
