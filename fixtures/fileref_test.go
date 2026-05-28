package fixtures

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
}

func TestResolveFileRef(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "golden.txt"), "hello world\n")

	t.Run("literal value passes through", func(t *testing.T) {
		ref, err := ResolveFileRef(dir, "plain literal")
		require.NoError(t, err)
		assert.False(t, ref.IsFile)
		assert.Equal(t, "plain literal", ref.Raw)
	})

	t.Run("escaped @ becomes literal", func(t *testing.T) {
		ref, err := ResolveFileRef(dir, `\@literal`)
		require.NoError(t, err)
		assert.False(t, ref.IsFile)
		assert.Equal(t, "@literal", ref.Raw)
	})

	t.Run("relative @path resolves against sourceDir", func(t *testing.T) {
		ref, err := ResolveFileRef(dir, "@golden.txt")
		require.NoError(t, err)
		assert.True(t, ref.IsFile)
		assert.Equal(t, "hello world\n", ref.Contents)
		assert.Equal(t, filepath.Join(dir, "golden.txt"), ref.Path)
	})

	t.Run("absolute @path is honored", func(t *testing.T) {
		abs := filepath.Join(dir, "golden.txt")
		ref, err := ResolveFileRef("/nowhere", "@"+abs)
		require.NoError(t, err)
		assert.True(t, ref.IsFile)
		assert.Equal(t, "hello world\n", ref.Contents)
	})

	t.Run("missing file returns error", func(t *testing.T) {
		_, err := ResolveFileRef(dir, "@does-not-exist.txt")
		require.Error(t, err)
	})
}

func TestUnifiedDiff(t *testing.T) {
	t.Run("identical returns empty", func(t *testing.T) {
		assert.Empty(t, UnifiedDiff("a\nb\n", "a\nb\n", "want", "got"))
	})

	t.Run("differing produces unified diff with labels", func(t *testing.T) {
		d := UnifiedDiff("alpha\nbeta\n", "alpha\nGAMMA\n", "expected.md", "stdout")
		assert.Contains(t, d, "--- expected.md")
		assert.Contains(t, d, "+++ stdout")
		assert.Contains(t, d, "-beta")
		assert.Contains(t, d, "+GAMMA")
	})
}

func TestEvaluateStreamFileRef(t *testing.T) {
	dir := t.TempDir()
	goldenPath := filepath.Join(dir, "golden.md")
	writeFile(t, goldenPath, "rendered output\n")

	t.Run("matching @file passes", func(t *testing.T) {
		exp := Expectations{Stdout: "@golden.md"}
		fixture := FixtureResult{
			Name:     "match",
			Status:   "pending",
			Metadata: map[string]interface{}{},
			Test:     FixtureTest{Name: "match", SourceDir: dir},
		}
		result := exp.Evaluate(fixture, exec.ExecResult{
			Stdout:   "rendered output\n",
			ExitCode: 0,
		}, EvaluateOptions{SourceDir: dir})
		assert.Equal(t, task.StatusPASS, result.Status)
	})

	t.Run("mismatched @file fails with unified diff", func(t *testing.T) {
		exp := Expectations{Stdout: "@golden.md"}
		fixture := FixtureResult{
			Name:     "mismatch",
			Status:   "pending",
			Metadata: map[string]interface{}{},
			Test:     FixtureTest{Name: "mismatch", SourceDir: dir},
		}
		result := exp.Evaluate(fixture, exec.ExecResult{
			Stdout:   "rendered DIFFERENT\n",
			ExitCode: 0,
		}, EvaluateOptions{SourceDir: dir})
		assert.Equal(t, task.StatusFAIL, result.Status)
		assert.Contains(t, result.Error, "stdout differs")
		assert.Contains(t, result.Error, "---")
		assert.Contains(t, result.Error, "+++")
		assert.Contains(t, result.Error, "-rendered output")
		assert.Contains(t, result.Error, "+rendered DIFFERENT")
	})

	t.Run("missing @file surfaces resolve error", func(t *testing.T) {
		exp := Expectations{Stdout: "@nope.md"}
		fixture := FixtureResult{
			Name:     "missing",
			Status:   "pending",
			Metadata: map[string]interface{}{},
			Test:     FixtureTest{Name: "missing", SourceDir: dir},
		}
		result := exp.Evaluate(fixture, exec.ExecResult{Stdout: "anything"}, EvaluateOptions{SourceDir: dir})
		assert.Equal(t, task.StatusERR, result.Status)
		assert.Contains(t, result.Error, "load expected stdout")
	})

	t.Run("literal stdout still compared exactly", func(t *testing.T) {
		exp := Expectations{Stdout: "exact"}
		fixture := FixtureResult{
			Name:     "literal",
			Status:   "pending",
			Metadata: map[string]interface{}{},
			Test:     FixtureTest{Name: "literal", SourceDir: dir},
		}
		good := exp.Evaluate(fixture, exec.ExecResult{Stdout: "exact"}, EvaluateOptions{SourceDir: dir})
		assert.Equal(t, task.StatusPASS, good.Status)

		bad := exp.Evaluate(fixture, exec.ExecResult{Stdout: "other"}, EvaluateOptions{SourceDir: dir})
		assert.Equal(t, task.StatusFAIL, bad.Status)
		assert.Contains(t, bad.Error, "stdout differs")
	})
}

func TestEvaluateStreamUpdateGolden(t *testing.T) {
	dir := t.TempDir()
	goldenPath := filepath.Join(dir, "out.md")
	writeFile(t, goldenPath, "stale\n")

	exp := Expectations{Stdout: "@out.md"}
	fixture := FixtureResult{
		Name:     "update",
		Status:   "pending",
		Metadata: map[string]interface{}{},
		Test:     FixtureTest{Name: "update", SourceDir: dir},
	}
	result := exp.Evaluate(fixture, exec.ExecResult{Stdout: "fresh\n"}, EvaluateOptions{
		SourceDir:    dir,
		UpdateGolden: true,
	})
	assert.Equal(t, task.StatusPASS, result.Status)
	assert.Equal(t, true, result.Metadata["golden_updated_stdout"])

	body, err := os.ReadFile(goldenPath)
	require.NoError(t, err)
	assert.Equal(t, "fresh\n", string(body))
}

func TestEvaluateStreamUpdateGoldenSkipsLiteral(t *testing.T) {
	exp := Expectations{Stdout: "literal-value"}
	fixture := FixtureResult{
		Name:     "no-write",
		Status:   "pending",
		Metadata: map[string]interface{}{},
		Test:     FixtureTest{Name: "no-write"},
	}
	result := exp.Evaluate(fixture, exec.ExecResult{Stdout: "different"}, EvaluateOptions{
		UpdateGolden: true,
	})
	assert.Equal(t, task.StatusFAIL, result.Status)
	assert.NotContains(t, result.Metadata, "golden_updated_stdout")
}
