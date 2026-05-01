package commit

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateFixupOptionsAllowsEmpty(t *testing.T) {
	require.NoError(t, validateFixupOptions(Options{}))
}

func TestValidateFixupOptionsRejectsConflicts(t *testing.T) {
	cases := []struct {
		name string
		opts Options
		want error
	}{
		{"with commit-all", Options{Fixup: FixupAuto, CommitAll: true}, ErrFixupWithCommitAll},
		{"with interactive", Options{Fixup: FixupAuto, Interactive: true}, ErrFixupWithInteractive},
		{"with message", Options{Fixup: FixupAuto, Message: "feat: x"}, ErrFixupWithMessage},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.ErrorIs(t, validateFixupOptions(tc.opts), tc.want)
		})
	}
}

func TestRouteFilesByLastTouchGroupsByCommit(t *testing.T) {
	repo := initFixupRepo(t)
	base := strings.TrimSpace(gitOutput(t, repo, "rev-parse", "HEAD~3"))

	// Edit files that were introduced by older commits + add a fresh file.
	writeFile(t, repo, "a.txt", "a1\nedit\n")
	writeFile(t, repo, "b.txt", "b1\nedit\n")
	writeFile(t, repo, "c.txt", "fresh\n")
	gitRun(t, repo, "add", "a.txt", "b.txt", "c.txt")

	routes, leftovers, err := routeFilesByLastTouch(repo, []string{"a.txt", "b.txt", "c.txt"}, base)
	require.NoError(t, err)
	require.Len(t, routes, 2)

	hashesByFile := map[string]string{}
	for _, r := range routes {
		for _, f := range r.Files {
			hashesByFile[f] = r.Hash
		}
	}
	require.Contains(t, hashesByFile, "a.txt")
	require.Contains(t, hashesByFile, "b.txt")
	assert.NotEqual(t, hashesByFile["a.txt"], hashesByFile["b.txt"], "different files should map to different commits")
	assert.Equal(t, []string{"c.txt"}, leftovers)
}

func TestRouteFilesByLastTouchAllLeftoversWhenNoMatches(t *testing.T) {
	repo := initFixupRepo(t)
	base := strings.TrimSpace(gitOutput(t, repo, "rev-parse", "HEAD"))

	// All staged files are brand new — none touched by base..HEAD (which is empty).
	writeFile(t, repo, "x.txt", "new\n")
	gitRun(t, repo, "add", "x.txt")

	routes, leftovers, err := routeFilesByLastTouch(repo, []string{"x.txt"}, base)
	require.NoError(t, err)
	assert.Empty(t, routes)
	assert.Equal(t, []string{"x.txt"}, leftovers)
}

