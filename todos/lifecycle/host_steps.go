package lifecycle

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos/types"
)

// StepState is one lifecycle step as a caller choosing between them sees it:
// whether it applies to this todo now, whether it is the one the lifecycle
// would pick, why, and what the last run of it did.
//
// It is the single answer behind `gavel todos steps`, the dashboard's step
// picker, and the step a run defaults to — three surfaces that must not be able
// to disagree about which step comes next.
type StepState struct {
	Step Step
	// Applicable is the step's `when` predicate over this todo.
	Applicable bool
	// Suggested marks the one step Next would run: the first non-auxiliary
	// applicable step. An auxiliary step is never suggested even when it applies.
	Suggested bool
	// Reason is the predicate that decided Applicable, in the definition's own
	// words, so a caller can show why a step is offered or withheld.
	Reason string
	// LastRun is the step's most recent recorded run, nil when it has never run.
	LastRun *StepRun
	// Done reports that the step's last run finished successfully.
	Done bool
}

// Steps explains every declared step for a todo, in definition order, auxiliary
// steps included — a caller listing what a todo can do wants them, even though
// Next never picks one.
func (h *Host) Steps(ctx context.Context, todo *types.TODO) ([]StepState, error) {
	if h.Def == nil {
		return nil, fmt.Errorf("lifecycle host: no lifecycle loaded")
	}
	lc, err := h.Context(ctx, todo)
	if err != nil {
		return nil, err
	}
	applicable, err := h.Def.Explain(lc)
	if err != nil {
		return nil, err
	}
	next, hasNext, err := h.Def.Next(lc)
	if err != nil {
		return nil, err
	}
	last := map[string]StepRun{}
	for _, run := range lc.Runs {
		last[run.Step] = run
	}
	states := make([]StepState, 0, len(h.Def.Definition().Steps))
	for _, step := range h.Def.Definition().Steps {
		state := StepState{
			Step:       step,
			Applicable: applicable[step.Name],
			Suggested:  hasNext && next.Name == step.Name,
			Reason:     stepReason(step, applicable[step.Name]),
		}
		if run, ok := last[step.Name]; ok {
			state.LastRun = &run
			state.Done = run.State == RunSucceeded
		}
		states = append(states, state)
	}
	return states, nil
}

// StepFor resolves the step a run will execute: the one the caller named, or —
// for an empty name — the step the lifecycle would run next. The reason is the
// text a CLI or dialog shows next to the choice.
//
// A named step is returned whether or not it applies: naming a step is an
// explicit override of the lifecycle's suggestion, and refusing it would leave
// no way to re-plan a todo the definition considers planned. An unknown name is
// an error that enumerates the lifecycle's own steps, because the set is
// project-specific and not discoverable from the error otherwise.
func (h *Host) StepFor(ctx context.Context, todo *types.TODO, name string) (Step, string, error) {
	if h.Def == nil {
		return Step{}, "", fmt.Errorf("lifecycle host: no lifecycle loaded")
	}
	def := h.Def.Definition()
	if name = strings.TrimSpace(name); name != "" {
		step, ok := def.Step(name)
		if !ok {
			return Step{}, "", fmt.Errorf("step %q is not part of lifecycle %s; steps: %s",
				name, def.Name, strings.Join(def.StepNames(), ", "))
		}
		return step, "requested with --step " + name, nil
	}
	lc, err := h.Context(ctx, todo)
	if err != nil {
		return Step{}, "", err
	}
	step, ok, err := h.Def.Next(lc)
	if err != nil {
		return Step{}, "", err
	}
	if !ok {
		return Step{}, "", fmt.Errorf(
			"no lifecycle step applies to todo %s (status %s); name one with --step, or move the todo on",
			todoName(todo), todo.Status)
	}
	return step, stepReason(step, true), nil
}

// stepReason renders a step's predicate as the sentence a caller reads.
func stepReason(step Step, applies bool) string {
	when := strings.Join(strings.Fields(step.When), " ")
	if when == "" {
		return "always applies (no `when` predicate)"
	}
	if applies {
		return "applies: " + when
	}
	return "does not apply: " + when
}
