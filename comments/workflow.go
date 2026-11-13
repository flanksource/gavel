package comments

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

var ErrInvalid = errors.New("invalid comment input")

type Phase string

const (
	Active          Phase = "active"
	AwaitingOutcome Phase = "awaiting_outcome"
	Final           Phase = "final"
)

type Workflow struct {
	ID       string                `json:"id"`
	Initial  string                `json:"initial"`
	Statuses map[string]Phase      `json:"statuses"`
	Actions  map[string]Transition `json:"actions"`
}

func copyWorkflow(w Workflow) Workflow {
	w.Statuses = maps.Clone(w.Statuses)
	w.Actions = maps.Clone(w.Actions)
	for action, t := range w.Actions {
		t.From = slices.Clone(t.From)
		w.Actions[action] = t
	}
	return w
}

type Transition struct {
	From      []string `json:"from"`
	To        string   `json:"to"`
	HumanOnly bool     `json:"humanOnly"`
}

func (w Workflow) Apply(status, action string, actor Actor) (string, Phase, error) {
	t, ok := w.Actions[action]
	if !ok || !slices.Contains(t.From, status) {
		return "", "", fmt.Errorf("%w: %s cannot %s from %s", ErrInvalid, w.ID, action, status)
	}
	if t.HumanOnly && actor.Kind != "user" {
		return "", "", fmt.Errorf("%w: %s requires a human", ErrInvalid, action)
	}
	phase, ok := w.Statuses[t.To]
	if !ok {
		return "", "", fmt.Errorf("%w: workflow destination %s", ErrInvalid, t.To)
	}
	return t.To, phase, nil
}

func HumanReview() Workflow {
	return Workflow{ID: "human-review", Initial: "open", Statuses: map[string]Phase{"open": Active, "in_progress": Active, "resolved": AwaitingOutcome, "closed": Final}, Actions: map[string]Transition{
		"acknowledge": {From: []string{"open", "in_progress"}, To: "in_progress"},
		"resolve":     {From: []string{"open", "in_progress"}, To: "resolved"},
		"close":       {From: []string{"open", "in_progress", "resolved"}, To: "closed", HumanOnly: true},
		"reopen":      {From: []string{"resolved", "closed"}, To: "open", HumanOnly: true},
	}}
}
func NativePR() Workflow {
	return Workflow{ID: "native-pr", Initial: "open", Statuses: map[string]Phase{"open": Active, "resolved": Final}, Actions: map[string]Transition{
		"resolve": {From: []string{"open"}, To: "resolved"}, "reopen": {From: []string{"resolved"}, To: "open"},
	}}
}

type Actor struct {
	Kind  string `json:"kind"`
	ID    string `json:"id"`
	Label string `json:"label,omitempty"`
}
