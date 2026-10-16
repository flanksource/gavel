// Tests Runner, serve's side of a remote run: it admits and records the run in Captain's store
// from what the sandbox job reports, and returns the result as promptrun.Run would.
package remote

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/flanksource/captain/pkg/ai/approval"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// inProcessJobs runs each job's sandbox side in this process, sending to the relay as a sandbox
// job sends to serve's run API.
type inProcessJobs struct {
	relay *Relay
	agent *scriptedAgent
	// skipResult ends the job before it reports a result, as a job that crashed does.
	skipResult bool
	dir        string
	// admittedBeforeStart is whether the run was admitted when its job started.
	admitted            *bool
	admittedBeforeStart bool
}

type inProcessJob struct {
	done     chan struct{}
	exitCode int
	cancel   context.CancelFunc
}

func (j *inProcessJob) Wait(ctx context.Context) (int, error) {
	select {
	case <-j.done:
		return j.exitCode, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func (j *inProcessJob) Cancel(context.Context) error {
	j.cancel()
	return nil
}

func (s *inProcessJobs) Start(ctx context.Context, id uuid.UUID, job Job) (Started, error) {
	if s.admitted != nil {
		s.admittedBeforeStart = *s.admitted
	}
	runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	started := &inProcessJob{done: make(chan struct{}), cancel: cancel}
	go func() {
		defer close(started.done)
		send := func(message Message) error {
			// Messages reach serve as JSON, so nothing they carry is shared with the job.
			data, err := json.Marshal(message)
			if err != nil {
				return err
			}
			var received Message
			if err := json.Unmarshal(data, &received); err != nil {
				return err
			}
			return s.relay.Deliver(runCtx, id, received)
		}
		finished, err := Execute(runCtx, job, s.dir, s.agent, send)
		if err == nil && !s.skipResult {
			err = send(ResultMessage(finished.Result, finished.Err))
		}
		if err != nil || s.skipResult {
			started.exitCode = 1
		}
	}()
	return started, nil
}

func openRecording(t *testing.T) *promptrun.Recording {
	t.Helper()
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_remote_runner"}).DSN())
	t.Setenv(database.EnvDisable, "")
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	captain, err := captaindb.Use(opened.Gorm())
	require.NoError(t, err)
	root := uuid.New()
	return &promptrun.Recording{
		DB: captain,
		Placement: promptrun.Placement{Sessions: []captaindb.CreateSessionInput{
			{ID: root, Source: "gavel", Provider: "todos", HostID: "remote-test", CWD: "/srv/repo", Title: "Fix the build"},
			{ID: uuid.New(), Source: "gavel", Provider: "agent-claude", HostID: "remote-test", ParentSessionID: &root, CWD: "/srv/repo", Title: "Fix the build · run"},
		}},
		Origin: "gavel.todos", SpecProfile: "run", AdmissionKey: "admission-" + uuid.NewString(), PromptMarkdown: "fix the build",
		Runtime: captaindb.PromptRunRuntime{Mode: "run", Driver: "agent-claude"},
	}
}

func remoteRun(record *promptrun.Recording) (lifecycle.RemoteRun, *[]api.Event, *bool) {
	spec := api.Spec{Model: api.Model{Name: "claude-sonnet-5", Provider: api.Anthropic, Mode: api.ModeAgent}, Prompt: api.Prompt{User: "fix the build"}}
	spec.SetCwd("/srv/repo")
	var mu sync.Mutex
	var events []api.Event
	admitted := false
	return lifecycle.RemoteRun{
		Input: promptrun.Input{
			Resolved: api.ResolvedSpec{Spec: spec}, Timeout: time.Minute, Repo: "/srv/repo", Record: record,
			OnEvent: func(_ int, event api.Event) {
				mu.Lock()
				defer mu.Unlock()
				events = append(events, event)
			},
		},
		Admitted: func() error {
			admitted = true
			return nil
		},
	}, &events, &admitted
}

func TestRunnerRecordsTheRunFromTheSandbox(t *testing.T) {
	record := openRecording(t)
	relay := &Relay{}
	jobs := &inProcessJobs{relay: relay, dir: t.TempDir(), agent: &scriptedAgent{events: []api.Event{
		{Kind: api.EventSystem, SessionID: uuid.NewString()},
		{Kind: api.EventText, Text: "done"},
		{Kind: api.EventResult, Success: true, CostUSD: 0.25},
	}}}
	run, events, admitted := remoteRun(record)
	jobs.admitted = admitted

	result, err := (&Runner{Jobs: jobs, Relay: relay}).Run(context.Background(), run)

	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, result.PromptRunID)
	assert.Equal(t, "done", result.Response.Text)
	assert.True(t, jobs.admittedBeforeStart, "the host learns the run was admitted before its job starts, as a local run's does before setup")
	assert.NotEmpty(t, *events, "the host saw the run's events")
	stored, err := record.DB.GetPromptRun(context.Background(), result.PromptRunID)
	require.NoError(t, err)
	assert.Equal(t, captaindb.PromptRunStateSucceeded, stored.State)
	assert.NotNil(t, stored.StartedAt)
	assert.NotNil(t, stored.ExecutionSessionID, "the session the sandbox's agent reported is bound")
}

func TestRunnerFailsARunWhoseJobExitsWithoutAResult(t *testing.T) {
	record := openRecording(t)
	relay := &Relay{}
	jobs := &inProcessJobs{relay: relay, dir: t.TempDir(), skipResult: true, agent: &scriptedAgent{events: []api.Event{{Kind: api.EventResult, Success: true}}}}
	run, _, _ := remoteRun(record)

	result, err := (&Runner{Jobs: jobs, Relay: relay}).Run(context.Background(), run)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "exited with code 1 before it reported a result")
	stored, getErr := record.DB.GetPromptRun(context.Background(), result.PromptRunID)
	require.NoError(t, getErr)
	assert.Equal(t, captaindb.PromptRunStateFailed, stored.State)
}

