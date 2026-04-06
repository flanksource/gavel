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

	violations, err := j.parseViolations([]byte(report), nil)
	require.NoError(t, err)
	require.Len(t, violations, 2)

	require.Equal(t, "/workspace/pkg/handler.go", violations[0].File)
	require.Equal(t, 10, violations[0].Line)
	require.Equal(t, 0, violations[0].Column)
	require.Equal(t, "jscpd", violations[0].Source)
	require.Contains(t, *violations[0].Message, "15 lines")
	require.NotContains(t, *violations[0].Message, "tokens")
	require.Contains(t, *violations[0].Message, "lines 10-25")
	require.Contains(t, *violations[0].Message, "pkg/other.go:42-57")
	require.Equal(t, "duplicate-go", violations[0].Rule.Method)
	require.NotNil(t, violations[0].Code)
	require.Equal(t, "func doThing() {}", *violations[0].Code)

	require.Equal(t, "/workspace/src/utils.ts", violations[1].File)
	require.Equal(t, 5, violations[1].Line)
	require.Equal(t, 0, violations[1].Column)
	require.Contains(t, *violations[1].Message, "8 lines")
	require.NotContains(t, *violations[1].Message, "tokens")
	require.Contains(t, *violations[1].Message, "src/helpers.ts:20-28")
	require.Equal(t, "duplicate-typescript", violations[1].Rule.Method)
}

func TestParseViolationsSameFile(t *testing.T) {
	j := &JSCPD{}
	j.WorkDir = "/workspace"

	report := `{
		"statistics": {},
		"duplicates": [{
			"format": "go",
			"lines": 10,
			"fragment": "func A() {}",
			"tokens": 0,
			"firstFile": {
				"name": "pkg/handler.go",
				"start": 10, "end": 20,
				"startLoc": {"line": 10, "column": 1, "position": 0},
				"endLoc": {"line": 20, "column": 1, "position": 100}
			},
			"secondFile": {
				"name": "pkg/handler.go",
				"start": 50, "end": 60,
				"startLoc": {"line": 50, "column": 1, "position": 500},
				"endLoc": {"line": 60, "column": 1, "position": 600}
			}
		}]
	}`

	violations, err := j.parseViolations([]byte(report), nil)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Contains(t, *violations[0].Message, "also at lines 50-60")
	require.NotContains(t, *violations[0].Message, "handler.go:")
}

func TestParseViolationsEmpty(t *testing.T) {
	j := &JSCPD{}
	j.WorkDir = "/workspace"

	report := `{"statistics": {}, "duplicates": []}`

	violations, err := j.parseViolations([]byte(report), nil)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestParseViolationsInvalidJSON(t *testing.T) {
	j := &JSCPD{}
	j.WorkDir = "/workspace"

	_, err := j.parseViolations([]byte(`not json`), nil)
	require.Error(t, err)
}

func TestParseViolationsFiltersExcluded(t *testing.T) {
	j := &JSCPD{}
	j.WorkDir = "/workspace"

	report := `{
		"statistics": {},
		"duplicates": [
			{
				"format": "go",
				"lines": 10,
				"fragment": "func A() {}",
				"tokens": 80,
				"firstFile": {
					"name": "pkg/handler.go",
					"start": 1, "end": 10,
					"startLoc": {"line": 1, "column": 1, "position": 0},
					"endLoc": {"line": 10, "column": 1, "position": 100}
				},
				"secondFile": {
					"name": "pkg/handler_test.go",
					"start": 1, "end": 10,
					"startLoc": {"line": 1, "column": 1, "position": 0},
					"endLoc": {"line": 10, "column": 1, "position": 100}
				}
			},
			{
				"format": "go",
				"lines": 10,
				"fragment": "func B() {}",
				"tokens": 80,
				"firstFile": {
					"name": "pkg/util.go",
					"start": 1, "end": 10,
					"startLoc": {"line": 1, "column": 1, "position": 0},
					"endLoc": {"line": 10, "column": 1, "position": 100}
				},
				"secondFile": {
					"name": "pkg/other.go",
					"start": 1, "end": 10,
					"startLoc": {"line": 1, "column": 1, "position": 0},
					"endLoc": {"line": 10, "column": 1, "position": 100}
				}
			},
			{
				"format": "go",
				"lines": 5,
				"fragment": "func C() {}",
				"tokens": 40,
				"firstFile": {
					"name": "vendor/lib/foo.go",
					"start": 1, "end": 5,
					"startLoc": {"line": 1, "column": 1, "position": 0},
					"endLoc": {"line": 5, "column": 1, "position": 50}
				},
				"secondFile": {
					"name": "pkg/bar.go",
					"start": 1, "end": 5,
					"startLoc": {"line": 1, "column": 1, "position": 0},
					"endLoc": {"line": 5, "column": 1, "position": 50}
				}
			}
		]
	}`

	excludes := []string{"**/*_test.go", "vendor/**"}
	violations, err := j.parseViolations([]byte(report), excludes)
	require.NoError(t, err)
	require.Len(t, violations, 1, "should filter out test file and vendor duplicates")
	require.Equal(t, "/workspace/pkg/util.go", violations[0].File)
}

func TestMatchesAny(t *testing.T) {
	tests := []struct {
		path     string
		patterns []string
		expected bool
	}{
		{"pkg/handler_test.go", []string{"**/*_test.go"}, true},
		{"pkg/handler.go", []string{"**/*_test.go"}, false},
		{"vendor/lib/foo.go", []string{"vendor/**"}, true},
		{"pkg/foo.go", []string{"vendor/**"}, false},
		{"node_modules/pkg/index.js", []string{"node_modules/**"}, true},
		{"src/index.js", []string{"node_modules/**"}, false},
		{"foo.min.js", []string{"*.min.*"}, true},
		{"foo.js", []string{"*.min.*"}, false},
		{"pkg/foo.go", nil, false},
		{"pkg/foo.go", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			require.Equal(t, tt.expected, matchesAny(tt.path, tt.patterns))
		})
	}
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

	has := func(patterns []string, pattern string) bool {
		for _, e := range patterns {
			if e == pattern {
				return true
			}
		}
		return false
	}

	require.True(t, has(excludes, "**/.git/**"), "should include builtin .git exclude")
	require.True(t, has(excludes, "**/node_modules/**"), "should include builtin node_modules exclude")
	require.True(t, has(excludes, "**/*_test.go"), "should include jscpd-specific test exclude")
	require.True(t, has(excludes, "**/go.sum"), "should include lock file exclude")

	j.Ignores = []string{"custom/**", "*.generated.ts"}
	withIgnores := j.buildExcludes()
	require.True(t, has(withIgnores, "**/custom/**"), "should include CLI ignore pattern")
	require.True(t, has(withIgnores, "**/*.generated.ts"), "should include CLI ignore pattern")
	require.True(t, has(withIgnores, "**/.git/**"), "should still include builtins")
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
