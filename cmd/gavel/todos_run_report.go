package main

import (
	"fmt"
	"strings"

	"github.com/flanksource/clicky"
	clickyapi "github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
)

// renderRunResult is the terminal report for one finished lifecycle step run:
// what it cost, the status the lifecycle decided, and everything the agent said.
//
// It is rendered from the run's own ExecutionResult and printed before the
// outcome is persisted, so a failing write cannot take the account of the run
// with it. Before this the CLI printed one hardcoded-green
// "<step> finished — <status>" line, and the summary, the error, the questions,
// the verdict and a response that failed to decode were all dropped.
func renderRunResult(step string, outcome *lifecycle.StepOutcome, status string, runErr error) string {
	if outcome == nil || outcome.Execution == nil {
		return ""
	}
	execution := outcome.Execution
	var lines []string
	add := func(text clickyapi.Text) { lines = append(lines, text.ANSI()) }

	add(execution.Pretty())
	add(runStatusText(step, status))
	summary := strings.TrimSpace(execution.Summary)
	if summary != "" {
		add(clicky.Text(summary))
	}
	// The error message doubles as the summary for an agent-reported failure;
	// printing it twice would read as two separate problems.
	if message := strings.TrimSpace(execution.ErrorMessage); message != "" && message != summary {
		add(clicky.Text(message, "text-red-600"))
	}
	if runErr != nil && strings.TrimSpace(runErr.Error()) != strings.TrimSpace(execution.ErrorMessage) {
		add(clicky.Text(runErr.Error(), "text-red-600"))
	}
	lines = append(lines, renderRunQuestions(execution)...)
	lines = append(lines, renderRunVerdict(execution)...)
	if execution.Plan != nil {
		add(clicky.Text("plan: ", "text-gray-500").Append(string(execution.Plan.Status), "text-blue-600"))
	}
	lines = append(lines, renderRunDoD(execution)...)
	lines = append(lines, renderUndecodedResponse(execution)...)
	return strings.Join(lines, "\n")
}

// runStatusText reports the status through types.Status.Pretty, so a failed run
// is red with its own icon instead of the green every outcome used to print in.
// `keep` is not a todo status — it is the step declining to move one.
func runStatusText(step, status string) clickyapi.Text {
	prefix := clicky.Text(step+" finished — ", "text-gray-500")
	if status == lifecycle.OutcomeKeep || status == "" {
		return prefix.Append("status unchanged", "text-gray-600")
	}
	return prefix.Add(types.Status(status).Pretty())
}

// renderRunQuestions prints what an ask outcome is blocked on. An ask used to
// print "finished — ask" and swallow the questions themselves, leaving the
// operator to go and read the TODO to find out what was being asked.
func renderRunQuestions(execution *todos.ExecutionResult) []string {
	if execution.EndStatus != types.EndAsk || len(execution.Questions) == 0 {
		return nil
	}
	lines := []string{clicky.Text("Questions blocking progress:", "text-purple-600 font-bold").ANSI()}
	for i, question := range execution.Questions {
		lines = append(lines, clicky.Text(fmt.Sprintf("  %d. ", i+1), "text-gray-500").
			Append(strings.TrimSpace(question.Text), "text-purple-600").ANSI())
		if context := strings.TrimSpace(question.Context); context != "" {
			lines = append(lines, clicky.Text("     "+context, "text-gray-500").ANSI())
		}
		for _, option := range question.Options {
			lines = append(lines, clicky.Text("     - "+strings.TrimSpace(option), "text-gray-600").ANSI())
		}
	}
	return lines
}

// renderRunVerdict prints a triage run's verdict and the edits it asked gavel to
// make. The envelope is a request, not a record: this is printed whether or not
// the writes that follow it land, which is the point.
func renderRunVerdict(execution *todos.ExecutionResult) []string {
	env := execution.Triage
	if env == nil {
		return nil
	}
	line := clicky.Text("triage verdict: ", "text-gray-500").Append(string(env.Verdict), "text-blue-600 font-bold")
	if fields := todos.TriageEnvelopeFields(env); len(fields) > 0 {
		line = line.Append("  writes: "+strings.Join(fields, ", "), "text-gray-500")
	}
	if folds := env.RetirementTargets(); len(folds) > 0 {
		line = line.Append("  folds: "+strings.Join(folds, ", "), "text-gray-500")
	}
	if duplicate := strings.TrimSpace(env.DuplicateOf); duplicate != "" {
		line = line.Append("  duplicate of: "+duplicate, "text-gray-500")
	}
	lines := []string{line.ANSI()}
	if comment := strings.TrimSpace(env.Comment); comment != "" {
		lines = append(lines, clicky.Text("  "+comment, "text-gray-600").ANSI())
	}
	return lines
}

// renderRunDoD prints the definition-of-done verdict, with the reason the last
// verification report gave for it.
func renderRunDoD(execution *todos.ExecutionResult) []string {
	dod := execution.DoD
	if dod == nil || !dod.Ran {
		return nil
	}
	line := clicky.Text("definition of done: ", "text-gray-500")
	if dod.Passed {
		line = line.Append("passed", "text-green-600 font-bold")
	} else {
		line = line.Append("failed", "text-red-600 font-bold")
	}
	if dod.Report != nil {
		if reason := strings.TrimSpace(dod.Report.Reason); reason != "" {
			line = line.Append("  "+reason, "text-gray-500")
		}
	}
	return []string{line.ANSI()}
}

// renderUndecodedResponse prints the agent's reply verbatim when no envelope was
// captured from it. That is the case the whole report exists for: a reply that
// answered the question in the wrong shape used to leave nothing behind but the
// word "failed".
func renderUndecodedResponse(execution *todos.ExecutionResult) []string {
	response := strings.TrimSpace(execution.ResponseText)
	if execution.EndStatus != "" || response == "" {
		return nil
	}
	return []string{
		clicky.Text("The response did not match the step's schema; it is reproduced as it arrived:", "text-yellow-600").ANSI(),
		clicky.Text(response, "text-gray-600").ANSI(),
	}
}

// printRunResult writes the report, unless the run produced nothing to report.
func printRunResult(step string, outcome *lifecycle.StepOutcome, status string, runErr error) {
	if report := renderRunResult(step, outcome, status, runErr); report != "" {
		fmt.Println(report)
	}
}
