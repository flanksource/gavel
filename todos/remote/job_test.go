// Tests the remote run wire format: a job carries a step's prompt run and hook data to the
// sandbox, and messages carry its progress and result back, all as JSON.
package remote

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	capverify "github.com/flanksource/captain/pkg/ai/agent/verify"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errBoom = errors.New("boom")

func roundTrip[T any](t *testing.T, value T) T {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	var out T
	require.NoError(t, json.Unmarshal(data, &out))
	return out
}

func TestAJobRebuildsThePromptRunInTheSandboxsCheckout(t *testing.T) {
	spec := api.Spec{Model: api.Model{Name: "claude-sonnet-5", Provider: api.Anthropic, Mode: api.ModeAgent}, Prompt: api.Prompt{User: "fix the build"}}
	spec.SetCwd("/srv/gavel/workspaces/org/repo/svc")
	run := lifecycle.RemoteRun{
		Input: promptrun.Input{
			Resolved:          api.ResolvedSpec{Spec: spec},
			Config:            captainai.Config{Model: spec.Model, NoCache: true, SessionID: "sess-1", APIKey: "never-sent"},
			CallerOwnsCommits: true,
			Verify:            capverify.Options{Timeout: 5 * time.Minute},
			Timeout:           30 * time.Minute,
			Repo:              "/srv/gavel/workspaces/org/repo",
		},
		Hooks: lifecycle.RunHooks{Commits: []api.Commit{{On: api.CommitOnTurn}}, Meta: commit.AgentRunMetadata{IssueID: "todo-1"}, Env: map[string]string{"GAVEL_ISSUE_ID": "todo-1"}},
	}

	job := roundTrip(t, JobFor(run))
	assert.Empty(t, job.Config.APIKey, "credentials never travel in a job")

	in := job.Input("/workspace/src")
	assert.Equal(t, "/workspace/src/svc", in.Resolved.Spec.Cwd(), "the run's directory, under the sandbox's checkout")
	assert.Equal(t, "/workspace/src", in.Repo)
	assert.Equal(t, "fix the build", in.Resolved.Spec.Prompt.User)
	assert.Equal(t, "sess-1", in.Config.SessionID)
	assert.True(t, in.Config.NoCache)
	assert.True(t, in.CallerOwnsCommits)
	assert.Equal(t, 5*time.Minute, in.Verify.Timeout)
	assert.Equal(t, 30*time.Minute, in.Timeout)
	assert.Equal(t, run.Hooks, job.Hooks)
	assert.Nil(t, in.Record, "the sandbox records nothing")
}

func TestMessagesCrossAsJSONWithoutBackendNativeValues(t *testing.T) {
	event := api.Event{Kind: api.EventToolUse, Tool: "Bash", Input: map[string]any{"command": "make"}, ToolCallID: "call-1", Raw: struct{ Native bool }{true}}
	got := roundTrip(t, EventMessage(2, event))
	require.NotNil(t, got.Event)
	event.Raw = nil
	assert.Equal(t, event, got.DecodedEvent())
	assert.Equal(t, 2, got.Iteration)

	report := api.NewNodeReport(api.VerifyKindFixture, "fixture", api.VerifyNode{Name: "acceptance", Running: true})
	report.Iteration = 1
	assert.Equal(t, report, *roundTrip(t, ProgressMessage(report)).Progress)

	result := promptrun.Result{
		Response:        &api.Response{Text: "done", StructuredData: map[string]any{"summary": "Built it."}, Raw: "native", Workspace: &api.Workspace{Cwd: "/workspace/src"}},
		Verdicts:        []agent.VerifyResult{{Valid: true, Report: &report, Iteration: 1}},
		Report:          &report,
		Passed:          true,
		SessionID:       "provider-session",
		Model:           "claude-sonnet-5",
		Usage:           api.Usage{InputTokens: 7, OutputTokens: 3},
		CostUSD:         0.25,
		Duration:        90 * time.Second,
		StructuredData:  map[string]any{"summary": "Built it."},
		TerminalOutcome: &api.TerminalOutcome{Kind: api.TerminalOutcomeQuestions},
		Loop: &captainai.LoopResult{StopReason: "condition-met", TotalCost: 0.25, Iterations: []*captainai.LoopIteration{
			{Iteration: 1, SessionID: "provider-session", CostUSD: 0.25, Success: true, Events: []api.Event{{Kind: api.EventText, Text: "done", Raw: "native"}}},
			{Iteration: 2, Err: errBoom},
		}},
	}
	got = roundTrip(t, ResultMessage(result, nil))
	decoded := got.Result.Unwrap()
	require.NotNil(t, decoded.Loop)
	require.Len(t, decoded.Loop.Iterations, 2)
	assert.EqualError(t, decoded.Loop.Iterations[1].Err, "boom", "an iteration's error crosses as its message")
	result.Response.Raw = nil
	result.Loop.Iterations[0].Events[0].Raw = nil
	result.Loop.Iterations[1].Err = decoded.Loop.Iterations[1].Err
	assert.Equal(t, result, decoded)
	assert.Empty(t, got.Error)

	failed := roundTrip(t, ResultMessage(promptrun.Result{}, errBoom))
	assert.Equal(t, "boom", failed.Error)
}
