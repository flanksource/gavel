package commit

import (
	"context"
	"strings"
	"testing"

	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveSinceRef(t *testing.T) {
	repo := initIssueRepo(t)
	cases := []struct {
		name  string
		in    string
		want  string
		isErr bool
	}{
		{"tilde", "~2", "HEAD~2", false},
		{"bare int", "2", "HEAD~2", false},
		{"literal head", "HEAD~1", "HEAD~1", false},
		{"unknown ref", "no-such-ref", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveSinceRef(repo, tc.in)
			if tc.isErr {
				assert.ErrorIs(t, err, ErrSinceInvalidRef)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestBuildRebaseTodoReordersDuplicates(t *testing.T) {
	a := models.Commit{Hash: "aaaaaaa1", Subject: "add a"}
	b := models.Commit{Hash: "bbbbbbb2", Subject: "add b"}
	c := models.Commit{Hash: "ccccccc3", Subject: "add c"}
	ordered := []models.Commit{a, b, c}
	dups := []issueGroup{{IssueID: "1", Commits: []models.Commit{a, c}}}
	msgFiles := map[string]string{"1": "/scratch/msg-0.txt"}

	got := buildRebaseTodo(ordered, dups, msgFiles)

	want := "pick aaaaaaa1 add a\n" +
		"fixup ccccccc3 add c\n" +
		"exec git commit --amend -F /scratch/msg-0.txt\n" +
		"pick bbbbbbb2 add b\n"
	assert.Equal(t, want, got)
}

func TestRunIssueIdSquashMergesDuplicates(t *testing.T) {
	repo := initIssueRepo(t)
	t.Setenv(testEnvVar, "1")

	preCount := commitCount(t, repo)

	result, err := runIssueIdSquash(context.Background(), Options{
		WorkDir:   repo,
		Since:     "HEAD~3",
		AssumeYes: true,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, preCount-1, commitCount(t, repo), "issue 1's two commits should collapse to one")

	// The merged issue-1 commit (now HEAD~1) carries the simplified message and
	// a single Gavel-Issue-Id: 1 trailer; issue 2 (HEAD) is untouched.
	mergedID := strings.TrimSpace(gitOutput(t, repo, "log", "-1", "--format=%(trailers:key=Gavel-Issue-Id,valueonly)", "HEAD~1"))
	assert.Equal(t, "1", mergedID)
	mergedSubject := strings.TrimSpace(gitOutput(t, repo, "log", "-1", "--format=%s", "HEAD~1"))
	assert.Equal(t, stubMessage, mergedSubject)

	headID := strings.TrimSpace(gitOutput(t, repo, "log", "-1", "--format=%(trailers:key=Gavel-Issue-Id,valueonly)", "HEAD"))
	assert.Equal(t, "2", headID)

	fullLog := gitOutput(t, repo, "log", "--format=%B", "HEAD~2..HEAD")
	assert.Equal(t, 1, strings.Count(fullLog, "Gavel-Issue-Id: 1"), "merged commit keeps exactly one issue-1 trailer")
}

func TestRunIssueIdSquashNoDuplicates(t *testing.T) {
	repo := initCommitRepo(t)
	commitWithIssue(t, repo, "a.txt", "a\n", "feat: add a", "1")
	commitWithIssue(t, repo, "b.txt", "b\n", "feat: add b", "2")
	t.Setenv(testEnvVar, "1")

	preCount := commitCount(t, repo)
	_, err := runIssueIdSquash(context.Background(), Options{
		WorkDir:   repo,
		Since:     "HEAD~2",
		AssumeYes: true,
	})
	assert.ErrorIs(t, err, ErrSinceNoDuplicates)
	assert.Equal(t, preCount, commitCount(t, repo), "history must be unchanged")
}

func TestRunIssueIdSquashDryRun(t *testing.T) {
	repo := initIssueRepo(t)
	t.Setenv(testEnvVar, "1")

	preCount := commitCount(t, repo)
	result, err := runIssueIdSquash(context.Background(), Options{
		WorkDir: repo,
		Since:   "HEAD~3",
		DryRun:  true,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.DryRun)
	assert.Equal(t, preCount, commitCount(t, repo), "dry-run must not rewrite history")
}

func TestRunIssueIdSquashNonTTYNeedsConfirm(t *testing.T) {
	repo := initIssueRepo(t)
	t.Setenv(testEnvVar, "1")

	prevTTY := stdinIsTerminal
	stdinIsTerminal = func() bool { return false }
	t.Cleanup(func() { stdinIsTerminal = prevTTY })

	preCount := commitCount(t, repo)
	_, err := runIssueIdSquash(context.Background(), Options{
		WorkDir: repo,
		Since:   "HEAD~3",
	})
	assert.ErrorIs(t, err, ErrSinceNeedsConfirm)
	assert.Equal(t, preCount, commitCount(t, repo), "history must be unchanged without confirmation")
}

func TestPublishedCommitsGuard(t *testing.T) {
	repo := initIssueRepo(t)
	// Pretend the second-from-last commit and older are already pushed.
	gitRun(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD~1")

	published, err := publishedCommits(repo, "HEAD~3")
	require.NoError(t, err)
	assert.Len(t, published, 2, "HEAD~2 and HEAD~1 are reachable from origin/main")

	safe, err := publishedCommits(repo, "origin/main")
	require.NoError(t, err)
	assert.Empty(t, safe, "origin/main..HEAD is entirely unpushed")
}

func TestRunIssueIdSquashRefusesPublished(t *testing.T) {
	repo := initIssueRepo(t)
	gitRun(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD~1")
	t.Setenv(testEnvVar, "1")

	preCount := commitCount(t, repo)
	_, err := runIssueIdSquash(context.Background(), Options{
		WorkDir:   repo,
		Since:     "HEAD~3",
		AssumeYes: true,
	})
	assert.ErrorIs(t, err, ErrSincePushed)
	assert.Equal(t, preCount, commitCount(t, repo), "published history must not be rewritten")
}

// initIssueRepo builds a repo whose history (oldest→newest) is:
//
//	initial commit -> README.md
//	feat: add a    -> a.txt  (Gavel-Issue-Id: 1)
//	feat: add b    -> b.txt  (Gavel-Issue-Id: 2)
//	feat: add c    -> c.txt  (Gavel-Issue-Id: 1)
//
// so HEAD~3..HEAD has two non-adjacent commits sharing Gavel-Issue-Id 1.
func initIssueRepo(t *testing.T) string {
	t.Helper()
	repo := initCommitRepo(t)
	commitWithIssue(t, repo, "a.txt", "a\n", "feat: add a", "1")
	commitWithIssue(t, repo, "b.txt", "b\n", "feat: add b", "2")
	commitWithIssue(t, repo, "c.txt", "c\n", "feat: add c", "1")
	return repo
}

func commitWithIssue(t *testing.T, repo, file, content, subject, issue string) {
	t.Helper()
	writeFile(t, repo, file, content)
	gitRun(t, repo, "add", file)
	gitRun(t, repo, "commit", "-m", subject, "-m", "Gavel-Issue-Id: "+issue)
}

func commitCount(t *testing.T, repo string) int {
	t.Helper()
	out := strings.TrimSpace(gitOutput(t, repo, "rev-list", "--count", "HEAD"))
	n := 0
	for _, r := range out {
		n = n*10 + int(r-'0')
	}
	return n
}
