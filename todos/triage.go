package todos

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos/labels"
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
// then the comment, and last the retirements. Every write is guarded on
// todo_issues.version and each one refreshes it, so they cannot be reordered or
// batched. Retirement goes last because it is terminal — a soft-deleted TODO
// takes no further edits.
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
	if err := validateTriageLabels(todo, env, AllowedLabels(ctx, provider)); err != nil {
		return err
	}
	// Every TODO this verdict closes is resolved before anything is written: a fold
	// that turns out to name a missing or already-closed TODO must not leave the
	// survivor rewritten and the fold half-applied.
	retiring, err := resolveTriageRetirements(ctx, provider, todo, env)
	if err != nil {
		return err
	}

	if err := applyTriageContent(ctx, provider, todo, env); err != nil {
		return err
	}
	if err := applyTriageState(ctx, provider, todo, env, retiring); err != nil {
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
	return applyTriageRetirements(ctx, provider, retiring, env.Comment)
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

// validateTriageLabels holds the agent to the workspace's label vocabulary,
// before anything is written.
//
// The vocabulary is closed on purpose: a free-text label is how a backlog ends up
// filterable by neither `bug` nor `bugs`, and an agent with no list in front of it
// will coin a synonym every run. Removals are NOT checked — a TODO may carry a
// label nothing defines (the resolver colours those from a hash), and a removal
// naming a label the TODO does not have is a harmless no-op, not a mistake worth
// failing a whole triage over.
func validateTriageLabels(todo *types.TODO, env *types.TriageEnvelope, allowed []string) error {
	for _, label := range env.AddLabels {
		if labels.Contains(allowed, label) {
			continue
		}
		return fmt.Errorf("triage proposed label %q for %s, which is not in the workspace taxonomy; allowed labels: %s",
			strings.TrimSpace(label), triageRef(todo), strings.Join(allowed, ", "))
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
	// A label edit replaces the whole set, so the delta is merged against the
	// TODO's own labels here. An unchanged result is dropped rather than written:
	// every write consumes an optimistic-lock version, and "add a label it already
	// has" must not burn one.
	if env.ChangesLabels() {
		if next := labels.Apply(todo.Labels, env.AddLabels, env.RemoveLabels); !labels.Equal(next, todo.Labels) {
			edit.Labels = &next
		}
	}
	if edit.IsEmpty() {
		return nil
	}
	if err := provider.Edit(ctx, todo, edit); err != nil {
		return fmt.Errorf("apply triage content to %s: %w", triageRef(todo), err)
	}
	if edit.Labels != nil {
		todo.Labels = *edit.Labels
	}
	return nil
}

// applyTriageState writes the status and priority. The envelope has already been
// validated, so an invalid value cannot reach here — but a `done` verdict never
// writes a status even if the agent supplied one: whether the work is finished is
// decided by running the definition of done, not by the agent's opinion of it.
//
// A merge-into verdict raises the priority to the most urgent TODO it absorbs,
// whether or not the agent sent one. Absorbing work does not make it less urgent,
// and the agent judging the survivor's severity has not necessarily read every
// folded TODO's.
func applyTriageState(ctx context.Context, provider Provider, todo *types.TODO, env *types.TriageEnvelope, retiring []triageRetirement) error {
	var update StateUpdate
	if raw := strings.TrimSpace(env.Status); raw != "" && env.Verdict != types.VerdictDone {
		status := types.Status(raw)
		if err := types.ValidateAssignableStatus(status); err != nil {
			return fmt.Errorf("triage status for %s: %w", triageRef(todo), err)
		}
		update.Status = &status
	}
	priority := todo.Priority
	if raw := strings.TrimSpace(env.Priority); raw != "" {
		priority = types.Priority(raw)
		if err := types.ValidatePriority(priority); err != nil {
			return fmt.Errorf("triage priority for %s: %w", triageRef(todo), err)
		}
		update.Priority = &priority
	}
	if env.Verdict == types.VerdictMergeInto {
		if raised := retirementPriority(priority, retiring); raised != "" && raised != todo.Priority {
			update.Priority = &raised
		}
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
	if update.Priority != nil {
		todo.Priority = *update.Priority
	}
	return nil
}

// applyTriageLinks records the related references the agent found, as related_to:
// depends_on means "blocked until", which is a claim triage is not in a position
// to make.
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
// itself removed. An agent asked to name a related TODO will sometimes name the
// one it is looking at; linking a TODO to itself is noise, not an error worth
// failing the whole triage over.
//
// duplicateOf and merges are deliberately absent: those edges are written by the
// retirement path, which links the TODO being closed to its survivor. Writing
// them here too would attempt the same edge twice, and todo_issue_relationships
// is unique on (workspace, issue, target, relation).
func triageLinkTargets(todo *types.TODO, env *types.TriageEnvelope) []string {
	self := selfRefs(todo)
	seen := map[string]bool{}
	var targets []string
	for _, ref := range env.Related {
		ref = strings.TrimSpace(ref)
		if ref == "" || seen[ref] || self[strings.ToLower(ref)] {
			continue
		}
		seen[ref] = true
		targets = append(targets, ref)
	}
	return targets
}
