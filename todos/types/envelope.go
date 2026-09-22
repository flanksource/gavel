package types

import (
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos/labels"
)

// EndStatus is the agent-reported outcome of a run/plan session, carried in the
// structured-output envelope every todo prompt requests.
type EndStatus string

const (
	// EndCompleted means the agent finished the requested work.
	EndCompleted EndStatus = "completed"
	// EndFailed means the agent could not complete the work.
	EndFailed EndStatus = "failed"
	// EndAsk means the agent is blocked on questions a human must answer.
	EndAsk EndStatus = "ask"
)

// PlanStatus describes what a plan run did to the plan file.
type PlanStatus string

const (
	// PlanNew means the agent authored a plan for the first time.
	PlanNew PlanStatus = "new"
	// PlanUpdated means the agent revised an existing plan.
	PlanUpdated PlanStatus = "updated"
	// PlanUnchanged means the existing plan still stands; the todo is ready to
	// execute without another review.
	PlanUnchanged PlanStatus = "unchanged"
)

// AgentQuestion is one question the agent needs answered before it can
// continue. Options, when present, are suggested answers.
type AgentQuestion struct {
	Text    string   `json:"text" jsonschema:"required" jsonschema_description:"The question that blocks progress"`
	Context string   `json:"context,omitempty" jsonschema_description:"Why the question matters (files, trade-offs)"`
	Options []string `json:"options,omitempty" jsonschema_description:"Suggested answers, if any"`
}

// PlanResult reports the plan a plan-mode session produced. Path is the agent's
// native plan-mode file when it has one; Content carries the captured markdown.
type PlanResult struct {
	Status  PlanStatus `json:"status,omitempty" jsonschema:"required,enum=new,enum=updated,enum=unchanged" jsonschema_description:"new = first plan, updated = revised, unchanged = existing plan still stands"`
	Path    string     `json:"path,omitempty" jsonschema_description:"Absolute path of the native plan-mode file this session wrote, when available"`
	Content string     `json:"content,omitempty" jsonschema_description:"Inline markdown plan content for backends that do not write a native plan file"`
}

// ResultEnvelope is the structured final result every run-mode agent session
// must emit.
type ResultEnvelope struct {
	Summary   string          `json:"summary" jsonschema:"required,maxLength=1000" jsonschema_description:"What was done, found, or attempted in 2–4 sentences"`
	EndStatus EndStatus       `json:"endStatus" jsonschema:"required,enum=completed,enum=failed,enum=ask"`
	Questions []AgentQuestion `json:"questions,omitempty" jsonschema_description:"Required when endStatus is ask"`
}

// Validate fails loud on an envelope that cannot drive a status transition.
func (e *ResultEnvelope) Validate() error {
	if strings.TrimSpace(e.Summary) == "" {
		return fmt.Errorf("envelope summary is empty")
	}
	switch e.EndStatus {
	case EndCompleted, EndFailed, EndAsk:
	default:
		return fmt.Errorf("envelope endStatus %q is not one of completed, failed, ask", e.EndStatus)
	}
	if e.EndStatus == EndAsk && len(e.Questions) == 0 {
		return fmt.Errorf("envelope endStatus is ask but no questions were provided")
	}
	return nil
}

// TriageVerdict is the fate a triage run assigns a TODO. The verdicts are the
// ones the gavel-triage workflow defines; they decide which fields of the
// envelope gavel is expected to act on.
type TriageVerdict string

const (
	// VerdictReady means the TODO is implementable as written; only its priority
	// may need correcting.
	VerdictReady TriageVerdict = "ready"
	// VerdictShape means the work is real but under-specified: the body and the
	// verification fixture are rewritten.
	VerdictShape TriageVerdict = "shape"
	// VerdictInvestigate means the solution is genuinely unknown, so the TODO is
	// left for a planning run rather than reshaped.
	VerdictInvestigate TriageVerdict = "investigate"
	// VerdictDone means the agent believes the work is already implemented. It is
	// a claim, never a status write: the definition-of-done check proves it.
	VerdictDone TriageVerdict = "done"
	// VerdictRetire means the TODO is obsolete or won't be done, so it is closed
	// with the rationale recorded on it. A duplicate is not retired: it belongs to
	// VerdictDuplicateOf or VerdictMergeInto, which record where the work went
	// instead of dropping it.
	VerdictRetire TriageVerdict = "retire"
	// VerdictMergeInto means the TODO being triaged is the SURVIVOR of a set of
	// overlapping TODOs: Merges names the ones folded into it, and the envelope's
	// body carries the combined description. It is the only verdict that closes a
	// TODO other than the one being triaged.
	VerdictMergeInto TriageVerdict = "merge-into"
	// VerdictDuplicateOf means another TODO already covers this work entirely, so
	// nothing needs carrying over: this one is linked to the survivor and closed.
	VerdictDuplicateOf TriageVerdict = "duplicate-of"
)

