package oxlint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// report mirrors a real `oxlint --format=json` document (oxlint 1.69.0).
const report = `{
	"diagnostics": [
		{
			"message": "Expected a conditional expression and instead saw an assignment",
			"code": "eslint(no-cond-assign)",
			"severity": "warning",
			"help": "Consider wrapping the assignment in additional parentheses",
			"url": "https://oxc.rs/docs/guide/usage/linter/rules/eslint/no-cond-assign.html",
			"filename": "src/sample.js",
			"labels": [{"span": {"offset": 17, "length": 1, "line": 2, "column": 7}}]
		},
		{
			"message": "` + "`debugger`" + ` statement is not allowed",
			"code": "eslint(no-debugger)",
			"severity": "error",
			"help": "Remove the debugger statement",
			"url": "https://oxc.rs/docs/guide/usage/linter/rules/eslint/no-debugger.html",
			"filename": "src/sample.js",
			"labels": [{"span": {"offset": 26, "length": 9, "line": 3, "column": 3}}]
		}
	],
	"number_of_files": 1,
	"number_of_rules": 95
}`

func TestParseViolations(t *testing.T) {
	o := &Oxlint{}
	o.WorkDir = "/workspace"

	violations, err := o.parseViolations([]byte(report))
	require.NoError(t, err)
	require.Len(t, violations, 2)

	require.Equal(t, "/workspace/src/sample.js", violations[0].File)
	require.Equal(t, 2, violations[0].Line)
	require.Equal(t, 7, violations[0].Column)
	require.Equal(t, "oxlint", violations[0].Source)
	require.Equal(t, "eslint/no-cond-assign", violations[0].Rule.Method)
	require.Equal(t, "oxlint", violations[0].Rule.Package)
	require.Equal(t, "Expected a conditional expression and instead saw an assignment", *violations[0].Message)

	require.Equal(t, 3, violations[1].Line)
	require.Equal(t, "eslint/no-debugger", violations[1].Rule.Method)

	require.Equal(t, 1, o.GetFileCount())
	require.Equal(t, 95, o.GetRuleCount())
}

func TestParseViolationsAbsoluteFilename(t *testing.T) {
	o := &Oxlint{}
	o.WorkDir = "/workspace"

	abs := `{"diagnostics": [{"message": "x", "code": "eslint(no-debugger)", "severity": "error", "filename": "/abs/path/file.ts", "labels": [{"span": {"line": 4, "column": 1}}]}], "number_of_files": 1, "number_of_rules": 1}`
	violations, err := o.parseViolations([]byte(abs))
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Equal(t, "/abs/path/file.ts", violations[0].File, "absolute filenames must not be re-rooted under workDir")
}

func TestParseViolationsNoLabels(t *testing.T) {
	o := &Oxlint{}
	o.WorkDir = "/workspace"

	noLabels := `{"diagnostics": [{"message": "x", "code": "eslint(no-debugger)", "severity": "error", "filename": "a.js", "labels": []}], "number_of_files": 1, "number_of_rules": 1}`
	violations, err := o.parseViolations([]byte(noLabels))
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Equal(t, 0, violations[0].Line, "missing labels leave location unset rather than erroring")
	require.Equal(t, 0, violations[0].Column)
}

func TestParseViolationsEmpty(t *testing.T) {
	o := &Oxlint{}
	o.WorkDir = "/workspace"

	violations, err := o.parseViolations([]byte(`{"diagnostics": [], "number_of_files": 0, "number_of_rules": 95}`))
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestParseViolationsInvalidJSON(t *testing.T) {
	o := &Oxlint{}
	o.WorkDir = "/workspace"

	_, err := o.parseViolations([]byte(`not json`))
	require.Error(t, err)
}

func TestRuleName(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"eslint(no-debugger)", "eslint/no-debugger"},
		{"typescript(no-explicit-any)", "typescript/no-explicit-any"},
		{"oxc(only-used-in-recursion)", "oxc/only-used-in-recursion"},
		{"bare-rule", "bare-rule"},
		{"", "unknown"},
		{"(no-plugin)", "no-plugin"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			require.Equal(t, tt.expected, ruleName(tt.code))
		})
	}
}

func TestBuildArgs(t *testing.T) {
	o := &Oxlint{}
	o.ForceJSON = true

	args := o.buildArgs()
	require.Equal(t, []string{"--format=json", "."}, args)

	o.Fix = true
	args = o.buildArgs()
	require.Contains(t, args, "--fix")

	o.Files = []string{"src/a.ts"}
	o.Fix = false
	args = o.buildArgs()
	require.Equal(t, []string{"--format=json", "src/a.ts"}, args)
}
