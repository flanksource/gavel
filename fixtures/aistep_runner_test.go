package fixtures_test

import (
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/types"
)

// TestRunnerRunsAIVerificationStep exercises the full public path that
// `gavel fixtures <file>` uses: an `ai:` front-matter file parses into an AI
// step, the runner dispatches it through the registered AIStepRunner hook, and
// the structured verify result (under MOCK) is asserted by the step's CEL.
func TestRunnerRunsAIVerificationStep(t *testing.T) {
	path := writeFixtureFile(t, `
---
ai:
  model: claude-code-sonnet
verify:
  threshold: 80
---

# Verify the change

Review the recent parser change.

`+"```prompt"+`
Focus on the new parser path.
`+"```"+`

## Acceptance Criteria
- [ ] tests-added
- [ ] no secrets are logged

- cel: json.score >= 80
`)

	runner, err := fixtures.NewRunner(fixtures.RunnerOptions{
		Paths:   []string{path},
		WorkDir: filepath.Dir(path),
	})
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}

	tree := runner.RunTesting(t)
	if tree == nil {
		t.Fatal("expected fixture tree")
	}
	if got := tree.GetStats().Passed; got != 1 {
		t.Fatalf("expected 1 passing AI verification step, got %d (stats=%#v)", got, tree.GetStats())
	}
}
