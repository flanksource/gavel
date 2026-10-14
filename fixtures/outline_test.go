package fixtures

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestOutlineParsesFixtureKindsWithoutExecuting(t *testing.T) {
	dir := t.TempDir()
	standardPath := filepath.Join(dir, "standard.fixture.md")
	aiPath := filepath.Join(dir, "ai.fixture.md")

	content := fmt.Sprintf(`---
build: "touch %s"
daemon: "touch %s"
exec: bash
skip: "touch %s"
---

# Suite

| Name | CLI | Args | CEL |
|------|-----|------|-----|
| table exec | bash | -c 'touch %s' | exitCode == 0 |

### command: command block

`+"```bash"+`
touch %s
`+"```"+`

## Standalone

`+"```bash"+`
touch %s
`+"```"+`

## Run tests

`+"```yaml test"+`
paths: [./pkg/...]
`+"```"+`

## Lint

`+"```lint"+`
linters: [golangci-lint]
`+"```"+`
`, filepath.Join(dir, "ran-build"), filepath.Join(dir, "ran-daemon"), filepath.Join(dir, "ran-skip"), filepath.Join(dir, "ran-table"), filepath.Join(dir, "ran-command"), filepath.Join(dir, "ran-standalone"))

	if err := os.WriteFile(standardPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write standard fixture: %v", err)
	}
	if err := os.WriteFile(aiPath, []byte(`---
ai:
  model: mock
verify:
  threshold: 80
---

# Verify feature

`+"```prompt"+`
Review without executing anything.
`+"```"+`

- [ ] Criterion one
- [x] Criterion two
- cel: json.score >= 80

## Lint

`+"```lint"+`
linters: [golangci-lint]
`+"```"+`

## Test

`+"```yaml test"+`
paths: [./pkg/...]
`+"```"+`

## Formatting

`+"```bash"+`
touch `+filepath.Join(dir, "ran-ai-exec")+`
`+"```"+`
`), 0o600); err != nil {
		t.Fatalf("write ai fixture: %v", err)
	}

	for _, name := range []string{"ai", RunnerKindTest, RunnerKindLint} {
		t.Cleanup(DefaultKinds.Replace(fakeKind{name: name, execute: func(context.Context, *StepContext) (StepOutcome, error) {
			t.Fatalf("%s kind should not be executed by Outline", name)
			return StepOutcome{}, nil
		}}))
	}

	report, err := Outline(OutlineOptions{
		Paths:   []string{standardPath, aiPath},
		WorkDir: dir,
	})
	if err != nil {
		t.Fatalf("Outline: %v", err)
	}

	if report.Files != 2 {
		t.Fatalf("Files = %d, want 2", report.Files)
	}
	if report.Fixtures != 9 {
		t.Fatalf("Fixtures = %d, want 9", report.Fixtures)
	}
	for kind, want := range map[string]int{
		"exec": 4,
		"test": 2,
		"lint": 2,
		"ai":   1,
	} {
		if got := report.Counts[kind]; got != want {
			t.Fatalf("Counts[%q] = %d, want %d; counts=%v", kind, got, want, report.Counts)
		}
	}
	ai := findOutlineKind(report.Tree, "ai")
	if ai == nil {
		t.Fatalf("no ai outline node found: %+v", report.Tree)
	}
	if len(ai.Children) != 2 {
		t.Fatalf("AI criteria children = %d, want 2: %+v", len(ai.Children), ai.Children)
	}
	if ai.Children[0].Type != "criterion" || ai.Children[0].Name != "Criterion one" {
		t.Fatalf("criterion child 0 = %+v", ai.Children[0])
	}

	matches, err := filepath.Glob(filepath.Join(dir, "ran-*"))
	if err != nil {
		t.Fatalf("glob sentinel files: %v", err)
	}
	if len(matches) > 0 {
		t.Fatalf("outline executed fixture side effects: %v", matches)
	}
}

func findOutlineKind(nodes []*OutlineNode, kind string) *OutlineNode {
	for _, node := range nodes {
		if node.Kind == kind {
			return node
		}
		if child := findOutlineKind(node.Children, kind); child != nil {
			return child
		}
	}
	return nil
}
