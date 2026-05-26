package fixtures_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/types"
	"github.com/onsi/gomega"
)

func TestRunnerRunTestingMirrorsMarkdownTables(t *testing.T) {
	path := writeFixtureFile(t, `
---
exec: bash
args: ["-c", "echo {{.word}}"]
---

# Formula Fixtures

## Table Cases

| Name | word | CEL Validation |
|------|------|----------------|
| first row | alpha | stdout.contains("alpha") |
| second row | beta | stdout.contains("beta") |
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
		t.Fatalf("expected fixture tree")
	}

	table := findFixtureNode(tree, fixtures.TableNode, "Table 1")
	if table == nil {
		t.Fatalf("expected generated table node in fixture tree")
	}
	if len(table.Children) != 2 {
		t.Fatalf("expected two table row children, got %d", len(table.Children))
	}
	if table.Children[0].Origin == nil || table.Children[0].Origin.Kind != "table-row" || table.Children[0].Origin.RowIndex != 1 {
		t.Fatalf("expected first row origin metadata, got %#v", table.Children[0].Origin)
	}
}

func TestRunnerRunGomega(t *testing.T) {
	path := writeFixtureFile(t, `
# Command Fixture

`+"```bash"+`
echo gomega
`+"```"+`

- contains: gomega
`)

	runner, err := fixtures.NewRunner(fixtures.RunnerOptions{
		Paths:   []string{path},
		WorkDir: filepath.Dir(path),
	})
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}

	tree := runner.RunGomega(gomega.NewWithT(t))
	if tree == nil || tree.GetStats().Passed != 1 {
		t.Fatalf("expected one passing fixture, got %#v", tree)
	}
}

func TestRegistryReportsMissingExecRegistration(t *testing.T) {
	registry := fixtures.NewRegistry()
	_, err := registry.GetForFixture(fixtures.FixtureTest{
		Name: "missing exec registration",
		ExecFixtureBase: fixtures.ExecFixtureBase{
			Exec: "echo",
		},
	})
	if err == nil {
		t.Fatalf("expected missing exec registration error")
	}
	if !strings.Contains(err.Error(), `import _ "github.com/flanksource/gavel/fixtures/types"`) {
		t.Fatalf("expected actionable import guidance, got %q", err.Error())
	}
}

func writeFixtureFile(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixtures.md")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(body)+"\n"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

func findFixtureNode(node *fixtures.FixtureNode, typ fixtures.NodeType, name string) *fixtures.FixtureNode {
	if node == nil {
		return nil
	}
	if node.Type == typ && node.Name == name {
		return node
	}
	for _, child := range node.Children {
		if found := findFixtureNode(child, typ, name); found != nil {
			return found
		}
	}
	return nil
}
