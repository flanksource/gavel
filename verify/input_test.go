package verify

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyArg(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "pkg", "verify"), 0o755)
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0o644)

	tests := []struct {
		arg      string
		wantType string
		wantVal  string
	}{
		// PR URLs
		{"https://github.com/org/repo/pull/123", "pr", "123"},
		{"https://github.com/flanksource/gavel/pull/42", "pr", "42"},
		// PR hash refs
		{"#123", "pr", "123"},
		{"#1", "pr", "1"},
		// Bare digits -> PR
		{"42", "pr", "42"},
		{"99999", "pr", "99999"},
		// Date ranges
		{"2024-01-01..2024-06-15", "date-range", "2024-01-01..2024-06-15"},
		// Commit ranges (contain ..)
		{"main..HEAD", "range", "main..HEAD"},
		{"abc123..def456", "range", "abc123..def456"},
		{"v1.0...v2.0", "range", "v1.0...v2.0"},
		// Globs
		{"*.go", "file", "*.go"},
		{"src/**/*.ts", "file", "src/**/*.ts"},
		// Existing directory
		{"pkg/verify", "directory", "pkg/verify"},
		// Existing file
		{"main.go", "file", "main.go"},
		// Path-like (contains /)
		{"path/to/new.go", "file", "path/to/new.go"},
		// Hex SHAs
		{"abc1234", "commit", "abc1234"},
		{"deadbeefcafe1234567890abcdef1234567890ab", "commit", "deadbeefcafe1234567890abcdef1234567890ab"},
		// HEAD offsets
		{"HEAD~3", "commit", "HEAD~3"},
		{"HEAD^2", "commit", "HEAD^2"},
		// Ref offsets
		{"main~2", "commit", "main~2"},
		{"feature^1", "commit", "feature^1"},
		// Bare branch names (default)
		{"main", "branch", "main"},
		{"feature-branch", "branch", "feature-branch"},
		{"release", "branch", "release"},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			typ, val := ClassifyArg(tt.arg, tmpDir)
			assert.Equal(t, tt.wantType, typ, "type mismatch for %q", tt.arg)
			assert.Equal(t, tt.wantVal, val, "value mismatch for %q", tt.arg)
		})
	}
}

func TestClassifyArgNoRepoPath(t *testing.T) {
	// Without repoPath, can't stat files. Falls through to branch for bare names.
	typ, _ := ClassifyArg("somefile.go", "")
	assert.Equal(t, "branch", typ)

	// Path-like args still detected as files by the / check
	typ, _ = ClassifyArg("pkg/somefile.go", "")
	assert.Equal(t, "file", typ)
}

func TestResolveScopeArgs(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "a.go"), []byte("package a"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "b.go"), []byte("package b"), 0o644)

	t.Run("PR arg", func(t *testing.T) {
		s, err := ResolveScope([]string{"#42"}, "", tmpDir)
		require.NoError(t, err)
		assert.Equal(t, "pr", s.Type)
		assert.Equal(t, 42, s.PRNumber)
	})

	t.Run("commit range arg", func(t *testing.T) {
		s, err := ResolveScope([]string{"main..HEAD"}, "", tmpDir)
		require.NoError(t, err)
		assert.Equal(t, "range", s.Type)
		assert.Equal(t, "main..HEAD", s.CommitRange)
	})

	t.Run("commit SHA arg", func(t *testing.T) {
		s, err := ResolveScope([]string{"abc1234"}, "", tmpDir)
		require.NoError(t, err)
		assert.Equal(t, "commit", s.Type)
		assert.Equal(t, "abc1234", s.Commit)
	})

	t.Run("branch arg", func(t *testing.T) {
		s, err := ResolveScope([]string{"develop"}, "", tmpDir)
		require.NoError(t, err)
		assert.Equal(t, "branch", s.Type)
		assert.Equal(t, "develop", s.Branch)
	})

	t.Run("date range arg", func(t *testing.T) {
		s, err := ResolveScope([]string{"2024-01-01..2024-06-01"}, "", tmpDir)
		require.NoError(t, err)
		assert.Equal(t, "date-range", s.Type)
		assert.Equal(t, "2024-01-01", s.Since)
		assert.Equal(t, "2024-06-01", s.Until)
	})

	t.Run("multiple files", func(t *testing.T) {
		s, err := ResolveScope([]string{"a.go", "b.go"}, "", tmpDir)
		require.NoError(t, err)
		assert.Equal(t, "files", s.Type)
		assert.Equal(t, []string{"a.go", "b.go"}, s.Files)
	})

	t.Run("mixed singular and files errors", func(t *testing.T) {
		_, err := ResolveScope([]string{"#42", "a.go"}, "", tmpDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot mix")
	})

	t.Run("two singular args errors", func(t *testing.T) {
		_, err := ResolveScope([]string{"#42", "main..HEAD"}, "", tmpDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot combine")
	})
}
