package types

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/lint"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/runconfig"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/repomap"
	"github.com/goccy/go-yaml"
)

// stepDisplay carries the fixture-tree rendering knobs a runner step supports,
// read from the same YAML body as the engine options. Failing children are
// shown by default; passing children only when show-passed is set.
type stepDisplay struct {
	ShowPassed *bool `yaml:"show-passed"`
	ShowFailed *bool `yaml:"show-failed"`
	// skippedAreFindings shows skipped leaves as findings (a lint step's info
	// violations) instead of eliding them with the passes.
	skippedAreFindings bool
}

func (d stepDisplay) showPassed() bool { return d.ShowPassed != nil && *d.ShowPassed }
func (d stepDisplay) showFailed() bool { return d.ShowFailed == nil || *d.ShowFailed }

// RunTestStep runs a `yaml test` fixture step: the YAML body is unmarshalled
// onto testrunner.RunOptions and the test engine runs. Each test becomes a
// child node (failures always, passes when show-passed) so the fixture tree
// rolls up the results.
func RunTestStep(fixture fixtures.FixtureTest, opts fixtures.RunOptions) fixtures.FixtureResult {
	now := time.Now()
	result := fixtures.FixtureResult{Name: fixture.Name, Type: "test", Start: &now, Test: fixture}

	var ro testrunner.RunOptions
	if err := yaml.Unmarshal([]byte(fixture.RunnerStep.Config), &ro); err != nil {
		return result.Errorf(err, "parse `yaml test` options")
	}
	var disp stepDisplay
	_ = yaml.Unmarshal([]byte(fixture.RunnerStep.Config), &disp)
	fields, err := runconfig.YAMLFields([]byte(fixture.RunnerStep.Config))
	if err != nil {
		return result.Errorf(err, "read test option presence")
	}
	ro.OptionFields = fields
	ro.OptionFields["ui"], ro.OptionFields["fixtures"] = true, true

	// A fixture test step is headless: never launch the dashboard or recurse
	// back into fixture discovery (fixture → test → fixture).
	ro.UI = false
	ro.Context = opts.Context
	runtime, err := fixtureStepRuntime(opts)
	if err != nil {
		return result.Errorf(err, "fixture environment")
	}
	ro.Runtime = runtime.Merge(ro.Runtime)
	ro.Fixtures = false
	ro.OutputTee = nil
	if ro.WorkDir == "" {
		ro.WorkDir = resolveStepWorkDir(fixture, opts)
	}
	// A verification scoped to the change under review narrows the run to the
	// packages those files affect, on top of whatever selector the step declares.
	ro.ChangedFiles = append(ro.ChangedFiles, opts.Changed...)
	if len(opts.Changed) > 0 {
		ro.OptionFields["changed-files"] = true
	}
	var summary parsers.TestSummary
	ro.SummaryOut = &summary
	var progressDone chan struct{}
	var progressErr error
	if opts.Progress != nil {
		updates := make(chan []parsers.Test, 16)
		ro.Updates = updates
		progressDone = make(chan struct{})
		go func() {
			defer close(progressDone)
			for update := range updates {
				current := parsers.Tests(update).Sum()
				done := current.Total - current.Pending - current.Running
				if err := opts.Progress(done, current.Total); err != nil && progressErr == nil {
					progressErr = err
				}
			}
		}()
	}

	raw, runErr := testrunner.Run(ro)
	if progressDone != nil {
		<-progressDone
	}
	tests, _ := raw.([]parsers.Test)

	result.CWD = ro.WorkDir
	result.Duration = time.Since(now)
	result.Metadata = map[string]interface{}{"summary": summary}
	result.Children, result.Summary = runnerStepChildren(tests, disp)
	if progressErr != nil {
		return result.Errorf(progressErr, "report test progress")
	}

	artifact, err := saveRunArtifact(runArtifactOptions{
		WorkDir: ro.WorkDir,
		Kind:    "test",
		Label:   fixture.Name,
		Started: now,
		Tests:   tests,
		Err:     runErr,
	})
	if err != nil {
		return result.Errorf(err, "record run artifact")
	}
	result.Run = artifact

	if runErr != nil && summary.Total == 0 {
		return result.Errorf(runErr, "test run failed")
	}
	applyStepVerdict(&result, *result.Summary, "tests")
	return result
}

