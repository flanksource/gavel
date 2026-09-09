package fixtures_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/types"
)

// TestRunnerRunsTestStep exercises the full `gavel fixtures` path for a
// `yaml test` fence: the block body unmarshals onto testrunner.RunOptions, the
// test engine runs a throwaway package with one passing and one failing test,
// and each test surfaces as a child node under the step. With show-passed the
// passing test is attached and counted too.
func TestRunnerRunsTestStep(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module runsteptest\n\ngo 1.21\n")
	writeFile(t, filepath.Join(dir, "sample_test.go"), `package sample

import "testing"

func TestPasses(t *testing.T) {}

func TestFails(t *testing.T) { t.Fatal("intentional failure") }
`)
	fixturePath := filepath.Join(dir, "steps.md")
	writeFile(t, fixturePath, "# Steps\n\n## Run sample\n\n"+
		"```yaml test\npaths: [.]\nframework: [go]\nshow-passed: true\npre-build: false\ntest-timeout: 60s\n```\n")

	runner, err := fixtures.NewRunner(fixtures.RunnerOptions{
		Paths:   []string{fixturePath},
		WorkDir: dir,
	})
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}

	// Run() returns the tree plus a non-nil error because a child test failed;
	// that is the expected signal, not a harness failure — inspect the tree.
	tree, _ := runner.Run()
	if tree == nil {
		t.Fatal("expected fixture tree")
	}

	stats := tree.GetStats()
	if stats.Failed < 1 {
		t.Fatalf("expected at least one failed test, stats=%#v", stats)
	}
	if stats.Passed < 1 {
		t.Fatalf("expected the passing test counted with show-passed, stats=%#v", stats)
	}

	fail := findChildByStatus(tree, task.StatusFAIL)
	if fail == nil {
		t.Fatalf("expected a failing child node under the test step")
	}
	pass := findChildByStatus(tree, task.StatusPASS)
	if pass == nil {
		t.Fatalf("expected a passing child node (show-passed: true)")
	}
}

// TestRunnerTestStepHidesPassesByDefault confirms passing tests do not appear
// as child nodes when show-passed is omitted (failures still do).
func TestRunnerTestStepHidesPassesByDefault(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module runsteptest\n\ngo 1.21\n")
	writeFile(t, filepath.Join(dir, "sample_test.go"), `package sample

import "testing"

func TestPasses(t *testing.T) {}

func TestFails(t *testing.T) { t.Fatal("intentional failure") }
`)
	fixturePath := filepath.Join(dir, "steps.md")
	writeFile(t, fixturePath, "# Steps\n\n## Run sample\n\n"+
		"```yaml test\npaths: [.]\nframework: [go]\npre-build: false\ntest-timeout: 60s\n```\n")

	runner, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{fixturePath}, WorkDir: dir})
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	tree, _ := runner.Run()
	if pass := findChildByStatus(tree, task.StatusPASS); pass != nil {
		t.Fatalf("passing tests should be hidden without show-passed, got node %q", pass.Name)
	}
	if fail := findChildByStatus(tree, task.StatusFAIL); fail == nil {
		t.Fatalf("failing test should always be shown")
	}
}

// TestRunnerRunsLintStep exercises the `yaml lint` path end to end: the body
// unmarshals onto lint.Options, the lint engine runs, and a clean target (no
// files any selected linter flags) produces a passing step with no children.
func TestRunnerRunsLintStep(t *testing.T) {
	dir := t.TempDir()
	fixturePath := filepath.Join(dir, "steps.md")
	// golangci-lint on a dir with no go.mod / no .golangci config is skipped,
	// so the step is deterministically clean regardless of what is installed.
	writeFile(t, fixturePath, "# Lint\n\n## Lint clean\n\n"+
		"```yaml lint\nlinters: [golangci-lint]\ntimeout: 60s\n```\n")

	runner, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{fixturePath}, WorkDir: dir})
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	tree, err := runner.Run()
	if err != nil {
		t.Fatalf("lint step run: %v", err)
	}
	stats := tree.GetStats()
	if stats.Failed != 0 || stats.Error != 0 {
		t.Fatalf("expected a clean lint step, stats=%#v", stats)
	}
	if fail := findChildByStatus(tree, task.StatusFAIL); fail != nil {
		t.Fatalf("clean lint step should have no violation children, got %q", fail.Name)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func findChildByStatus(node *fixtures.FixtureNode, status task.Status) *fixtures.FixtureNode {
	if node == nil {
		return nil
	}
	if node.Results != nil && node.Results.Status == status && node.Type == fixtures.TestNode {
		return node
	}
	for _, child := range node.Children {
		if found := findChildByStatus(child, status); found != nil {
			return found
		}
	}
	return nil
}
