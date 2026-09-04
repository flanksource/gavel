package commit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResolveGitRoot_FromSubdir confirms a subdirectory resolves to the
// repository top-level, which the commit flow relies on to normalize WorkDir.
func TestResolveGitRoot_FromSubdir(t *testing.T) {
	repo := initCommitRepo(t)
	sub := filepath.Join(repo, "a", "b")
	require.NoError(t, os.MkdirAll(sub, 0o755))

	got, err := resolveGitRoot(sub)
	require.NoError(t, err)

	wantRoot, err := filepath.EvalSymlinks(repo)
	require.NoError(t, err)
	gotRoot, err := filepath.EvalSymlinks(got)
	require.NoError(t, err)
	assert.Equal(t, wantRoot, gotRoot)
}

// TestResolveGitRoot_NotARepo fails loudly outside a git repository rather than
// returning a silent empty/zero value.
func TestResolveGitRoot_NotARepo(t *testing.T) {
	_, err := resolveGitRoot(t.TempDir())
	require.Error(t, err)
}

// TestGitCheckIgnore_HonorsAllExcludeSources proves the switch to
// `git check-ignore` picks up exclude sources go-git's ReadPatterns misses —
// .git/info/exclude in particular — and that --no-index still flags a *tracked*
// file that matches an ignore rule.
func TestGitCheckIgnore_HonorsAllExcludeSources(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, ".gitignore", "*.log\n")
	writeFile(t, repo, ".git/info/exclude", "*.secret\n")

	// A tracked file that also matches *.log: --no-index must still flag it,
	// otherwise committed-but-gitignored bundles never get stripped.
	writeFile(t, repo, "tracked.log", "x\n")
	gitRun(t, repo, "add", "-f", "tracked.log")

	got, err := gitCheckIgnore(repo, []string{"a.log", "b.secret", "c.txt", "tracked.log"})
	require.NoError(t, err)

	assert.Contains(t, got, "a.log", "matched by .gitignore")
	assert.Contains(t, got, "b.secret", "matched by .git/info/exclude")
	assert.Contains(t, got, "tracked.log", "--no-index flags tracked-but-ignored files")
	assert.NotContains(t, got, "c.txt")
}

// TestGitCheckIgnore_NoneIgnored confirms exit code 1 (no path matched) is a
// normal empty result, not an error.
func TestGitCheckIgnore_NoneIgnored(t *testing.T) {
	repo := initCommitRepo(t)
	got, err := gitCheckIgnore(repo, []string{"a.txt", "b.go"})
	require.NoError(t, err)
	assert.Empty(t, got)
}

// TestUnstageGitIgnored_InWorktree is the headline regression: a tracked file
// that the repo's .gitignore excludes must be stripped from the commit even when
// gavel runs inside a linked git worktree (where .git is a file, not a dir).
func TestUnstageGitIgnored_InWorktree(t *testing.T) {
	main := initCommitRepo(t)
	writeFileInDir(t, main, "dist/bundle.js", "old\n")
	writeFile(t, main, ".gitignore", "dist/\n")
	gitRun(t, main, "add", "-f", "dist/bundle.js")
	gitRun(t, main, "add", ".gitignore")
	gitRun(t, main, "commit", "-m", "seed tracked-but-ignored bundle")

	wt := filepath.Join(t.TempDir(), "wt")
	gitRun(t, main, "worktree", "add", wt, "HEAD")
	require.FileExists(t, filepath.Join(wt, ".git"), "worktree records .git as a file")

	writeFileInDir(t, wt, "dist/bundle.js", "new\n")

	require.NoError(t, stageFiles(wt, StageUnstaged, verify.CommitConfig{}))

	assert.NotContains(t, mustStagedFiles(t, wt), "dist/bundle.js",
		"tracked-but-gitignored file must be stripped from the commit inside a worktree")
}
