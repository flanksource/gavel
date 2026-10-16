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
	"sync"
	"testing"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/approval"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/todos/remote"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
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
	// agent answers the run's turns; nil is planAgent.
	agent    captainai.Provider
	mu       sync.Mutex
	canceled int
}

type inProcessHandle struct {
	executor *inProcessExecutor
	done     chan struct{}
	cancel   context.CancelFunc
	exitCode int
	err      error
}

func (e *inProcessExecutor) Name() string { return "in-process" }

func (e *inProcessExecutor) Submit(ctx context.Context, job sandbox.Job) (sandbox.Handle, error) {
	runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	handle := &inProcessHandle{executor: e, done: make(chan struct{}), cancel: cancel}
	go func() {
		defer close(handle.done)
		handle.err = e.run(runCtx, job.RunID)
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
	send := func(message remote.Message) error {
		data, err := json.Marshal(message)
		if err != nil {
			return err
		}
		var received remote.Message
		if err := json.Unmarshal(data, &received); err != nil {
			return err
		}
		return e.relay.Deliver(ctx, id, received)
	}
	var agent captainai.Provider = planAgent{}
	if e.agent != nil {
		agent = e.agent
	}
	call := func(message remote.Message) (remote.Reply, error) { return e.relay.Call(ctx, id, message) }
	finished, err := remote.Execute(ctx, job, dir, agent, remote.Serve{Send: send, Call: call})
	if err != nil {
		return err
	}
	return send(remote.ResultMessage(finished.Result, finished.Err))
}

func (h *inProcessHandle) Logs(context.Context) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}
func (h *inProcessHandle) Wait(context.Context) (sandbox.Result, error) {
	<-h.done
	return sandbox.Result{ExitCode: h.exitCode}, nil
}
func (h *inProcessHandle) Cancel(context.Context) error {
	h.executor.mu.Lock()
	h.executor.canceled++
	h.executor.mu.Unlock()
	h.cancel()
	return nil
}
func (h *inProcessHandle) Ref() string { return "in-process" }

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

func TestEveryHostedRunGoesToASandbox(t *testing.T) {
	hosted := todoSource{Dir: "/srv/gavel/workspaces/org/repo", Hosted: &HostedWorkspace{Repo: sandbox.Repo{ID: uuid.New()}, SHA: "abc123"}}
	todo := &types.TODO{ID: uuid.NewString()}
	override, err := continuationOverride(nil)
	require.NoError(t, err)
	continuation := run.Continuation{Dir: hosted.Dir, Todo: todo, Step: "run", Override: override}

	withSandbox := &Server{}
	withSandbox.SetAgentRuns(AgentRuns{Dispatcher: &sandbox.Dispatcher{}, Relay: &remote.Relay{}})
	req, err := withSandbox.continuationRequest(context.Background(), hosted, continuation)
	require.NoError(t, err)
	assert.NotNil(t, req.Remote, "a hosted continuation runs in a sandbox")
	require.NotNil(t, req.PRs, "and reads linked PRs with the organization's GitHub App")
	_, _, err = req.PRs(context.Background(), "https://dev.azure.com/acme/app/_git/app/pullrequest/7")
	require.EqualError(t, err, "hosted todos work on GitHub PRs; https://dev.azure.com/acme/app/_git/app/pullrequest/7 isn't one")
	assert.True(t, req.Approvals, "a sandboxed run asks serve for approvals")

	_, err = (&Server{}).continuationRequest(context.Background(), hosted, continuation)
	require.ErrorIs(t, err, errAgentRunsNeedSandbox)

	req, err = (&Server{}).continuationRequest(context.Background(), todoSource{Dir: t.TempDir()}, continuation)
	require.NoError(t, err)
	assert.Nil(t, req.Remote, "a local continuation runs here")
	assert.True(t, req.Approvals)
}

// workingAgent works until its run is stopped.
type workingAgent struct{}

