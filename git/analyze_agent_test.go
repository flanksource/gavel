package git

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AnalyzeCommitHistory used to build its own agent from ai.DefaultConfig() when
// options.AI was set. That is why `git analyze --ai --ai-model X` and every
// `git amend-commits` run silently used the hardcoded model: the field carrying
// the caller's agent was unexported and written in exactly one place, here.
//
// The agent is now the caller's to resolve and own, so asking for AI analysis
// without supplying one is a loud error rather than a silent default.
func TestAnalyzeCommitHistoryRequiresACallerSuppliedAgent(t *testing.T) {
	ctx, err := NewAnalyzerContext(context.Background(), ".")
	require.NoError(t, err)

	_, err = AnalyzeCommitHistory(ctx, []models.Commit{{Hash: "abc123"}}, AnalyzeOptions{AI: true})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no agent was supplied")
}

// The non-AI path must keep working with no agent at all — it is the default for
// `git analyze` and must not be dragged into the new requirement.
func TestAnalyzeCommitHistoryWithoutAINeedsNoAgent(t *testing.T) {
	ctx, err := NewAnalyzerContext(context.Background(), ".")
	require.NoError(t, err)

	_, err = AnalyzeCommitHistory(ctx, nil, AnalyzeOptions{})

	assert.NoError(t, err)
}

// countingAgent records that the caller's agent is the one actually used, rather
// than an agent the git package built for itself behind the caller's back.
type countingAgent struct{ prompts int }

func (a *countingAgent) ExecutePrompt(context.Context, ai.PromptRequest) (*ai.PromptResponse, error) {
	a.prompts++
	return &ai.PromptResponse{Result: `{"type":"fix","subject":"correct the thing"}`}, nil
}

func (a *countingAgent) ExecuteBatch(context.Context, []ai.PromptRequest) (map[string]*ai.PromptResponse, error) {
	return nil, errors.New("unexpected batch execution")
}
func (a *countingAgent) GetCosts() ai.Costs { return nil }
func (a *countingAgent) Close() error       { return nil }

// The whole point of the change: the agent the caller resolved is the agent that
// runs. Previously AnalyzeCommitHistory discarded any caller intent and built its
// own from ai.DefaultConfig().
func TestAnalyzeCommitHistoryUsesTheCallerSuppliedAgent(t *testing.T) {
	ctx, err := NewAnalyzerContext(context.Background(), ".")
	require.NoError(t, err)

	agent := &countingAgent{}
	// A real commit in this repo: AnalyzeCommit resolves each changed file's
	// scope through `git show <hash>:<path>`, so a synthetic hash would send it
	// hunting for blobs that do not exist.
	commit := models.Commit{
		Hash:    "93f113d105e27b6629f652c066ffa6379d11fa8f",
		Subject: "wip",
		Patch: `diff --git a/verify/config.go b/verify/config.go
--- a/verify/config.go
+++ b/verify/config.go
@@ -1,2 +1,3 @@
 package verify
+// added a line
`,
	}

	_, err = AnalyzeCommitHistory(ctx, []models.Commit{commit}, AnalyzeOptions{
		AI:        true,
		Agent:     agent,
		AITimeout: 10 * time.Second,
	})

	require.NoError(t, err)
	assert.Equal(t, 1, agent.prompts, "the caller's agent must receive the analysis prompt")
}
