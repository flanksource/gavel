package commit

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/flanksource/gavel/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupChangesByDirSingleDirFits(t *testing.T) {
	changes := []stagedChange{
		{Path: "linters/eslint/eslint.go", Adds: 3, Dels: 1},
		{Path: "linters/tsc/tsc.go", Adds: 5, Dels: 2},
		{Path: "linters/ruff/ruff.go", Adds: 4, Dels: 5},
	}

	groups := groupChangesByDir(changes, 7, 100)
	require.Len(t, groups, 1)
	assert.Equal(t, "linters", groups[0].Label)
	assert.ElementsMatch(t,
		[]string{"linters/eslint/eslint.go", "linters/tsc/tsc.go", "linters/ruff/ruff.go"},
		groups[0].Files())
}

func TestGroupChangesByDirMultipleDirs(t *testing.T) {
	changes := []stagedChange{
		{Path: "linters/eslint/eslint.go", Adds: 2, Dels: 1},
		{Path: "cmd/gavel/lint.go", Adds: 2, Dels: 1},
	}

	groups := groupChangesByDir(changes, 7, 100)
	require.Len(t, groups, 2)
	assert.Equal(t, "cmd", groups[0].Label)
	assert.Equal(t, "linters", groups[1].Label)
}

func TestGroupChangesByDirDescendsOnMaxFiles(t *testing.T) {
	changes := []stagedChange{
		{Path: "linters/eslint/a.go"},
		{Path: "linters/eslint/b.go"},
		{Path: "linters/eslint/c.go"},
		{Path: "linters/tsc/a.go"},
		{Path: "linters/tsc/b.go"},
		{Path: "linters/ruff/a.go"},
		{Path: "linters/ruff/b.go"},
		{Path: "linters/ruff/c.go"},
		{Path: "linters/ruff/d.go"},
	}

	groups := groupChangesByDir(changes, 4, 0)
	require.Len(t, groups, 3)
	assert.Equal(t, "linters/eslint", groups[0].Label)
	assert.Equal(t, "linters/ruff", groups[1].Label)
	assert.Equal(t, "linters/tsc", groups[2].Label)
	assert.Len(t, groups[1].Files(), 4)
}

func TestGroupChangesByDirDescendsOnMaxLines(t *testing.T) {
	changes := []stagedChange{
		{Path: "linters/eslint/a.go", Adds: 40, Dels: 20},
		{Path: "linters/eslint/b.go", Adds: 15, Dels: 10},
		{Path: "linters/tsc/a.go", Adds: 50, Dels: 10},
	}

	// linters bucket sums to 145 lines (>100), so it splits into sub-dirs.
	// linters/eslint is 85 lines, linters/tsc is 60 lines — both fit individually.
	groups := groupChangesByDir(changes, 0, 100)
	require.Len(t, groups, 2)
	assert.Equal(t, "linters/eslint", groups[0].Label)
	assert.Equal(t, "linters/tsc", groups[1].Label)
}

func TestGroupChangesByDirDescendsFurtherWhenSubdirStillExceedsLines(t *testing.T) {
	changes := []stagedChange{
		{Path: "linters/eslint/a.go", Adds: 200, Dels: 50},
		{Path: "linters/eslint/b.go", Adds: 40, Dels: 10},
		{Path: "linters/tsc/a.go", Adds: 5, Dels: 2},
	}

	// linters/ splits into eslint and tsc. linters/eslint is 300 lines (over
	// budget) but flat — stays as one oversized group rather than one commit
	// per file.
	groups := groupChangesByDir(changes, 0, 100)
	require.Len(t, groups, 2)
	assert.Equal(t, "linters/eslint", groups[0].Label)
	assert.Equal(t, "linters/tsc", groups[1].Label)
	assert.Len(t, groups[0].Files(), 2)
}

func TestGroupChangesByDirRepoRootFiles(t *testing.T) {
	changes := []stagedChange{
		{Path: "Makefile", Adds: 2, Dels: 1},
		{Path: "go.mod", Adds: 3, Dels: 0},
	}

	groups := groupChangesByDir(changes, 7, 100)
	require.Len(t, groups, 1)
	assert.Equal(t, rootGroupLabel, groups[0].Label)
}

func TestGroupChangesByDirHandlesRenames(t *testing.T) {
	changes := []stagedChange{
		{Path: "linters/eslint/renamed.go", PreviousPath: "linters/old/renamed.go", Adds: 0, Dels: 0},
		{Path: "cmd/gavel/lint.go", Adds: 1, Dels: 0},
	}

	groups := groupChangesByDir(changes, 7, 100)
	require.Len(t, groups, 2)
	labels := []string{groups[0].Label, groups[1].Label}
	assert.Contains(t, labels, "cmd")
	assert.Contains(t, labels, "linters")
}

