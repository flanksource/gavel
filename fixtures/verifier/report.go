package verifier

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"

	"github.com/flanksource/gavel/fixtures"
)

// Kind and Name label every report this package produces: the registry kind it
// was dispatched under, and the hook name a reader sees next to the verdict.
const (
	Kind = "fixture"
	Name = "fixture"
)

// Report maps one fixture run onto captain's typed verification report. It is
// the single FixtureResult → VerifyNode mapping in gavel: the CLI's
// `--format captain-verify-report`, the in-process definition-of-done verifier
// and `gavel todos check` all render from what this returns, so a field added
// here reaches every surface at once.
//
// The snapshot supplies the run's wall clock and, decisively, whether every
// declared step actually ran. A step still queued when the run ended never
// produced a verdict — that is a scheduling failure, not a red test — so the
// report is `errored` rather than `failed`, which is a host-stamped terminal
// state and therefore keeps the tree of what did run alongside it.
func Report(results []fixtures.FixtureResult, snapshot *fixtures.ExecutionSnapshot) api.VerifyReport {
	report := api.VerifyReport{Kind: Kind, Name: Name}
	if snapshot != nil {
		report.StartedAt, report.FinishedAt = snapshot.StartedAt, snapshot.EndedAt
		if snapshot.StartedAt != nil && snapshot.EndedAt != nil {
			report.Duration = snapshot.EndedAt.Sub(*snapshot.StartedAt)
		}
	}

	report.Tests = make([]api.VerifyNode, 0, len(results))
	for i := range results {
		report.Tests = append(report.Tests, nodeFrom(&results[i]))
	}
	report.Checklist = checklistFrom(results)
	report.Summary = api.SummarizeNodes(report.Tests)

	if stalled := neverRan(snapshot); len(stalled) > 0 {
		report.State = api.VerifyStateErrored
		report.Reason = fmt.Sprintf("%d definition-of-done step(s) never ran: %s",
			len(stalled), strings.Join(stalled, ", "))
		report.Feedback = report.Reason
		return report
	}

	report.State = api.StateForReport(report.Tests)
	report.Ran = report.State != api.VerifyStateQueued
	report.Passed = report.State == api.VerifyStatePassed
	report.Reason = reasonFor(report)
	report.Feedback = Feedback(report.Tests, report.Checklist)
	return report
}

// RunningReport renders an in-flight execution snapshot as a report a reader can
// draw while the checks are still going. Only the tree's shape and states are
// known mid-run, so the nodes carry no output — the verdict report replaces it.
//
// Passed stays false throughout: a snapshot records what has happened, and
// nothing has judged it yet. The state is the one the tree justifies (queued
// before anything starts, running while work is in flight), because a report
// whose state its own leaves contradict is rejected — including a live one.
//
// The second result is false once the tree has nothing left to run. Such a
// snapshot is no longer a running report: its leaves already add up to a
// verdict, and a report carrying that verdict's state with passed=false would
// contradict itself. Only Report may stamp it, from the results.
func RunningReport(snapshot fixtures.ExecutionSnapshot) (api.VerifyReport, bool) {
	report := api.VerifyReport{
		Kind: Kind, Name: Name,
		StartedAt: snapshot.StartedAt,
		Tests:     snapshotNodes(snapshot.Root),
	}
	report.Summary = api.SummarizeNodes(report.Tests)
	report.State = api.StateForReport(report.Tests)
	report.Ran = report.State != api.VerifyStateQueued
	return report, report.Summary.Running > 0 || report.Summary.Pending > 0
}

// snapshotNodes maps a progress tree onto verify nodes. The report's state is
// pinned to running by the caller, so the leaves are reported as running or
// pending rather than as verdicts that have not been reached yet.
func snapshotNodes(node *fixtures.ExecutionNode) []api.VerifyNode {
	if node == nil {
		return nil
	}
	nodes := make([]api.VerifyNode, 0, len(node.Children))
	for _, child := range node.Children {
		mapped := api.VerifyNode{Name: child.Name, Framework: string(child.Kind), Duration: child.Duration}
		if len(child.Children) > 0 {
			mapped.Children = snapshotNodes(child)
		} else {
			applySnapshotState(&mapped, child.State)
			if child.Total > 0 {
				mapped.Progress = &api.VerifyNodeProgress{Done: child.Done, Total: child.Total}
			}
		}
		nodes = append(nodes, mapped)
	}
	return nodes
}

func applySnapshotState(node *api.VerifyNode, state fixtures.ExecutionState) {
	switch state {
	case fixtures.ExecutionPassed:
		node.Passed = true
	case fixtures.ExecutionFailed, fixtures.ExecutionErrored:
		node.Failed = true
	case fixtures.ExecutionWarned:
		node.Warned = true
	case fixtures.ExecutionSkipped, fixtures.ExecutionCancelled:
		node.Skipped = true
	case fixtures.ExecutionTimedOut:
		node.TimedOut = true
	case fixtures.ExecutionRunning:
		node.Running = true
	default:
		node.Pending = true
	}
}

