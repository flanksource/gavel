package main

import (
	"strings"
	"testing"

	captainapi "github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/lifecycle"
	"github.com/flanksource/gavel/todos"
	lifecycle1 "github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
)

func triageOutcomeFor(execution *todos.ExecutionResult) *lifecycle1.StepOutcome {
	return &lifecycle1.StepOutcome{Step: lifecycle.Step{Name: "triage"}, Execution: execution}
}

// The reported status is the lifecycle's, rendered through types.Status.Pretty —
// the old line was hardcoded green, so a failed run finished in the same colour
// as a successful one.
func TestRenderRunResultColoursTheStatusByOutcome(t *testing.T) {
	execution := &todos.ExecutionResult{ExecutorName: "cli-claude", Summary: "Could not read the fixture."}
	report := renderRunResult("ref", "triage", triageOutcomeFor(execution), string(types.StatusFailed), nil)

	for _, want := range []string{"triage finished", "cli-claude", "Could not read the fixture."} {
		if !strings.Contains(report, want) {
			t.Errorf("report %q should contain %q", report, want)
		}
	}
	if !strings.Contains(report, types.StatusFailed.Pretty().ANSI()) {
		t.Errorf("report %q should render the status through types.Status.Pretty", report)
	}
	if strings.Contains(report, types.StatusCompleted.Pretty().ANSI()) {
		t.Errorf("report %q rendered a failure as a success", report)
	}
}

// `keep` is not a todo status: the triage step deliberately leaves it alone, and
// printing it as one would name a transition that never happened.
func TestRenderRunResultReportsAKeptStatus(t *testing.T) {
	report := renderRunResult("ref", "triage", triageOutcomeFor(&todos.ExecutionResult{ExecutorName: "cli-claude"}), lifecycle.OutcomeKeep, nil)

	if !strings.Contains(report, "status unchanged") {
		t.Errorf("report %q should say the status was kept", report)
	}
	if lines := strings.Count(report, "\n") + 1; lines != 2 {
		t.Errorf("report has %d lines, want the metrics line and the status line only:\n%s", lines, report)
	}
}

// The tools wrote during the run, so the report prints what they wrote from the
// execution rather than reading it back off a TODO.
func TestRenderRunResultPrintsTheEventAndToolWrites(t *testing.T) {
	execution := &todos.ExecutionResult{
		ExecutorName: "cli-claude",
		Envelope:     &types.Envelope{Success: true, Event: "merged", Comment: "Same work as the two folded in."},
		ToolWrites:   []todos.ToolWrite{{Name: "todo_merge", Ref: "ff0011aa"}, {Name: "todo_merge", Ref: "cc3344bb"}},
	}
	report := renderRunResult("ref", "triage", triageOutcomeFor(execution), lifecycle.OutcomeKeep, nil)

	for _, want := range []string{"merged", "todo_merge ff0011aa", "todo_merge cc3344bb"} {
		if !strings.Contains(report, want) {
			t.Errorf("report %q should contain %q", report, want)
		}
	}
}

// The case the report exists for: an agent that answered well in the wrong shape
// left nothing behind but the word "failed".
func TestRenderRunResultPrintsAnUndecodedResponse(t *testing.T) {
	const prose = "This TODO is already implemented in todos/triage.go."
	execution := &todos.ExecutionResult{
		ExecutorName: "cli-claude",
		ErrorMessage: "decode envelope: unexpected end of JSON input",
		ResponseText: prose,
	}
	report := renderRunResult("ref", "triage", triageOutcomeFor(execution), string(types.StatusFailed), nil)

	for _, want := range []string{prose, "did not match the step's schema", "unexpected end of JSON input"} {
		if !strings.Contains(report, want) {
			t.Errorf("report %q should contain %q", report, want)
		}
	}
}

// A decoded envelope's own text is already reported as its summary; repeating the
// raw response under a schema-mismatch banner would claim a failure that did not
// happen.
func TestRenderRunResultOmitsTheResponseWhenAnEnvelopeDecoded(t *testing.T) {
	execution := &todos.ExecutionResult{
		ExecutorName: "cli-claude",
		Envelope:     &types.Envelope{Success: true, Event: "ready", Comment: "Triaged it."},
		Summary:      "Triaged it.",
		ResponseText: `{"success":true,"event":"ready","comment":"Triaged it."}`,
	}
	report := renderRunResult("ref", "triage", triageOutcomeFor(execution), lifecycle.OutcomeKeep, nil)

	if strings.Contains(report, "did not match the step's schema") {
		t.Errorf("report %q should not claim a schema mismatch", report)
	}
	if strings.Contains(report, `"event"`) {
		t.Errorf("report %q should not echo the raw envelope", report)
	}
}

func TestRenderRunResultPrintsTheDefinitionOfDoneVerdict(t *testing.T) {
	execution := &todos.ExecutionResult{
		ExecutorName: "cli-claude",
		DoD:          &todos.DoDOutcome{Ran: true, Report: &captainapi.VerifyReport{Reason: "2 of 3 fixtures failed"}},
	}
	report := renderRunResult("ref", "verify", triageOutcomeFor(execution), string(types.StatusUnverified), nil)

	for _, want := range []string{"definition of done", "failed", "2 of 3 fixtures failed"} {
		if !strings.Contains(report, want) {
			t.Errorf("report %q should contain %q", report, want)
		}
	}
}

func TestRenderRunResultIsEmptyWithoutAnExecution(t *testing.T) {
	if report := renderRunResult("ref", "triage", nil, string(types.StatusFailed), nil); report != "" {
		t.Errorf("report for a run that produced nothing = %q, want empty", report)
	}
}
