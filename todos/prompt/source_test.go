package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTestFile(t *testing.T, dir, name string, numLines int) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	var sb strings.Builder
	for i := 1; i <= numLines; i++ {
		fmt.Fprintf(&sb, "line %d content\n", i)
	}
	require.NoError(t, os.WriteFile(path, []byte(sb.String()), 0o644))
	return path
}

func TestReadSourceLines_WholeFile(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "small.go", 10)

	result, err := ReadSourceLines(dir, types.PathRef{File: "small.go"})
	require.NoError(t, err)
	assert.Contains(t, result, "line 1 content")
	assert.Contains(t, result, "line 10 content")
}

func TestReadSourceLines_WholeFileTruncated(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "large.go", 300)

	result, err := ReadSourceLines(dir, types.PathRef{File: "large.go"})
	require.NoError(t, err)
	assert.Contains(t, result, "line 200 content")
	assert.NotContains(t, result, "line 201 content")
}

func TestReadSourceLines_SingleLine(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "code.go", 50)

	result, err := ReadSourceLines(dir, types.PathRef{File: "code.go", Line: 25})
	require.NoError(t, err)
	assert.Contains(t, result, "line 25 content")
	assert.Contains(t, result, "line 15 content")
	assert.Contains(t, result, "line 35 content")
}

func TestReadSourceLines_Range(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "code.go", 50)

	result, err := ReadSourceLines(dir, types.PathRef{File: "code.go", Line: 10, EndLine: 20})
	require.NoError(t, err)
	assert.Contains(t, result, "line 10 content")
	assert.Contains(t, result, "line 20 content")
	assert.NotContains(t, result, "line 9 content")
	assert.NotContains(t, result, "line 21 content")
}

func TestReadSourceLines_FileNotFound(t *testing.T) {
	result, err := ReadSourceLines(t.TempDir(), types.PathRef{File: "nonexistent.go"})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestReadSourceLines_EdgeOfFile(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "short.go", 5)

	result, err := ReadSourceLines(dir, types.PathRef{File: "short.go", Line: 3})
	require.NoError(t, err)
	assert.Contains(t, result, "line 1 content")
	assert.Contains(t, result, "line 5 content")
}

func TestReadSourceLines_AbsolutePathInsideWorkDir(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "pkg/code.go", 5)

	result, err := ReadSourceLines(dir, types.PathRef{File: path})
	require.NoError(t, err)
	assert.Contains(t, result, "line 5 content")
}

// A todo's path refs come from its frontmatter, which a hosted todo's author writes, so a ref
// names only a file in the todo's repository: anything else would put another file serve can
// read into the prompt.
func TestReadSourceLines_RefusesAFileOutsideWorkDir(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "repo")
	require.NoError(t, os.Mkdir(dir, 0o755))
	outside := writeTestFile(t, parent, "secret.txt", 3)
	require.NoError(t, os.Symlink(outside, filepath.Join(dir, "link.txt")))

	for _, ref := range []string{"../secret.txt", outside, "link.txt"} {
		result, err := ReadSourceLines(dir, types.PathRef{File: ref})
		assert.ErrorContains(t, err, "escapes", ref)
		assert.Empty(t, result, ref)
	}
}
