package status

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	captaincli "github.com/flanksource/captain/pkg/cli"
	clickyai "github.com/flanksource/gavel/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatherWithAIFileSummaries(t *testing.T) {
	repo := initStatusRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(repo, "staged.go"), []byte("package x\n\nfunc Added() {}\n"), 0o644))
	gitRun(t, repo, "add", "staged.go")

	prev := summarizeFileChangeWithAIFunc
	summarizeFileChangeWithAIFunc = func(_ context.Context, _ SummaryOptions, file FileStatus) (string, error) {
		return "add helper function", nil
	}
	t.Cleanup(func() {
		summarizeFileChangeWithAIFunc = prev
	})

	result, err := Gather(repo, Options{
		NoRepomap: true,
		Summary:   &SummaryOptions{},
		Context:   context.Background(),
	})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)
	assert.Equal(t, "add helper function", result.Files[0].AISummary)
}

func TestSummarizeFileChangeWithAISendsDiffOnlyPrompt(t *testing.T) {
	restoreIO := stubAISummaryIO(
		func(string, string, bool) (string, error) {
			return "diff --git a/a.go b/a.go\n+added helper\n", nil
		},
		func(string, string) (string, error) {
			t.Fatal("readUntrackedStatusFileFunc should not be called for tracked files")
			return "", nil
		},
	)
	defer restoreIO()

	agent := &capturePromptAgent{}
	prompt, err := ResolveSummaryPrompt(SummaryPromptOptions{Dir: t.TempDir(), Saved: captainconfig.Config{AI: captainconfig.AIDefaults{DefaultModel: "agent:claude-sonnet-5"}}})
	require.NoError(t, err)
	var projected captaincli.AIRuntimeResolved
	summary, err := summarizeFileChangeWithAI(context.Background(), SummaryOptions{
		Prompt: prompt, NewAgent: func(runtime captaincli.AIRuntimeResolved) (clickyai.Agent, error) {
			projected = runtime
			return agent, nil
		},
	}, FileStatus{
		Path:  "a.go",
		State: StateStaged,
		Adds:  3,
		Dels:  1,
	})
	require.NoError(t, err)
	assert.Equal(t, "tighten handler flow", summary)
	assert.Contains(t, agent.prompt, "Staged diff:")
	assert.NotContains(t, agent.prompt, "File metadata:")
	assert.NotContains(t, agent.prompt, "path:")
	assert.NotContains(t, agent.prompt, "state:")
	assert.NotContains(t, agent.prompt, "adds:")
	assert.NotContains(t, agent.prompt, "dels:")
	assert.Equal(t, projected.Request, agent.spec)
	assert.Equal(t, "status summary: a.go", agent.name)
	assert.True(t, agent.closed)
}

func TestStreamAISummariesPreservesFileOrdering(t *testing.T) {
	prev := summarizeFileChangeWithAIFunc
	summarizeFileChangeWithAIFunc = func(_ context.Context, _ SummaryOptions, file FileStatus) (string, error) {
		switch file.Path {
		case "slow.go":
			time.Sleep(40 * time.Millisecond)
			return "slow summary", nil
		case "fast.go":
			return "fast summary", nil
		default:
			return "", nil
		}
	}
	t.Cleanup(func() {
		summarizeFileChangeWithAIFunc = prev
	})

	result := &Result{
		Files: []FileStatus{
			{Path: "slow.go", State: StateStaged, StagedKind: KindModified},
			{Path: "fast.go", State: StateStaged, StagedKind: KindModified},
		},
	}
	result.PrepareAISummaries()

	var running int
	for update := range StreamAISummaries(context.Background(), SummaryOptions{Files: result.Files, MaxWorkers: 2}) {
		if update.Status == AISummaryStatusRunning {
			running++
		}
		result.ApplyAISummaryUpdate(update)
	}

	assert.Equal(t, 2, running)
	assert.Equal(t, "slow summary", result.Files[0].AISummary)
	assert.Equal(t, "fast summary", result.Files[1].AISummary)
	assert.Equal(t, AISummaryStatusDone, result.Files[0].AIStatus)
	assert.Equal(t, AISummaryStatusDone, result.Files[1].AIStatus)
}

type capturePromptAgent struct {
	prompt string
	spec   api.Spec
	name   string
	closed bool
}

func (a *capturePromptAgent) GetConfig() clickyai.AgentConfig {
	return clickyai.AgentConfig{}
}
func (a *capturePromptAgent) ExecutePrompt(_ context.Context, req clickyai.PromptRequest) (*clickyai.PromptResponse, error) {
	a.prompt = req.Spec.Prompt.User
	a.spec, a.name = req.Spec, req.Name
	return &clickyai.PromptResponse{StructuredData: json.RawMessage(`{"summary":"tighten handler flow"}`)}, nil
}
func (a *capturePromptAgent) ExecuteBatch(context.Context, []clickyai.PromptRequest) (map[string]*clickyai.PromptResponse, error) {
	return nil, nil
}
func (a *capturePromptAgent) GetCosts() clickyai.Costs { return clickyai.Costs{} }
func (a *capturePromptAgent) Close() error             { a.closed = true; return nil }

func stubAISummaryIO(
	diff func(workDir, path string, cached bool) (string, error),
	read func(workDir, path string) (string, error),
) func() {
	prevDiff := diffForStatusFileFunc
	prevRead := readUntrackedStatusFileFunc
	diffForStatusFileFunc = diff
	readUntrackedStatusFileFunc = read
	return func() {
		diffForStatusFileFunc = prevDiff
		readUntrackedStatusFileFunc = prevRead
	}
}
