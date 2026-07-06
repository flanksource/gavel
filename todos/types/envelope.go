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
// native plan-mode file when the backend has one. Content is for backends like
// Codex that keep the latest plan inline instead of writing a plan file.
type PlanResult struct {
	Status  PlanStatus `json:"status,omitempty" jsonschema:"required,enum=new,enum=updated,enum=unchanged" jsonschema_description:"new = first plan, updated = revised, unchanged = existing plan still stands"`
	Path    string     `json:"path,omitempty" jsonschema_description:"Absolute path of the native plan-mode file this session wrote, when available"`
	Content string     `json:"content,omitempty" jsonschema_description:"Inline markdown plan content for backends that do not write a native plan file"`
}

// ResultEnvelope is the structured final result every run-mode agent session
// must emit (and is re-requested when a session ends without one or times out).
type ResultEnvelope struct {
	Summary   string          `json:"summary" jsonschema:"required" jsonschema_description:"What was done, found, or attempted — a few sentences"`
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

// PlanEnvelope is the plan-mode structured result: the run envelope plus the
// required plan definition.
type PlanEnvelope struct {
	ResultEnvelope
	Plan PlanResult `json:"plan" jsonschema:"required" jsonschema_description:"The plan this session produced; required when endStatus is completed"`
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
	switch e.Plan.Status {
	case PlanNew, PlanUpdated, PlanUnchanged:
	default:
		return fmt.Errorf("plan.status %q is not one of new, updated, unchanged", e.Plan.Status)
	}
	if e.Plan.Status != PlanUnchanged && strings.TrimSpace(e.Plan.Path) == "" && strings.TrimSpace(e.Plan.Content) == "" {
		return fmt.Errorf("plan.status %q requires plan.path or plan.content", e.Plan.Status)
	}
	return nil
}