func TestGroupChangesByDirKeepsFlatDirAsSingleOversizedGroup(t *testing.T) {
	changes := []stagedChange{
		{Path: "commit/commit.go", Adds: 16, Dels: 11},
		{Path: "commit/commit_all_test.go", Adds: 164, Dels: 130},
		{Path: "commit/planner.go", Adds: 64, Dels: 190},
		{Path: "commit/ai-commit-group.md", Adds: 0, Dels: 27},
	}

	// commit/ has 4 flat files totalling 602 lines (>500), but no
	// subdirectories to split by. Keep as one oversized group, not four.
	groups := groupChangesByDir(changes, 7, 500)
	require.Len(t, groups, 1)
	assert.Equal(t, "commit", groups[0].Label)
	assert.Len(t, groups[0].Files(), 4)
}

func TestGroupChangesByDirEmitsOversizedLeaf(t *testing.T) {
	changes := []stagedChange{
		{Path: "linters/eslint/huge.go", Adds: 400, Dels: 200},
	}

	groups := groupChangesByDir(changes, 7, 100)
	require.Len(t, groups, 1)
	assert.Equal(t, []string{"linters/eslint/huge.go"}, groups[0].Files())
}

func TestGroupChangesByDirExcludesNewFilesFromLineBudget(t *testing.T) {
	changes := []stagedChange{
		{Path: "linters/eslint/new.go", Status: "inserted", Adds: 900, Dels: 0},
		{Path: "linters/eslint/existing.go", Status: "updated", Adds: 30, Dels: 10},
		{Path: "linters/tsc/existing.go", Status: "updated", Adds: 20, Dels: 5},
	}

	// Budget ignores the 900-line new file; remaining 65 lines fit under 100,
	// so the whole linters bucket stays as one group.
	groups := groupChangesByDir(changes, 0, 100)
	require.Len(t, groups, 1)
	assert.Equal(t, "linters", groups[0].Label)
	assert.Len(t, groups[0].Files(), 3)
}

func TestGroupChangesByDirIsDeterministicallyOrdered(t *testing.T) {
	changes := []stagedChange{
		{Path: "zeta/file.go"},
		{Path: "alpha/file.go"},
		{Path: "mu/file.go"},
	}

	groups := groupChangesByDir(changes, 7, 100)
	require.Len(t, groups, 3)
	assert.Equal(t, "alpha", groups[0].Label)
	assert.Equal(t, "mu", groups[1].Label)
	assert.Equal(t, "zeta", groups[2].Label)
}

func TestRunCommitAllSplitsStagedChanges(t *testing.T) {
	repo := initCommitRepo(t)
	writeFileInDir(t, repo, "alpha/a.txt", "one\n")
	writeFileInDir(t, repo, "beta/b.txt", "two\n")
	gitRun(t, repo, "add", "alpha/a.txt", "beta/b.txt")

	t.Setenv(testEnvVar, "1")

	result, err := Run(context.Background(), Options{WorkDir: repo, CommitAll: true})
	require.NoError(t, err)
	require.Len(t, result.Commits, 2)
	assert.Equal(t, []string{"alpha/a.txt"}, result.Commits[0].Files)
	assert.Equal(t, []string{"beta/b.txt"}, result.Commits[1].Files)
	assert.NotEmpty(t, result.Commits[0].Hash)
	assert.NotEmpty(t, result.Commits[1].Hash)
	assert.Equal(t, "3", strings.TrimSpace(gitOutput(t, repo, "rev-list", "--count", "HEAD")))
	assert.Empty(t, strings.TrimSpace(gitOutput(t, repo, "status", "--short")))
}

func TestRunCommitAllStagesAllWhenNothingIsStaged(t *testing.T) {
	repo := initCommitRepo(t)
	writeFileInDir(t, repo, "alpha/a.txt", "one\n")
	writeFileInDir(t, repo, "beta/b.txt", "two\n")

	t.Setenv(testEnvVar, "1")

	result, err := Run(context.Background(), Options{WorkDir: repo, CommitAll: true})
	require.NoError(t, err)
	require.Len(t, result.Commits, 2)
	assert.ElementsMatch(t, []string{"alpha/a.txt", "beta/b.txt"}, result.Staged)
	assert.Equal(t, "3", strings.TrimSpace(gitOutput(t, repo, "rev-list", "--count", "HEAD")))
}

func TestRunCommitAllUsesOnlyExistingStagedSet(t *testing.T) {
	repo := initCommitRepo(t)
	writeFileInDir(t, repo, "alpha/a.txt", "staged\n")
	writeFileInDir(t, repo, "beta/b.txt", "unstaged\n")
	gitRun(t, repo, "add", "alpha/a.txt")

	t.Setenv(testEnvVar, "1")

	result, err := Run(context.Background(), Options{WorkDir: repo, CommitAll: true})
	require.NoError(t, err)
	require.Len(t, result.Commits, 1)
	assert.Equal(t, []string{"alpha/a.txt"}, result.Staged)
	assert.Contains(t, gitOutput(t, repo, "status", "--short"), "?? beta/")
}

