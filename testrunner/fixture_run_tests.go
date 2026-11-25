package testrunner

import (
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// FixtureRunTests projects a finished fixture run, the tree runner.Run
// returned, onto the tests a snapshot of it holds.
func FixtureRunTests(runner *fixtures.Runner, tree *fixtures.FixtureNode, workDir string) ([]parsers.Test, error) {
	snapshot := runner.ExecutionSnapshot()
	tests, err := executionSnapshotToTests(snapshot)
	if err != nil {
		return nil, err
	}
	if err := mergeFixtureResultDetails(tests, tree, workDir, liveResultKeys(snapshot)); err != nil {
		return tests, err
	}
	mergeFixtureMeasurements(tests, runner.BenchmarkReport())
	return tests, nil
}
