package testui

import (
	"strings"
	"testing"

	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/formatters"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// snapshotWithTestsAndLint builds a Snapshot holding two suites (one with a
// failing leaf) and a single lint finding, so the tree assertions below can
// exercise both top-level sections.
func snapshotWithTestsAndLint() Snapshot {
	return Snapshot{
		Tests: []parsers.Test{
			{
				Name:   "pkg/foo",
				Passed: true,
				Children: []parsers.Test{
					{Name: "TestBar", Passed: true, File: "foo.go", Line: 10},
					{Name: "TestBaz", Failed: true, File: "foo.go", Line: 20, Message: "boom"},
				},
			},
			{
				Name:     "pkg/qux",
				Passed:   true,
				Children: []parsers.Test{{Name: "TestQux", Passed: true}},
			},
		},
		Lint: []*linters.LinterResult{
			{
				Linter: "golangci",
				Violations: []models.Violation{
					{
						File:    "foo.go",
						Line:    10,
						Rule:    &models.Rule{Method: "errcheck"},
						Message: models.StringPtr("unchecked error"),
					},
				},
			},
		},
	}
}

// childLabels renders each top-level child node's Pretty() to plain text so a
// test can assert on the section headers without depending on styling.
func childLabels(nodes []api.TreeNode) []string {
	labels := make([]string, 0, len(nodes))
	for _, n := range nodes {
		labels = append(labels, n.Pretty().String())
	}
	return labels
}

func TestSnapshotGetChildren_TopLevelSections(t *testing.T) {
	snap := snapshotWithTestsAndLint()
	children := snap.GetChildren()

	if len(children) != 2 {
		t.Fatalf("expected 2 top-level sections (tests, lint), got %d", len(children))
	}

	labels := childLabels(children)
	if !strings.Contains(labels[0], "Tests") {
		t.Errorf("first section should be Tests, got %q", labels[0])
	}
	if !strings.Contains(labels[1], "Lint") {
		t.Errorf("second section should be Lint, got %q", labels[1])
	}
}

func TestSnapshotGetChildren_TestsSectionNestsSuites(t *testing.T) {
	snap := snapshotWithTestsAndLint()
	children := snap.GetChildren()

	testsSection := children[0]
	suites := testsSection.GetChildren()
	if len(suites) != 2 {
		t.Fatalf("expected 2 suites under Tests section, got %d", len(suites))
	}

	fooLeaves := suites[0].GetChildren()
	if len(fooLeaves) != 2 {
		t.Fatalf("expected 2 leaf tests under pkg/foo, got %d", len(fooLeaves))
	}
}

func TestSnapshotGetChildren_LintSectionNestsFindings(t *testing.T) {
	snap := snapshotWithTestsAndLint()
	children := snap.GetChildren()

	lintSection := children[1]
	if len(lintSection.GetChildren()) == 0 {
		t.Fatal("expected lint section to have child nodes for the finding")
	}
}

// TestSnapshotMarkdownRendersTree drives the real clicky markdown formatter so
// the assertion proves the snapshot exports as a nested tree (indented section
// labels + the lint finding under a Lint subtree) rather than concatenated
// suites. It is the regression guard for the flat-output bug this change fixes.
func TestSnapshotMarkdownRendersTree(t *testing.T) {
	snap := snapshotWithTestsAndLint()
	out, err := formatters.NewFormatManager().FormatWithOptions(
		formatters.FormatOptions{Format: "markdown"}, snap)
	if err != nil {
		t.Fatalf("format markdown: %v", err)
	}

	for _, want := range []string{"Tests", "Lint", "TestBaz", "errcheck"} {
		if !strings.Contains(out, want) {
			t.Errorf("markdown output missing %q\n---\n%s", want, out)
		}
	}

	// A tree (not a flat concat) indents children beneath their section, so the
	// Lint header must sit at a shallower indent than the rule line under it.
	lintIdx := strings.Index(out, "Lint")
	ruleIdx := strings.Index(out, "errcheck")
	if lintIdx < 0 || ruleIdx < 0 || ruleIdx < lintIdx {
		t.Fatalf("expected errcheck rule to render after the Lint section header\n%s", out)
	}
	if indentOfLine(out, ruleIdx) <= indentOfLine(out, lintIdx) {
		t.Errorf("expected lint finding to be indented under the Lint section\n%s", out)
	}
}

// indentOfLine returns the count of leading whitespace runes on the line that
// contains byte offset pos.
func indentOfLine(s string, pos int) int {
	start := strings.LastIndexByte(s[:pos], '\n') + 1
	indent := 0
	for _, r := range s[start:] {
		if r == ' ' || r == '\t' {
			indent++
			continue
		}
		break
	}
	return indent
}

func TestSnapshotGetChildren_OmitsEmptySections(t *testing.T) {
	snap := Snapshot{
		Tests: []parsers.Test{
			{Name: "pkg/foo", Passed: true, Children: []parsers.Test{{Name: "TestBar", Passed: true}}},
		},
	}
	children := snap.GetChildren()
	if len(children) != 1 {
		t.Fatalf("expected only the Tests section when no lint ran, got %d", len(children))
	}
	if !strings.Contains(children[0].Pretty().String(), "Tests") {
		t.Errorf("the single section should be Tests, got %q", children[0].Pretty().String())
	}
}
