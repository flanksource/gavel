package ui

import (
	"context"
	"time"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
)

// todoLifecycleState is a todo's lifecycle as the detail view shows it: every
// declared step with whether it applies now, the step the lifecycle would run
// next, and why. It is computed server-side because the definition — with its
// CEL predicates — is data the browser has no copy of.
type todoLifecycleState struct {
	Steps []todoLifecycleStep `json:"steps"`
	// Next is the step the lifecycle would run for this todo now; empty when no
	// step applies.
	Next   string `json:"next,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// todoLifecycleStep is one step's state for this todo.
type todoLifecycleStep struct {
	Name       string `json:"name"`
	Label      string `json:"label"`
	Prompt     string `json:"prompt"`
	Applicable bool   `json:"applicable"`
	Suggested  bool   `json:"suggested,omitempty"`
	Done       bool   `json:"done,omitempty"`
	Auxiliary  bool   `json:"auxiliary,omitempty"`
	ReadOnly   bool   `json:"readOnly,omitempty"`
	// Reason is the predicate that decided Applicable, in the definition's words.
	Reason  string            `json:"reason,omitempty"`
	LastRun *todoLifecycleRun `json:"lastRun,omitempty"`
}

// todoLifecycleRun is a step's most recent recorded run.
type todoLifecycleRun struct {
	State       string     `json:"state"`
	Outcome     string     `json:"outcome,omitempty"`
	PromptRunID string     `json:"promptRunId,omitempty"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	FinishedAt  *time.Time `json:"finishedAt,omitempty"`
}

// todoDetail builds the detail-shaped summary of a todo, including its
// lifecycle. Every HTTP handler that returns a detail-shaped todo (as opposed
// to a list-row summary) must go through this rather than calling
// summarizeTodo(todo, true) directly, so the lifecycle strip never goes stale
// in the client's todo cache after a write. A lifecycle that cannot be
// evaluated is reported as an error rather than silently omitted.
func todoDetail(ctx context.Context, provider todos.Provider, dir string, todo *types.TODO) (todoSummary, error) {
	sum := summarizeTodo(todo, true)
	lifecycle, err := todoLifecycleFor(ctx, provider, dir, todo)
	if err != nil {
		return todoSummary{}, err
	}
	sum.Lifecycle = lifecycle
	return sum, nil
}

// todoLifecycleFor asks the workspace's lifecycle host where a todo stands.
func todoLifecycleFor(ctx context.Context, provider todos.Provider, dir string, todo *types.TODO) (*todoLifecycleState, error) {
	host, err := lifecycle.NewHost(provider, dir, lifecycle.HostDashboard)
	if err != nil {
		return nil, err
	}
	states, err := host.Steps(ctx, todo)
	if err != nil {
		return nil, err
	}
	out := &todoLifecycleState{Steps: make([]todoLifecycleStep, 0, len(states))}
	for _, state := range states {
		step := todoLifecycleStep{
			Name:       state.Step.Name,
			Label:      stepLabel(state.Step.Name),
			Prompt:     state.Step.Prompt,
			Applicable: state.Applicable,
			Suggested:  state.Suggested,
			Done:       state.Done,
			Auxiliary:  state.Step.Auxiliary,
			ReadOnly:   lifecycle.Class(state.Step) != types.ModeRun,
			Reason:     state.Reason,
		}
		if last := state.LastRun; last != nil {
			step.LastRun = &todoLifecycleRun{
				State: last.State, Outcome: last.Outcome, PromptRunID: last.PromptRunID,
				StartedAt: last.StartedAt, FinishedAt: last.FinishedAt,
			}
		}
		if state.Suggested {
			out.Next, out.Reason = state.Step.Name, state.Reason
		}
		out.Steps = append(out.Steps, step)
	}
	return out, nil
}
