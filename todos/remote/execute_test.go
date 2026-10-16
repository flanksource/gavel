// Tests Execute, the sandbox side of a remote run: it runs the job's prompt run in its checkout
// with gavel's hooks rebuilt, and reports the start, every event and the result in order.
package remote

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// scriptedAgent streams one scripted turn in place of an agent CLI.
type scriptedAgent struct {
	mu       sync.Mutex
	events   []captainai.Event
	requests []api.Spec
}

func (a *scriptedAgent) GetModel() string        { return "claude-sonnet-5" }
func (a *scriptedAgent) GetRuntime() api.Runtime { return api.RuntimeOf(api.Anthropic, api.ModeAgent) }
func (a *scriptedAgent) Execute(context.Context, api.Spec) (*api.Response, error) {
	return nil, errors.New("the scripted agent only streams")
}
func (a *scriptedAgent) ExecuteStream(_ context.Context, req api.Spec) (<-chan captainai.Event, error) {
	a.mu.Lock()
	a.requests = append(a.requests, req)
	a.mu.Unlock()
	ch := make(chan captainai.Event, len(a.events))
	for _, ev := range a.events {
		ch <- ev
	}
	close(ch)
	return ch, nil
}

func scriptedJob(root string) Job {
	spec := api.Spec{Model: api.Model{Name: "claude-sonnet-5", Provider: api.Anthropic, Mode: api.ModeAgent}, Prompt: api.Prompt{User: "fix the build"}}
	spec.SetCwd(root)
	return Job{
		Resolved: api.ResolvedSpec{Spec: spec},
		Config:   JobConfig{Model: spec.Model},
		Hooks: lifecycle.RunHooks{
			Meta: commit.AgentRunMetadata{IssueID: "todo-1"},
			Env:  map[string]string{commit.EnvIssueID: "todo-1"},
		},
		Timeout: time.Minute,
		Root:    root,
	}
}

func TestExecuteReportsTheRunInOrder(t *testing.T) {
	dir := t.TempDir()
	agent := &scriptedAgent{events: []captainai.Event{
		{Kind: captainai.EventSystem, SessionID: "provider-session"},
		{Kind: captainai.EventText, Text: "done"},
		{Kind: captainai.EventResult, Success: true, CostUSD: 0.25, Usage: &api.Usage{InputTokens: 7, OutputTokens: 3}},
	}}
	var sent []Message
	finished, err := Execute(context.Background(), scriptedJob("/srv/gavel/workspaces/org/repo"), dir, agent, Serve{Send: func(message Message) error {
		sent = append(sent, message)
		return nil
	}})
	require.NoError(t, err)

	require.NotEmpty(t, sent)
	assert.Equal(t, KindStarted, sent[0].Kind)
	require.NotNil(t, sent[0].Spec)
	assert.Equal(t, dir, sent[0].Spec.Cwd(), "the run works in the sandbox's checkout")
	assert.Contains(t, sent[0].Spec.Setup.Env, commit.EnvIssueID+"=todo-1", "gavel's hooks ran before the start")

	require.NoError(t, finished.Err)
	assert.Equal(t, "done", finished.Result.Response.Text, "the result is the caller's to report")
	assert.Equal(t, 0.25, finished.Result.CostUSD)

	var kinds []captainai.EventKind
	for _, message := range sent[1:] {
		require.Equal(t, KindEvent, message.Kind)
		kinds = append(kinds, message.DecodedEvent().Kind)
	}
	assert.Contains(t, kinds, captainai.EventSystem)
	assert.Contains(t, kinds, captainai.EventText)
	require.Len(t, agent.requests, 1)
	assert.Equal(t, "fix the build", agent.requests[0].Prompt.User)
}

func TestExecuteStopsWhenServeStopsListening(t *testing.T) {
	agent := &scriptedAgent{events: []captainai.Event{{Kind: captainai.EventText, Text: "done"}, {Kind: captainai.EventResult, Success: true}}}
	_, err := Execute(context.Background(), scriptedJob("/srv/root"), t.TempDir(), agent, Serve{Send: func(Message) error {
		return errors.New("connection reset")
	}})
	require.EqualError(t, err, "report the run's start to serve: connection reset")
}

func TestAJobsToolsAndApprovalsCallServe(t *testing.T) {
	job := scriptedJob("/srv/repo")
	job.Tools = []api.ToolDefinition{{Name: "todo_search", Description: "Search the backlog.", DefaultPermission: api.ToolPolicyAllow}}
	job.Approvals = true
	var calls []Message
	serve := Serve{
		Send: func(Message) error { return nil },
		Call: func(message Message) (Reply, error) {
			calls = append(calls, message)
			if message.Kind == KindTool {
				return Reply{Output: json.RawMessage(`{"found":"bug"}`)}, nil
			}
			return Reply{Decision: &api.ApprovalDecision{Allow: true}}, nil
		},
	}
	in := job.Input(t.TempDir())
	callServe(&in, job, serve)

	require.Len(t, in.Config.Tools, 1)
	assert.Equal(t, "Search the backlog.", in.Config.Tools[0].Description)
	output, err := in.Config.Tools[0].Handler(context.Background(), map[string]any{"query": "bug"})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"found": "bug"}, output)

	require.NotNil(t, in.Config.OnApproval)
	decision, err := in.Config.OnApproval(context.Background(), api.ApprovalRequest{Tool: "triage_review", Kind: api.ApprovalKindTool})
	require.NoError(t, err)
	assert.True(t, decision.Allow)

	require.Len(t, calls, 2)
	assert.Equal(t, ToolCallMessage("todo_search", map[string]any{"query": "bug"}), calls[0])
	assert.Equal(t, KindApproval, calls[1].Kind)

	refused := Serve{Call: func(Message) (Reply, error) { return Reply{Error: "the run has no tool \"todo_search\""}, nil }}
	in = job.Input(t.TempDir())
	callServe(&in, job, refused)
	_, err = in.Config.Tools[0].Handler(context.Background(), nil)
	require.EqualError(t, err, `tool todo_search: the run has no tool "todo_search"`)
}
