package verify

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSingleGavelConfig(t *testing.T) {
	t.Run("reads one file without layering", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".gavel.yaml")
		require.NoError(t, os.WriteFile(path, []byte(`
commit:
  gitignore:
    - "*.log"
  allow:
    - "keep.log"
  precommit:
    mode: false
`), 0o644))

		cfg, err := LoadSingleGavelConfig(path)
		require.NoError(t, err)
		assert.Equal(t, []string{"*.log"}, cfg.Commit.GitIgnore)
		assert.Equal(t, []string{"keep.log"}, cfg.Commit.Allow)
		assert.Equal(t, CheckMode("skip"), cfg.Commit.Precommit.Mode)
	})

	t.Run("missing file returns os.ErrNotExist", func(t *testing.T) {
		_, err := LoadSingleGavelConfig(filepath.Join(t.TempDir(), "missing.yaml"))
		require.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("malformed yaml returns parse error", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".gavel.yaml")
		require.NoError(t, os.WriteFile(path, []byte("not: [valid yaml"), 0o644))
		_, err := LoadSingleGavelConfig(path)
		require.Error(t, err)
	})
}

func TestLoadGavelConfigTrace_FilePath(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	targetDir := filepath.Join(repo, "pkg", "api")
	targetFile := filepath.Join(targetDir, "handler.go")

	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))
	require.NoError(t, os.MkdirAll(targetDir, 0o755))
	require.NoError(t, os.WriteFile(targetFile, []byte("package api\n"), 0o644))
	t.Setenv("HOME", home)

	require.NoError(t, os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte(`ai:
  model: gemini
pre:
  - name: home
    run: echo home
`), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(repo, ".gavel.yaml"), []byte(`ai:
  model: claude
pre:
  - name: repo
    run: echo repo
ssh:
  cmd: make ci
`), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(targetDir, ".gavel.yaml"), []byte(`ai:
  model: codex
pre:
  - name: target
    run: echo target
lint:
  ignore:
    - file: pkg/api/**
`), 0o644))

	trace, err := LoadGavelConfigTrace(targetFile)
	require.NoError(t, err)

	assert.Equal(t, targetFile, trace.TargetPath)
	assert.Equal(t, targetDir, trace.TargetDir)
	assert.Equal(t, repo, trace.GitRoot)

	require.Len(t, trace.Sources, 3)
	assert.Equal(t, "user-home", trace.Sources[0].Origin)
	assert.Equal(t, filepath.Join(home, ".gavel.yaml"), trace.Sources[0].Path)
	assert.Equal(t, "git-root", trace.Sources[1].Origin)
	assert.Equal(t, filepath.Join(repo, ".gavel.yaml"), trace.Sources[1].Path)
	assert.Equal(t, "parent-directory", trace.Sources[2].Origin)
	assert.Equal(t, filepath.Join(targetDir, ".gavel.yaml"), trace.Sources[2].Path)

	assert.Equal(t, "codex", trace.Merged.AI.Model.Name)
	require.Len(t, trace.Merged.Pre, 3)
	assert.Equal(t, "home", trace.Merged.Pre[0].Name)
	assert.Equal(t, "repo", trace.Merged.Pre[1].Name)
	assert.Equal(t, "target", trace.Merged.Pre[2].Name)
	assert.Equal(t, "make ci", trace.Merged.SSH.Cmd)
	require.Len(t, trace.Merged.Lint.Ignore, 1)
	assert.Equal(t, "pkg/api/**", trace.Merged.Lint.Ignore[0].File)
}

func TestLoadGavelConfigTrace_DedupesGitRootTarget(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	t.Setenv("HOME", home)

	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte(`ai:
  model: gemini
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".gavel.yaml"), []byte(`ai:
  model: claude
`), 0o644))

	trace, err := LoadGavelConfigTrace(repo)
	require.NoError(t, err)

	require.Len(t, trace.Sources, 2)
	assert.Equal(t, "user-home", trace.Sources[0].Origin)
	assert.Equal(t, "git-root", trace.Sources[1].Origin)
	assert.Equal(t, "claude", trace.Merged.AI.Model.Name)
}
