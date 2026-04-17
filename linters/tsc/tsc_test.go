package tsc

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/require"
)

func TestParseViolations(t *testing.T) {
	data, err := os.ReadFile("testdata/tsc-diagnostics.json")
	require.NoError(t, err)

	workDir := "/workspace"
	violations, err := parseViolations(data, workDir, io.Discard)
	require.NoError(t, err)
	require.Len(t, violations, 3)

	first := violations[0]
	require.Equal(t, filepath.Join(workDir, "src/App.tsx"), first.File, "relative file joined onto workDir")
	require.Equal(t, 42, first.Line)
	require.Equal(t, 7, first.Column)
	require.Equal(t, models.SeverityError, first.Severity)
	require.Equal(t, "tsc", first.Source)
	require.NotNil(t, first.Rule)
	require.Equal(t, "TS2322", first.Rule.Method)
	require.Equal(t, "tsc", first.Rule.Package)
	require.NotNil(t, first.Message)
	require.Equal(t, "Type 'string' is not assignable to type 'number'.", *first.Message)

	second := violations[1]
	require.Equal(t, "/abs/path/src/handler.ts", second.File, "absolute file path preserved as-is")
	require.Equal(t, models.SeverityWarning, second.Severity)
	require.Equal(t, "TS6133", second.Rule.Method)

	third := violations[2]
	require.Equal(t, models.SeverityInfo, third.Severity, "unknown category maps to info")
	require.Equal(t, "TS7053", third.Rule.Method)
}

// TestParseViolations_RulePerCode ensures each TS error code is its own rule
// so that users can exclude codes individually (rule: TS2322) or globally
// (source: tsc) via LintIgnoreRule.
func TestParseViolations_RulePerCode(t *testing.T) {
	raw := []byte(`[
		{"file":"a.ts","line":1,"column":1,"code":2322,"category":"Error","message":"x"},
		{"file":"b.ts","line":2,"column":2,"code":6133,"category":"Error","message":"y"},
		{"file":"c.ts","line":3,"column":3,"code":2322,"category":"Warning","message":"z"}
	]`)
	violations, err := parseViolations(raw, "/workspace", io.Discard)
	require.NoError(t, err)
	require.Len(t, violations, 3)

	require.Equal(t, "TS2322", violations[0].Rule.Method)
	require.Equal(t, "TS6133", violations[1].Rule.Method)
	require.Equal(t, "TS2322", violations[2].Rule.Method,
		"same code reported at different categories maps to the same rule so one ignore entry covers both")
}

func TestParseViolations_EmptyInputs(t *testing.T) {
	for _, in := range []string{"", "   ", "[]", "\n"} {
		v, err := parseViolations([]byte(in), "/workspace", io.Discard)
		require.NoError(t, err, "input %q", in)
		require.Empty(t, v, "input %q", in)
	}
}

func TestParseViolations_MalformedJSON(t *testing.T) {
	raw := []byte("not json at all")
	_, err := parseViolations(raw, "/workspace", io.Discard)
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse tsc JSON output")
	require.Contains(t, err.Error(), "not json at all", "raw output must appear in error for debugging")
}

func TestCategoryToSeverity(t *testing.T) {
	cases := map[string]models.ViolationSeverity{
		"Error":      models.SeverityError,
		"Warning":    models.SeverityWarning,
		"Message":    models.SeverityInfo,
		"Suggestion": models.SeverityInfo,
		"":           models.SeverityInfo,
	}
	for in, want := range cases {
		require.Equal(t, want, categoryToSeverity(in), "category %q", in)
	}
}

func TestResolveScript_WritesAndReuses(t *testing.T) {
	workDir := t.TempDir()
	tr := &TSC{}
	tr.WorkDir = workDir

	first, err := tr.resolveScript()
	require.NoError(t, err)
	require.FileExists(t, first)

	info1, err := os.Stat(first)
	require.NoError(t, err)

	// Force a distinct mtime so a rewrite would be observable.
	oldTime := info1.ModTime().Add(-5 * 1e9)
	require.NoError(t, os.Chtimes(first, oldTime, oldTime))

	second, err := tr.resolveScript()
	require.NoError(t, err)
	require.Equal(t, first, second, "same content produces same path")

	info2, err := os.Stat(second)
	require.NoError(t, err)
	require.Equal(t, oldTime.Unix(), info2.ModTime().Unix(), "unchanged content must not trigger a rewrite")

	contents, err := os.ReadFile(second)
	require.NoError(t, err)
	require.Equal(t, tscJSONScript, contents)
}

func TestResolveScript_RewritesWhenContentDiffers(t *testing.T) {
	workDir := t.TempDir()
	tr := &TSC{}
	tr.WorkDir = workDir

	path, err := tr.resolveScript()
	require.NoError(t, err)

	// Corrupt the file on disk and confirm resolveScript repairs it.
	require.NoError(t, os.WriteFile(path, []byte("tampered"), 0o644))

	again, err := tr.resolveScript()
	require.NoError(t, err)
	require.Equal(t, path, again)

	contents, err := os.ReadFile(again)
	require.NoError(t, err)
	require.Equal(t, tscJSONScript, contents, "file was rewritten to match embedded script")
}

func TestResolveScript_EmptyWorkDirFails(t *testing.T) {
	tr := &TSC{}
	_, err := tr.resolveScript()
	require.Error(t, err)
}