func TestRunCommitAllRunsHooksOnceUpfront(t *testing.T) {
	repo := initCommitRepo(t)
	writeFileInDir(t, repo, "alpha/a.txt", "one\n")
	writeFileInDir(t, repo, "beta/b.txt", "two\n")
	gitRun(t, repo, "add", "alpha/a.txt", "beta/b.txt")

	t.Setenv(testEnvVar, "1")

	result, err := Run(context.Background(), Options{
		WorkDir:   repo,
		CommitAll: true,
		Config: verify.CommitConfig{
			Hooks: []verify.CommitHook{
				{Name: "count", Run: "printf x >> hook.log"},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Hooks, 1)
	assert.Equal(t, "x", readFile(t, filepath.Join(repo, "hook.log")))
}

func TestRunCommitAllDryRunDoesNotCreateCommits(t *testing.T) {
	repo := initCommitRepo(t)
	writeFileInDir(t, repo, "alpha/a.txt", "one\n")
	writeFileInDir(t, repo, "beta/b.txt", "two\n")
	gitRun(t, repo, "add", "alpha/a.txt", "beta/b.txt")

	t.Setenv(testEnvVar, "1")

	result, err := Run(context.Background(), Options{WorkDir: repo, CommitAll: true, DryRun: true})
	require.NoError(t, err)
	require.Len(t, result.Commits, 2)
	assert.Empty(t, result.Commits[0].Hash)
	assert.Empty(t, result.Commits[1].Hash)
	assert.Equal(t, "1", strings.TrimSpace(gitOutput(t, repo, "rev-list", "--count", "HEAD")))
}

func TestRunCommitAllDryRunPrintsPreview(t *testing.T) {
	repo := initCommitRepo(t)
	writeFileInDir(t, repo, "alpha/a.txt", "one\n")
	writeFileInDir(t, repo, "beta/b.txt", "two\n")
	gitRun(t, repo, "add", "alpha/a.txt", "beta/b.txt")

	t.Setenv(testEnvVar, "1")

	var buf bytes.Buffer
	previous := dryRunOutput
	dryRunOutput = &buf
	defer func() {
		dryRunOutput = previous
	}()

	_, err := Run(context.Background(), Options{WorkDir: repo, CommitAll: true, DryRun: true})
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "\x1b[")
	clean := stripANSIForTest(out)
	assert.Contains(t, clean, "DRY RUN")
	assert.Contains(t, clean, "would create 2 commit(s)")
	assert.Contains(t, clean, "dry-run/1 of 2")
	assert.Contains(t, clean, "dry-run/2 of 2")
	assert.NotContains(t, clean, "Files:")
}

func TestRunCommitAllSplitsLargeDirByMaxFiles(t *testing.T) {
	repo := initCommitRepo(t)
	writeFileInDir(t, repo, "linters/eslint/a.go", "a\n")
	writeFileInDir(t, repo, "linters/eslint/b.go", "b\n")
	writeFileInDir(t, repo, "linters/tsc/a.go", "a\n")
	gitRun(t, repo, "add", "linters")

	t.Setenv(testEnvVar, "1")

	result, err := Run(context.Background(), Options{
		WorkDir:   repo,
		CommitAll: true,
		MaxFiles:  2,
	})
	require.NoError(t, err)
	require.Len(t, result.Commits, 2)
	assert.Equal(t, []string{"linters/eslint/a.go", "linters/eslint/b.go"}, result.Commits[0].Files)
	assert.Equal(t, []string{"linters/tsc/a.go"}, result.Commits[1].Files)
}

func TestRunCommitAllRejectsExplicitMessage(t *testing.T) {
	repo := initCommitRepo(t)

	_, err := Run(context.Background(), Options{
		WorkDir:   repo,
		CommitAll: true,
		Message:   "feat: nope",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCommitAllWithMessage)
}

func initCommitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, dir, "init")
	gitRun(t, dir, "config", "user.email", "test@example.com")
	gitRun(t, dir, "config", "user.name", "Test User")
	gitRun(t, dir, "config", "commit.gpgsign", "false")
	writeFile(t, dir, "README.md", "# test\n")
	gitRun(t, dir, "add", "README.md")
	gitRun(t, dir, "commit", "-m", "initial commit")
	return dir
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed: %s", args, out)
}

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed: %s", args, out)
	return string(out)
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
}

func writeFileInDir(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	out, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(out)
}

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;:]*[A-Za-z]`)

func stripANSIForTest(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}