// RunLintStep runs a `yaml lint` fixture step: the YAML body is unmarshalled
// onto lint.Options and the lint engine runs. Each violation, of any severity,
// becomes a child node shown by default, as does a linter that failed to run;
// clean linters appear when show-passed is set.
func RunLintStep(fixture fixtures.FixtureTest, opts fixtures.RunOptions) fixtures.FixtureResult {
	now := time.Now()
	result := fixtures.FixtureResult{Name: fixture.Name, Type: "lint", Start: &now, Test: fixture}

	var lo lint.Options
	if err := yaml.Unmarshal([]byte(fixture.RunnerStep.Config), &lo); err != nil {
		return result.Errorf(err, "parse `yaml lint` options")
	}
	var disp stepDisplay
	_ = yaml.Unmarshal([]byte(fixture.RunnerStep.Config), &disp)
	fields, err := runconfig.YAMLFields([]byte(fixture.RunnerStep.Config))
	if err != nil {
		return result.Errorf(err, "read lint option presence")
	}
	lo.OptionFields = fields
	lo.OptionFields["ui"] = true

	lo.UI = false
	lo.OutputTee = nil
	if lo.WorkDir == "" {
		lo.WorkDir = resolveStepWorkDir(fixture, opts)
	}
	// A step that names its own files has already scoped itself; otherwise the
	// change under review is the scope.
	if len(lo.Files) == 0 {
		lo.Files = opts.Changed
	}
	if len(lo.Files) > 0 {
		lo.OptionFields["files"] = true
	}
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}
	lo.Context = ctx
	runtime, err := fixtureStepRuntime(opts)
	if err != nil {
		return result.Errorf(err, "fixture environment")
	}
	lo.Runtime = runtime.Merge(lo.Runtime)

	results, runErr := lint.Run(ctx, lo)

	result.CWD = lo.WorkDir
	result.Duration = time.Since(now)
	applyLintResults(&result, results, disp)

	artifact, err := saveRunArtifact(runArtifactOptions{
		WorkDir: lo.WorkDir,
		Kind:    "lint",
		Label:   fixture.Name,
		Started: now,
		Lint:    results,
		Err:     runErr,
	})
	if err != nil {
		return result.Errorf(err, "record run artifact")
	}
	result.Run = artifact

	if runErr != nil {
		return result.Errorf(runErr, "lint run failed")
	}
	return result
}

// applyLintResults records every linter that ran as the step's children,
// summary and verdict. Findings of every severity are shown; a linter that
// crashed or timed out is a failed row of its own, which fails the step without
// hiding the other linters' results.
func applyLintResults(result *fixtures.FixtureResult, results []*linters.LinterResult, disp stepDisplay) {
	disp.skippedAreFindings = true
	result.Children, result.Summary = runnerStepChildren(lintTests(results), disp)
	var violations int
	for _, lr := range results {
		if !lr.Skipped {
			violations += len(lr.Violations)
		}
	}
	result.Metadata = map[string]interface{}{"violations": violations}
	applyStepVerdict(result, *result.Summary, "lint results")
}

// resolveStepWorkDir defaults a runner step to the git root of the fixture, so
// paths in the YAML body are repo-root-relative and consistent with the
// daemon:/build: front-matter and `gavel test` / `gavel lint`. Falls back to the
// fixture's own directory (then opts.WorkDir) when it is not inside a git repo.
func resolveStepWorkDir(fixture fixtures.FixtureTest, opts fixtures.RunOptions) string {
	dir := fixture.SourceDir
	if dir == "" {
		dir = opts.WorkDir
	}
	if root := repomap.FindGitRoot(dir); root != "" {
		return root
	}
	return dir
}

