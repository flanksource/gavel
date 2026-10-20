package commit

import (
	"context"
	"os/exec"
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

// conflictedRebaseRepo leaves a repo stopped mid-rebase on a scratch branch
// built from the PR head, the state `pr status --ai-fix` hands the agent: the
// branch's go.mod bump conflicts with main's while being replayed onto it.
func conflictedRebaseRepo(t *testing.T) string {
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
	cmd := exec.Command("git", "rebase", "main")
	cmd.Dir = repo
	out, err := cmd.CombinedOutput()
	require.Error(t, err, "the rebase must stop on the conflict: %s", out)
	_, inProgress, err := github.RebaseInProgress(repo)
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

// A stopped rebase is never committed over, whether its conflict is still open
// or resolved and staged but not yet continued: either commit would conclude
// the replayed commit under a generated message.
func TestRunAfterAgentLeavesAStoppedRebaseInPlace(t *testing.T) {
	tests := []struct {
		name    string
		resolve func(t *testing.T, repo string)
		wantErr string
	}{
		{
			name:    "conflict still marked",
			resolve: func(*testing.T, string) {},
			wantErr: "rebase in progress (unresolved: " + conflictedPath + ")",
		},
		{
			name: "resolved but not continued",
			resolve: func(t *testing.T, repo string) {
				writeFile(t, repo, conflictedPath, resolvedContent)
				gitRun(t, repo, "add", conflictedPath)
			},
			wantErr: "rebase in progress; continue it with `git rebase --continue`",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := conflictedRebaseRepo(t)
			tc.resolve(t, repo)
			head := gitOutput(t, repo, "rev-parse", "HEAD")
			t.Setenv(testEnvVar, "1")
			pushed := stubBranchPush(t)

			_, err := RunAfterAgent(context.Background(), AgentRun{
				WorkDir:    repo,
				Files:      []string{conflictedPath},
				Push:       true,
				PushBranch: prHeadBranch,
			})

			require.ErrorIs(t, err, ErrRebaseInProgress)
			assert.EqualError(t, err, tc.wantErr)
			assert.Equal(t, head, gitOutput(t, repo, "rev-parse", "HEAD"))
			assert.Empty(t, *pushed)
			_, inProgress, err := github.RebaseInProgress(repo)
			require.NoError(t, err)
			assert.True(t, inProgress, "a stopped rebase must survive for the agent to continue")
		})
	}
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
