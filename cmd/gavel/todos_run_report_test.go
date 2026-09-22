package main

import (
	"strings"
	"testing"

	captainapi "github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
)

func triageOutcomeFor(execution *todos.ExecutionResult) *lifecycle.StepOutcome {
	return &lifecycle.StepOutcome{Step: lifecycle.Step{Name: "triage"}, Execution: execution}
}

// The reported status is the lifecycle's, rendered through types.Status.Pretty —
// the old line was hardcoded green, so a failed run finished in the same colour
// as a successful one.
func TestRenderRunResultColoursTheStatusByOutcome(t *testing.T) {
	execution := &todos.ExecutionResult{ExecutorName: "cli-claude", Summary: "Could not read the fixture."}
	report := renderRunResult("triage", triageOutcomeFor(execution), string(types.StatusFailed), nil)

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
	report := renderRunResult("triage", triageOutcomeFor(&todos.ExecutionResult{ExecutorName: "cli-claude"}), lifecycle.OutcomeKeep, nil)

	if !strings.Contains(report, "status unchanged") {
		t.Errorf("report %q should say the status was kept", report)
	}
	if lines := strings.Count(report, "\n") + 1; lines != 2 {
		t.Errorf("report has %d lines, want the metrics line and the status line only:\n%s", lines, report)
	}
}

// An ask used to print "finished — ask" and swallow the questions, leaving the
// operator to go and read the TODO to find out what was being asked.
func TestRenderRunResultPrintsQuestionsForAnAsk(t *testing.T) {
	execution := &todos.ExecutionResult{
		ExecutorName: "cli-claude",
		EndStatus:    types.EndAsk,
		Questions: []types.AgentQuestion{{
			Text:    "Which provider owns the write?",
			Context: "todos/triage.go holds the lock",
			Options: []string{"the native provider", "the lifecycle host"},
		}},
	}
	report := renderRunResult("triage", triageOutcomeFor(execution), string(types.StatusAsk), nil)

	for _, want := range []string{"Which provider owns the write?", "todos/triage.go holds the lock", "the lifecycle host"} {
		if !strings.Contains(report, want) {
			t.Errorf("report %q should contain %q", report, want)
		}
	}
}

// The verdict is a request the host applies afterwards, so it is printed from the
// envelope rather than read back off a TODO the write may not have reached.
func TestRenderRunResultPrintsTheTriageVerdict(t *testing.T) {
	execution := &todos.ExecutionResult{
		ExecutorName: "cli-claude",
		EndStatus:    types.EndCompleted,
		Triage: &types.TriageEnvelope{
			Verdict: types.VerdictMergeInto,
			Body:    "Combined description.",
			Comment: "Same work as the two folded in.",
			Merges:  []string{"ff0011aa", "cc3344bb"},
		},
	}
	report := renderRunResult("triage", triageOutcomeFor(execution), lifecycle.OutcomeKeep, nil)

	for _, want := range []string{"merge-into", "body", "ff0011aa", "cc3344bb", "Same work as the two folded in."} {
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
	report := renderRunResult("triage", triageOutcomeFor(execution), string(types.StatusFailed), nil)

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
		EndStatus:    types.EndCompleted,
		Summary:      "Triaged it.",
		ResponseText: `{"summary":"Triaged it.","endStatus":"completed"}`,
	}
	report := renderRunResult("triage", triageOutcomeFor(execution), lifecycle.OutcomeKeep, nil)

	if strings.Contains(report, "did not match the step's schema") {
		t.Errorf("report %q should not claim a schema mismatch", report)
	}
	if strings.Contains(report, `"endStatus"`) {
		t.Errorf("report %q should not echo the raw envelope", report)
	}
}

func TestRenderRunResultPrintsTheDefinitionOfDoneVerdict(t *testing.T) {
	execution := &todos.ExecutionResult{
		ExecutorName: "cli-claude",
		EndStatus:    types.EndCompleted,
		DoD:          &todos.DoDOutcome{Ran: true, Report: &captainapi.VerifyReport{Reason: "2 of 3 fixtures failed"}},
	}
	report := renderRunResult("verify", triageOutcomeFor(execution), string(types.StatusUnverified), nil)

	for _, want := range []string{"definition of done", "failed", "2 of 3 fixtures failed"} {
		if !strings.Contains(report, want) {
			t.Errorf("report %q should contain %q", report, want)
		}
	}
}

func TestRenderRunResultIsEmptyWithoutAnExecution(t *testing.T) {
	if report := renderRunResult("triage", nil, string(types.StatusFailed), nil); report != "" {
		t.Errorf("report for a run that produced nothing = %q, want empty", report)
	}
}