// KnownTriageVerdicts returns every verdict a triage envelope may carry.
func KnownTriageVerdicts() []TriageVerdict {
	return []TriageVerdict{
		VerdictReady, VerdictShape, VerdictInvestigate, VerdictDone,
		VerdictRetire, VerdictMergeInto, VerdictDuplicateOf,
	}
}

// TriageEnvelope is the triage structured result: the run envelope plus the
// verdict and the edits gavel applies on the agent's behalf. The agent itself is
// read-only — it proposes, gavel writes — so every field here is an instruction,
// validated before it reaches storage.
//
// Every field is a top-level scalar (or a scalar array) for the same reason
// PlanEnvelope's are: nested objects become $ref nodes that some backends refuse
// to emit. See TestTriageEnvelopeSchemaUsesFlatScalarFields.
type TriageEnvelope struct {
	ResultEnvelope
	Verdict      TriageVerdict `json:"verdict" jsonschema:"required,enum=ready,enum=shape,enum=investigate,enum=done,enum=retire,enum=merge-into,enum=duplicate-of" jsonschema_description:"ready = implementable as written, shape = rewrite body and fixture, investigate = needs a planning run, done = believed already implemented, retire = obsolete or won't be done and this TODO is closed, merge-into = this TODO is the survivor and merges names the ones folded into it, duplicate-of = duplicateOf already covers this entirely"`
	Title        string        `json:"title,omitempty" jsonschema_description:"Replacement title, when the current one does not describe the work"`
	Body         string        `json:"body,omitempty" jsonschema_description:"The compacted description: problem statement, then ## Acceptance Criteria, then ## Scope. Required when verdict is shape"`
	Verification string        `json:"verification,omitempty" jsonschema_description:"The rewritten ## Verification fixture markdown, without the outer heading"`
	Priority     string        `json:"priority,omitempty" jsonschema:"enum=high,enum=medium,enum=low" jsonschema_description:"Severity: high = blocks other work or is broken for users, medium = ordinary queued work, low = nice to have. Omit when the current priority is already right"`
	Status       string        `json:"status,omitempty" jsonschema:"enum=draft,enum=pending,enum=verified,enum=completed,enum=skipped" jsonschema_description:"Only directly-assignable statuses; run projections such as review or in_progress are rejected"`
	AddLabels    []string      `json:"addLabels,omitempty" jsonschema_description:"Labels to add, chosen ONLY from the labels listed in the prompt. A label that is not listed is rejected"`
	RemoveLabels []string      `json:"removeLabels,omitempty" jsonschema_description:"Labels the TODO currently carries that no longer describe the work"`
	Merges       []string      `json:"merges,omitempty" jsonschema_description:"Short ids of the TODOs to fold INTO this one, which are then linked here and closed. Required when verdict is merge-into"`
	DuplicateOf  string        `json:"duplicateOf,omitempty" jsonschema_description:"Short id of the surviving TODO that already covers this work. Required when verdict is duplicate-of, and used by no other verdict"`
	Related      []string      `json:"related,omitempty" jsonschema_description:"Short ids of related TODOs to link"`
	Comment      string        `json:"comment,omitempty" jsonschema_description:"Rationale recorded on the TODO. Required when verdict is retire"`
}

