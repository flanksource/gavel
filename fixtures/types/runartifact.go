package types

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// maxRecordedFailures bounds the failures inlined on a RunArtifact. A
// FixtureResult ends up inside a prompt run's result_json and is re-served on
// every dashboard poll, so the full tree stays in the .gavel snapshot and only
// the head of the failure list travels with the verdict.
const maxRecordedFailures = 10

type runArtifactOptions struct {
	WorkDir string
	Kind    string // test | lint | test+lint
	Label   string // fixture step name, keeps same-second steps in distinct files
	Started time.Time
	Tests   []parsers.Test
	Lint    []*linters.LinterResult
	Err     error // engine failure, recorded rather than swallowed
}

// saveRunArtifact writes a runner step's engine output to the .gavel per-run
// artifact store and returns the summary + reference recorded on the fixture
// result. It deliberately uses SavePerRun, never Save: a verification run must
// not move .gavel/last.json or the sha/branch pointers, which belong to the
// last `gavel test` / `gavel lint` invocation.
func saveRunArtifact(opts runArtifactOptions) (*fixtures.RunArtifact, error) {
	snap := &testui.Snapshot{
		Metadata: &testui.SnapshotMetadata{
			Kind:    opts.Kind,
			Started: opts.Started,
			Ended:   time.Now().UTC(),
		},
		Status: testui.SnapshotStatus{LintRun: opts.Kind != "test"},
		Tests:  opts.Tests,
		Lint:   opts.Lint,
	}
	if opts.Err != nil {
		snap.Error = opts.Err.Error()
	}

	path, err := snapshots.SavePerRun(opts.WorkDir, snap, opts.Started, opts.Label)
	if err != nil {
		return nil, fmt.Errorf("write run artifact: %w", err)
	}

	sum := parsers.Tests(opts.Tests).Sum()
	artifact := &fixtures.RunArtifact{
		RunID:   strings.TrimSuffix(filepath.Base(path), ".json"),
		Path:    path,
		Kind:    opts.Kind,
		Total:   sum.Total,
		Passed:  sum.Passed,
		Failed:  sum.Failed,
		Warned:  sum.Warned,
		Skipped: sum.Skipped,
	}
	if opts.Err != nil {
		artifact.Error = opts.Err.Error()
	}

	failures := append(testFailures(opts.Tests), lintFailures(opts.Lint, artifact)...)
	if len(failures) > maxRecordedFailures {
		artifact.Truncated = len(failures) - maxRecordedFailures
		failures = failures[:maxRecordedFailures]
	}
	artifact.Failures = failures
	return artifact, nil
}

func testFailures(tests []parsers.Test) []fixtures.RunFailure {
	var out []fixtures.RunFailure
	for _, t := range flattenTests(tests) {
		if !t.Failed {
			continue
		}
		out = append(out, fixtures.RunFailure{
			Name:    t.Name,
			Suite:   strings.Join(t.Suite, " > "),
			Status:  "failed",
			Message: t.Message,
		})
	}
	return out
}

// lintFailures records one entry per violation and, as a side effect, the
// per-linter counts on the artifact — both walk the same results.
func lintFailures(results []*linters.LinterResult, artifact *fixtures.RunArtifact) []fixtures.RunFailure {
	var out []fixtures.RunFailure
	for _, lr := range results {
		if lr.Skipped {
			continue
		}
		artifact.Linters = append(artifact.Linters, lr.Linter)
		artifact.LintViolations += len(lr.Violations)
		for _, v := range lr.Violations {
			message := ""
			if v.Message != nil {
				message = *v.Message
			}
			out = append(out, fixtures.RunFailure{
				Name:    fmt.Sprintf("%s %s:%d", lr.Linter, v.File, v.Line),
				Status:  "violation",
				Message: message,
			})
		}
	}
	return out
}
