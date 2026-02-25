package claude

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	restore, err := gitStash(dir, false)
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

	restore, err := gitStash(dir, false)
	require.NoError(t, err)

	// dirty.txt should be stashed (not present)
	_, err = os.Stat(filepath.Join(dir, "dirty.txt"))
	assert.True(t, os.IsNotExist(err), "dirty.txt should be stashed")

	restore()

	// dirty.txt should be restored
	_, err = os.Stat(filepath.Join(dir, "dirty.txt"))
	assert.NoError(t, err, "dirty.txt should be restored after stash pop")
}

func TestGitStash_DirtyFlag(t *testing.T) {
	dir := initTestRepo(t)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("dirty"), 0644))

	restore, err := gitStash(dir, true)
	require.NoError(t, err)

	// dirty.txt should still be present (stash skipped)
	_, err = os.Stat(filepath.Join(dir, "dirty.txt"))
	assert.NoError(t, err, "dirty.txt should remain when dirty=true")

	restore()
}

func TestGitStash_StagedChanges(t *testing.T) {
	dir := initTestRepo(t)

	// Stage a change
	require.NoError(t, os.WriteFile(filepath.Join(dir, "staged.txt"), []byte("staged"), 0644))
	cmd := exec.Command("git", "add", "staged.txt")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	restore, err := gitStash(dir, false)
	require.NoError(t, err)

	// staged.txt should be stashed
	_, err = os.Stat(filepath.Join(dir, "staged.txt"))
	assert.True(t, os.IsNotExist(err), "staged.txt should be stashed")

	restore()

	_, err = os.Stat(filepath.Join(dir, "staged.txt"))
	assert.NoError(t, err, "staged.txt should be restored after stash pop")
}

func TestGitCheckoutBranch_SameBranch(t *testing.T) {
	dir := initTestRepo(t)

	// Get current branch name
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(t, err)
	currentBranch := strings.TrimSpace(string(out))

	restore, err := gitCheckoutBranch(dir, currentBranch)
	require.NoError(t, err)

	// Should be a noop
	restore()

	// Still on the same branch
	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = dir
	out, err = cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, currentBranch, strings.TrimSpace(string(out)))
}

func TestGitCheckoutBranch_SwitchAndRestore(t *testing.T) {
	dir := initTestRepo(t)

	// Create a target branch
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v failed: %s", args, out)
	}
	run("branch", "feature-branch")

	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(t, err)
	originalBranch := strings.TrimSpace(string(out))

	restore, err := gitCheckoutBranch(dir, "feature-branch")
	require.NoError(t, err)

	// Should be on the target branch
	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = dir
	out, err = cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "feature-branch", strings.TrimSpace(string(out)))

	restore()

	// Should be back on original branch
	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = dir
	out, err = cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, originalBranch, strings.TrimSpace(string(out)))
}

func TestGitCommitChanges_NoChanges(t *testing.T) {
	dir := initTestRepo(t)
	before, err := gitSnapshot(dir)
	require.NoError(t, err)

	todo := &types.TODO{}
	sha, err := gitCommitChanges(t.Context(), dir, todo, before)
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
	before, err := gitSnapshot(dir)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "new.go"), []byte("package main\n"), 0644))

	todo := &types.TODO{}
	todo.Title = "test-todo"
	sha, err := gitCommitChanges(t.Context(), dir, todo, before)
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
	before, err := gitSnapshot(dir)
	require.NoError(t, err)

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
	sha, err := gitCommitChanges(t.Context(), dir, todo, before)
	assert.NoError(t, err)
	assert.NotEmpty(t, sha)

	// Check that the commit message starts with "fixup!"
	cmd = exec.Command("git", "log", "-1", "--format=%s")
	cmd.Dir = dir
	msgOut, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(msgOut), "fixup!")
}

func TestGitSnapshot(t *testing.T) {
	dir := initTestRepo(t)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0644))

	snap, err := gitSnapshot(dir)
	require.NoError(t, err)
	assert.Contains(t, snap, "new.txt")
	assert.Equal(t, "??", snap["new.txt"])

	// Clean tree returns empty snapshot
	dir2 := initTestRepo(t)
	snap2, err := gitSnapshot(dir2)
	require.NoError(t, err)
	assert.Empty(t, snap2)
}

func TestGitChangedFiles_ExcludesTodos(t *testing.T) {
	before := map[string]string{}
	after := map[string]string{
		"src/main.go":       "??",
		".todos/task.md":    "??",
		".todos":            "??",
		".todos/done.md":    " M",
		"pkg/handler.go":    " M",
	}

	changed := gitChangedFiles(before, after)
	assert.ElementsMatch(t, []string{"src/main.go", "pkg/handler.go"}, changed)
}

func TestGitChangedFiles_IgnoresUnchanged(t *testing.T) {
	before := map[string]string{
		"existing.go": " M",
	}
	after := map[string]string{
		"existing.go": " M",
		"new.go":      "??",
	}

	changed := gitChangedFiles(before, after)
	assert.Equal(t, []string{"new.go"}, changed)
}

func TestGitCommitChanges_ExcludesTodosDir(t *testing.T) {
	dir := initTestRepo(t)
	before, err := gitSnapshot(dir)
	require.NoError(t, err)

	// Create both a real file and a .todos file
	require.NoError(t, os.WriteFile(filepath.Join(dir, "real.go"), []byte("package main\n"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".todos"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".todos", "task.md"), []byte("# todo"), 0644))

	todo := &types.TODO{}
	todo.Title = "test-todo"
	sha, err := gitCommitChanges(t.Context(), dir, todo, before)
	assert.NoError(t, err)
	assert.NotEmpty(t, sha)

	// Verify only real.go was committed, not .todos/task.md
	cmd := exec.Command("git", "show", "--name-only", "--format=", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(t, err)
	files := strings.TrimSpace(string(out))
	assert.Contains(t, files, "real.go")
	assert.NotContains(t, files, ".todos")
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
