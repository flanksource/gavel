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

func TestValidateCommitPlanRejectsUnknownFile(t *testing.T) {
	changes := []stagedChange{{Path: "known.txt"}}

	_, err := validateCommitPlan([]commitGroupSpec{{Files: []string{"unknown.txt"}}}, changes)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCommitAllPlan)
	assert.Contains(t, err.Error(), "unknown.txt")
}

func TestValidateCommitPlanRejectsDuplicateFile(t *testing.T) {
	changes := []stagedChange{{Path: "known.txt"}}

	_, err := validateCommitPlan([]commitGroupSpec{
		{Files: []string{"known.txt"}},
		{Files: []string{"known.txt"}},
	}, changes)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCommitAllPlan)
	assert.Contains(t, err.Error(), "multiple groups")
}

func TestValidateCommitPlanRejectsMissingFile(t *testing.T) {
	changes := []stagedChange{{Path: "a.txt"}, {Path: "b.txt"}}

	_, err := validateCommitPlan([]commitGroupSpec{{Files: []string{"a.txt"}}}, changes)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCommitAllPlan)
	assert.Contains(t, err.Error(), "b.txt")
}

func TestParseCommitGroupResponseAcceptsTypedStructuredData(t *testing.T) {
	groups, ok, err := parseCommitGroupResponse(&commitGroupPlanSchema{
		Groups: []commitGroupSpec{{Label: "first", Files: []string{"a.txt"}}},
	}, "")
	require.NoError(t, err)
	require.True(t, ok)
	require.Len(t, groups, 1)
	assert.Equal(t, "first", groups[0].Label)
}

func TestParseCommitGroupResponseAcceptsGenericStructuredData(t *testing.T) {
	groups, ok, err := parseCommitGroupResponse(map[string]any{
		"groups": []any{
			map[string]any{
				"label": "first",
				"files": []any{"a.txt", "b.txt"},
			},
		},
	}, "")
	require.NoError(t, err)
	require.True(t, ok)
	require.Len(t, groups, 1)
	assert.Equal(t, []string{"a.txt", "b.txt"}, groups[0].Files)
}

func TestParseCommitGroupResponseAcceptsYAMLResult(t *testing.T) {
	raw := "```yaml\ngroups:\n  - label: first\n    files:\n      - a.txt\n      - b.txt\n```"
	groups, ok, err := parseCommitGroupResponse(nil, raw)
	require.NoError(t, err)
	require.True(t, ok)
	require.Len(t, groups, 1)
	assert.Equal(t, "first", groups[0].Label)
	assert.Equal(t, []string{"a.txt", "b.txt"}, groups[0].Files)
}

func TestRunCommitAllSplitsStagedChanges(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "a.txt", "one\n")
	writeFile(t, repo, "b.txt", "two\n")
	gitRun(t, repo, "add", "a.txt", "b.txt")

	t.Setenv(testEnvVar, "1")
	restore := stubCommitPlanner(func(context.Context, Options, []stagedChange) ([]commitGroupSpec, error) {
		return []commitGroupSpec{
			{Label: "first", Files: []string{"a.txt"}},
			{Label: "second", Files: []string{"b.txt"}},
		}, nil
	})
	defer restore()

	result, err := Run(context.Background(), Options{WorkDir: repo, CommitAll: true})
	require.NoError(t, err)
	require.Len(t, result.Commits, 2)
	assert.Equal(t, []string{"a.txt"}, result.Commits[0].Files)
	assert.Equal(t, []string{"b.txt"}, result.Commits[1].Files)
	assert.NotEmpty(t, result.Commits[0].Hash)
	assert.NotEmpty(t, result.Commits[1].Hash)
	assert.Equal(t, "3", strings.TrimSpace(gitOutput(t, repo, "rev-list", "--count", "HEAD")))
	assert.Empty(t, strings.TrimSpace(gitOutput(t, repo, "status", "--short")))
}

func TestRunCommitAllStagesAllWhenNothingIsStaged(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "a.txt", "one\n")
	writeFile(t, repo, "b.txt", "two\n")

	t.Setenv(testEnvVar, "1")

	result, err := Run(context.Background(), Options{WorkDir: repo, CommitAll: true})
	require.NoError(t, err)
	require.Len(t, result.Commits, 2)
	assert.ElementsMatch(t, []string{"a.txt", "b.txt"}, result.Staged)
	assert.Equal(t, "3", strings.TrimSpace(gitOutput(t, repo, "rev-list", "--count", "HEAD")))
}

