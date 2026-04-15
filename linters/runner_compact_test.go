package linters

import (
	"strings"
	"testing"

	"github.com/flanksource/clicky"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/models"
)

func TestStripANSILinter(t *testing.T) {
	cases := map[string]string{
		"plain text":         "plain text",
		"\x1b[31mred\x1b[0m": "red",
		"\x1b[1;38;2;107;113;128mbold gray\x1b[0m": "bold gray",
	}
	for in, want := range cases {
		if got := stripANSILinter(in); got != want {
			t.Errorf("stripANSILinter(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFirstViolationSnippet(t *testing.T) {
	msg := func(s string) *string { return &s }
	cases := []struct {
		name       string
		violations []models.Violation
		want       string
	}{
		{
			name: "file line rule message",
			violations: []models.Violation{{
				File:    "cmd/gavel/pr_list.go",
				Line:    254,
				Source:  "errcheck",
				Message: msg("Error return value of mb.Run is not checked"),
			}},
			want: "cmd/gavel/pr_list.go:254 errcheck: Error return value of mb.Run is not checked",
		},
		{
			name: "no line",
			violations: []models.Violation{{
				File:    "go.mod",
				Source:  "gomod",
				Message: msg("invalid module"),
			}},
			want: "go.mod gomod: invalid module",
		},
		{
			name: "message only",
			violations: []models.Violation{{
				Message: msg("something failed"),
			}},
			want: "something failed",
		},
		{
			name: "file only",
			violations: []models.Violation{{
				File: "broken.go",
			}},
			want: "broken.go",
		},
		{
			name:       "empty",
			violations: nil,
			want:       "",
		},
		{
			name: "picks first of many",
			violations: []models.Violation{
				{File: "a.go", Line: 1, Source: "rule1", Message: msg("m1")},
				{File: "b.go", Line: 2, Source: "rule2", Message: msg("m2")},
			},
			want: "a.go:1 rule1: m1",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := firstViolationSnippet(tc.violations); got != tc.want {
				t.Errorf("firstViolationSnippet(%+v) = %q, want %q", tc.violations, got, tc.want)
			}
		})
	}
}

func TestBuildLinterLabelCases(t *testing.T) {
	cases := []struct {
		name               string
		linter             string
		violationCount     int
		success            bool
		firstViolationLine string
		errorMessage       string
		wantContains       []string
		wantNotContains    []string
	}{
		{
			name:         "clean",
			linter:       "gofmt",
			success:      true,
			wantContains: []string{"gofmt"},
			wantNotContains: []string{
				"(failed)", "(1)", "err:",
			},
		},
		{
			name:               "success with violations",
			linter:             "golangci-lint",
			violationCount:     3,
			success:            true,
			firstViolationLine: "cmd/gavel/pr_list.go:254 errcheck: Error return value",
			wantContains: []string{
				"golangci-lint (3)",
				"cmd/gavel/pr_list.go:254",
				"errcheck",
			},
		},
		{
			name:         "failed with error",
			linter:       "eslint",
			success:      false,
			errorMessage: "Cannot find config file\nother line",
			wantContains: []string{
				"eslint (failed)",
				"err: Cannot find config file",
			},
			wantNotContains: []string{"other line"},
		},
		{
			name:         "failed with no error or violations",
			linter:       "vale",
			success:      false,
			errorMessage: "",
			wantContains: []string{"vale (failed)"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildLinterLabel(tc.linter, tc.violationCount, tc.success, tc.firstViolationLine, tc.errorMessage)
			for _, want := range tc.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("label missing %q: %q", want, got)
				}
			}
			for _, want := range tc.wantNotContains {
				if strings.Contains(got, want) {
					t.Errorf("label leaked %q: %q", want, got)
				}
			}
			if runeLenLinter(got) > labelBudget {
				t.Errorf("label exceeded budget: %q (%d > %d)", got, runeLenLinter(got), labelBudget)
			}
			if strings.Contains(got, "\x1b") {
				t.Errorf("label contained ANSI: %q", got)
			}
		})
	}
}

func TestJoinLabelLinter(t *testing.T) {
	cases := []struct {
		name    string
		prefix  string
		details string
		budget  int
		want    string
	}{
		{
			name:    "fits",
			prefix:  "golangci-lint (3)",
			details: "a.go:1 rule: msg",
			budget:  40,
			want:    "golangci-lint (3)  a.go:1 rule: msg",
		},
		{
			name:    "truncates details",
			prefix:  "eslint (failed)",
			details: strings.Repeat("x", 100),
			budget:  30,
			want:    "eslint (failed)  " + strings.Repeat("x", 30-len("eslint (failed)  ")-1) + "…",
		},
		{
			name:    "no details",
			prefix:  "gofmt",
			details: "",
			budget:  40,
			want:    "gofmt",
		},
		{
			name:    "strips ANSI",
			prefix:  "\x1b[1mgolangci-lint\x1b[0m (3)",
			details: "\x1b[31mfail\x1b[0m",
			budget:  40,
			want:    "golangci-lint (3)  fail",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := joinLabelLinter(tc.prefix, tc.details, tc.budget)
			if got != tc.want {
				t.Errorf("joinLabelLinter = %q, want %q", got, tc.want)
			}
			if runeLenLinter(got) > tc.budget {
				t.Errorf("exceeded budget %d: %q (%d)", tc.budget, got, runeLenLinter(got))
			}
		})
	}
}

type dryRunLinter struct{}

func (d dryRunLinter) Name() string { return "dummy" }
func (d dryRunLinter) Run(_ commonsContext.Context, _ *clicky.Task) ([]models.Violation, error) {
	return nil, nil
}
func (d dryRunLinter) DefaultIncludes() []string                   { return nil }
func (d dryRunLinter) DefaultExcludes() []string                   { return nil }
func (d dryRunLinter) SupportsJSON() bool                          { return false }
func (d dryRunLinter) JSONArgs() []string                          { return nil }
func (d dryRunLinter) SupportsFix() bool                           { return false }
func (d dryRunLinter) FixArgs() []string                           { return nil }
func (d dryRunLinter) ValidateConfig(_ *models.LinterConfig) error { return nil }
func (d dryRunLinter) DryRunCommand() (string, []string) {
	return "eslint", []string{"--format=json", "."}
}

func TestRunningCommandLabel(t *testing.T) {
	if got := runningCommandLabel(dryRunLinter{}); got != "eslint --format=json ." {
		t.Errorf("runningCommandLabel(dryRunLinter) = %q, want %q", got, "eslint --format=json .")
	}
}
