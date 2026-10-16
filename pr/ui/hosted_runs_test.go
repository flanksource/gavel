// Tests hosted agent runs: a run of a hosted todo is admitted and recorded on serve while its
// prompt runs in a sandbox job, which here runs in this process with a scripted agent.
package ui

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/todos/remote"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// planAgent answers every turn with a new plan, in place of an agent CLI.
type planAgent struct{}

func (planAgent) GetModel() string        { return "claude-sonnet-5" }
func (planAgent) GetRuntime() api.Runtime { return api.RuntimeOf(api.Anthropic, api.ModeAgent) }
func (planAgent) Execute(context.Context, api.Spec) (*api.Response, error) {
	return nil, errors.New("the plan agent only streams")
}
func (planAgent) ExecuteStream(context.Context, api.Spec) (<-chan captainai.Event, error) {
	ch := make(chan captainai.Event, 3)
	ch <- captainai.Event{Kind: captainai.EventSystem, SessionID: uuid.NewString()}
	ch <- captainai.Event{Kind: captainai.EventText, Text: `{"summary":"Planned it.","endStatus":"completed","planStatus":"new","planContent":"# Plan\n\n1. Fix the build."}`}
	ch <- captainai.Event{Kind: captainai.EventResult, Success: true}
	close(ch)
	return ch, nil
}

// inProcessExecutor runs each agent job's sandbox side in this process: the job's checkout of
// the pushed repository, its messages sent to the relay as JSON, as the run API receives them.
type inProcessExecutor struct {
	db      *gorm.DB
	relay   *remote.Relay
	bareDir func(repo uuid.UUID) string
}

type inProcessHandle struct {
	done     chan struct{}
	exitCode int
	err      error
}

func (e *inProcessExecutor) Name() string { return "in-process" }

func (e *inProcessExecutor) Submit(ctx context.Context, job sandbox.Job) (sandbox.Handle, error) {
	handle := &inProcessHandle{done: make(chan struct{})}
	go func() {
		defer close(handle.done)
		handle.err = e.run(context.WithoutCancel(ctx), job.RunID)
		if handle.err != nil {
			handle.exitCode = 1
		}
	}()
	return handle, nil
}

func (e *inProcessExecutor) run(ctx context.Context, id uuid.UUID) error {
	record, err := sandbox.LoadRun(ctx, e.db, id)
	if err != nil {
		return err
	}
	var job remote.Job
	if err := json.Unmarshal(record.Spec.Agent, &job); err != nil {
		return err
	}
	dir := filepath.Join(os.TempDir(), "gavel-hosted-run-"+id.String())
	defer os.RemoveAll(dir)
	if out, err := exec.Command("git", "clone", "-q", e.bareDir(record.Spec.RepoID), dir).CombinedOutput(); err != nil {
		return errors.New(string(out))
	}
	if out, err := exec.Command("git", "-C", dir, "checkout", "-q", "--detach", record.Spec.SHA).CombinedOutput(); err != nil {
		return errors.New(string(out))
	}
	return remote.Execute(ctx, job, dir, planAgent{}, func(message remote.Message) error {
		data, err := json.Marshal(message)
		if err != nil {
			return err
		}
		var received remote.Message
		if err := json.Unmarshal(data, &received); err != nil {
			return err
		}
		return e.relay.Deliver(ctx, id, received)
	})
}

func (h *inProcessHandle) Logs(context.Context) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}
func (h *inProcessHandle) Wait(context.Context) (sandbox.Result, error) {
	<-h.done
	return sandbox.Result{ExitCode: h.exitCode}, nil
}
func (h *inProcessHandle) Cancel(context.Context) error { return nil }
func (h *inProcessHandle) Ref() string                  { return "in-process" }

