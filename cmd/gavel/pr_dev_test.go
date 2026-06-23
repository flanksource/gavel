package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFiles(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, n := range names {
		require.NoError(t, os.WriteFile(filepath.Join(dir, n), []byte("{}"), 0o644))
	}
}

func TestIsDevDir(t *testing.T) {
	t.Run("dir with package.json and vite.config.ts is valid", func(t *testing.T) {
		dir := t.TempDir()
		writeFiles(t, dir, "package.json", "vite.config.ts")
		assert.True(t, isDevDir(dir))
	})

	t.Run("dir missing vite.config.ts is invalid", func(t *testing.T) {
		dir := t.TempDir()
		writeFiles(t, dir, "package.json")
		assert.False(t, isDevDir(dir))
	})

	t.Run("empty dir is invalid", func(t *testing.T) {
		assert.False(t, isDevDir(t.TempDir()))
	})
}

func TestResolveDevDirFromExplicitPath(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "package.json", "vite.config.ts")

	got, err := resolveDevDir(dir)
	require.NoError(t, err)
	assert.Equal(t, dir, got)
}
