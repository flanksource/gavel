// Tests workspace checkouts: serve keeps a read-only working tree of a repository's default
// branch, refreshed on each call, with symlinks written as plain files.
package sandbox

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckoutKeepsTheDefaultBranchCurrent(t *testing.T) {
	dataDir := t.TempDir()
	org, repo := uuid.New(), uuid.New()
	bare := filepath.Join(dataDir, "repos", org.String(), repo.String()+".git")
	work := newWork(t)
	commitFile(t, work, ".gavel.yaml", "todos: {}\n")
	require.NoError(t, os.Symlink("/etc/hostname", filepath.Join(work, "leak.yaml")))
	git(t, work, "add", "leak.yaml")
	git(t, work, "commit", "-q", "-m", "a symlink out of the tree")
	git(t, work, "clone", "-q", "--bare", work, bare)
	sources := &Sources{DataDir: dataDir}
	target := Repo{ID: repo, Org: org, Source: "push"}

	dir, err := sources.Checkout(context.Background(), target)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dataDir, "workspaces", org.String(), repo.String()), dir)
	config, err := os.ReadFile(filepath.Join(dir, ".gavel.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "todos: {}\n", string(config))

	info, err := os.Lstat(filepath.Join(dir, "leak.yaml"))
	require.NoError(t, err)
	assert.True(t, info.Mode().IsRegular(), "a committed symlink is a plain file holding its target")
	leak, err := os.ReadFile(filepath.Join(dir, "leak.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "/etc/hostname", string(leak))

	commitFile(t, work, ".gavel.yaml", "todos: {lifecycle: custom}\n")
	git(t, work, "push", "-q", bare, "main")
	dir, err = sources.Checkout(context.Background(), target)
	require.NoError(t, err)
	config, err = os.ReadFile(filepath.Join(dir, ".gavel.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "todos: {lifecycle: custom}\n", string(config), "the checkout follows the default branch")
}

func TestCheckoutOfAnEmptyRepository(t *testing.T) {
	dataDir := t.TempDir()
	org, repo := uuid.New(), uuid.New()
	bare := filepath.Join(dataDir, "repos", org.String(), repo.String()+".git")
	git(t, t.TempDir(), "init", "-q", "--bare", bare)

	_, err := (&Sources{DataDir: dataDir}).Checkout(context.Background(), Repo{ID: repo, Org: org, Source: "push", Name: "demo"})
	require.EqualError(t, err, "repository demo has no default branch yet: push a branch to it first")
}
