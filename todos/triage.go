package todos

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos/types"
)

// TriageOptions configures ApplyTriage.
type TriageOptions struct {
	// WorkDir is the directory a rewritten fixture's relative references resolve
	// against while it is being validated.
	WorkDir string
}

// triageRef names a TODO in an error. It prefers the short id because triage
// runs in bulk: a failure among fifty is only actionable if the message says
// which TODO to go and look at, and a full UUID is not what the CLI accepts.
func triageRef(todo *types.TODO) string {
	if todo != nil && strings.TrimSpace(todo.ShortID) != "" {
		return todo.ShortID
	}
	return TODOReference(todo)
}

// ApplyTriage writes a triage run's proposed edits onto the TODO.
//
// A triage agent is read-only: it reports what should change and gavel performs
// the writes. That is what makes the edits validatable — a status storage would
// decline, or a fixture that does not parse, fails here with a message naming the
// TODO, instead of being applied piecemeal by an agent that cannot see the
// optimistic-lock version.
//
// The write order mirrors `gavel todos edit`: content, then state, then links,
// then the comment. Every write is guarded on todo_issues.version and each one
// refreshes it, so they cannot be reordered or batched.
func ApplyTriage(ctx context.Context, provider Provider, todo *types.TODO, env *types.TriageEnvelope, opts TriageOptions) error {
	if env == nil {
		return fmt.Errorf("triage run for %s finished without a verdict", triageRef(todo))
	}
	if env.EndStatus != types.EndCompleted {
		// An ask/failed session reported no verdict to honour; the run's own status
		// transition already covers it.
		return nil
	}
	if err := validateTriageFixture(todo, env, opts.WorkDir); err != nil {
		return err
	}

	if err := applyTriageContent(ctx, provider, todo, env); err != nil {
		return err
	}
	if err := applyTriageState(ctx, provider, todo, env); err != nil {
		return err
	}
	if err := applyTriageLinks(ctx, provider, todo, env); err != nil {
		return err
	}
	if comment := strings.TrimSpace(env.Comment); comment != "" {
		if err := provider.Comment(ctx, todo, comment); err != nil {
			return fmt.Errorf("record triage rationale on %s: %w", triageRef(todo), err)
		}
	}
	return nil
}

// validateTriageFixture rejects a rewritten fixture that does not parse, before
// anything is written. A fixture is the only thing that can prove a TODO done, so
// storing one that cannot even be read would replace a working definition of done
// with a silent one — the failure would not surface until someone ran the check.
func validateTriageFixture(todo *types.TODO, env *types.TriageEnvelope, workDir string) error {
	fixture := strings.TrimSpace(env.Verification)
	if fixture == "" {
		return nil
	}
	_, err := ParseVerificationMarkdown(VerificationMarkdownOptions{
		Name:      "verification",
		Markdown:  fixture,
		SourceDir: workDir,
	})
	if err != nil {
		return fmt.Errorf("triage produced an unparseable verification fixture for %s: %w", triageRef(todo), err)
	}
	return nil
}

func applyTriageContent(ctx context.Context, provider Provider, todo *types.TODO, env *types.TriageEnvelope) error {
	var edit EditRequest
	if title := strings.TrimSpace(env.Title); title != "" {
		edit.Title = &title
	}
	if body := strings.TrimSpace(env.Body); body != "" {
		edit.Body = &body
	}
	if fixture := strings.TrimSpace(env.Verification); fixture != "" {
		edit.Verification = &fixture
	}
	if edit.IsEmpty() {
		return nil
	}
	if err := provider.Edit(ctx, todo, edit); err != nil {
		return fmt.Errorf("apply triage content to %s: %w", triageRef(todo), err)
	}
	return nil
}

// applyTriageState writes the status and priority. The envelope has already been
// validated, so an invalid value cannot reach here — but a `done` verdict never
// writes a status even if the agent supplied one: whether the work is finished is
// decided by running the definition of done, not by the agent's opinion of it.
func applyTriageState(ctx context.Context, provider Provider, todo *types.TODO, env *types.TriageEnvelope) error {
	var update StateUpdate
	if raw := strings.TrimSpace(env.Status); raw != "" && env.Verdict != types.VerdictDone {
		status := types.Status(raw)
		if err := types.ValidateAssignableStatus(status); err != nil {
			return fmt.Errorf("triage status for %s: %w", triageRef(todo), err)
		}
		update.Status = &status
	}
	if raw := strings.TrimSpace(env.Priority); raw != "" {
		priority := types.Priority(raw)
		if err := types.ValidatePriority(priority); err != nil {
			return fmt.Errorf("triage priority for %s: %w", triageRef(todo), err)
		}
		update.Priority = &priority
	}
	if update.Status == nil && update.Priority == nil {
		return nil
	}
	if err := provider.UpdateState(ctx, todo, update); err != nil {
		return fmt.Errorf("apply triage state to %s: %w", triageRef(todo), err)
	}
	if update.Status != nil {
		todo.Status = *update.Status
	}
	return nil
}

// applyTriageLinks records the duplicate and related references the agent found.
// Both are related_to: depends_on means "blocked until", which is a claim triage
// is not in a position to make.
//
// A link to a TODO that cannot be resolved is reported rather than skipped — a
// silently dropped duplicate link is how two TODOs stay duplicated.
func applyTriageLinks(ctx context.Context, provider Provider, todo *types.TODO, env *types.TriageEnvelope) error {
	relationships, ok := provider.(RelationshipProvider)
	targets := triageLinkTargets(todo, env)
	if len(targets) == 0 {
		return nil
	}
	if !ok {
		return fmt.Errorf("triage linked %s to %s but the TODO provider does not support links",
			triageRef(todo), strings.Join(targets, ", "))
	}
	for _, target := range targets {
		if _, err := relationships.Link(ctx, todo, target, types.RelationRelatedTo); err != nil {
			return fmt.Errorf("link %s to %s: %w", triageRef(todo), target, err)
		}
	}
	return nil
}

// triageLinkTargets collects the link references, deduplicated and with the TODO
// itself removed. An agent asked to name the survivor of a duplicate pair will
// sometimes name the one it is looking at; linking a TODO to itself is noise, not
// an error worth failing the whole triage over.
func triageLinkTargets(todo *types.TODO, env *types.TriageEnvelope) []string {
	self := map[string]bool{}
	if todo != nil {
		for _, ref := range []string{todo.ID, todo.ShortID, todo.Title} {
			if ref = strings.TrimSpace(ref); ref != "" {
				self[ref] = true
			}
		}
	}
	seen := map[string]bool{}
	var targets []string
	for _, ref := range append([]string{env.DuplicateOf}, env.Related...) {
		ref = strings.TrimSpace(ref)
		if ref == "" || seen[ref] || self[ref] {
			continue
		}
		seen[ref] = true
		targets = append(targets, ref)
	}
	return targets
}
