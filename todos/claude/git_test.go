package claude

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v failed: %s", args, out)
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644))
	run("add", "README.md")
	run("commit", "-m", "initial commit")
	return dir
}

func TestGitStash_CleanTree(t *testing.T) {
	dir := initTestRepo(t)
	restore, err := gitStash(dir)
	require.NoError(t, err)

	// restore should be a no-op
	restore()

	// Verify HEAD is unchanged
	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), "initial commit")
}

func TestGitStash_DirtyTree(t *testing.T) {
	dir := initTestRepo(t)

	// Make the tree dirty
	require.NoError(t, os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("dirty"), 0644))

	restore, err := gitStash(dir)
	require.NoError(t, err)

	// dirty.txt should be stashed (not present)
	_, err = os.Stat(filepath.Join(dir, "dirty.txt"))
	assert.True(t, os.IsNotExist(err), "dirty.txt should be stashed")

	restore()

	// dirty.txt should be restored
	_, err = os.Stat(filepath.Join(dir, "dirty.txt"))
	assert.NoError(t, err, "dirty.txt should be restored after stash pop")
}

func TestGitStash_StagedChanges(t *testing.T) {
	dir := initTestRepo(t)

	// Stage a change
	require.NoError(t, os.WriteFile(filepath.Join(dir, "staged.txt"), []byte("staged"), 0644))
	cmd := exec.Command("git", "add", "staged.txt")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	restore, err := gitStash(dir)
	require.NoError(t, err)

	// staged.txt should be stashed
	_, err = os.Stat(filepath.Join(dir, "staged.txt"))
	assert.True(t, os.IsNotExist(err), "staged.txt should be stashed")

	restore()

	_, err = os.Stat(filepath.Join(dir, "staged.txt"))
	assert.NoError(t, err, "staged.txt should be restored after stash pop")
}

func TestGitCommitChanges_NoChanges(t *testing.T) {
	dir := initTestRepo(t)
	todo := &types.TODO{}
	sha, err := gitCommitChanges(t.Context(), dir, todo)
	assert.NoError(t, err)
	assert.Empty(t, sha)

	// Should still be just 1 commit
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "1\n", string(out))
}

func TestGitCommitChanges_WithChanges(t *testing.T) {
	dir := initTestRepo(t)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "new.go"), []byte("package main\n"), 0644))

	todo := &types.TODO{}
	todo.Title = "test-todo"
	sha, err := gitCommitChanges(t.Context(), dir, todo)
	assert.NoError(t, err)
	assert.NotEmpty(t, sha)

	// Should have 2 commits now
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "2\n", string(out))
}

func TestGitCommitChanges_Fixup(t *testing.T) {
	dir := initTestRepo(t)

	// Get the initial commit hash
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	hashOut, err := cmd.Output()
	require.NoError(t, err)
	hash := string(hashOut[:len(hashOut)-1]) // trim newline

	// Create a new file
	require.NoError(t, os.WriteFile(filepath.Join(dir, "fixup.txt"), []byte("fixup"), 0644))

	todo := &types.TODO{}
	todo.WorkingCommit = hash
	sha, err := gitCommitChanges(t.Context(), dir, todo)
	assert.NoError(t, err)
	assert.NotEmpty(t, sha)

	// Check that the commit message starts with "fixup!"
	cmd = exec.Command("git", "log", "-1", "--format=%s")
	cmd.Dir = dir
	msgOut, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(msgOut), "fixup!")
}

func TestFormatCommitMsg(t *testing.T) {
	tests := []struct {
		name     string
		analysis models.CommitAnalysis
		expected string
	}{
		{
			name: "with type, scope and body",
			analysis: models.CommitAnalysis{
				Commit: models.Commit{
					CommitType: "feat",
					Scope:      "api",
					Subject:    "add endpoint",
					Body:       "detailed description",
				},
			},
			expected: "feat(api): add endpoint\n\ndetailed description",
		},
		{
			name: "type only",
			analysis: models.CommitAnalysis{
				Commit: models.Commit{
					CommitType: "fix",
					Subject:    "resolve crash",
				},
			},
			expected: "fix: resolve crash",
		},
		{
			name: "no type or scope",
			analysis: models.CommitAnalysis{
				Commit: models.Commit{
					Subject: "plain message",
				},
			},
			expected: "plain message",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, formatCommitMsg(tc.analysis))
		})
	}
}
