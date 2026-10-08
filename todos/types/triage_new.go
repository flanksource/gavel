package types

import (
	"fmt"
	"strings"
)

type TriageNewEnvelope struct {
	ResultEnvelope
	Title        string   `json:"title" jsonschema:"required,minLength=1"`
	Labels       []string `json:"labels" jsonschema:"required,minItems=1"`
	Action       string   `json:"action" jsonschema:"required,enum=keep,enum=duplicate-of,enum=merge-into,enum=child-of"`
	Target       string   `json:"target,omitempty"`
	Rationale    string   `json:"rationale,omitempty"`
	Body         string   `json:"body,omitempty"`
	Verification string   `json:"verification,omitempty"`
	ProposalID   string   `json:"proposalId,omitempty"`
	Plan         string   `json:"plan,omitempty"`
}

func (e *TriageNewEnvelope) Validate() error {
	if err := e.ResultEnvelope.Validate(); err != nil {
		return err
	}
	if e.EndStatus != EndCompleted {
		return nil
	}
	if strings.TrimSpace(e.Title) == "" {
		return fmt.Errorf("triage.new title is required")
	}
	if len(e.Labels) == 0 {
		return fmt.Errorf("triage.new labels are required")
	}
	seen := map[string]bool{}
	for _, label := range e.Labels {
		key := strings.ToLower(strings.TrimSpace(label))
		if key == "" || seen[key] {
			return fmt.Errorf("triage.new invalid or repeated label %q", label)
		}
		seen[key] = true
	}
	switch e.Action {
	case "keep":
		if e.Target != "" || e.ProposalID != "" {
			return fmt.Errorf("keep cannot name a relationship target or approval")
		}
	case "duplicate-of", "merge-into", "child-of":
		if strings.TrimSpace(e.Target) == "" || strings.TrimSpace(e.Rationale) == "" {
			return fmt.Errorf("%s requires target and rationale", e.Action)
		}
		if e.Action == "merge-into" && strings.TrimSpace(e.Body) == "" {
			return fmt.Errorf("merge-into requires combined body")
		}
	default:
		return fmt.Errorf("unknown triage.new action %q", e.Action)
	}
	return nil
}
