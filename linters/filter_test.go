package linters

import (
	"testing"

	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
	"github.com/stretchr/testify/assert"
)

func TestFilterIgnoredViolations(t *testing.T) {
	mkViolation := func(source, rule, file string) models.Violation {
		v := models.Violation{Source: source, File: file}
		if rule != "" {
			v.Rule = &models.Rule{Method: rule}
		}
		return v
	}

	tests := []struct {
		name       string
		violations []models.Violation
		rules      []verify.LintIgnoreRule
		wantKept   int
		wantFiltered int
	}{
		{
			name: "rule only filter",
			violations: []models.Violation{
				mkViolation("golangci-lint", "errcheck", "a.go"),
				mkViolation("golangci-lint", "unused", "b.go"),
				mkViolation("golangci-lint", "errcheck", "c.go"),
			},
			rules:        []verify.LintIgnoreRule{{Rule: "errcheck"}},
			wantKept:     1,
			wantFiltered: 2,
		},
		{
			name: "source only filter",
			violations: []models.Violation{
				mkViolation("eslint", "no-unused-vars", "a.ts"),
				mkViolation("ruff", "F401", "b.py"),
			},
			rules:        []verify.LintIgnoreRule{{Source: "eslint"}},
			wantKept:     1,
			wantFiltered: 1,
		},
		{
			name: "rule + file filter",
			violations: []models.Violation{
				mkViolation("golangci-lint", "errcheck", "pkg/foo.go"),
				mkViolation("golangci-lint", "errcheck", "pkg/bar.go"),
			},
			rules:        []verify.LintIgnoreRule{{Rule: "errcheck", File: "pkg/foo.go"}},
			wantKept:     1,
			wantFiltered: 1,
		},
		{
			name: "no matches keeps all",
			violations: []models.Violation{
				mkViolation("golangci-lint", "errcheck", "a.go"),
			},
			rules:        []verify.LintIgnoreRule{{Rule: "unused"}},
			wantKept:     1,
			wantFiltered: 0,
		},
		{
			name: "all filtered",
			violations: []models.Violation{
				mkViolation("eslint", "no-var", "a.ts"),
				mkViolation("eslint", "no-let", "b.ts"),
			},
			rules:        []verify.LintIgnoreRule{{Source: "eslint"}},
			wantKept:     0,
			wantFiltered: 2,
		},
		{
			name:         "empty rules no-op",
			violations:   []models.Violation{mkViolation("ruff", "F401", "a.py")},
			rules:        nil,
			wantKept:     1,
			wantFiltered: 0,
		},
		{
			name: "glob file pattern",
			violations: []models.Violation{
				mkViolation("golangci-lint", "errcheck", "pkg/sub/deep.go"),
				mkViolation("golangci-lint", "errcheck", "cmd/main.go"),
			},
			rules:        []verify.LintIgnoreRule{{Rule: "errcheck", File: "pkg/**/*.go"}},
			wantKept:     1,
			wantFiltered: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &LinterResult{Violations: tt.violations}
			filtered := FilterIgnoredViolations([]*LinterResult{result}, tt.rules)
			assert.Equal(t, tt.wantFiltered, filtered)
			assert.Len(t, result.Violations, tt.wantKept)
		})
	}
}
