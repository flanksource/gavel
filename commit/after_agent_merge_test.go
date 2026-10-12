package commit

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	prHeadBranch    = "feature"
	scratchBranch   = "gavel/pr-7-fix"
	conflictedPath  = "go.mod"
	resolvedContent = "require xz v0.5.20\n"
)

// conflictedMergeRepo leaves a repo mid-merge on a scratch branch built from the
// PR head, the state `pr status --ai-fix` hands the agent: main's side of go.mod
// conflicts with the branch's, and MERGE_HEAD is set.
func conflictedMergeRepo(t *testing.T) string {
	t.Helper()
	repo := initCommitRepo(t)
	gitRun(t, repo, "branch", "-M", "main")
	writeFile(t, repo, conflictedPath, "require xz v0.5.12\n")
	gitRun(t, repo, "add", conflictedPath)
	gitRun(t, repo, "commit", "-m", "base")

	gitRun(t, repo, "checkout", "-b", prHeadBranch)
	writeFile(t, repo, conflictedPath, "require xz v0.5.14\n")
	gitRun(t, repo, "commit", "-am", "bump on the branch")

	gitRun(t, repo, "checkout", "main")
	writeFile(t, repo, conflictedPath, resolvedContent)
	gitRun(t, repo, "commit", "-am", "bump on main")

	gitRun(t, repo, "checkout", "-b", scratchBranch, prHeadBranch)
	cmd := exec.Command("git", "merge", "--no-ff", "--no-commit", "main")
	cmd.Dir = repo
	out, err := cmd.CombinedOutput()
	require.Error(t, err, "the merge must stop on the conflict: %s", out)
	_, inProgress, err := github.MergeInProgress(repo)
	require.NoError(t, err)
	require.True(t, inProgress)
	return repo
}

// stubBranchPush records the refspec pushed and fails the test if the push flow
// falls back to searching for a PR: PushBranch names the target outright.
func stubBranchPush(t *testing.T) *string {
	t.Helper()
	pushed := new(string)
	pushDepsForTest = &pushDeps{
		searchPRs: func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
			t.Error("PushBranch must bypass the PR search")
			return nil, nil, nil
		},
		defaultBranch: func(github.Options) (string, error) { return "main", nil },
		rebaseOnto:    func(string, string) error { return nil },
		gitPush: func(_, refspec string) error {
			*pushed = refspec
			return nil
		},
		aheadCommits: loadAheadCommits,
	}
	t.Cleanup(func() { pushDepsForTest = nil })
	return pushed
}

func parentCount(t *testing.T, repo string) int {
	t.Helper()
	return len(strings.Fields(gitOutput(t, repo, "log", "-1", "--format=%P")))
}

func TestRunAfterAgentConcludesAResolvedMergeAndPushesItToThePRBranch(t *testing.T) {
	repo := conflictedMergeRepo(t)
	writeFile(t, repo, conflictedPath, resolvedContent)
	t.Setenv(testEnvVar, "1")
	pushed := stubBranchPush(t)

	result, err := RunAfterAgent(context.Background(), AgentRun{
		WorkDir:    repo,
		Files:      []string{conflictedPath},
		Push:       true,
		PushBranch: prHeadBranch,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Commits, 1)
	assert.Equal(t, 2, parentCount(t, repo), "the conclusion must be a merge commit, not a squash of main's side")
	assert.Equal(t, strings.TrimSpace(gitOutput(t, repo, "rev-parse", "HEAD")), result.Commits[0].Hash)
	assert.Equal(t, "HEAD:"+prHeadBranch, *pushed)
	_, inProgress, err := github.MergeInProgress(repo)
	require.NoError(t, err)
	assert.False(t, inProgress)
}

func TestRunAfterAgentLeavesAnUnresolvedMergeInPlace(t *testing.T) {
	repo := conflictedMergeRepo(t)
	head := gitOutput(t, repo, "rev-parse", "HEAD")
	t.Setenv(testEnvVar, "1")
	pushed := stubBranchPush(t)

	// The agent staged the file without removing the markers.
	_, err := RunAfterAgent(context.Background(), AgentRun{
		WorkDir:    repo,
		Files:      []string{conflictedPath},
		Push:       true,
		PushBranch: prHeadBranch,
	})

	require.ErrorIs(t, err, ErrMergeUnresolved)
	assert.Contains(t, err.Error(), conflictedPath)
	assert.Equal(t, head, gitOutput(t, repo, "rev-parse", "HEAD"))
	assert.Empty(t, *pushed)
	_, inProgress, err := github.MergeInProgress(repo)
	require.NoError(t, err)
	assert.True(t, inProgress, "an unresolved merge must survive for the next turn to finish")
}

func TestRunAfterAgentPushBranchBypassesThePRSearch(t *testing.T) {
	repo := initCommitRepo(t)
	gitRun(t, repo, "checkout", "-b", scratchBranch)
	writeFile(t, repo, "a.txt", "fixed\n")
	t.Setenv(testEnvVar, "1")
	pushed := stubBranchPush(t)

	_, err := RunAfterAgent(context.Background(), AgentRun{
		WorkDir:    repo,
		Files:      []string{"a.txt"},
		Push:       true,
		PushBranch: prHeadBranch,
	})

	require.NoError(t, err)
	assert.Equal(t, "HEAD:"+prHeadBranch, *pushed)
}
