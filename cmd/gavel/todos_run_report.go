package main

import (
	"fmt"
	"strings"

	captainapi "github.com/flanksource/captain/pkg/api"
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
func renderRunResult(todoRef, step string, outcome *lifecycle.StepOutcome, status string, runErr error) string {
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
	lines = append(lines, renderRunEvent(execution)...)
	if execution.Plan != nil {
		add(clicky.Text("plan: ", "text-gray-500").Append(string(execution.Plan.Status), "text-blue-600"))
	}
	lines = append(lines, renderRunDoD(execution)...)
	lines = append(lines, renderRunWorkspace(todoRef, execution.Workspace)...)
	lines = append(lines, renderUndecodedResponse(execution)...)
	return strings.Join(lines, "\n")
}

// renderRunWorkspace prints where the run's work lives: the branch, the agent's
// own Setup..Head range, what teardown did with the worktree, each commit, and
// — while the branch still holds commits — how to land them.
func renderRunWorkspace(todoRef string, workspace *captainapi.WorkspaceRecord) []string {
	if workspace == nil || (workspace.Worktree == nil && len(workspace.Commits) == 0) {
		return nil
	}
	line := clicky.Text("workspace:", "text-gray-500")
	worktree := workspace.Worktree
	if worktree != nil {
		line = line.Append(" "+worktree.Branch, "text-blue-600 font-bold")
		if worktree.Setup != "" && worktree.Head != "" {
			line = line.Append("  "+todos.ShortSHA(worktree.Setup)+".."+todos.ShortSHA(worktree.Head), "text-gray-600")
		}
		if fate := todos.WorktreeFate(worktree); fate != "" {
			style := "text-gray-500"
			if worktree.Kept {
				style = "text-yellow-600"
			}
			line = line.Append("  worktree "+fate, style)
		}
	}
	lines := []string{line.ANSI()}
	for _, commit := range workspace.Commits {
		lines = append(lines, clicky.Text("  "+todos.ShortSHA(commit.SHA), "text-yellow-600").
			Append(" "+todos.CommitSubject(commit.Message)).ANSI())
	}
	if len(workspace.Commits) > 0 && (worktree == nil || !worktree.BranchDeleted) {
		lines = append(lines, clicky.Text("land it: ", "text-gray-500").
			Append("gavel todos land "+todoRef+" --merge|--pr", "text-blue-600").ANSI())
	}
	return lines
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

// renderRunEvent prints the event the run reported and the writes its todo
// tools made — the run's changes landed as it went, so this is the record of
// what changed, not a request.
func renderRunEvent(execution *todos.ExecutionResult) []string {
	if execution.Envelope == nil {
		return nil
	}
	line := clicky.Text("event: ", "text-gray-500").Append(execution.Envelope.Event, "text-blue-600 font-bold")
	if len(execution.ToolWrites) > 0 {
		writes := make([]string, len(execution.ToolWrites))
		for i, write := range execution.ToolWrites {
			writes[i] = fmt.Sprintf("%s %s", write.Name, write.Ref)
		}
		line = line.Append("  writes: "+strings.Join(writes, ", "), "text-gray-500")
	}
	return []string{line.ANSI()}
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
	if execution.Envelope != nil || response == "" {
		return nil
	}
	return []string{
		clicky.Text("The response did not match the step's schema; it is reproduced as it arrived:", "text-yellow-600").ANSI(),
		clicky.Text(response, "text-gray-600").ANSI(),
	}
}

// printRunResult writes the report, unless the run produced nothing to report.
func printRunResult(todoRef, step string, outcome *lifecycle.StepOutcome, status string, runErr error) {
	if report := renderRunResult(todoRef, step, outcome, status, runErr); report != "" {
		fmt.Println(report)
	}
}
