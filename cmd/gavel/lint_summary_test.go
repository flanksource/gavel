package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
)

func mkViolation(file string, line int, rule, msg string) models.Violation {
	m := msg
	return models.Violation{
		File:    file,
		Line:    line,
		Rule:    &models.Rule{Method: rule},
		Message: &m,
	}
}

func TestLintSummaryView_GroupsByLinterAndRule(t *testing.T) {
	results := []*linters.LinterResult{
		{
			Linter:  "eslint",
			WorkDir: "/repo/frontend",
			Success: true,
			Violations: []models.Violation{
				mkViolation("src/a.ts", 1, "no-unused-vars", "x is unused"),
				mkViolation("src/b.ts", 3, "no-unused-vars", "y is unused"),
				mkViolation("src/c.ts", 5, "semi", "missing semicolon"),
			},
		},
		{
			Linter:  "eslint",
			WorkDir: "/repo/tools",
			Success: true,
			Violations: []models.Violation{
				mkViolation("script.ts", 2, "no-unused-vars", "z is unused"),
			},
		},
		{
			Linter:  "vale",
			WorkDir: "/repo",
			Skipped: true,
			Error:   "no vale config found in work dir",
		},
	}

	view := newLintSummaryView(results, 5)
	linterNodes := view.GetChildren()
	if len(linterNodes) != 2 {
		t.Fatalf("expected 2 linter groups, got %d", len(linterNodes))
	}

	// Linters sorted alphabetically: eslint first, vale second.
	eslintNode, ok := linterNodes[0].(*linterSummaryNode)
	if !ok {
		t.Fatalf("expected *linterSummaryNode, got %T", linterNodes[0])
	}
	if eslintNode.linter != "eslint" {
		t.Fatalf("expected first group=eslint, got %q", eslintNode.linter)
	}
	if len(eslintNode.violations) != 4 {
		t.Fatalf("expected eslint aggregated to 4 violations across 2 results, got %d", len(eslintNode.violations))
	}

	rules := eslintNode.GetChildren()
	if len(rules) != 2 {
		t.Fatalf("expected 2 rule groups under eslint, got %d", len(rules))
	}
	// Most frequent rule first: no-unused-vars (3) before semi (1).
	first := rules[0].(*ruleSummaryNode)
	if first.rule != "no-unused-vars" {
		t.Fatalf("expected first rule=no-unused-vars, got %q", first.rule)
	}
	if len(first.violations) != 3 {
		t.Fatalf("expected 3 no-unused-vars violations, got %d", len(first.violations))
	}

	valeNode := linterNodes[1].(*linterSummaryNode)
	if valeNode.linter != "vale" || !valeNode.skipped {
		t.Fatalf("expected vale to be shown as skipped, got %+v", valeNode)
	}
	if got := valeNode.GetChildren(); got != nil {
		t.Fatalf("skipped linter should have no children, got %d", len(got))
	}
}

func TestLintSummaryView_TruncatesAtLimit(t *testing.T) {
	var vs []models.Violation
	for i := 1; i <= 12; i++ {
		vs = append(vs, mkViolation(fmt.Sprintf("f%d.go", i), i, "errcheck", ""))
	}
	results := []*linters.LinterResult{{
		Linter:     "golangci-lint",
		WorkDir:    "/repo",
		Success:    true,
		Violations: vs,
	}}

	view := newLintSummaryView(results, 3)
	rules := view.GetChildren()[0].GetChildren()
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule group, got %d", len(rules))
	}
	locations := rules[0].GetChildren()
	// 3 locations + 1 "… N more" trailer
	if len(locations) != 4 {
		t.Fatalf("expected 3 locations + 1 trailer = 4 nodes, got %d", len(locations))
	}
	if _, ok := locations[3].(*moreLocationsNode); !ok {
		t.Fatalf("expected last child to be moreLocationsNode, got %T", locations[3])
	}
	trailer := locations[3].(*moreLocationsNode)
	if trailer.remaining != 9 {
		t.Fatalf("expected trailer remaining=9, got %d", trailer.remaining)
	}
}

func TestLintSummaryView_NoTrailerWhenUnderLimit(t *testing.T) {
	results := []*linters.LinterResult{{
		Linter:  "ruff",
		Success: true,
		Violations: []models.Violation{
			mkViolation("a.py", 1, "E501", "line too long"),
		},
	}}
	view := newLintSummaryView(results, 5)
	rule := view.GetChildren()[0].GetChildren()[0]
	children := rule.GetChildren()
	if len(children) != 1 {
		t.Fatalf("expected 1 location, got %d", len(children))
	}
	if _, ok := children[0].(*moreLocationsNode); ok {
		t.Fatal("did not expect moreLocationsNode when violations < limit")
	}
}

func TestLintSummaryView_DefaultLimitWhenZero(t *testing.T) {
	view := newLintSummaryView(nil, 0)
	if view.Limit != 5 {
		t.Fatalf("expected fallback limit=5, got %d", view.Limit)
	}
}

