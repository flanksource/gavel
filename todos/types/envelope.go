package types

import (
	"fmt"
	"slices"
	"strings"
)

// Envelope is the structured final result every todo prompt returns, whatever
// its step. Success is the agent's own claim; Event names what happened, from
// the vocabulary the lifecycle step declares; Comment is the markdown account of
// the run and becomes the run's single history entry.
//
// It carries no edits: a run changes a TODO through its tools, which write as
// they are called, so there is nothing left for gavel to apply afterwards.
type Envelope struct {
	Success bool   `json:"success"`
	Event   string `json:"event"`
	Comment string `json:"comment"`
}

// Validate fails loud on an envelope the step's outcomes could not classify: a
// blank comment, or an event outside the step's declared vocabulary.
func (e *Envelope) Validate(events []string) error {
	if len(events) == 0 {
		return fmt.Errorf("envelope cannot be validated: the step declares no events")
	}
	if strings.TrimSpace(e.Comment) == "" {
		return fmt.Errorf("envelope comment is empty")
	}
	if !slices.Contains(events, e.Event) {
		return fmt.Errorf("envelope event %q is not one of %s", e.Event, strings.Join(events, ", "))
	}
	return nil
}

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

// KnownPlanStatus reports whether value names a plan status.
func KnownPlanStatus(value string) bool {
	switch PlanStatus(value) {
	case PlanNew, PlanUpdated, PlanUnchanged:
		return true
	}
	return false
}

// PlanResult is the plan a plan-mode run produced. Path is the agent's native
// plan-mode file when it has one; Content carries the captured markdown, from
// the native plan outcome or the todo_plan tool.
type PlanResult struct {
	Status  PlanStatus `json:"status,omitempty"`
	Path    string     `json:"path,omitempty"`
	Content string     `json:"content,omitempty"`
}