func (workingAgent) GetModel() string        { return "claude-sonnet-5" }
func (workingAgent) GetRuntime() api.Runtime { return api.RuntimeOf(api.Anthropic, api.ModeAgent) }
func (workingAgent) Execute(context.Context, api.Spec) (*api.Response, error) {
	return nil, errors.New("the working agent only streams")
}
func (workingAgent) ExecuteStream(ctx context.Context, _ api.Spec) (<-chan captainai.Event, error) {
	ch := make(chan captainai.Event, 1)
	ch <- captainai.Event{Kind: captainai.EventText, Text: "working"}
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

func TestHostedTodoRunsBelongToTheTodosCreator(t *testing.T) {
	db, _, org := openOrgDB(t)
	t.Setenv("HOME", t.TempDir())
	creator, member := uuid.New(), uuid.New()
	for name, id := range map[string]uuid.UUID{"ada": creator, "grace": member} {
		require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, "user_"+name).Error)
	}
	dataDir := t.TempDir()
	repo := pushedRepo(t, db, org, dataDir, "app")
	workspace := "repo:" + repo.String()
	relay := &remote.Relay{}
	logs := blob.Local(t.TempDir())
	executor := &inProcessExecutor{db: db, relay: relay, agent: workingAgent{}, bareDir: func(id uuid.UUID) string {
		return filepath.Join(dataDir, "repos", org.String(), id.String()+".git")
	}}
	dispatcher := &sandbox.Dispatcher{
		DB: db, Image: "runner", Server: "https://gavel.example", TokenKey: []byte(strings.Repeat("k", 32)),
		Logs: func(uuid.UUID) blob.Store { return logs }, Limits: sandbox.Limits{Timeout: time.Minute}, Executor: executor,
	}
	srv := &Server{}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)
	srv.SetHostedWorkspaces(HostedWorkspaces{DB: db, Sources: &sandbox.Sources{DataDir: dataDir}})
	srv.SetAgentRuns(AgentRuns{Dispatcher: dispatcher, Relay: relay})
	base := serveActingAs(t, srv, org)

	_, created := hostedCall(t, base, http.MethodPost, "/api/todos", creator, "org:member", map[string]any{"dir": workspace, "title": "Shared", "visibility": "public"})
	todo := created["id"].(string)

	status, refused := hostedCall(t, base, http.MethodPost, "/api/todos/run", member, "org:member", map[string]any{"dir": workspace, "ref": todo, "step": "plan"})
	assert.Equal(t, http.StatusForbidden, status, "another member can't run the creator's todo: %v", refused)

	status, started := hostedCall(t, base, http.MethodPost, "/api/todos/run", creator, "org:member", map[string]any{"dir": workspace, "ref": todo, "step": "plan"})
	require.Equal(t, http.StatusOK, status, started)
	promptRun := started["promptRunId"].(string)

	stop := map[string]any{"ref": todo, "promptRunId": promptRun}
	status, refused = hostedCall(t, base, http.MethodPost, "/api/todos/session/stop?dir="+workspace, member, "org:member", stop)
	assert.Equal(t, http.StatusForbidden, status, "another member can't stop it: %v", refused)

	status, stopped := hostedCall(t, base, http.MethodPost, "/api/todos/session/stop?dir="+workspace, creator, "org:member", stop)
	require.Equal(t, http.StatusAccepted, status, stopped)
	var state struct{ State string }
	require.Eventually(t, func() bool {
		require.NoError(t, db.Raw(`SELECT state FROM captain_prompt_runs WHERE id = ?`, promptRun).Scan(&state).Error)
		return state.State == "cancelled"
	}, 30*time.Second, 50*time.Millisecond, "the stopped run is recorded as cancelled: %+v", state)
	executor.mu.Lock()
	assert.Equal(t, 1, executor.canceled, "stopping the run cancels its sandbox job")
	executor.mu.Unlock()
	var outcomes int64
	require.Eventually(t, func() bool {
		require.NoError(t, db.Raw(`SELECT count(*) FROM todo_issue_events WHERE issue_id = ? AND kind = 'lifecycle_outcome'`, todo).Scan(&outcomes).Error)
		return outcomes == 1
	}, 30*time.Second, 50*time.Millisecond, "the stopped run's outcome reached the todo")
}

