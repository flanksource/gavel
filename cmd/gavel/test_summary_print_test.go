package main

import (
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner/parsers"
)

func TestFormatTestRunSummary_EmptyIsNoOp(t *testing.T) {
	got := formatTestRunSummary(parsers.TestSummary{}, nil)
	if got != "" {
		t.Errorf("want empty string for zero summary + no lint, got %q", got)
	}
}

func TestFormatTestRunSummary_TestsOnlyIncludesHeaderAndCounts(t *testing.T) {
	summary := parsers.TestSummary{
		Total:    3,
		Passed:   1,
		Failed:   1,
		Skipped:  1,
		Duration: 250 * time.Millisecond,
	}
	got := formatTestRunSummary(summary, nil)
	wantSubstrings := []string{"Test summary:", "passed", "failed", "skipped", "total"}
	for _, s := range wantSubstrings {
		if !strings.Contains(got, s) {
			t.Errorf("want output to contain %q, got:\n%s", s, got)
		}
	}
	if strings.Contains(got, "Lint summary") {
		t.Errorf("no lint results were provided but output contains lint tree:\n%s", got)
	}
}

func TestFormatTestRunSummary_LintResultsRenderLintTree(t *testing.T) {
	lintResults := []*linters.LinterResult{
		{
			Linter:  "golangci-lint",
			WorkDir: ".",
			Violations: []models.Violation{
				{File: "a.go"},
			},
		},
	}
	got := formatTestRunSummary(parsers.TestSummary{}, lintResults)
	if !strings.Contains(got, "Lint summary") {
		t.Errorf("want lint tree header, got:\n%s", got)
	}
	if strings.Contains(got, "Test summary:") {
		t.Errorf("zero test summary was provided but output contains test summary header:\n%s", got)
	}
}

func TestFormatTestRunSummary_CombinedIncludesBoth(t *testing.T) {
	summary := parsers.TestSummary{Total: 1, Passed: 1}
	lintResults := []*linters.LinterResult{{
		Linter:     "golangci-lint",
		WorkDir:    ".",
		Violations: []models.Violation{{File: "x.go"}},
	}}
	got := formatTestRunSummary(summary, lintResults)
	if !strings.Contains(got, "Test summary:") {
		t.Errorf("missing test summary header in combined output:\n%s", got)
	}
	if !strings.Contains(got, "Lint summary") {
		t.Errorf("missing lint summary header in combined output:\n%s", got)
	}
	if strings.Index(got, "Test summary:") > strings.Index(got, "Lint summary") {
		t.Errorf("test summary must appear before lint tree:\n%s", got)
	}
}