// runnerStepChildren maps the engine's leaf results onto child nodes that keep
// each leaf's framework, and tallies every leaf — shown or elided — into the
// step's summary. Passed leaves, and skipped ones unless they are findings, are
// shown only with show-passed; every other state with show-failed. A leaf with no state flag is a
// placeholder row: neither shown nor counted, exactly as parsers.Test.Sum and
// captain's SummarizeNodes treat it.
func runnerStepChildren(tests parsers.Tests, disp stepDisplay) ([]*fixtures.FixtureNode, *api.VerifySummary) {
	var out []*fixtures.FixtureNode
	summary := &api.VerifySummary{}
	for _, t := range flattenTests(tests) {
		status, counts := testStatus(t)
		if counts.Total == 0 {
			continue
		}
		*summary = api.AddSummaries(*summary, counts)
		quiet := status == task.StatusPASS || (status == task.StatusSKIP && !disp.skippedAreFindings)
		if (quiet && !disp.showPassed()) || (!quiet && !disp.showFailed()) {
			continue
		}
		out = append(out, testChildNode(t, status, counts.TimedOut > 0))
	}
	return out, summary
}

func testChildNode(t parsers.Test, status task.Status, timedOut bool) *fixtures.FixtureNode {
	name := t.Name
	if len(t.Suite) > 0 {
		name = strings.Join(t.Suite, " > ") + " > " + t.Name
	}
	result := &fixtures.FixtureResult{
		Name:     name,
		Type:     string(t.Framework),
		Status:   status,
		TimedOut: timedOut,
		Duration: t.Duration,
		Stdout:   t.Stdout,
		Stderr:   t.Stderr,
		Test:     fixtures.FixtureTest{Name: name},
	}
	if status != task.StatusPASS {
		result.Error = t.Message
	}
	return &fixtures.FixtureNode{Name: name, Type: fixtures.TestNode, Results: result}
}

// flattenTests returns the leaf tests of a possibly-nested result set.
func flattenTests(tests parsers.Tests) []parsers.Test {
	var out []parsers.Test
	for _, t := range tests {
		if len(t.Children) > 0 {
			out = append(out, flattenTests(t.Children)...)
			continue
		}
		out = append(out, t)
	}
	return out
}

// testStatus maps a leaf's state flags onto the child's task status and the
// one summary bucket it counts in, with parsers.Test.Sum's precedence and a
// timeout ahead of everything: a timed-out test fails and counts as timed out.
// A flagless leaf yields an empty summary.
func testStatus(t parsers.Test) (task.Status, api.VerifySummary) {
	switch {
	case t.TimedOut:
		return task.StatusFAIL, api.VerifySummary{Total: 1, TimedOut: 1}
	case t.Failed:
		return task.StatusFAIL, api.VerifySummary{Total: 1, Failed: 1}
	case t.Warned:
		return task.StatusWarning, api.VerifySummary{Total: 1, Warned: 1}
	case t.Aborted, t.Skipped:
		return task.StatusSKIP, api.VerifySummary{Total: 1, Skipped: 1}
	case t.Passed:
		return task.StatusPASS, api.VerifySummary{Total: 1, Passed: 1}
	case t.Running:
		return task.StatusRunning, api.VerifySummary{Total: 1, Running: 1}
	case t.Pending:
		return task.StatusPending, api.VerifySummary{Total: 1, Pending: 1}
	default:
		return "", api.VerifySummary{}
	}
}

// lintTests converts every linter that ran into its test group; a skipped
// linter did not apply to the change and is left out.
func lintTests(results []*linters.LinterResult) parsers.Tests {
	var tests parsers.Tests
	for _, lr := range results {
		if !lr.Skipped {
			tests = append(tests, lr.Test())
		}
	}
	return tests
}

// applyStepVerdict stamps a runner step's status from its children's summary:
// a failure or timeout fails it, as does a result that never finished; a
// warning alone warns.
func applyStepVerdict(result *fixtures.FixtureResult, s api.VerifySummary, noun string) {
	switch {
	case s.Failed+s.TimedOut > 0:
		result.Status = task.StatusFAIL
		result.Error = fmt.Sprintf("%d/%d %s failed", s.Failed+s.TimedOut, s.Total, noun)
	case s.Pending+s.Running > 0:
		result.Status = task.StatusFAIL
		result.Error = fmt.Sprintf("%d/%d %s did not finish", s.Pending+s.Running, s.Total, noun)
	case s.Warned > 0:
		result.Status = task.StatusWarning
		result.Error = fmt.Sprintf("%d/%d %s warned", s.Warned, s.Total, noun)
	default:
		result.Status = task.StatusPASS
	}
}
