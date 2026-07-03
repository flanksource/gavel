package types

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/lint"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/repomap"
	"github.com/goccy/go-yaml"
)

func init() {
	fixtures.TestStepRunner = RunTestStep
	fixtures.LintStepRunner = RunLintStep
}

// stepDisplay carries the fixture-tree rendering knobs a runner step supports,
// read from the same YAML body as the engine options. Failing children are
// shown by default; passing children only when show-passed is set.
type stepDisplay struct {
	ShowPassed *bool `yaml:"show-passed"`
	ShowFailed *bool `yaml:"show-failed"`
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

	// A fixture test step is headless: never launch the dashboard or recurse
	// back into fixture discovery (fixture → test → fixture).
	ro.UI = false
	ro.Fixtures = false
	ro.Updates = nil
	ro.OutputTee = nil
	if ro.WorkDir == "" {
		ro.WorkDir = resolveStepWorkDir(fixture, opts)
	}
	var summary parsers.TestSummary
	ro.SummaryOut = &summary

	raw, runErr := testrunner.Run(ro)
	tests, _ := raw.([]parsers.Test)

	result.CWD = ro.WorkDir
	result.Duration = time.Since(now)
	result.Metadata = map[string]interface{}{"summary": summary}
	result.Children = testChildNodes(tests, disp)

	if runErr != nil && summary.Total == 0 {
		return result.Errorf(runErr, "test run failed")
	}
	if summary.Failed > 0 {
		result.Status = task.StatusFAIL
		result.Error = fmt.Sprintf("%d/%d tests failed", summary.Failed, summary.Total)
	} else {
		result.Status = task.StatusPASS
	}
	return result
}

// RunLintStep runs a `yaml lint` fixture step: the YAML body is unmarshalled
// onto lint.Options and the lint engine runs. Each violation becomes a child
// node (shown by default), clean linters appear when show-passed is set.
func RunLintStep(fixture fixtures.FixtureTest, opts fixtures.RunOptions) fixtures.FixtureResult {
	now := time.Now()
	result := fixtures.FixtureResult{Name: fixture.Name, Type: "lint", Start: &now, Test: fixture}

	var lo lint.Options
	if err := yaml.Unmarshal([]byte(fixture.RunnerStep.Config), &lo); err != nil {
		return result.Errorf(err, "parse `yaml lint` options")
	}
	var disp stepDisplay
	_ = yaml.Unmarshal([]byte(fixture.RunnerStep.Config), &disp)

	lo.UI = false
	lo.OutputTee = nil
	if lo.WorkDir == "" {
		lo.WorkDir = resolveStepWorkDir(fixture, opts)
	}
	ctx := context.Background()
	lo.Context = ctx

	results, err := lint.Run(ctx, lo)
	if err != nil {
		return result.Errorf(err, "lint run failed")
	}

	result.CWD = lo.WorkDir
	result.Duration = time.Since(now)
	result.Children = lintChildNodes(results, disp)

	var violations int
	for _, lr := range results {
		if lr.Skipped {
			continue
		}
		violations += len(lr.Violations)
	}
	result.Metadata = map[string]interface{}{"violations": violations}
	if violations > 0 {
		result.Status = task.StatusFAIL
		result.Error = fmt.Sprintf("%d lint violations", violations)
	} else {
		result.Status = task.StatusPASS
	}
	return result
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

func testChildNodes(tests parsers.Tests, disp stepDisplay) []*fixtures.FixtureNode {
	var out []*fixtures.FixtureNode
	for _, t := range flattenTests(tests) {
		status := testStatus(t)
		switch status {
		case task.StatusFAIL, task.StatusERR:
			if !disp.showFailed() {
				continue
			}
		default:
			if !disp.showPassed() {
				continue
			}
		}
		name := t.Name
		if len(t.Suite) > 0 {
			name = strings.Join(t.Suite, " > ") + " > " + t.Name
		}
		child := &fixtures.FixtureNode{
			Name: name,
			Type: fixtures.TestNode,
			Results: &fixtures.FixtureResult{
				Name:     name,
				Type:     "test",
				Status:   status,
				Duration: t.Duration,
				Stdout:   t.Stdout,
				Stderr:   t.Stderr,
				Test:     fixtures.FixtureTest{Name: name},
			},
		}
		if status == task.StatusFAIL && t.Message != "" {
			child.Results.Error = t.Message
		}
		out = append(out, child)
	}
	return out
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

func testStatus(t parsers.Test) task.Status {
	switch {
	case t.Failed:
		return task.StatusFAIL
	case t.Skipped:
		return task.StatusSKIP
	case t.Warned:
		return task.StatusWarning
	default:
		return task.StatusPASS
	}
}

func lintChildNodes(results []*linters.LinterResult, disp stepDisplay) []*fixtures.FixtureNode {
	var out []*fixtures.FixtureNode
	for _, lr := range results {
		if lr.Skipped {
			continue
		}
		if len(lr.Violations) == 0 {
			if disp.showPassed() {
				out = append(out, leafResult(lr.Linter, task.StatusPASS, ""))
			}
			continue
		}
		if !disp.showFailed() {
			continue
		}
		for _, v := range lr.Violations {
			name := fmt.Sprintf("%s %s:%d", lr.Linter, v.File, v.Line)
			msg := lr.Linter
			if v.Message != nil {
				msg = *v.Message
			}
			out = append(out, leafResult(name, task.StatusFAIL, msg))
		}
	}
	return out
}

func leafResult(name string, status task.Status, errMsg string) *fixtures.FixtureNode {
	return &fixtures.FixtureNode{
		Name: name,
		Type: fixtures.TestNode,
		Results: &fixtures.FixtureResult{
			Name:   name,
			Type:   "lint",
			Status: status,
			Error:  errMsg,
			Test:   fixtures.FixtureTest{Name: name},
		},
	}
}