// Validate extends ResultEnvelope.Validate with the triage contract. It rejects
// a status or priority storage would decline rather than letting the write be
// silently dropped, and it holds the two verdicts that mean nothing without
// their payload to that promise.
//
// Ask/failed sessions are exempt: an agent that stopped to ask a question has no
// verdict to honour.
func (e *TriageEnvelope) Validate() error {
	if err := e.ResultEnvelope.Validate(); err != nil {
		return err
	}
	if e.EndStatus != EndCompleted {
		return nil
	}
	if !e.knownVerdict() {
		return fmt.Errorf("triage verdict %q is not one of %s", e.Verdict, joinVerdicts(KnownTriageVerdicts()))
	}
	if e.Verdict == VerdictShape && strings.TrimSpace(e.Body) == "" {
		return fmt.Errorf("triage verdict %q requires a rewritten body", VerdictShape)
	}
	if e.Verdict == VerdictRetire && strings.TrimSpace(e.Comment) == "" {
		return fmt.Errorf("triage verdict %q requires a comment recording why", VerdictRetire)
	}
	if err := e.validateRetirements(); err != nil {
		return err
	}
	if raw := strings.TrimSpace(e.Status); raw != "" {
		if err := ValidateAssignableStatus(Status(raw)); err != nil {
			return fmt.Errorf("triage status: %w", err)
		}
	}
	if raw := strings.TrimSpace(e.Priority); raw != "" {
		if err := ValidatePriority(Priority(raw)); err != nil {
			return fmt.Errorf("triage priority: %w", err)
		}
	}
	return e.validateLabels()
}

// validateRetirements holds the two verdicts that close a TODO to the payload
// that makes them actionable, and keeps each field to the one verdict that owns
// it.
//
// These are the only verdicts gavel cannot walk back: a fold with no content
// discards the folded TODOs' descriptions, and a duplicate with no target leaves
// nothing to point at. Both are rejected here rather than half-applied. A field
// sent under the wrong verdict is rejected too — silently ignoring it would let
// an agent believe it had recorded a duplicate when nothing was written.
func (e *TriageEnvelope) validateRetirements() error {
	merges := trimmedNonEmpty(e.Merges)
	duplicateOf := strings.TrimSpace(e.DuplicateOf)

	switch e.Verdict {
	case VerdictMergeInto:
		if len(merges) == 0 {
			return fmt.Errorf("triage verdict %q requires merges to name at least one TODO to fold in", VerdictMergeInto)
		}
		if strings.TrimSpace(e.Body) == "" {
			return fmt.Errorf("triage verdict %q requires a body carrying the combined description", VerdictMergeInto)
		}
		if duplicateOf != "" {
			return fmt.Errorf("triage verdict %q uses merges, not duplicateOf", VerdictMergeInto)
		}
	case VerdictDuplicateOf:
		if duplicateOf == "" {
			return fmt.Errorf("triage verdict %q requires duplicateOf to name the surviving TODO", VerdictDuplicateOf)
		}
		if strings.TrimSpace(e.Comment) == "" {
			return fmt.Errorf("triage verdict %q requires a comment recording why", VerdictDuplicateOf)
		}
		if len(merges) > 0 {
			return fmt.Errorf("triage verdict %q uses duplicateOf, not merges", VerdictDuplicateOf)
		}
	default:
		if len(merges) > 0 {
			return fmt.Errorf("triage merges is only accepted with verdict %q, not %q", VerdictMergeInto, e.Verdict)
		}
		if duplicateOf != "" {
			return fmt.Errorf("triage duplicateOf is only accepted with verdict %q, not %q", VerdictDuplicateOf, e.Verdict)
		}
	}

	for i, ref := range merges {
		for _, earlier := range merges[:i] {
			if strings.EqualFold(earlier, ref) {
				return fmt.Errorf("triage merges names %q twice", ref)
			}
		}
	}
	if len(merges) != len(e.Merges) {
		return fmt.Errorf("triage merges contains a blank id")
	}
	return e.rejectClosingEdits()
}

// rejectClosingEdits refuses a verdict that both closes the TODO being triaged
// and rewrites it. The close is applied last, so a title, body or fixture written
// alongside it lands on something already on its way out, and a status write
// fights the close over what the TODO's final state is.
func (e *TriageEnvelope) rejectClosingEdits() error {
	if !e.ClosesSelf() {
		return nil
	}
	for _, field := range []struct{ name, value string }{
		{"body", e.Body}, {"verification", e.Verification}, {"title", e.Title}, {"status", e.Status},
	} {
		if strings.TrimSpace(field.value) != "" {
			return fmt.Errorf("triage verdict %q closes this TODO and cannot also set %s", e.Verdict, field.name)
		}
	}
	return nil
}

// RetiresTODOs reports whether acting on this envelope closes a TODO — this one
// for retire and duplicate-of, the folded ones for merge-into. It is the question
// a caller asks before deciding whether the TODO will still be there afterwards.
func (e *TriageEnvelope) RetiresTODOs() bool {
	return e.Verdict == VerdictMergeInto || e.ClosesSelf()
}

