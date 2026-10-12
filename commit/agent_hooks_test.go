package commit

import (
	"context"
	"testing"

	"github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	capcommit "github.com/flanksource/captain/pkg/ai/agent/commit"
	"github.com/flanksource/gavel/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubExistingPRPush points the push flow at an existing PR whose head branch is
// the repo's current branch, and records the refspec `git push` was asked for.
func stubExistingPRPush(t *testing.T, repo string) *string {
	t.Helper()
	// A feature branch, not the repo's default: pushing to a protected branch asks
	// for confirmation, which a test must never depend on.
	branch := "feat/agent-push"
	gitRun(t, repo, "checkout", "-b", branch)
	pushed := new(string)
	pushDepsForTest = &pushDeps{
		searchPRs: func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
			return github.PRSearchResults{
				{Number: 7, Source: branch, Target: "main", URL: "http://x/7"},
			}, nil, nil
		},
		defaultBranch: func(github.Options) (string, error) { return "main", nil },
		isAncestor:    gitIsAncestor,
		rebaseOnto:    func(string, string) error { return nil },
		gitPush: func(_, refspec string) error {
			*pushed = refspec
			return nil
		},
		pickPR:       choosePR,
		aheadCommits: loadAheadCommits,
	}
	t.Cleanup(func() { pushDepsForTest = nil })
	return pushed
}

// A verify loop that re-polls the remote (`gavel pr status`) only sees a fix once
// it is pushed, so AgentRun.Push must reach commit.Options.Push.
func TestRunAfterAgentPushesTheBranchWhenPushIsSet(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "a.txt", "fixed\n")
	t.Setenv(testEnvVar, "1")
	pushed := stubExistingPRPush(t, repo)

	result, err := RunAfterAgent(context.Background(), AgentRun{
		WorkDir: repo,
		Files:   []string{"a.txt"},
		Push:    true,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Commits, 1)
	assert.Equal(t, "HEAD:feat/agent-push", *pushed)
}

func TestRunAfterAgentDoesNotPushByDefault(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "a.txt", "fixed\n")
	t.Setenv(testEnvVar, "1")
	pushed := stubExistingPRPush(t, repo)

	_, err := RunAfterAgent(context.Background(), AgentRun{WorkDir: repo, Files: []string{"a.txt"}})
	require.NoError(t, err)
	assert.Empty(t, *pushed, "a run that did not ask to push must leave the remote alone")
}

// A post-run commit (the default `on: run` policy) fires after captain cleared
// hc.Turn, so the session trailer must come from the run's settled session id on
// the workspace — a todo run whose metadata had none committed with only
// Gavel-Issue-Id.
func TestAgentCommitterStampsTheRunSessionAfterTheTurnEnds(t *testing.T) {
	const issueID, sessionID = "issue-1", "sess-settled"
	repo := initCommitRepo(t)
	writeFile(t, repo, "a.txt", "agent edit\n")
	t.Setenv(testEnvVar, "1")
	hc := &agent.HookContext{Context: context.Background(), Phase: agent.PhaseRun, Response: &ai.Response{}}
	hc.Workspace().SessionID = sessionID

	commit := agentCommitter(AgentHooksOptions{Meta: AgentRunMetadata{IssueID: issueID}}, false)
	sha, err := commit(hc, capcommit.Plan{Dir: repo, Phase: agent.PhaseRun, Paths: []string{"a.txt"}})
	require.NoError(t, err)
	require.NotEmpty(t, sha)

	body := gitOutput(t, repo, "log", "-1", "--format=%B", sha)
	assert.Contains(t, body, TrailerIssueID+": "+issueID)
	assert.Contains(t, body, trailerSessionID+": "+sessionID)
}

// The path set captain's commit hook selected is the whole point of AgentRun.Files:
// unrelated working-tree changes must survive the agent's commit untouched.
func TestRunAfterAgentStagesOnlyTheNamedFiles(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "a.txt", "agent edit\n")
	writeFile(t, repo, "mine.txt", "user's own work\n")
	t.Setenv(testEnvVar, "1")

	_, err := RunAfterAgent(context.Background(), AgentRun{WorkDir: repo, Files: []string{"a.txt"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"a.txt"}, committedFiles(t, repo))
	assert.Contains(t, gitOutput(t, repo, "status", "--short"), "?? mine.txt")
}
