package linters

import (
	"testing"
	"time"

	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/stretchr/testify/assert"
)

func violation(file string, line int, rule, msg string, severity models.ViolationSeverity) models.Violation {
	v := mkViolation(file, line, rule, msg)
	v.Severity = severity
	return v
}

func TestLinterResultTestMapsViolationsBySeverity(t *testing.T) {
	const linter = "golangci-lint"
	result := &LinterResult{
		Linter:   linter,
		Success:  true,
		Duration: 2 * time.Second,
		Violations: []models.Violation{
			violation("pkg/a.go", 10, "errcheck", "error return value not checked", models.SeverityError),
			violation("pkg/a.go", 20, "unused", "func x is unused", models.SeverityWarning),
			violation("pkg/b.go", 3, "godot", "comment should end in a period", models.SeverityInfo),
			violation("pkg/c.go", 7, "govet", "unreachable code", ""),
			{File: "pkg/d.go", Message: ptr("no rule, no line")},
		},
	}

	assert.Equal(t, parsers.Test{
		Name:      linter,
		Framework: linter,
		Duration:  2 * time.Second,
		Children: parsers.Tests{
			{Name: "pkg/a.go:10 errcheck", Framework: linter, File: "pkg/a.go", Line: 10, Message: "error return value not checked", Failed: true},
			{Name: "pkg/a.go:20 unused", Framework: linter, File: "pkg/a.go", Line: 20, Message: "func x is unused", Warned: true},
			{Name: "pkg/b.go:3 godot", Framework: linter, File: "pkg/b.go", Line: 3, Message: "comment should end in a period", Skipped: true},
			{Name: "pkg/c.go:7 govet", Framework: linter, File: "pkg/c.go", Line: 7, Message: "unreachable code", Failed: true},
			{Name: "pkg/d.go", Framework: linter, File: "pkg/d.go", Message: "no rule, no line", Failed: true},
		},
	}, result.Test())
}

func TestLinterResultTestRejectsUnknownSeverity(t *testing.T) {
	result := &LinterResult{Linter: "eslint", Success: true, Violations: []models.Violation{
		violation("src/a.ts", 1, "no-console", "unexpected console", "critical"),
	}}

	leaf := result.Test().Children[0]
	assert.True(t, leaf.Failed, "an unmapped severity must never read as a pass")
	assert.Equal(t, `unknown lint severity "critical": unexpected console`, leaf.Message)
}

func TestLinterResultTestKeepsTheCrashOfARunWithViolations(t *testing.T) {
	const linter = "eslint"
	result := &LinterResult{Linter: linter, TimedOut: true, Error: "context deadline exceeded", Violations: []models.Violation{
		violation("src/a.ts", 4, "no-undef", "x is not defined", models.SeverityWarning),
	}}

	assert.Equal(t, parsers.Test{
		Name: linter, Framework: linter, Failed: true, TimedOut: true, Message: "context deadline exceeded",
		Children: parsers.Tests{
			{Name: "src/a.ts:4 no-undef", Framework: linter, File: "src/a.ts", Line: 4, Message: "x is not defined", Warned: true},
			{Name: linter, Framework: linter, Failed: true, TimedOut: true, Message: "context deadline exceeded"},
		},
	}, result.Test())
}

func TestLinterResultTestWithoutViolations(t *testing.T) {
	cases := map[string]struct {
		result LinterResult
		want   parsers.Test
	}{
		"clean run passes": {
			result: LinterResult{Linter: "ruff", Success: true},
			want:   parsers.Test{Name: "ruff", Framework: "ruff", Passed: true},
		},
		"skipped linter is skipped": {
			result: LinterResult{Linter: "tsc", Skipped: true, Error: "no tsconfig.json"},
			want:   parsers.Test{Name: "tsc", Framework: "tsc", Skipped: true, Message: "no tsconfig.json"},
		},
		"timed out linter fails and times out": {
			result: LinterResult{Linter: "eslint", TimedOut: true, Error: "context deadline exceeded"},
			want:   parsers.Test{Name: "eslint", Framework: "eslint", Failed: true, TimedOut: true, Message: "context deadline exceeded"},
		},
		"crashed linter fails": {
			result: LinterResult{Linter: "markdownlint", Error: "exit status 2"},
			want:   parsers.Test{Name: "markdownlint", Framework: "markdownlint", Failed: true, Message: "exit status 2"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.result.Test())
		})
	}
}

func ptr(s string) *string { return &s }
