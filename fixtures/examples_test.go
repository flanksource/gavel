package fixtures_test

import (
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/fixtures"
)

// TestExampleFixturesParse guards the shipped examples/ workflow templates: each
// must parse into a valid fixture tree carrying the runner steps it advertises,
// so the examples cannot silently rot as the parser evolves. It parses only (the
// templates reference project-specific commands like `make build` / curl), so it
// never runs the engines.
func TestExampleFixturesParse(t *testing.T) {
	files, err := filepath.Glob("../examples/*.fixture.md")
	if err != nil {
		t.Fatalf("glob examples: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no example fixtures found under ../examples")
	}

	// Expected step kinds per example, so a botched edit that drops a fence /
	// checklist (or mislabels a step) fails loudly.
	want := map[string]struct{ test, lint, ai bool }{
		"precommit.fixture.md":   {test: true, lint: true},
		"pre-release.fixture.md": {test: true, lint: true},
		"smoke-test.fixture.md":  {test: true},
		"help.fixture.md":        {},
	}

	seen := map[string]bool{}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			tree, err := fixtures.ParseMarkdownFixturesWithTree(f)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if tree == nil {
				t.Fatal("nil fixture tree")
			}

			var hasTest, hasLint, hasAI bool
			tree.Walk(func(n *fixtures.FixtureNode) {
				if n.Test == nil {
					return
				}
				if n.Test.IsTestStep() {
					hasTest = true
				}
				if n.Test.IsLintStep() {
					hasLint = true
				}
				if n.Test.IsAIStep() {
					hasAI = true
				}
			})

			exp, ok := want[filepath.Base(f)]
			if !ok {
				t.Fatalf("unlisted example %s; add it to the want map", filepath.Base(f))
			}
			seen[filepath.Base(f)] = true
			if hasTest != exp.test {
				t.Errorf("test step present=%v, want %v", hasTest, exp.test)
			}
			if hasLint != exp.lint {
				t.Errorf("lint step present=%v, want %v", hasLint, exp.lint)
			}
			if hasAI != exp.ai {
				t.Errorf("AI step present=%v, want %v", hasAI, exp.ai)
			}
		})
	}

	for name := range want {
		if !seen[name] {
			t.Errorf("expected example %s not found under ../examples", name)
		}
	}
}