// ClosesSelf reports whether the verdict closes the TODO being triaged rather
// than another one: retire drops the work, duplicate-of hands it to a survivor.
// Both end with the same soft delete, which is why they share the guards on what
// else the envelope may carry.
func (e *TriageEnvelope) ClosesSelf() bool {
	return e.Verdict == VerdictRetire || e.Verdict == VerdictDuplicateOf
}

// RetirementTargets returns the refs this envelope closes: the folded TODOs for
// merge-into, and an empty list for retire and duplicate-of, which close the TODO
// being triaged rather than another one.
func (e *TriageEnvelope) RetirementTargets() []string {
	if e.Verdict != VerdictMergeInto {
		return nil
	}
	return trimmedNonEmpty(e.Merges)
}

func trimmedNonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// validateLabels holds the label delta to what a delta can mean. Membership of
// the workspace taxonomy is NOT checked here — that needs the provider, so it
// lives in todos.ApplyTriage — but a blank token or a label on both sides of the
// delta is incoherent whatever the taxonomy says, and silently dropping either
// would apply a label set the agent did not ask for.
func (e *TriageEnvelope) validateLabels() error {
	for _, field := range []struct {
		name   string
		values []string
	}{{"addLabels", e.AddLabels}, {"removeLabels", e.RemoveLabels}} {
		for _, label := range field.values {
			if strings.TrimSpace(label) == "" {
				return fmt.Errorf("triage %s contains a blank label", field.name)
			}
		}
	}
	for _, label := range e.AddLabels {
		if labels.Contains(e.RemoveLabels, label) {
			return fmt.Errorf("triage label %q is both added and removed", strings.TrimSpace(label))
		}
	}
	return nil
}

// ChangesLabels reports whether the envelope carries a label delta at all.
func (e *TriageEnvelope) ChangesLabels() bool {
	return len(e.AddLabels) > 0 || len(e.RemoveLabels) > 0
}

// ChangesFixture reports whether acting on this envelope alters the TODO's
// definition of done, which is what decides whether it earns a verification run.
func (e *TriageEnvelope) ChangesFixture() bool {
	return strings.TrimSpace(e.Verification) != "" || e.Verdict == VerdictDone
}

func (e *TriageEnvelope) knownVerdict() bool {
	for _, known := range KnownTriageVerdicts() {
		if e.Verdict == known {
			return true
		}
	}
	return false
}

func joinVerdicts(verdicts []TriageVerdict) string {
	names := make([]string, 0, len(verdicts))
	for _, verdict := range verdicts {
		names = append(names, string(verdict))
	}
	return strings.Join(names, ", ")
}

// PlanEnvelope is the plan-mode structured result: the run envelope plus the
// required plan definition.
type PlanEnvelope struct {
	ResultEnvelope
	PlanStatus  PlanStatus `json:"planStatus" jsonschema:"required,enum=new,enum=updated,enum=unchanged" jsonschema_description:"new = first plan, updated = revised, unchanged = existing plan still stands"`
	PlanPath    string     `json:"planPath,omitempty" jsonschema_description:"Absolute path of the native plan-mode file this session wrote, when available"`
	PlanContent string     `json:"planContent,omitempty" jsonschema_description:"Inline markdown plan content for backends that do not write a native plan file"`
}

// Validate extends ResultEnvelope.Validate with the plan contract: a completed
// plan run must provide either a plan file path or inline content (except when
// the existing plan is unchanged). Ask/failed sessions may legitimately have
// produced no plan. Callers additionally verify any file at Plan.Path exists
// and is non-empty.
func (e *PlanEnvelope) Validate() error {
	if err := e.ResultEnvelope.Validate(); err != nil {
		return err
	}
	if e.EndStatus != EndCompleted {
		return nil
	}
	switch e.PlanStatus {
	case PlanNew, PlanUpdated, PlanUnchanged:
	default:
		return fmt.Errorf("planStatus %q is not one of new, updated, unchanged", e.PlanStatus)
	}
	if e.PlanStatus != PlanUnchanged && strings.TrimSpace(e.PlanPath) == "" && strings.TrimSpace(e.PlanContent) == "" {
		return fmt.Errorf("planStatus %q requires planPath or planContent", e.PlanStatus)
	}
	return nil
}