// nodeFrom maps one fixture result onto a verify node. A result with children is
// a group: it carries the children's evidence and no verdict of its own, which
// is what captain's SummarizeNodes counts on when it tallies leaves.
func nodeFrom(result *fixtures.FixtureResult) api.VerifyNode {
	node := api.VerifyNode{
		Name:      result.Name,
		Framework: framework(result),
		Message:   result.Error,
		Command:   result.Command,
		WorkDir:   result.CWD,
		Stdout:    result.Stdout,
		Stderr:    result.Stderr,
		Duration:  result.Duration,
		Context:   contextFrom(result),
		Detail:    detailFrom(result),
	}
	for _, child := range result.Children {
		if child == nil {
			continue
		}
		if child.Results != nil {
			node.Children = append(node.Children, nodeFrom(child.Results))
			continue
		}
		node.Children = append(node.Children, api.VerifyNode{Name: child.Name, Children: childNodes(child)})
	}
	if len(node.Children) == 0 {
		applyStatus(&node, result.Status)
	}
	return node
}

func childNodes(node *fixtures.FixtureNode) []api.VerifyNode {
	var nodes []api.VerifyNode
	for _, child := range node.Children {
		if child == nil {
			continue
		}
		if child.Results != nil {
			nodes = append(nodes, nodeFrom(child.Results))
			continue
		}
		nodes = append(nodes, api.VerifyNode{Name: child.Name, Children: childNodes(child)})
	}
	return nodes
}

// applyStatus translates a clicky task status into the independent state flags
// clicky-ui's Test carries. An unrecognised status is reported as failed rather
// than as an unflagged row: captain's SummarizeNodes does not count a flagless
// leaf at all, so a status nobody mapped would vanish from the verdict.
func applyStatus(node *api.VerifyNode, status task.Status) {
	switch status {
	case task.StatusPASS, task.StatusSuccess:
		node.Passed = true
	case task.StatusFAIL, task.StatusFailed, task.StatusERR:
		node.Failed = true
	case task.StatusWarning:
		node.Warned = true
	case task.StatusSKIP, task.StatusCancelled:
		node.Skipped = true
	case task.StatusRunning:
		node.Running = true
	case task.StatusPending:
		node.Pending = true
	default:
		node.Failed = true
		if node.Message == "" {
			node.Message = fmt.Sprintf("fixture reported no recognised status (%q)", string(status))
		}
	}
}

func framework(result *fixtures.FixtureResult) string {
	if result.Type != "" {
		return result.Type
	}
	return Kind
}

// contextFrom carries the command-shaped and CEL-shaped evidence a reader needs
// to understand a failure without re-running it. Nil when the result recorded
// none of it, so a checklist item does not render an empty command panel.
func contextFrom(result *fixtures.FixtureResult) *api.VerifyNodeContext {
	if result.Command == "" && result.CWD == "" && result.ExitCode == 0 &&
		result.CELExpression == "" && len(result.CELVars) == 0 &&
		result.Expected == nil && result.Actual == nil {
		return nil
	}
	return &api.VerifyNodeContext{
		Command:       result.Command,
		ExitCode:      result.ExitCode,
		Cwd:           result.CWD,
		CELExpression: result.CELExpression,
		CELVars:       result.CELVars,
		Expected:      result.Expected,
		Actual:        result.Actual,
	}
}

// detailFrom carries the evidence that has no field of its own on a verify node:
// the CEL evaluation trace, the run artifact a `yaml test` / `yaml lint` step
// points at, and the changed files the node was scoped to. An unencodable
// detail is dropped rather than failing the report — the verdict is what
// matters, and the trace is an aid.
func detailFrom(result *fixtures.FixtureResult) json.RawMessage {
	detail := map[string]any{}
	if result.CELTrace != "" {
		detail["cel_trace"] = result.CELTrace
	}
	if result.Run != nil {
		detail["run"] = result.Run
	}
	if len(result.Recordings) > 0 {
		detail["recordings"] = result.Recordings
	}
	if changed, ok := result.Metadata["changed_files"]; ok {
		detail["changed_files"] = changed
	}
	if len(detail) == 0 {
		return nil
	}
	raw, err := json.Marshal(detail)
	if err != nil {
		return nil
	}
	return raw
}

// checklistFrom collects the acceptance-criteria verdicts an `ai` step recorded
// under Metadata["checklist"], across every step in the run.
func checklistFrom(results []fixtures.FixtureResult) []api.VerifyChecklistItem {
	var checklist []api.VerifyChecklistItem
	for i := range results {
		for _, verdict := range checklistVerdicts(&results[i]) {
			passed := verdict.Passed
			checklist = append(checklist, api.VerifyChecklistItem{
				Item: verdict.Item, Passed: &passed, Message: verdict.Message,
			})
		}
	}
	return checklist
}

func checklistVerdicts(result *fixtures.FixtureResult) []fixtures.ChecklistResult {
	if result.Metadata == nil {
		return nil
	}
	verdicts, _ := result.Metadata["checklist"].([]fixtures.ChecklistResult)
	return verdicts
}

// reasonFor is the one-line verdict a reader sees before the tree.
func reasonFor(report api.VerifyReport) string {
	s := report.Summary
	if report.Passed {
		return fmt.Sprintf("%d/%d definition-of-done checks passed", s.Passed, s.Total)
	}
	var parts []string
	for _, bucket := range []struct {
		count int
		label string
	}{
		{s.Failed, "failed"}, {s.TimedOut, "timed out"}, {s.Warned, "warned"},
		{s.Skipped, "skipped"}, {s.Pending, "never ran"}, {s.Running, "still running"},
	} {
		if bucket.count > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", bucket.count, bucket.label))
		}
	}
	if len(parts) == 0 {
		return "no definition-of-done checks ran"
	}
	return fmt.Sprintf("%d definition-of-done checks: %s", s.Total, strings.Join(parts, ", "))
}