func TestHostedTodoRunsInASandboxJob(t *testing.T) {
	db, _, org := openOrgDB(t)
	t.Setenv("HOME", t.TempDir())
	creator := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', 'user_ada')`, creator).Error)
	dataDir := t.TempDir()
	repo := pushedRepo(t, db, org, dataDir, "app")
	workspace := "repo:" + repo.String()

	relay := &remote.Relay{}
	logs := blob.Local(t.TempDir())
	dispatcher := &sandbox.Dispatcher{
		DB: db, Image: "runner", Server: "https://gavel.example", TokenKey: []byte(strings.Repeat("k", 32)),
		Logs: func(uuid.UUID) blob.Store { return logs }, Limits: sandbox.Limits{Timeout: time.Minute},
		Executor: &inProcessExecutor{db: db, relay: relay, bareDir: func(id uuid.UUID) string {
			return filepath.Join(dataDir, "repos", org.String(), id.String()+".git")
		}},
	}
	srv := &Server{}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)
	srv.SetHostedWorkspaces(HostedWorkspaces{DB: db, Sources: &sandbox.Sources{DataDir: dataDir}})
	srv.SetAgentRuns(AgentRuns{Dispatcher: dispatcher, Relay: relay})
	base := serveActingAs(t, srv, org)

	status, created := hostedCall(t, base, http.MethodPost, "/api/todos", creator, "org:member", map[string]any{"dir": workspace, "title": "Fix the build"})
	require.Equal(t, http.StatusCreated, status, created)
	todo := created["id"].(string)

	status, started := hostedCall(t, base, http.MethodPost, "/api/todos/run", creator, "org:member", map[string]any{"dir": workspace, "ref": todo, "step": "plan"})
	require.Equal(t, http.StatusOK, status, started)
	assert.Equal(t, "started", started["status"])
	require.NotEmpty(t, started["promptRunId"])

	var run struct {
		Status string
		Kind   string
		UserID uuid.UUID
		SHA    string
	}
	require.Eventually(t, func() bool {
		require.NoError(t, db.Raw(`SELECT status, kind, user_id, sha FROM sandbox_runs WHERE todo_id = ?`, todo).Scan(&run).Error)
		return run.Status == "succeeded"
	}, 30*time.Second, 50*time.Millisecond, "the sandbox job ran: %+v", run)
	assert.Equal(t, "agent", run.Kind)
	assert.Equal(t, creator, run.UserID)
	assert.NotEmpty(t, run.SHA)

	var promptRun struct{ State string }
	require.Eventually(t, func() bool {
		require.NoError(t, db.Raw(`SELECT state FROM captain_prompt_runs WHERE id = ?`, started["promptRunId"]).Scan(&promptRun).Error)
		return promptRun.State == "succeeded"
	}, 30*time.Second, 50*time.Millisecond, "Captain recorded the run's outcome: %+v", promptRun)

	// The outcome event is the run's last write: once it is there, the todo has its outcome.
	var outcomes int64
	require.Eventually(t, func() bool {
		require.NoError(t, db.Raw(`SELECT count(*) FROM todo_issue_events WHERE issue_id = ? AND kind = 'lifecycle_outcome'`, todo).Scan(&outcomes).Error)
		return outcomes == 1
	}, 30*time.Second, 50*time.Millisecond, "the run's outcome reached the todo")
	var outcome struct {
		RunState  string
		EndStatus string
	}
	require.NoError(t, db.Raw(`SELECT payload->'run'->>'state' AS run_state, payload->'envelope'->>'endStatus' AS end_status
		FROM todo_issue_events WHERE issue_id = ? AND kind = 'lifecycle_outcome'`, todo).Scan(&outcome).Error)
	assert.Equal(t, "succeeded", outcome.RunState, "the plan the sandboxed agent wrote was accepted")
	assert.Equal(t, "completed", outcome.EndStatus)
}

func TestHostedTodoRunsNeedASandbox(t *testing.T) {
	db, _, org := openOrgDB(t)
	t.Setenv("HOME", t.TempDir())
	creator := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', 'user_ada')`, creator).Error)
	dataDir := t.TempDir()
	repo := pushedRepo(t, db, org, dataDir, "app")
	base := hostedTodosServer(t, db, org, dataDir)
	workspace := "repo:" + repo.String()
	_, created := hostedCall(t, base, http.MethodPost, "/api/todos", creator, "org:member", map[string]any{"dir": workspace, "title": "Fix the build"})

	status, refused := hostedCall(t, base, http.MethodPost, "/api/todos/run", creator, "org:member", map[string]any{"dir": workspace, "ref": created["id"], "step": "plan"})
	assert.Equal(t, http.StatusNotImplemented, status)
	assert.Equal(t, "agent runs need a sandbox: start gavel serve with --sandbox", refused["error"])
}