func TestHostedTriageNewRunsWithItsToolsOnServe(t *testing.T) {
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
		Executor: &inProcessExecutor{db: db, relay: relay, agent: workingAgent{}, bareDir: func(id uuid.UUID) string {
			return filepath.Join(dataDir, "repos", org.String(), id.String()+".git")
		}},
	}
	srv := &Server{}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)
	srv.SetHostedWorkspaces(HostedWorkspaces{DB: db, Sources: &sandbox.Sources{DataDir: dataDir}})
	srv.SetAgentRuns(AgentRuns{Dispatcher: dispatcher, Relay: relay})
	base := serveActingAs(t, srv, org)
	_, created := hostedCall(t, base, http.MethodPost, "/api/todos", creator, "org:member", map[string]any{"dir": workspace, "title": "Fix the build"})
	todo := created["id"].(string)

	status, started := hostedCall(t, base, http.MethodPost, "/api/todos/run", creator, "org:member", map[string]any{"dir": workspace, "ref": todo, "step": "triage.new"})
	require.Equal(t, http.StatusOK, status, started)

	var spec string
	require.Eventually(t, func() bool {
		require.NoError(t, db.Raw(`SELECT spec->>'agent' FROM sandbox_runs WHERE todo_id = ?`, todo).Scan(&spec).Error)
		return spec != ""
	}, 30*time.Second, 50*time.Millisecond)
	var job remote.Job
	require.NoError(t, json.Unmarshal([]byte(spec), &job))
	var tools []string
	for _, tool := range job.Tools {
		tools = append(tools, tool.Name)
	}
	assert.ElementsMatch(t, []string{"todo_search", "todo_get", "triage_prepare", "triage_review"}, tools, "the job calls serve for triage.new's tools")
	assert.True(t, job.Approvals, "and asks serve for the person's approval of the proposal")

	stop := map[string]any{"ref": todo, "promptRunId": started["promptRunId"]}
	status, stopped := hostedCall(t, base, http.MethodPost, "/api/todos/session/stop?dir="+workspace, creator, "org:member", stop)
	require.Equal(t, http.StatusAccepted, status, stopped)
	var outcomes int64
	require.Eventually(t, func() bool {
		require.NoError(t, db.Raw(`SELECT count(*) FROM todo_issue_events WHERE issue_id = ? AND kind = 'lifecycle_outcome'`, todo).Scan(&outcomes).Error)
		return outcomes == 1
	}, 30*time.Second, 50*time.Millisecond, "the stopped run's outcome reached the todo")
}