func TestResolveFixupBasePrefersOriginMain(t *testing.T) {
	repo := initFixupRepo(t)
	gitRun(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD~3")

	base, err := resolveFixupBase(repo)
	require.NoError(t, err)
	assert.Equal(t, "origin/main", base)
}

func TestResolveFixupBaseFallsBackToMaster(t *testing.T) {
	repo := initFixupRepo(t)
	gitRun(t, repo, "update-ref", "refs/remotes/origin/master", "HEAD~3")

	base, err := resolveFixupBase(repo)
	require.NoError(t, err)
	assert.Equal(t, "origin/master", base)
}

func TestResolveFixupBaseErrorsWhenNothingMatches(t *testing.T) {
	repo := initFixupRepo(t)

	_, err := resolveFixupBase(repo)
	assert.ErrorIs(t, err, ErrFixupNoBase)
}

func TestRunFixupExplicitHashCreatesFixupCommit(t *testing.T) {
	repo := initFixupRepo(t)
	gitRun(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD~3")
	target := strings.TrimSpace(gitOutput(t, repo, "rev-parse", "HEAD~2")) // commit "feat: add a"

	writeFile(t, repo, "a.txt", "a1\nfix\n")
	gitRun(t, repo, "add", "a.txt")

	t.Setenv(testEnvVar, "1")

	result, err := Run(context.Background(), Options{
		WorkDir: repo,
		Fixup:   target,
	})
	require.NoError(t, err)
	require.Len(t, result.Commits, 1)
	assert.Equal(t, []string{"a.txt"}, result.Commits[0].Files)

	headMsg := strings.TrimSpace(gitOutput(t, repo, "log", "-1", "--format=%s", "HEAD"))
	assert.True(t, strings.HasPrefix(headMsg, "fixup!"), "expected fixup! commit, got %q", headMsg)
	assert.Equal(t, "5", strings.TrimSpace(gitOutput(t, repo, "rev-list", "--count", "HEAD")))
}

func TestRunFixupAutoRoutesPlusLeftover(t *testing.T) {
	repo := initFixupRepo(t)
	gitRun(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD~3")

	writeFile(t, repo, "a.txt", "a1\nfix\n")
	writeFile(t, repo, "b.txt", "b1\nfix\n")
	writeFile(t, repo, "c.txt", "fresh\n")
	gitRun(t, repo, "add", "a.txt", "b.txt", "c.txt")

	t.Setenv(testEnvVar, "1")

	result, err := Run(context.Background(), Options{
		WorkDir: repo,
		Fixup:   FixupAuto,
	})
	require.NoError(t, err)
	require.Len(t, result.Commits, 3, "2 fixups + 1 leftover")

	subjects := strings.TrimSpace(gitOutput(t, repo, "log", "--format=%s", "HEAD~3..HEAD"))
	fixupCount := strings.Count(subjects, "fixup!")
	assert.Equal(t, 2, fixupCount, "expected 2 fixup! commits in last 3, got log:\n%s", subjects)
	assert.Equal(t, "7", strings.TrimSpace(gitOutput(t, repo, "rev-list", "--count", "HEAD")))
}

func TestRunFixupAutosquashCollapsesHistory(t *testing.T) {
	repo := initFixupRepo(t)
	gitRun(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD~3")

	writeFile(t, repo, "a.txt", "a1\nfix\n")
	gitRun(t, repo, "add", "a.txt")

	t.Setenv(testEnvVar, "1")

	target := strings.TrimSpace(gitOutput(t, repo, "rev-parse", "HEAD~2"))
	preCount := strings.TrimSpace(gitOutput(t, repo, "rev-list", "--count", "HEAD"))

	_, err := Run(context.Background(), Options{
		WorkDir:    repo,
		Fixup:      target,
		Autosquash: true,
	})
	require.NoError(t, err)

	postCount := strings.TrimSpace(gitOutput(t, repo, "rev-list", "--count", "HEAD"))
	assert.Equal(t, preCount, postCount, "autosquash should leave the same commit count, not add one")

	// The collapsed target should now contain the fix line.
	blob := gitOutput(t, repo, "show", "HEAD~2:a.txt")
	assert.Contains(t, blob, "fix\n")
}

func TestRunFixupNoBaseErrorsClearly(t *testing.T) {
	repo := initFixupRepo(t)
	writeFile(t, repo, "a.txt", "a1\nfix\n")
	gitRun(t, repo, "add", "a.txt")

	t.Setenv(testEnvVar, "1")

	_, err := Run(context.Background(), Options{
		WorkDir: repo,
		Fixup:   FixupAuto,
	})
	assert.ErrorIs(t, err, ErrFixupNoBase)
}

func TestRunFixupInvalidTarget(t *testing.T) {
	repo := initFixupRepo(t)
	writeFile(t, repo, "a.txt", "a1\nfix\n")
	gitRun(t, repo, "add", "a.txt")

	t.Setenv(testEnvVar, "1")

	_, err := Run(context.Background(), Options{
		WorkDir: repo,
		Fixup:   "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
	})
	assert.ErrorIs(t, err, ErrFixupInvalidTarget)
}

// initFixupRepo builds a repo with this history (oldest→newest):
//
//	initial commit  -> README.md
//	feat: add a     -> a.txt = "a1\n"
//	feat: add b     -> b.txt = "b1\n"
//	feat: add d     -> d.txt = "d1\n"
//
// Tests typically point origin/main at HEAD~3 so base..HEAD covers the three
// feat commits.
func initFixupRepo(t *testing.T) string {
	t.Helper()
	repo := initCommitRepo(t)
	gitRun(t, repo, "config", "rebase.autosquash", "true")

	steps := []struct{ file, content, msg string }{
		{"a.txt", "a1\n", "feat: add a"},
		{"b.txt", "b1\n", "feat: add b"},
		{"d.txt", "d1\n", "feat: add d"},
	}
	for _, s := range steps {
		require.NoError(t, os.WriteFile(filepath.Join(repo, s.file), []byte(s.content), 0o644))
		gitRun(t, repo, "add", s.file)
		gitRun(t, repo, "commit", "-m", s.msg)
	}
	return repo
}
