package commit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunGitRebaseAutostashesUnstagedChanges proves that a dirty working tree
// (files not part of the commit) no longer aborts the rebase. The commit push
// flow stages and commits only selected files, then rebases onto the base
// branch — without --autostash git refuses with "cannot rebase: You have
// unstaged changes", which is the failure this test guards against.
func TestRunGitRebaseAutostashesUnstagedChanges(t *testing.T) {
	repo := initCommitRepo(t)
	base := strings.TrimSpace(gitOutput(t, repo, "rev-parse", "--abbrev-ref", "HEAD"))

	// Diverge: a feature branch with its own commit, while base advances on a
	// different file so the rebase replays a real commit without conflicting.
	gitRun(t, repo, "checkout", "-b", "feature")
	writeFile(t, repo, "feature.txt", "f1\n")
	gitRun(t, repo, "add", "feature.txt")
	gitRun(t, repo, "commit", "-m", "feat: add feature")

	gitRun(t, repo, "checkout", base)
	writeFile(t, repo, "base.txt", "b1\n")
	gitRun(t, repo, "add", "base.txt")
	gitRun(t, repo, "commit", "-m", "chore: advance base")

	gitRun(t, repo, "checkout", "feature")
	// Leave an unstaged modification in the working tree — the condition that
	// previously broke the rebase.
	writeFile(t, repo, "feature.txt", "f1\nunstaged edit\n")

	clean, err := runGitRebase(repo, base, "")
	require.NoError(t, err)
	assert.True(t, clean, "rebase should complete cleanly with --autostash")

	// The rebased feature commit is now on top of base's new commit.
	assert.FileExists(t, filepath.Join(repo, "base.txt"))

	// The autostashed change is restored: file still carries the unstaged edit
	// and shows as modified in the working tree.
	content, readErr := os.ReadFile(filepath.Join(repo, "feature.txt"))
	require.NoError(t, readErr)
	assert.Contains(t, string(content), "unstaged edit")
	status := gitOutput(t, repo, "status", "--porcelain", "feature.txt")
	assert.Contains(t, status, "feature.txt", "unstaged edit should survive the rebase")
}
