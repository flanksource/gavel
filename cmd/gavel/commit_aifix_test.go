package main

import (
	"testing"

	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
)

func TestUniqueViolationFiles(t *testing.T) {
	results := []*linters.LinterResult{
		{
			Violations: []models.Violation{
				{File: "a.go"},
				{File: "b.go"},
				{File: "a.go"}, // duplicate
				{File: ""},     // empty path skipped
			},
		},
		{Skipped: true, Violations: []models.Violation{{File: "skipped.go"}}}, // skipped linter ignored
		nil, // nil result tolerated
		{
			Violations: []models.Violation{
				{File: "c.go"},
				{File: "b.go"}, // cross-linter dedupe
			},
		},
	}

	got := uniqueViolationFiles(results)
	assert.Equal(t, []string{"a.go", "b.go", "c.go"}, got)
}

func TestCountViolations(t *testing.T) {
	results := []*linters.LinterResult{
		{Violations: []models.Violation{{File: "a"}, {File: "b"}}},
		{Skipped: true, Violations: []models.Violation{{File: "x"}}}, // skipped not counted
		nil,
		{Violations: []models.Violation{{File: "c"}}},
	}
	assert.Equal(t, 3, countViolations(results))
	assert.Equal(t, 0, countViolations(nil))
}

// commitGateRequest mirrors applyLintGate's mapping in the commit package.
// If either side drifts, the AI-fix re-lint will run a different linter set
// than the original commit gate — silent and very confusing. This test is a
// guard, not a behaviour test.
func TestCommitGateRequest(t *testing.T) {
	cases := []struct {
		name  string
		gates commitpkg.LintGates
		want  []string
	}{
		{"both", commitpkg.LintGates{FullLint: true, Secrets: true}, nil},
		{"full only", commitpkg.LintGates{FullLint: true, Secrets: false}, []string{"!betterleaks"}},
		{"secrets only", commitpkg.LintGates{FullLint: false, Secrets: true}, []string{"betterleaks"}},
		{"neither", commitpkg.LintGates{}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, commitGateRequest(tc.gates))
		})
	}
}
