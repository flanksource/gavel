package todos

import (
	"fmt"
	"path/filepath"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"

	// Registers the fixture engine's `yaml test` / `yaml lint` step runners
	// (fixtures.TestStepRunner / LintStepRunner) via its init.
	_ "github.com/flanksource/gavel/fixtures/types"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/repomap"
	"github.com/goccy/go-yaml"
)

// The run loop's verifiers are gavel fixtures: each check is the same
// `yaml test` / `yaml lint` step a fixture file would declare, executed through
// the fixture engine's step runners. A failing step's children become the
// feedback the agent sees on the next iteration.

const maxVerifierFeedbackLines = 50

// fixtureVerifier adapts one fixture runner step to captain's agent.VerifyPlugin.
type fixtureVerifier struct {
	name    string
	fixture fixtures.FixtureTest
	runner  func(fixtures.FixtureTest, fixtures.RunOptions) fixtures.FixtureResult
	workDir string
}

func (v *fixtureVerifier) Name() string { return v.name }

func (v *fixtureVerifier) Verify(rc *agent.RunContext, _ *captainai.LoopIteration) (agent.Verdict, error) {
	res := v.runner(v.fixture, fixtures.RunOptions{WorkDir: v.workDir})
	switch res.Status {
	case task.StatusPASS, task.StatusSKIP, task.StatusWarning:
		return agent.Verdict{OK: true}, nil
	}
	return agent.Verdict{
		OK:       false,
		Reason:   res.Error,
		Feedback: fixtureFeedback(res),
	}, nil
}

// fixtureFeedback renders a failing fixture result as the compact failure
// summary fed back to the agent: the step error plus each failing child.
func fixtureFeedback(res fixtures.FixtureResult) string {
	var lines []string
	if res.Error != "" {
		lines = append(lines, fmt.Sprintf("%s: %s", res.Name, res.Error))
	}
	for _, child := range res.Children {
		if child == nil || child.Results == nil {
			continue
		}
		r := child.Results
		if r.Status == task.StatusPASS || r.Status == task.StatusSKIP {
			continue
		}
		line := "- " + r.Name
		if r.Error != "" {
			line += ": " + r.Error
		}
		lines = append(lines, line)
		if len(lines) >= maxVerifierFeedbackLines {
			lines = append(lines, fmt.Sprintf("… (%d more failures truncated)", len(res.Children)))
			break
		}
	}
	return strings.Join(lines, "\n")
}

// BuildCheckVerifiers resolves the todos' checks configuration (.gavel.yaml
// `checks`, frontmatter `checks:`, or the --check flag via force) into
// fixture-backed verify plugins for the run loop, plus the loop's iteration
// budget (initial run + feedback rounds). Empty when checks are disabled.
func BuildCheckVerifiers(workDir string, todosInGroup []*types.TODO, force bool) ([]agent.Plugin, int, error) {
	gitRoot := checksWorkDirFor(workDir, todosInGroup)
	project, err := verify.LoadGavelConfig(gitRoot)
	if err != nil {
		return nil, 0, fmt.Errorf("checks: load .gavel.yaml: %w", err)
	}
	cfg := types.ResolveAgentChecks(project.Checks, firstChecksConfig(todosInGroup), force)
	if !cfg.IsEnabled() {
		return nil, 0, nil
	}

	var plugins []agent.Plugin
	if cfg.Test != nil {
		v, err := newStepVerifier("checks:test", fixtures.RunnerKindTest, cfg.Test, fixtures.TestStepRunner, gitRoot)
		if err != nil {
			return nil, 0, err
		}
		plugins = append(plugins, v)
	}
	if cfg.Lint != nil {
		v, err := newStepVerifier("checks:lint", fixtures.RunnerKindLint, cfg.Lint, fixtures.LintStepRunner, gitRoot)
		if err != nil {
			return nil, 0, err
		}
		plugins = append(plugins, v)
	}
	return plugins, cfg.MaxIterations + 1, nil
}

// newStepVerifier marshals an options struct into the fixture step's YAML body
// (the same wire contract as a fixture file's `yaml test` / `yaml lint` fence).
func newStepVerifier(name, kind string, options any, runner func(fixtures.FixtureTest, fixtures.RunOptions) fixtures.FixtureResult, workDir string) (agent.Plugin, error) {
	if runner == nil {
		return nil, fmt.Errorf("%s: fixture step runner not registered (import gavel/fixtures/types)", name)
	}
	body, err := yaml.Marshal(options)
	if err != nil {
		return nil, fmt.Errorf("%s: marshal step options: %w", name, err)
	}
	return &fixtureVerifier{
		name:    name,
		workDir: workDir,
		runner:  runner,
		fixture: fixtures.FixtureTest{
			Name:      name,
			SourceDir: workDir,
			RunnerStep: &fixtures.RunnerStepSpec{
				Kind:   kind,
				Config: string(body),
			},
		},
	}, nil
}

// checksWorkDirFor resolves the directory checks run in: the git root of the
// group's working directory, mirroring how the commit step derives its dir.
func checksWorkDirFor(workDir string, todosInGroup []*types.TODO) string {
	dir := workDir
	for _, todo := range todosInGroup {
		if todo == nil || todo.CWD == "" {
			continue
		}
		if filepath.IsAbs(todo.CWD) {
			dir = filepath.Clean(todo.CWD)
		} else if workDir != "" {
			dir = filepath.Clean(filepath.Join(workDir, todo.CWD))
		} else {
			dir = filepath.Clean(todo.CWD)
		}
		break
	}
	if root := repomap.FindGitRoot(dir); root != "" {
		return root
	}
	return dir
}

// firstChecksConfig returns the first TODO frontmatter `checks` block in the
// group, or nil when none set one. Frontmatter overrides the project default.
func firstChecksConfig(todosInGroup []*types.TODO) *types.AgentChecksConfig {
	for _, todo := range todosInGroup {
		if todo != nil && todo.Checks != nil {
			return todo.Checks
		}
	}
	return nil
}