func TestRunCommitAllUsesOnlyExistingStagedSet(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "a.txt", "staged\n")
	writeFile(t, repo, "b.txt", "unstaged\n")
	gitRun(t, repo, "add", "a.txt")

	t.Setenv(testEnvVar, "1")

	result, err := Run(context.Background(), Options{WorkDir: repo, CommitAll: true})
	require.NoError(t, err)
	require.Len(t, result.Commits, 1)
	assert.Equal(t, []string{"a.txt"}, result.Staged)
	assert.Contains(t, gitOutput(t, repo, "status", "--short"), "?? b.txt")
}

func TestRunCommitAllRunsHooksOnceUpfront(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "a.txt", "one\n")
	writeFile(t, repo, "b.txt", "two\n")
	gitRun(t, repo, "add", "a.txt", "b.txt")

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
	writeFile(t, repo, "a.txt", "one\n")
	writeFile(t, repo, "b.txt", "two\n")
	gitRun(t, repo, "add", "a.txt", "b.txt")

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
	writeFile(t, repo, "a.txt", "one\n")
	writeFile(t, repo, "b.txt", "two\n")
	gitRun(t, repo, "add", "a.txt", "b.txt")

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
	assert.Contains(t, clean, "commit dry-run/1 of 2")
	assert.Contains(t, clean, "Files:")
	assert.Contains(t, clean, "a.txt")
	assert.Contains(t, clean, "b.txt")
}

func TestMergeGroupsToMax(t *testing.T) {
	groups := []commitGroupSpec{
		{Label: "a", Files: []string{"a.txt"}},
		{Label: "b", Files: []string{"b.txt"}},
		{Label: "c", Files: []string{"c.txt"}},
		{Label: "d", Files: []string{"d.txt"}},
		{Label: "e", Files: []string{"e.txt"}},
	}

	t.Run("zero means unlimited", func(t *testing.T) {
		result := mergeGroupsToMax(groups, 0)
		assert.Len(t, result, 5)
	})

	t.Run("max greater than count is noop", func(t *testing.T) {
		result := mergeGroupsToMax(groups, 10)
		assert.Len(t, result, 5)
	})

	t.Run("max merges trailing groups", func(t *testing.T) {
		result := mergeGroupsToMax(groups, 2)
		require.Len(t, result, 2)
		assert.Equal(t, []string{"a.txt"}, result[0].Files)
		assert.Equal(t, []string{"b.txt", "c.txt", "d.txt", "e.txt"}, result[1].Files)
	})

	t.Run("max equal to count is noop", func(t *testing.T) {
		result := mergeGroupsToMax(groups, 5)
		assert.Len(t, result, 5)
	})

	t.Run("max 1 collapses all into one", func(t *testing.T) {
		result := mergeGroupsToMax(groups, 1)
		require.Len(t, result, 1)
		assert.Equal(t, []string{"a.txt", "b.txt", "c.txt", "d.txt", "e.txt"}, result[0].Files)
	})
}

func TestRunCommitAllWithMax(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "a.txt", "one\n")
	writeFile(t, repo, "b.txt", "two\n")
	writeFile(t, repo, "c.txt", "three\n")
	gitRun(t, repo, "add", "a.txt", "b.txt", "c.txt")

	t.Setenv(testEnvVar, "1")

	result, err := Run(context.Background(), Options{WorkDir: repo, CommitAll: true, Max: 2})
	require.NoError(t, err)
	require.Len(t, result.Commits, 2)
	assert.Equal(t, []string{"a.txt"}, result.Commits[0].Files)
	assert.Equal(t, []string{"b.txt", "c.txt"}, result.Commits[1].Files)
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

func readFile(t *testing.T, path string) string {
	t.Helper()
	out, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(out)
}

func stubCommitPlanner(fn func(context.Context, Options, []stagedChange) ([]commitGroupSpec, error)) func() {
	previous := planCommitGroupsFunc
	planCommitGroupsFunc = fn
	return func() {
		planCommitGroupsFunc = previous
	}
}

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;:]*[A-Za-z]`)

func stripANSIForTest(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}
