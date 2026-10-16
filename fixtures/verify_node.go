package fixtures

import (
	"encoding/json"
	"fmt"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"
)

// verifyNodeKind is the framework of a verify node whose result names no type.
const verifyNodeKind = "fixture"

// VerifyNodeFrom maps one fixture result onto a verify node. A result with children is
// a group: it carries the children's evidence and no state flags of its own,
// which is what captain's SummarizeNodes counts on when it tallies leaves; a
// step that failed on its own keeps that verdict through keepStepVerdict.
func VerifyNodeFrom(result *FixtureResult) api.VerifyNode {
	node := api.VerifyNode{
		Name:      result.Name,
		Framework: verifyNodeFramework(result),
		Message:   result.Error,
		Command:   result.Command,
		WorkDir:   result.CWD,
		Stdout:    result.Stdout,
		Stderr:    result.Stderr,
		Duration:  result.Duration,
		Context:   contextFrom(result),
		Detail:    ResultDetail(result),
		Summary:   result.Summary,
	}
	for _, child := range result.Children {
		if child == nil {
			continue
		}
		if child.Results != nil {
			node.Children = append(node.Children, VerifyNodeFrom(child.Results))
			continue
		}
		node.Children = append(node.Children, api.VerifyNode{Name: child.Name, Children: childNodes(child)})
	}
	if len(node.Children) == 0 {
		applyLeafStatus(&node, result)
		return node
	}
	own := api.VerifyNode{Message: result.Error}
	applyLeafStatus(&own, result)
	keepStepVerdict(&node, own)
	return node
}

func applyLeafStatus(node *api.VerifyNode, result *FixtureResult) {
	applyTaskStatus(node, result.Status)
	node.TimedOut = node.Failed && result.TimedOut
}

// keepStepVerdict stops a group from reading as passed when the step itself
// failed (a non-zero exit, a folded teardown) but none of its children did: the
// group's summary becomes its children's tally plus the step's own failed leaf,
// so captain's SummarizeNodes and StateForReport see the failure.
func keepStepVerdict(node *api.VerifyNode, own api.VerifyNode) {
	if !own.Failed {
		return
	}
	tally := api.SummarizeNodes(node.Children)
	if node.Summary != nil {
		tally = *node.Summary
	}
	if tally.Failed+tally.TimedOut > 0 {
		return
	}
	summary := api.AddSummaries(tally, api.SummarizeNodes([]api.VerifyNode{own}))
	node.Summary = &summary
}

func childNodes(node *FixtureNode) []api.VerifyNode {
	var nodes []api.VerifyNode
	for _, child := range node.Children {
		if child == nil {
			continue
		}
		if child.Results != nil {
			nodes = append(nodes, VerifyNodeFrom(child.Results))
			continue
		}
		nodes = append(nodes, api.VerifyNode{Name: child.Name, Children: childNodes(child)})
	}
	return nodes
}

// applyTaskStatus translates a clicky task status into the independent state flags
// clicky-ui's Test carries. An unrecognised status is reported as failed rather
// than as an unflagged row: captain's SummarizeNodes does not count a flagless
// leaf at all, so a status nobody mapped would vanish from the verdict.
func applyTaskStatus(node *api.VerifyNode, status task.Status) {
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

func verifyNodeFramework(result *FixtureResult) string {
	if result.Type != "" {
		return result.Type
	}
	return verifyNodeKind
}

// contextFrom carries the command-shaped and CEL-shaped evidence a reader needs
// to understand a failure without re-running it. Nil when the result recorded
// none of it, so a checklist item does not render an empty command panel.
func contextFrom(result *FixtureResult) *api.VerifyNodeContext {
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

// ResultDetail carries the evidence that has no field of its own on a verify node:
// the CEL evaluation trace, the run artifact a `yaml test` / `yaml lint` step
// points at, the changed files the node was scoped to, and a typed step's
// step/outputs/exports/wait/attempts/background record plus a background
// watcher's events. An unencodable detail is dropped rather than failing the
// report — the verdict is what matters, and the trace is an aid.
func ResultDetail(result *FixtureResult) json.RawMessage {
	detail := map[string]any{}
	if result.CELTrace != "" {
		detail["cel_trace"] = result.CELTrace
	}
	if result.Run != nil {
		detail["run"] = result.Run
	}
	if len(result.Traces) > 0 {
		detail["traces"] = result.Traces
	}
	if changed, ok := result.Metadata["changed_files"]; ok {
		detail["changed_files"] = changed
	}
	// A typed (YAML/JSON) step's record: which step it was, what it produced and
	// exported, how its wait went, every Execute attempt and what a background
	// watcher saw.
	for _, key := range []string{"step", "outputs", "exports", "wait", "attempts", "events", "background"} {
		if value, ok := result.Metadata[key]; ok {
			if key == "attempts" {
				if attempts, typed := value.([]AttemptReport); typed {
					value = attemptDetails(attempts)
				}
			}
			detail[key] = value
		}
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

func attemptDetails(attempts []AttemptReport) []map[string]any {
	details := make([]map[string]any, 0, len(attempts))
	for _, attempt := range attempts {
		entry := map[string]any{"attempt": attempt.Attempt, "status": attempt.Status, "durationMs": attempt.DurationMs}
		if attempt.Error != "" {
			entry["error"] = attempt.Error
		}
		if attempt.Result != nil {
			entry["result"] = VerifyNodeFrom(attempt.Result)
		}
		details = append(details, entry)
	}
	return details
}
