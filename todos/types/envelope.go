package types

import (
	"fmt"
	"strings"
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

// TriageVerdict is the fate a triage run assigns a TODO. The five verdicts are
// the ones the gavel-triage workflow defines; they decide which fields of the
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
	// VerdictRetire means the TODO is obsolete, a duplicate, or won't be done.
	VerdictRetire TriageVerdict = "retire"
)

// KnownTriageVerdicts returns every verdict a triage envelope may carry.
func KnownTriageVerdicts() []TriageVerdict {
	return []TriageVerdict{VerdictReady, VerdictShape, VerdictInvestigate, VerdictDone, VerdictRetire}
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
	Verdict      TriageVerdict `json:"verdict" jsonschema:"required,enum=ready,enum=shape,enum=investigate,enum=done,enum=retire" jsonschema_description:"ready = implementable as written, shape = rewrite body and fixture, investigate = needs a planning run, done = believed already implemented, retire = obsolete or duplicate"`
	Title        string        `json:"title,omitempty" jsonschema_description:"Replacement title, when the current one does not describe the work"`
	Body         string        `json:"body,omitempty" jsonschema_description:"The compacted description: problem statement, then ## Acceptance Criteria, then ## Scope. Required when verdict is shape"`
	Verification string        `json:"verification,omitempty" jsonschema_description:"The rewritten ## Verification fixture markdown, without the outer heading"`
	Priority     string        `json:"priority,omitempty" jsonschema:"enum=high,enum=medium,enum=low"`
	Status       string        `json:"status,omitempty" jsonschema:"enum=draft,enum=pending,enum=verified,enum=completed,enum=skipped" jsonschema_description:"Only directly-assignable statuses; run projections such as review or in_progress are rejected"`
	DuplicateOf  string        `json:"duplicateOf,omitempty" jsonschema_description:"Short id of the surviving TODO this one duplicates"`
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
	return nil
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
