package testrunner

import (
	"fmt"
	"path/filepath"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner/parsers"
)

func mergeFixtureResultDetails(tests []parsers.Test, tree *fixtures.FixtureNode, workDir string) {
	byLocation := make(map[string][]parsers.Test)
	location := func(test parsers.Test) string {
		file := test.File
		if !filepath.IsAbs(file) {
			file = filepath.Join(workDir, file)
		}
		return fmt.Sprintf("%s:%d:%s", filepath.Clean(file), test.Line, test.Name)
	}
	var collect func([]parsers.Test)
	collect = func(items []parsers.Test) {
		for _, test := range items {
			if len(test.Children) > 0 {
				collect(test.Children)
			} else if test.Context != nil {
				key := location(test)
				byLocation[key] = append(byLocation[key], test)
			}
		}
	}
	for _, child := range tree.Children {
		collect(fixtureNodeToTests(child))
	}
	var enrich func([]parsers.Test)
	enrich = func(items []parsers.Test) {
		for i := range items {
			test := &items[i]
			if test.Fixture != nil && test.File != "" {
				key := location(*test)
				if matches := byLocation[key]; len(matches) > 0 {
					result := matches[0]
					byLocation[key] = matches[1:]
					test.Stdout, test.Stderr, test.Context = result.Stdout, result.Stderr, result.Context
					test.Message, test.FixtureProfile = result.Message, result.FixtureProfile
					test.GoProfiles, test.SQLProfile = result.GoProfiles, result.SQLProfile
				}
			}
			enrich(test.Children)
		}
	}
	enrich(tests)
}

func mergeFixtureMeasurements(tests []parsers.Test, report *fixtures.BenchmarkReport) {
	measured := make(map[string]parsers.Test)
	var collect func([]parsers.Test)
	collect = func(items []parsers.Test) {
		for _, test := range items {
			if test.Fixture != nil {
				measured[test.Fixture.Key] = test
			}
			collect(test.Children)
		}
	}
	collect(fixtures.BenchmarkReportToTests(report))
	var enrich func([]parsers.Test)
	enrich = func(items []parsers.Test) {
		for i := range items {
			test := &items[i]
			if test.Fixture != nil {
				if result, ok := measured[test.Fixture.Key]; ok {
					test.Fixture.CommandMS = result.Fixture.CommandMS
					test.Fixture.Violations = result.Fixture.Violations
					if result.FixtureProfile != nil {
						test.FixtureProfile = result.FixtureProfile
					}
					if len(result.GoProfiles) > 0 {
						test.GoProfiles = result.GoProfiles
					}
					if result.SQLProfile != nil {
						test.SQLProfile = result.SQLProfile
					}
				}
			}
			enrich(test.Children)
		}
	}
	enrich(tests)
}

func executionSnapshotToTests(snapshot fixtures.ExecutionSnapshot) []parsers.Test {
	if snapshot.Root == nil {
		return nil
	}
	tests := make([]parsers.Test, 0, len(snapshot.Root.Children))
	lifecycle := parsers.Test{Name: "Fixture lifecycle", Framework: parsers.Fixture}
	for _, child := range snapshot.Root.Children {
		test := executionNodeToTest(child)
		switch child.Kind {
		case fixtures.ExecutionKindSetup, fixtures.ExecutionKindBuild, fixtures.ExecutionKindDaemon,
			fixtures.ExecutionKindDaemonStop, fixtures.ExecutionKindCleanup:
			lifecycle.Children = append(lifecycle.Children, test)
		default:
			tests = append(tests, test)
		}
	}
	if len(lifecycle.Children) > 0 {
		tests = append([]parsers.Test{lifecycle}, tests...)
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
	if len(node.Children) > 0 {
		summary := parsers.Tests(test.Children).Sum()
		test.Summary = &summary
	} else if node.Kind != fixtures.ExecutionKindFile && node.Kind != fixtures.ExecutionKindSection && node.Kind != fixtures.ExecutionKindTable {
		test.Fixture = &parsers.FixtureExecution{Key: node.Key, Kind: string(node.Kind), State: string(node.State)}
	}
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
