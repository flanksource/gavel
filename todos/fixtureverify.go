package todos

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
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

// The run loop's definition of done is a gavel fixture: the configured `checks`
// test/lint steps and the todo's `## Verification` section, run through the
// fixture engine and aggregated into one TestResults tree that the CEL retry
// predicate (celVerifier, celverify.go) reads. A failing/warned node becomes the
// feedback the agent sees on the next iteration.

const maxVerifierFeedbackLines = 50

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

// BuildCheckVerifiers assembles the run's single definition-of-done verifier and
// the loop's iteration budget (initial run + feedback rounds). The DoD is the
// todo's `## Verification` fixture (always, when present) plus the configured
// `checks` test/lint suite (.gavel.yaml `checks`, frontmatter `checks:`, enabled
// whenever the run carries a Workflow.Verify), aggregated into one TestResults
// tree; the resolved `checks.retry` CEL predicate decides verified vs unverified.
// It returns no plugins (and a zero budget) only when the todo has no definition
// of done at all — such a run ends `completed`.
//
// verifySpec is the run's api.Spec Workflow.Verify (nil = no verify): its presence
// force-enables the checks suite and its MaxIterations overrides the loop cap.
func BuildCheckVerifiers(workDir string, todosInGroup []*types.TODO, verifySpec *api.Verify) ([]agent.Verify, int, error) {
	gitRoot := checksWorkDirFor(workDir, todosInGroup)
	project, err := verify.LoadGavelConfig(gitRoot)
	if err != nil {
		return nil, 0, fmt.Errorf("checks: load .gavel.yaml: %w", err)
	}
	cfg := types.ResolveAgentChecks(project.Checks, firstChecksConfig(todosInGroup), verifySpec != nil)

	var steps []stepFixture
	// The `checks` test/lint suite is opt-in (enabled config / --check).
	if cfg.IsEnabled() {
		if cfg.Test != nil {
			s, err := newStepFixture("checks:test", fixtures.RunnerKindTest, cfg.Test, fixtures.TestStepRunner, gitRoot)
			if err != nil {
				return nil, 0, err
			}
			steps = append(steps, s)
		}
		if cfg.Lint != nil {
			s, err := newStepFixture("checks:lint", fixtures.RunnerKindLint, cfg.Lint, fixtures.LintStepRunner, gitRoot)
			if err != nil {
				return nil, 0, err
			}
			steps = append(steps, s)
		}
	}
	// The todo's own `## Verification` fixture is the definition of done — it
	// gates the loop whenever present, independent of the `checks` toggle.
	var verNodes []*fixtures.FixtureNode
	var criteria []string
	for _, todo := range todosInGroup {
		if todo == nil {
			continue
		}
		verNodes = append(verNodes, todo.Verification...)
		for _, c := range todo.AcceptanceCriteria {
			criteria = append(criteria, c.Text)
		}
	}

	// Acceptance criteria become an LLM checklist step whose per-item
	// {item, passed, message} results are exposed to the predicate as
	// results.checklist.
	aiStep, err := newChecklistStep(criteria, gitRoot)
	if err != nil {
		return nil, 0, err
	}

	if len(steps) == 0 && len(verNodes) == 0 && aiStep == nil {
		return nil, 0, nil
	}
	maxIter := cfg.MaxIterations
	// An explicit Workflow.Verify.maxIterations overrides the .gavel.yaml cap.
	if verifySpec != nil && verifySpec.MaxIterations > 0 {
		maxIter = verifySpec.MaxIterations
	}
	if maxIter <= 0 {
		maxIter = types.DefaultMaxCheckIterations
	}
	verifier := &celVerifier{
		name:      "definition-of-done",
		retryExpr: cfg.Retry,
		steps:     steps,
		nodes:     verNodes,
		aiStep:    aiStep,
		workDir:   gitRoot,
	}
	return []agent.Verify{verifier}, maxIter + 1, nil
}

// newChecklistStep builds the acceptance-criteria checklist step: an ai fixture
// that has the LLM evaluate each criterion against the change and return a
// per-item {item, passed, message} result. nil when there are no criteria.
func newChecklistStep(criteria []string, workDir string) (*stepFixture, error) {
	if len(criteria) == 0 {
		return nil, nil
	}
	if fixtures.AIStepRunner == nil {
		return nil, fmt.Errorf("acceptance-criteria checklist: AI step runner not registered (import gavel/fixtures/types)")
	}
	items := make([]fixtures.ChecklistItem, 0, len(criteria))
	for _, c := range criteria {
		items = append(items, fixtures.ChecklistItem{Text: c})
	}
	return &stepFixture{
		runner: fixtures.AIStepRunner,
		fixture: fixtures.FixtureTest{
			Name:      "acceptance-criteria",
			SourceDir: workDir,
			AIStep: &fixtures.AIStepSpec{
				Description: "Verify the implementation satisfies each acceptance criterion.",
				Criteria:    items,
			},
		},
	}, nil
}

// newStepFixture marshals an options struct into a fixture step's YAML body (the
// same wire contract as a fixture file's `yaml test` / `yaml lint` fence) and
// pairs it with its registered runner for the aggregate DoD verifier.
func newStepFixture(name, kind string, options any, runner func(fixtures.FixtureTest, fixtures.RunOptions) fixtures.FixtureResult, workDir string) (stepFixture, error) {
	if runner == nil {
		return stepFixture{}, fmt.Errorf("%s: fixture step runner not registered (import gavel/fixtures/types)", name)
	}
	body, err := yaml.Marshal(options)
	if err != nil {
		return stepFixture{}, fmt.Errorf("%s: marshal step options: %w", name, err)
	}
	return stepFixture{
		runner: runner,
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
