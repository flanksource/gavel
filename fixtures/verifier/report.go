package verifier

import (
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"

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
		report.Tests = append(report.Tests, fixtures.VerifyNodeFrom(&results[i]))
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
				mapped.Progress = &api.VerifyNodeProgress{Phase: string(child.Kind), Done: child.Done, Total: child.Total}
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
		node.Failed, node.TimedOut = true, true
	case fixtures.ExecutionRunning:
		node.Running = true
	default:
		node.Pending = true
	}
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
