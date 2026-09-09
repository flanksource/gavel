package main

import (
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/lifecycle"
)

func TestTodosStepsCommandReplacesPrompts(t *testing.T) {
	registered := map[string]bool{}
	for _, command := range todosCmd.Commands() {
		registered[command.Name()] = true
	}
	if !registered["steps"] {
		t.Error("expected todos steps to be registered")
	}
	// `prompts` listed the run/plan/triage prompt catalog, which is no longer the
	// axis a caller chooses on: a project's lifecycle names the steps.
	if registered["prompts"] {
		t.Error("retired todos prompts command is still registered")
	}
	if todosStepsCmd.Args == nil {
		t.Fatal("expected todos steps to accept an optional todo argument")
	}
	if err := todosStepsCmd.Args(todosStepsCmd, []string{"3f2a1b"}); err != nil {
		t.Errorf("todos steps <id> rejected: %v", err)
	}
	if err := todosStepsCmd.Args(todosStepsCmd, nil); err != nil {
		t.Errorf("todos steps with no argument rejected: %v", err)
	}
	if err := todosStepsCmd.Args(todosStepsCmd, []string{"a", "b"}); err == nil {
		t.Error("todos steps accepted two arguments")
	}
}

// With no todo the command describes the lifecycle itself: which steps exist,
// which prompt each renders, and which are auxiliary (never chosen for you).
func TestRenderLifecycleStepsListsEveryStepAndItsPrompt(t *testing.T) {
	def := lifecycle.Lifecycle{
		Name: "todos",
		Steps: []lifecycle.Step{
			{Name: "triage", Prompt: "todos.triage", Auxiliary: true},
			{Name: "run", Prompt: "todos.run", When: "subject.status == 'pending'"},
		},
	}

	out := renderLifecycleSteps(def).String()
	for _, want := range []string{"todos", "triage", "todos.triage", "auxiliary", "run", "todos.run", "subject.status == 'pending'"} {
		if !strings.Contains(out, want) {
			t.Errorf("lifecycle listing is missing %q:\n%s", want, out)
		}
	}
}

// With a todo the command answers "what can I do with this one now": whether
// each step applies, which one comes next, why, and how its last run ended.
func TestRenderTodoStepsReportsApplicabilityNextAndLastRun(t *testing.T) {
	states := []lifecycle.StepState{
		{
			Step:       lifecycle.Step{Name: "plan", Prompt: "todos.plan"},
			Applicable: false,
			Reason:     "does not apply: !subject.plan.exists",
			LastRun:    &lifecycle.StepRun{Step: "plan", State: lifecycle.RunSucceeded},
			Done:       true,
		},
		{
			Step:       lifecycle.Step{Name: "run", Prompt: "todos.run"},
			Applicable: true,
			Suggested:  true,
			Reason:     "applies: subject.plan.approved",
		},
	}

	out := renderTodoSteps("Ship the thing", states).String()
	for _, want := range []string{
		"Ship the thing",
		"plan", "does not apply: !subject.plan.exists", "succeeded",
		"run", "applies: subject.plan.approved", "next",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("todo step listing is missing %q:\n%s", want, out)
		}
	}
	// A step that has never run must not claim a state it does not have.
	if strings.Count(out, "succeeded") != 1 {
		t.Errorf("last-run state reported for a step that never ran:\n%s", out)
	}
}