func TestHostedApprovalsBelongToTheTodosOwner(t *testing.T) {
	db, _, org := openOrgDB(t)
	t.Setenv("HOME", t.TempDir())
	creator, member := uuid.New(), uuid.New()
	for name, id := range map[string]uuid.UUID{"ada": creator, "grace": member} {
		require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, "user_"+name).Error)
	}
	dataDir := t.TempDir()
	repo := pushedRepo(t, db, org, dataDir, "app")
	workspace := "repo:" + repo.String()
	relay := &remote.Relay{}
	logs := blob.Local(t.TempDir())
	dispatcher := &sandbox.Dispatcher{
		DB: db, Image: "runner", Server: "https://gavel.example", TokenKey: []byte(strings.Repeat("k", 32)),
		Logs: func(uuid.UUID) blob.Store { return logs }, Limits: sandbox.Limits{Timeout: time.Minute},
		Executor: &inProcessExecutor{db: db, relay: relay, agent: workingAgent{}, bareDir: func(id uuid.UUID) string {
			return filepath.Join(dataDir, "repos", org.String(), id.String()+".git")
		}},
	}
	srv := &Server{}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)
	srv.SetHostedWorkspaces(HostedWorkspaces{DB: db, Sources: &sandbox.Sources{DataDir: dataDir}})
	srv.SetAgentRuns(AgentRuns{Dispatcher: dispatcher, Relay: relay})
	base := serveActingAs(t, srv, org)
	_, created := hostedCall(t, base, http.MethodPost, "/api/todos", creator, "org:member", map[string]any{"dir": workspace, "title": "Shared", "visibility": "public"})
	todo := created["id"].(string)
	status, started := hostedCall(t, base, http.MethodPost, "/api/todos/run", creator, "org:member", map[string]any{"dir": workspace, "ref": todo, "step": "plan"})
	require.Equal(t, http.StatusOK, status, started)
	promptRun := uuid.MustParse(started["promptRunId"].(string))

	// The running agent asks for an approval, as a sandboxed run's does through serve's broker.
	captain, err := captaindb.Use(db)
	require.NoError(t, err)
	var stored *captaindb.PromptRun
	require.Eventually(t, func() bool {
		stored, err = captain.GetPromptRun(context.Background(), promptRun)
		require.NoError(t, err)
		return stored.State == captaindb.PromptRunStateRunning
	}, 30*time.Second, 50*time.Millisecond, "the agent is working")
	decided := make(chan api.ApprovalDecision, 1)
	go func() {
		broker := &approval.Broker{DB: captain, SessionID: stored.SessionID, PromptRunID: promptRun, RequestedBy: "gavel-dashboard", Timeout: time.Minute,
			Notify: func(context.Context, api.Event) error { return nil }}
		decision, err := broker.OnApproval(context.Background(), api.ApprovalRequest{Tool: "Bash", Kind: api.ApprovalKindTool, Input: map[string]any{"command": "make"}, ToolUseID: "call-1"})
		assert.NoError(t, err)
		decided <- decision
	}()
	approvals := "/api/todos/session/approvals?dir=" + workspace + "&sessionId=" + stored.SessionID.String()
	var pending []any
	require.Eventually(t, func() bool {
		_, listed := hostedCall(t, base, http.MethodGet, approvals, creator, "org:member", nil)
		pending, _ = listed["approvals"].([]any)
		return len(pending) == 1
	}, 30*time.Second, 50*time.Millisecond, "the creator sees the pending approval")
	approvalID := pending[0].(map[string]any)["approvalId"].(string)

	answer := map[string]any{"approvalId": approvalID, "action": "approve"}
	status, refused := hostedCall(t, base, http.MethodPost, "/api/todos/session/approve?dir="+workspace, member, "org:member", answer)
	assert.Equal(t, http.StatusForbidden, status, "another member can't answer it: %v", refused)

	status, answered := hostedCall(t, base, http.MethodPost, "/api/todos/session/approve?dir="+workspace, creator, "org:member", answer)
	require.Equal(t, http.StatusOK, status, answered)
	assert.True(t, (<-decided).Allow, "the run gets the creator's approval")

	stop := map[string]any{"ref": todo, "promptRunId": promptRun.String()}
	status, stopped := hostedCall(t, base, http.MethodPost, "/api/todos/session/stop?dir="+workspace, creator, "org:member", stop)
	require.Equal(t, http.StatusAccepted, status, stopped)
	var outcomes int64
	require.Eventually(t, func() bool {
		require.NoError(t, db.Raw(`SELECT count(*) FROM todo_issue_events WHERE issue_id = ? AND kind = 'lifecycle_outcome'`, todo).Scan(&outcomes).Error)
		return outcomes == 1
	}, 30*time.Second, 50*time.Millisecond, "the stopped run's outcome reached the todo")
}
