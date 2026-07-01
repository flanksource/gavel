package commit

import (
	"context"
	"fmt"
	"strings"
	"testing"

	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsTokenLimitError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"anthropic prompt too long", fmt.Errorf("execute AI commit message prompt: prompt is too long: 250000 tokens > 200000 maximum"), true},
		{"openai maximum context length", fmt.Errorf("this model's maximum context length is 128000 tokens"), true},
		{"generic context_length_exceeded", fmt.Errorf("api error: context_length_exceeded"), true},
		{"too many tokens", fmt.Errorf("request has too many tokens for this model"), true},
		{"unrelated network error", fmt.Errorf("network connection refused"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isTokenLimitError(tc.err))
		})
	}
}

func TestRunSingleCommitChunksByDirectoryOnTokenLimit(t *testing.T) {
	repo := initCommitRepo(t)
	writeFileInDir(t, repo, "alpha/a.txt", "one\n")
	writeFileInDir(t, repo, "beta/b.txt", "two\n")
	gitRun(t, repo, "add", "alpha/a.txt", "beta/b.txt")

	prevAgent := newAgentFunc
	newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) { return nil, nil }
	defer func() { newAgentFunc = prevAgent }()

	// The whole-tree diff spans both directories and "overflows" the context;
	// each single-directory chunk succeeds. This forces the runSingleCommit
	// fallback to split the commit by directory.
	prevMsg := analyzeCommitMessageWithAIFunc
	analyzeCommitMessageWithAIFunc = func(ctx context.Context, commit models.CommitAnalysis, agent clickyai.Agent, opts git.AnalyzeOptions) (models.CommitAnalysis, error) {
		if strings.Contains(commit.Patch, "alpha/") && strings.Contains(commit.Patch, "beta/") {
			return models.CommitAnalysis{}, fmt.Errorf("execute AI commit message prompt: prompt is too long: 250000 tokens > 200000 maximum")
		}
		commit.CommitType = "chore"
		if strings.Contains(commit.Patch, "alpha/") {
			commit.Subject = "update alpha"
		} else {
			commit.Subject = "update beta"
		}
		return commit, nil
	}
	defer func() { analyzeCommitMessageWithAIFunc = prevMsg }()

	result, err := Run(context.Background(), Options{WorkDir: repo, PrecommitMode: CheckModeSkip})
	require.NoError(t, err)
	require.Len(t, result.Commits, 2)
	assert.Equal(t, []string{"alpha/a.txt"}, result.Commits[0].Files)
	assert.Equal(t, []string{"beta/b.txt"}, result.Commits[1].Files)
	assert.Equal(t, "chore: update alpha", result.Commits[0].Message)
	assert.Equal(t, "chore: update beta", result.Commits[1].Message)
	assert.NotEmpty(t, result.Commits[0].Hash)
	assert.NotEmpty(t, result.Commits[1].Hash)
	assert.Equal(t, "3", strings.TrimSpace(gitOutput(t, repo, "rev-list", "--count", "HEAD")))
	assert.Empty(t, strings.TrimSpace(gitOutput(t, repo, "status", "--short")))
}
