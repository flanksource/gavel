package testrunner

import (
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner/parsers"
)

func executionSnapshotToTests(snapshot fixtures.ExecutionSnapshot) []parsers.Test {
	if snapshot.Root == nil {
		return nil
	}
	tests := make([]parsers.Test, 0, len(snapshot.Root.Children))
	for _, child := range snapshot.Root.Children {
		tests = append(tests, executionNodeToTest(child))
	}
	return tests
}

func executionNodeToTest(node *fixtures.ExecutionNode) parsers.Test {
	test := parsers.Test{
		Name:      node.Name,
		Framework: parsers.Fixture,
		TaskID:    node.Key,
		Duration:  node.Duration,
		Message:   node.Error,
		Detail:    *node,
	}
	if node.Origin != nil {
		test.File = node.Origin.File
		test.Line = node.Origin.Line
	}
	for _, child := range node.Children {
		test.Children = append(test.Children, executionNodeToTest(child))
	}
	applyExecutionState(&test, node.State)
	if node.Total > 0 {
		test.Progress = &parsers.TestProgress{Phase: string(node.Kind), Done: node.Done, Total: node.Total}
	}
	return test
}

func applyExecutionState(test *parsers.Test, state fixtures.ExecutionState) {
	switch state {
	case fixtures.ExecutionQueued:
		test.Pending = true
	case fixtures.ExecutionRunning:
		test.Running = true
	case fixtures.ExecutionPassed:
		test.Passed = true
	case fixtures.ExecutionFailed, fixtures.ExecutionErrored:
		test.Failed = true
	case fixtures.ExecutionWarned:
		test.Warned = true
	case fixtures.ExecutionSkipped, fixtures.ExecutionCancelled:
		test.Skipped = true
	case fixtures.ExecutionTimedOut:
		test.Failed = true
		test.TimedOut = true
	}
}
