package prwatch

import (
	"strings"
	"testing"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos"
	todotypes "github.com/flanksource/gavel/todos/types"
)

func TestBuildPRStatusVerificationBlock(t *testing.T) {
	block := BuildPRStatusVerificationBlock(PRStatusVerification{
		PRNumber:   123,
		Repo:       "flanksource/gavel",
		CommentIDs: []int64{3, 1, 3},
		Actions:    []string{"*"},
	})

	want := "gavel pr status 123 --repo flanksource/gavel --comments 1,3 --actions '*'"
	if !strings.Contains(block, want) {
		t.Fatalf("block missing command %q:\n%s", want, block)
	}
	if !strings.Contains(block, "```exec") {
		t.Fatalf("block should be an exec fixture:\n%s", block)
	}
}

func TestBuildPRStatusVerificationBlockQuotesActions(t *testing.T) {
	block := BuildPRStatusVerificationBlock(PRStatusVerification{
		PRNumber: 9,
		Actions:  []string{"Go Test / pg 15"},
	})

	if !strings.Contains(block, "--actions 'Go Test / pg 15'") {
		t.Fatalf("action selector was not shell-quoted:\n%s", block)
	}
}

func TestBuildPRStatusVerificationBlockEmptyWithoutSelectors(t *testing.T) {
	if got := BuildPRStatusVerificationBlock(PRStatusVerification{PRNumber: 1}); got != "" {
		t.Fatalf("expected empty block without selectors, got %q", got)
	}
}

// TestPRStatusVerificationParsesToRunnableExec is the coverage that was missing:
// it parses the generated block the way a real TODO body is parsed and asserts a
// runnable exec node is produced. A "### command:" heading would have made the
// ```exec``` fence a skipped, non-runnable block.
func TestPRStatusVerificationParsesToRunnableExec(t *testing.T) {
	block := BuildPRStatusVerificationBlock(PRStatusVerification{
		PRNumber: 123,
		Repo:     "flanksource/gavel",
		Actions:  []string{"lint"},
	})
	if strings.Contains(block, "### command:") {
		t.Fatalf("block must not use a `### command:` heading (skips the exec fence):\n%s", block)
	}

	body := UpsertPRStatusVerification("Task body.", PRStatusVerification{
		PRNumber: 123,
		Repo:     "flanksource/gavel",
		Actions:  []string{"lint"},
	})
	todo, err := todos.ParseTODOContent("verify pr", body, t.TempDir(), todotypes.TODOFrontmatter{})
	if err != nil {
		t.Fatalf("parse TODO: %v", err)
	}

	var commands []string
	var walk func(*fixtures.FixtureNode)
	walk = func(node *fixtures.FixtureNode) {
		if node == nil {
			return
		}
		if node.Test != nil {
			if eb := node.Test.ExecBase(); !eb.IsEmpty() {
				commands = append(commands, strings.Join(append([]string{eb.Exec}, eb.Args...), " "))
			}
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	for _, node := range todo.Verification {
		walk(node)
	}

	if len(commands) != 1 {
		t.Fatalf("expected exactly one runnable exec node, got %d: %v", len(commands), commands)
	}
	if !strings.Contains(commands[0], "gavel pr status 123") {
		t.Fatalf("runnable exec node missing the pr-status command: %q", commands[0])
	}
}