func TestLintSummaryView_CollapsesPerFile(t *testing.T) {
	// Same rule, same file, many lines -> a single child with count suffix.
	var vs []models.Violation
	for i := 1; i <= 10; i++ {
		vs = append(vs, mkViolation("same.ts", i, "TS1005", "',' expected."))
	}
	// One violation in a second file to confirm distinct files stay distinct.
	vs = append(vs, mkViolation("other.ts", 42, "TS1005", "',' expected."))
	results := []*linters.LinterResult{{
		Linter:     "tsc",
		Success:    true,
		Violations: vs,
	}}

	view := newLintSummaryView(results, 5)
	rule := view.GetChildren()[0].GetChildren()[0].(*ruleSummaryNode)
	if len(rule.violations) != 11 {
		t.Fatalf("expected 11 raw violations retained, got %d", len(rule.violations))
	}
	locations := rule.GetChildren()
	if len(locations) != 2 {
		t.Fatalf("expected 2 file-level children (one per file, no trailer), got %d", len(locations))
	}
	first, ok := locations[0].(*locationSummaryNode)
	if !ok {
		t.Fatalf("expected *locationSummaryNode, got %T", locations[0])
	}
	if first.violation.File != "same.ts" || first.count != 10 {
		t.Fatalf("expected same.ts with count=10, got file=%q count=%d", first.violation.File, first.count)
	}
	second := locations[1].(*locationSummaryNode)
	if second.violation.File != "other.ts" || second.count != 1 {
		t.Fatalf("expected other.ts with count=1, got file=%q count=%d", second.violation.File, second.count)
	}
}

func TestLintSummaryView_TrailerUsesFileCountNotViolationCount(t *testing.T) {
	// 3 files, rule limit=2 -> 2 file children + trailer of 1 (not N-2).
	vs := []models.Violation{
		mkViolation("a.ts", 1, "X", "boom"), mkViolation("a.ts", 2, "X", "boom"), mkViolation("a.ts", 3, "X", "boom"),
		mkViolation("b.ts", 1, "X", "boom"),
		mkViolation("c.ts", 1, "X", "boom"),
	}
	results := []*linters.LinterResult{{Linter: "tsc", Success: true, Violations: vs}}
	rule := newLintSummaryView(results, 2).GetChildren()[0].GetChildren()[0]
	children := rule.GetChildren()
	if len(children) != 3 {
		t.Fatalf("expected 2 file children + trailer = 3, got %d", len(children))
	}
	trailer, ok := children[2].(*moreLocationsNode)
	if !ok {
		t.Fatalf("expected trailer as last child, got %T", children[2])
	}
	if trailer.remaining != 1 {
		t.Fatalf("expected trailer remaining=1 (one file past the limit), got %d", trailer.remaining)
	}
}

func TestLintSummaryView_SurfacesLinterError(t *testing.T) {
	errMsg := "eslint configuration error:\nOops! Something went wrong! :(\n\nError: Cannot find package 'typescript-eslint'"
	results := []*linters.LinterResult{{
		Linter:  "eslint",
		WorkDir: "/repo",
		Success: false,
		Error:   errMsg,
	}}

	view := newLintSummaryView(results, 5)
	linterNodes := view.GetChildren()
	if len(linterNodes) != 1 {
		t.Fatalf("expected 1 linter node, got %d", len(linterNodes))
	}
	node, ok := linterNodes[0].(*linterSummaryNode)
	if !ok {
		t.Fatalf("expected *linterSummaryNode, got %T", linterNodes[0])
	}
	if node.skipped {
		t.Fatal("failed linter must not be rendered as skipped")
	}
	if node.errorMsg != errMsg {
		t.Fatalf("expected errorMsg to carry full text, got %q", node.errorMsg)
	}

	// Pretty should include the first error line inline; full text is a child.
	pretty := node.Pretty().ANSI()
	if !strings.Contains(pretty, "❌") {
		t.Fatalf("expected ❌ icon in Pretty(), got %q", pretty)
	}
	if !strings.Contains(pretty, "eslint configuration error:") {
		t.Fatalf("expected first-line preview in Pretty(), got %q", pretty)
	}

	children := node.GetChildren()
	if len(children) != 1 {
		t.Fatalf("expected 1 child (linterErrorNode), got %d", len(children))
	}
	errNode, ok := children[0].(*linterErrorNode)
	if !ok {
		t.Fatalf("expected *linterErrorNode child, got %T", children[0])
	}
	if errNode.message != errMsg {
		t.Fatalf("expected error child to carry full message")
	}
}

func TestLintSummaryView_ErrorAndViolationsCoexist(t *testing.T) {
	// Partial-failure case: linter ran, emitted some violations AND errored.
	results := []*linters.LinterResult{{
		Linter:     "tsc",
		WorkDir:    "/repo",
		Success:    false,
		Error:      "type checker crashed",
		Violations: []models.Violation{mkViolation("a.ts", 1, "TS1005", "',' expected.")},
	}}
	node := newLintSummaryView(results, 5).GetChildren()[0].(*linterSummaryNode)
	if node.errorMsg == "" {
		t.Fatal("expected errorMsg to be populated")
	}
	children := node.GetChildren()
	if len(children) != 2 {
		t.Fatalf("expected 2 children (error node + rule group), got %d", len(children))
	}
	if _, ok := children[0].(*linterErrorNode); !ok {
		t.Fatalf("expected linterErrorNode first, got %T", children[0])
	}
	if _, ok := children[1].(*ruleSummaryNode); !ok {
		t.Fatalf("expected ruleSummaryNode second, got %T", children[1])
	}
}

func TestLintSummaryView_HandlesViolationsWithoutRule(t *testing.T) {
	results := []*linters.LinterResult{{
		Linter:  "jscpd",
		Success: true,
		Violations: []models.Violation{
			{File: "a.go", Line: 1, Message: models.StringPtr("duplicate code")},
			{File: "b.go", Line: 2, Message: models.StringPtr("duplicate code")},
		},
	}}
	view := newLintSummaryView(results, 5)
	rules := view.GetChildren()[0].GetChildren()
	if len(rules) != 1 {
		t.Fatalf("expected single (no rule) bucket, got %d", len(rules))
	}
	if r := rules[0].(*ruleSummaryNode); r.rule != "(no rule)" {
		t.Fatalf("expected rule=(no rule), got %q", r.rule)
	}
}
