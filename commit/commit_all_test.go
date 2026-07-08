package commit

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/flanksource/gavel/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubGroupPerFile stubs the AI grouping seam so the -A integration tests
// exercise the commit machinery without an LLM: each staged change becomes its
// own commit group, ordered by path for deterministic assertions.
func stubGroupPerFile(t *testing.T) {
	t.Helper()
	orig := groupChangesByAIFunc
	t.Cleanup(func() { groupChangesByAIFunc = orig })
	groupChangesByAIFunc = func(_ context.Context, _ Options, source stagedSource) ([]commitGroup, error) {
		sorted := append([]stagedChange(nil), source.Changes...)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].Path < sorted[j].Path })
		groups := make([]commitGroup, 0, len(sorted))
		for _, c := range sorted {
			groups = append(groups, commitGroup{Label: c.Path, Changes: []stagedChange{c}})
		}
		return groups, nil
	}
}

func TestRunCommitAllSplitsStagedChanges(t *testing.T) {
	repo := initCommitRepo(t)
	writeFileInDir(t, repo, "alpha/a.txt", "one\n")
	writeFileInDir(t, repo, "beta/b.txt", "two\n")
	gitRun(t, repo, "add", "alpha/a.txt", "beta/b.txt")

	t.Setenv(testEnvVar, "1")
	stubGroupPerFile(t)

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
	stubGroupPerFile(t)

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
	stubGroupPerFile(t)

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
	stubGroupPerFile(t)

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
	stubGroupPerFile(t)

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
	stubGroupPerFile(t)

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

// TestRunCommitAllExcludesGitIgnoredAndGavelIgnored drives the full
// `gavel commit -A` path with untracked files blocked by .gitignore and by
// .gavel.yaml commit.gitignore. It reproduces the reported bug — `-A` staging
// ignored files via `git add -A` — and guards that both ignore sources are
// honored: only the normal file is committed, the ignored ones stay untracked.
func TestRunCommitAllExcludesGitIgnoredAndGavelIgnored(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, ".gitignore", "*.secret\n")
	gitRun(t, repo, "add", ".gitignore")
	gitRun(t, repo, "commit", "-m", "add gitignore")

	writeFileInDir(t, repo, "app/main.go", "package main\n") // normal -> committed
	writeFileInDir(t, repo, "app/token.secret", "TOKEN=1\n") // .gitignore -> excluded
	writeFileInDir(t, repo, "secrets/keys.env", "KEY=1\n")   // .gavel commit.gitignore -> excluded

	t.Setenv(testEnvVar, "1")
	stubGroupPerFile(t)

	result, err := Run(context.Background(), Options{
		WorkDir:   repo,
		CommitAll: true,
		Config:    verify.CommitConfig{GitIgnore: []string{"*.env"}},
	})
	require.NoError(t, err)

	assert.Contains(t, result.Staged, "app/main.go")
	assert.NotContains(t, result.Staged, "app/token.secret", ".gitignore'd file must not be committed by -A")
	assert.NotContains(t, result.Staged, "secrets/keys.env", ".gavel.yaml commit.gitignore file must not be committed by -A")

	// The committed tree (clean index after -A) tracks only the normal file.
	tracked := gitOutput(t, repo, "ls-files")
	assert.Contains(t, tracked, "app/main.go")
	assert.NotContains(t, tracked, "app/token.secret", ".gitignore'd file must not be committed by -A")
	assert.NotContains(t, tracked, "secrets/keys.env", ".gavel.yaml commit.gitignore file must not be committed by -A")

	// Both ignored files are left untouched in the working tree, not deleted.
	assert.FileExists(t, filepath.Join(repo, "app/token.secret"))
	assert.FileExists(t, filepath.Join(repo, "secrets/keys.env"))
}

func initCommitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, dir, "init")
	gitRun(t, dir, "config", "user.email", "test@example.com")
	gitRun(t, dir, "config", "user.name", "Test User")
	gitRun(t, dir, "config", "commit.gpgsign", "false")
	// Isolate from the developer's global excludes (~/.config/git/ignore),
	// which commonly lists .env etc. and would otherwise make `git add .env`
	// fail with "paths are ignored" in gitignore-check tests.
	gitRun(t, dir, "config", "core.excludesFile", "/dev/null")
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
