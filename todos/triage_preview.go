package todos

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos/types"
)

// TriagePreview is what a triage verdict would have done to the backlog, for a
// verdict that closes TODOs and was held back instead of applied.
//
// retire, merge-into and duplicate-of are the verdicts that close a TODO, and
// they are decided by an agent from a one-line backlog excerpt. A preview is the
// tier between "trust it" and "find out afterwards".
type TriagePreview struct {
	Verdict types.TriageVerdict
	// Survivor is the ref of the TODO the work lives on in. It is empty for a
	// retire, which hands the work to nobody.
	Survivor string
	// Retires is one line per TODO that would be closed.
	Retires []TriagePreviewRetirement
	// Withheld names the fields that would have been written to the triaged TODO.
	// A preview is all-or-nothing: applying the merged body while leaving the folded
	// TODOs open would duplicate the content it was meant to combine.
	Withheld []string
}

// TriagePreviewRetirement is one TODO a verdict would close.
type TriagePreviewRetirement struct {
	Ref    string
	Title  string
	Reason string
}

// PreviewTriage resolves a closing verdict and reports what it would do, writing
// nothing.
//
// It runs the same resolution ApplyTriage does, so a preview fails on exactly
// what an apply would fail on — an unresolvable ref, a self-fold, an
// already-closed target. A preview that silently accepted what the real thing
// would refuse would be worse than no preview.
func PreviewTriage(ctx context.Context, provider Provider, todo *types.TODO, env *types.TriageEnvelope, opts TriageOptions) (*TriagePreview, error) {
	if env == nil {
		return nil, fmt.Errorf("triage run for %s finished without a verdict", triageRef(todo))
	}
	if env.EndStatus != types.EndCompleted || !env.RetiresTODOs() {
		return nil, nil
	}
	if err := validateTriageFixture(todo, env, opts.WorkDir); err != nil {
		return nil, err
	}
	if err := validateTriageLabels(todo, env, AllowedLabels(ctx, provider)); err != nil {
		return nil, err
	}
	retiring, err := resolveTriageRetirements(ctx, provider, todo, env)
	if err != nil {
		return nil, err
	}

	preview := &TriagePreview{Verdict: env.Verdict, Withheld: TriageEnvelopeFields(env)}
	for _, retirement := range retiring {
		if retirement.Survivor != nil {
			preview.Survivor = triageRef(retirement.Survivor)
		}
		preview.Retires = append(preview.Retires, TriagePreviewRetirement{
			Ref:    triageRef(retirement.Retired),
			Title:  strings.TrimSpace(retirement.Retired.Title),
			Reason: retirement.Reason,
		})
	}
	return preview, nil
}

// TriageEnvelopeFields names the fields the envelope asks gavel to write to the
// triaged TODO. A preview reports them as what did not land; an applied verdict
// reports them as what did — one list, so the two can never disagree.
func TriageEnvelopeFields(env *types.TriageEnvelope) []string {
	var fields []string
	for _, field := range []struct{ name, value string }{
		{"title", env.Title}, {"body", env.Body}, {"verification", env.Verification},
		{"priority", env.Priority}, {"status", env.Status}, {"comment", env.Comment},
	} {
		if strings.TrimSpace(field.value) != "" {
			fields = append(fields, field.name)
		}
	}
	if env.ChangesLabels() {
		fields = append(fields, "labels")
	}
	if len(env.Related) > 0 {
		fields = append(fields, "related")
	}
	return fields
}

// RenderTriageApplied is the counterpart report for a verdict that was applied:
// which fields were written and which TODOs were closed. Without it a triage run
// that rewrote a body and closed two issues logged nothing at all, while the
// preview of the same verdict logged everything.
func RenderTriageApplied(todo *types.TODO, env *types.TriageEnvelope) string {
	if env == nil {
		return ""
	}
	line := fmt.Sprintf("triage verdict %q applied to %s", env.Verdict, triageRef(todo))
	if fields := TriageEnvelopeFields(env); len(fields) > 0 {
		line += ": " + strings.Join(fields, ", ")
	}
	closed := env.RetirementTargets()
	if env.ClosesSelf() {
		// retire and duplicate-of close the TODO being triaged, not another one.
		closed = []string{triageRef(todo)}
	}
	if len(closed) > 0 {
		line += " — closed " + strings.Join(closed, ", ")
	}
	return line
}

// String renders the preview as the run's report. It names the command that
// applies it, because a preview a reader cannot act on just moves the decision.
func (p *TriagePreview) String() string {
	if p == nil {
		return ""
	}
	var out strings.Builder
	fmt.Fprintf(&out, "triage verdict %q was previewed, not applied — nothing was written.\n", p.Verdict)
	for _, retirement := range p.Retires {
		line := fmt.Sprintf("  would close %s", retirement.Ref)
		if retirement.Title != "" {
			line += fmt.Sprintf(" (%s)", retirement.Title)
		}
		if p.Survivor != "" {
			line += fmt.Sprintf(" — %s %s", strings.ToLower(retirement.Reason), p.Survivor)
		}
		out.WriteString(line + "\n")
	}
	if len(p.Withheld) > 0 {
		fmt.Fprintf(&out, "  would also write: %s\n", strings.Join(p.Withheld, ", "))
	}
	out.WriteString("  re-run the step without --preview to apply it")
	return out.String()
}
