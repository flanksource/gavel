package jscpd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseViolations(t *testing.T) {
	j := &JSCPD{}
	j.WorkDir = "/workspace"

	report := `{
		"statistics": {},
		"duplicates": [
			{
				"format": "go",
				"lines": 15,
				"fragment": "func doThing() {}",
				"tokens": 120,
				"firstFile": {
					"name": "pkg/handler.go",
					"start": 10,
					"end": 25,
					"startLoc": {"line": 10, "column": 1, "position": 200},
					"endLoc": {"line": 25, "column": 2, "position": 500}
				},
				"secondFile": {
					"name": "pkg/other.go",
					"start": 42,
					"end": 57,
					"startLoc": {"line": 42, "column": 3, "position": 800},
					"endLoc": {"line": 57, "column": 2, "position": 1100}
				}
			},
			{
				"format": "typescript",
				"lines": 8,
				"fragment": "const x = 1",
				"tokens": 60,
				"firstFile": {
					"name": "src/utils.ts",
					"start": 5,
					"end": 13,
					"startLoc": {"line": 5, "column": 0, "position": 50},
					"endLoc": {"line": 13, "column": 1, "position": 150}
				},
				"secondFile": {
					"name": "src/helpers.ts",
					"start": 20,
					"end": 28,
					"startLoc": {"line": 20, "column": 0, "position": 300},
					"endLoc": {"line": 28, "column": 1, "position": 400}
				}
			}
		]
	}`

	violations, err := j.parseViolations([]byte(report))
	require.NoError(t, err)
	require.Len(t, violations, 2)

	require.Equal(t, "/workspace/pkg/handler.go", violations[0].File)
	require.Equal(t, 10, violations[0].Line)
	require.Equal(t, 1, violations[0].Column)
	require.Equal(t, "jscpd", violations[0].Source)
	require.Contains(t, *violations[0].Message, "15 lines")
	require.Contains(t, *violations[0].Message, "pkg/other.go:42")
	require.Equal(t, "duplicate-go", violations[0].Rule.Method)

	require.Equal(t, "/workspace/src/utils.ts", violations[1].File)
	require.Equal(t, 5, violations[1].Line)
	require.Contains(t, *violations[1].Message, "8 lines")
	require.Contains(t, *violations[1].Message, "src/helpers.ts:20")
	require.Equal(t, "duplicate-typescript", violations[1].Rule.Method)
}

func TestParseViolationsEmpty(t *testing.T) {
	j := &JSCPD{}
	j.WorkDir = "/workspace"

	report := `{"statistics": {}, "duplicates": []}`

	violations, err := j.parseViolations([]byte(report))
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestParseViolationsInvalidJSON(t *testing.T) {
	j := &JSCPD{}
	j.WorkDir = "/workspace"

	_, err := j.parseViolations([]byte(`not json`))
	require.Error(t, err)
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		workDir  string
		name     string
		expected string
	}{
		{"/workspace", "pkg/file.go", "/workspace/pkg/file.go"},
		{"/workspace", "/absolute/path.go", "/absolute/path.go"},
		{"/workspace", "../../relative/path.go", "/relative/path.go"},
		{"/workspace", "./local/file.go", "/workspace/local/file.go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, normalizePath(tt.workDir, tt.name))
		})
	}
}

func TestBuildExcludes(t *testing.T) {
	j := &JSCPD{}
	j.WorkDir = "/workspace"

	excludes := j.buildExcludes()
	require.NotEmpty(t, excludes)

	has := func(pattern string) bool {
		for _, e := range excludes {
			if e == pattern {
				return true
			}
		}
		return false
	}

	require.True(t, has(".git/**"), "should include builtin .git exclude")
	require.True(t, has("node_modules/**"), "should include builtin node_modules exclude")
	require.True(t, has("**/*_test.go"), "should include jscpd-specific test exclude")
	require.True(t, has("**/go.sum"), "should include lock file exclude")
}

func TestDefaultExcludes(t *testing.T) {
	j := &JSCPD{}
	excludes := j.DefaultExcludes()
	require.NotEmpty(t, excludes)

	has := func(pattern string) bool {
		for _, e := range excludes {
			if e == pattern {
				return true
			}
		}
		return false
	}

	require.True(t, has("**/*_test.go"))
	require.True(t, has("**/*.pb.go"))
	require.True(t, has("**/testdata/**"))
}
