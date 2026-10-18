package testrunner

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// mergeFixtureResultDetails enriches the progress tree's tests with what each
// fixture's final result recorded: output, context, profiles and, merged
// into the progress node's detail, the typed step record (step, attempts, …)
// fixtures.ResultDetail projects for the verification report. A background
// step has two results at its location — the one its start landed on the step
// and the verdict its teardown attached to the file — and the teardown
// verdict, which carries what the step observed (a watcher's events), is laid
// over the start result.
func mergeFixtureResultDetails(tests []parsers.Test, tree *fixtures.FixtureNode, workDir string) error {
	location := func(test parsers.Test) string {
		file := test.File
		if !filepath.IsAbs(file) {
			file = filepath.Join(workDir, file)
		}
		return fmt.Sprintf("%s:%d:%s", filepath.Clean(file), test.Line, test.Name)
	}
	results, teardowns := map[string][]parsers.Test{}, map[string][]parsers.Test{}
	for _, child := range tree.Children {
		collectFixtureResults(child, "", func(test parsers.Test, teardown bool) {
			queue := results
			if teardown {
				queue = teardowns
			}
			queue[location(test)] = append(queue[location(test)], test)
		})
	}
	var enrich func([]parsers.Test) error
	enrich = func(items []parsers.Test) error {
		for i := range items {
			test := &items[i]
			if test.Fixture != nil && test.File != "" {
				key := location(*test)
				for _, queue := range []map[string][]parsers.Test{results, teardowns} {
					if matches := queue[key]; len(matches) > 0 {
						queue[key] = matches[1:]
						if err := applyFixtureResult(test, matches[0]); err != nil {
							return err
						}
					}
				}
			}
			if err := enrich(test.Children); err != nil {
				return err
			}
		}
		return nil
	}
	return enrich(tests)
}

// collectFixtureResults visits every result node under node as
// fixtureNodeToTests renders it, flagging the verdicts teardown attached
// (fixtures marks them Metadata["teardown"]).
func collectFixtureResults(node *fixtures.FixtureNode, inheritedFile string, visit func(test parsers.Test, teardown bool)) {
	if node.Results != nil {
		teardown, _ := node.Results.Metadata["teardown"].(bool)
		visit(fixtureNodeToTestsWithFile(node, inheritedFile)[0], teardown)
		return
	}
	file := fixtureNodeFile(node, inheritedFile)
	for _, child := range node.Children {
		collectFixtureResults(child, file, visit)
	}
}

func applyFixtureResult(test *parsers.Test, result parsers.Test) error {
	test.Stdout, test.Stderr, test.Context = result.Stdout, result.Stderr, result.Context
	test.Message, test.FixtureProfile = result.Message, result.FixtureProfile
	test.SQLProfile = result.SQLProfile
	return mergeDetail(test, result.Detail)
}

// mergeDetail lays a result's detail (fixtures.ResultDetail) over the
// progress node test.Detail holds, keeping the node's own fields.
func mergeDetail(test *parsers.Test, detail any) error {
	raw, ok := detail.(json.RawMessage)
	if detail == nil || ok && len(raw) == 0 {
		return nil
	}
	if !ok {
		return fmt.Errorf("fixture %s: result detail is %T, want json.RawMessage", test.Name, detail)
	}
	merged := map[string]json.RawMessage{}
	if test.Detail != nil {
		base, err := json.Marshal(test.Detail)
		if err != nil {
			return fmt.Errorf("fixture %s: encode progress detail: %w", test.Name, err)
		}
		if err := json.Unmarshal(base, &merged); err != nil {
			return fmt.Errorf("fixture %s: progress detail is not an object: %w", test.Name, err)
		}
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return fmt.Errorf("fixture %s: result detail is not an object: %w", test.Name, err)
	}
	for key, value := range fields {
		merged[key] = value
	}
	test.Detail = merged
	return nil
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

// executionDetail is an ExecutionNode as its test's display detail: everything
// but its key, which is already the test's TaskID, and its children, which are
// already the test's own.
type executionDetail struct {
	Name       string                  `json:"name"`
	Kind       fixtures.ExecutionKind  `json:"kind"`
	State      fixtures.ExecutionState `json:"state"`
	Origin     *fixtures.FixtureOrigin `json:"origin,omitempty"`
	StartedAt  *time.Time              `json:"started_at,omitempty"`
	FinishedAt *time.Time              `json:"finished_at,omitempty"`
	Duration   time.Duration           `json:"duration,omitempty"`
	Done       int                     `json:"done,omitempty"`
	Total      int                     `json:"total,omitempty"`
	Error      string                  `json:"error,omitempty"`
	PIDs       []int                   `json:"pids,omitempty"`
	Traces     []parsers.TraceResult   `json:"traces,omitempty"`
}

func executionNodeToTest(node *fixtures.ExecutionNode) parsers.Test {
	test := parsers.Test{
		Name:      node.Name,
		Framework: parsers.Fixture,
		TaskID:    node.Key,
		Duration:  node.Duration,
		Message:   node.Error,
		Detail: executionDetail{
			Name: node.Name, Kind: node.Kind, State: node.State, Origin: node.Origin,
			StartedAt: node.StartedAt, FinishedAt: node.FinishedAt, Duration: node.Duration,
			Done: node.Done, Total: node.Total, Error: node.Error, PIDs: node.PIDs, Traces: node.Traces,
		},
	}
	if node.Origin != nil {
		test.File = node.Origin.File
		test.Line = node.Origin.Line
	}
	if len(node.Traces) > 0 {
		test.Context = parsers.FixtureContext{Traces: node.Traces}
	}
	for _, child := range node.Children {
		test.Children = append(test.Children, executionNodeToTest(child))
	}
	applyExecutionState(&test, node.State)
	if len(node.Children) > 0 {
		summary := parsers.Tests(test.Children).Sum()
		test.Summary = &summary
	} else if node.Kind != fixtures.ExecutionKindFile && node.Kind != fixtures.ExecutionKindSection && node.Kind != fixtures.ExecutionKindTable {
		test.Fixture = &parsers.FixtureExecution{Key: node.Key, Kind: string(node.Kind), State: string(node.State), PIDs: node.PIDs}
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