func TestRelayRefusesARunNobodyIsWaitingFor(t *testing.T) {
	err := (&Relay{}).Deliver(context.Background(), uuid.New(), Message{Kind: KindStarted})
	require.True(t, errors.Is(err, ErrNoListener))
}

// callingJobs is a job that, once started, calls a tool and asks for an approval through the
// relay, as a job's agent does mid-run, then reports a result.
type callingJobs struct {
	relay    *Relay
	toolCall Reply
	approval Reply
}

func (c *callingJobs) Start(ctx context.Context, id uuid.UUID, _ Job) (Started, error) {
	runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	started := &inProcessJob{done: make(chan struct{}), cancel: cancel}
	go func() {
		defer close(started.done)
		_ = c.relay.Deliver(runCtx, id, StartedMessage(api.Spec{}))
		var err error
		if c.toolCall, err = c.relay.Call(runCtx, id, ToolCallMessage("todo_search", map[string]any{"query": "bug"})); err != nil {
			c.toolCall.Error = err.Error()
		}
		if c.approval, err = c.relay.Call(runCtx, id, ApprovalMessage(api.ApprovalRequest{Tool: "triage_review", Kind: api.ApprovalKindTool, Input: map[string]any{"proposal": "p-1"}, ToolUseID: "call-1"})); err != nil {
			c.approval.Error = err.Error()
		}
		_ = c.relay.Deliver(runCtx, id, ResultMessage(promptrun.Result{Response: &api.Response{Text: "triaged"}}, nil))
	}()
	return started, nil
}

func TestRunnerAnswersAJobsToolCallsAndApprovals(t *testing.T) {
	record := openRecording(t)
	relay := &Relay{}
	jobs := &callingJobs{relay: relay}
	run, _, _ := remoteRun(record)
	run.Input.Config.Tools = []api.ToolDefinition{{Name: "todo_search", Handler: func(_ context.Context, input map[string]any) (any, error) {
		return map[string]any{"found": input["query"]}, nil
	}}}
	run.Input.Approvals = &promptrun.ApprovalOptions{RequestedBy: "gavel-dashboard"}
	onEvent := run.Input.OnEvent
	run.Input.OnEvent = func(iteration int, event api.Event) {
		onEvent(iteration, event)
		if event.Kind == api.EventPermission && event.ApprovalID != "" {
			// A person approves it on the dashboard.
			go func() {
				_, err := approval.Resolve(context.Background(), record.DB, approval.ResolveInput{
					RequestID: uuid.MustParse(event.ApprovalID), Approved: true, ResolvedBy: "ada",
				})
				assert.NoError(t, err)
			}()
		}
	}

	_, err := (&Runner{Jobs: jobs, Relay: relay}).Run(context.Background(), run)
	require.NoError(t, err)

	assert.Empty(t, jobs.toolCall.Error)
	assert.JSONEq(t, `{"found":"bug"}`, string(jobs.toolCall.Output), "the tool ran on serve")
	assert.Empty(t, jobs.approval.Error)
	require.NotNil(t, jobs.approval.Decision)
	assert.True(t, jobs.approval.Decision.Allow, "the person's approval reached the job")
}
