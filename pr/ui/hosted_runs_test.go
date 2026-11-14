// Tests hosted agent runs: a run of a hosted todo is admitted and recorded on serve while its
// prompt runs whole on substrate, which here runs in this process with a scripted agent.
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
	capsetup "github.com/flanksource/captain/pkg/ai/agent/setup"
	"github.com/flanksource/captain/pkg/ai/approval"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/captain/pkg/sandbox/substraterun"
	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/internal/safefs"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// planAgent ends every turn by presenting a new plan, in place of an agent CLI.
type planAgent struct{}

func (planAgent) GetModel() string        { return "claude-sonnet-5" }
func (planAgent) GetRuntime() api.Runtime { return api.RuntimeOf(api.Anthropic, api.ModeAgent) }
func (planAgent) Execute(context.Context, api.Spec) (*api.Response, error) {
	return nil, errors.New("the plan agent only streams")
}
func (planAgent) ExecuteStream(context.Context, api.Spec) (<-chan captainai.Event, error) {
	ch := make(chan captainai.Event, 3)
	ch <- captainai.Event{Kind: captainai.EventSystem, SessionID: uuid.NewString()}
	ch <- captainai.Event{Kind: captainai.EventToolUse, Tool: "ExitPlanMode", Input: map[string]any{"plan": "# Plan\n\n1. Fix the build."}}
	ch <- captainai.Event{Kind: captainai.EventResult, Success: true}
	close(ch)
	return ch, nil
}

// inProcessSubstrate stands in for captain's substrate sandbox: it runs each agent run whole in this
// process, as gavel's program does in an environment, from the commit and repository the run's
// selection names, with gavel's hooks rebuilt from the run's host data and the test's agent. The
// run's tools and approvals are the caller's own, which an environment's program calls back. It
// keeps each input it was handed, and counts the runs that were stopped.
type inProcessSubstrate struct {
	agent captainai.Provider
	mu    sync.Mutex
	// options are the selections' options, by environment.
	options  map[string]map[string]any
	runs     []promptrun.Input
	contexts []context.Context
	canceled int
}

// selected is the sandbox the selection names: this one, keeping its options for its run.
func (s *inProcessSubstrate) selected(config api.SandboxConfig) (api.Sandbox, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.options == nil {
		s.options = map[string]map[string]any{}
	}
	s.options[config.Options["environment"].(string)] = config.Options
	return &substrateRun{substrate: s, options: config.Options}, nil
}

// handed is the input of the run the substrate was handed last, and the context it runs under,
// which its tools are called back with, once it was handed one.
func (s *inProcessSubstrate) handed() (promptrun.Input, context.Context, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.runs) == 0 {
		return promptrun.Input{}, nil, false
	}
	return s.runs[len(s.runs)-1], s.contexts[len(s.contexts)-1], true
}

type substrateRun struct {
	substrate *inProcessSubstrate
	options   map[string]any
}

func (r *substrateRun) Kind() api.SandboxKind { return api.SandboxSubstrate }
func (r *substrateRun) Prepare(context.Context, *api.Spec) (*api.SandboxSession, error) {
	return &api.SandboxSession{}, nil
}
func (r *substrateRun) Close() error              { return nil }
func (r *substrateRun) IsolatesWorkspace() bool   { return true }
func (r *substrateRun) ProvidesEgressProxy() bool { return true }
func (r *substrateRun) RunRemote(ctx context.Context, in promptrun.Input, started func(api.Spec) error) (promptrun.Result, error) {
	s := r.substrate
	s.mu.Lock()
	s.runs = append(s.runs, in)
	s.contexts = append(s.contexts, ctx)
	s.mu.Unlock()
	defer func() {
		if ctx.Err() != nil {
			s.mu.Lock()
			s.canceled++
			s.mu.Unlock()
		}
	}()
	dir := filepath.Join(os.TempDir(), "gavel-hosted-run-"+r.options["environment"].(string))
	defer safefs.SafeRemove(dir, safefs.RemoveContext{AllowedRoot: os.TempDir()}) //nolint:errcheck
	if out, err := exec.Command("git", "clone", "-q", in.Repo, dir).CombinedOutput(); err != nil {
		return promptrun.Result{}, errors.New(string(out))
	}
	if out, err := exec.Command("git", "-C", dir, "checkout", "-q", "--detach", r.options["ref"].(string)).CombinedOutput(); err != nil {
		return promptrun.Result{}, errors.New(string(out))
	}
	var hooks lifecycle.RunHooks
	if err := json.Unmarshal(in.HostData, &hooks); err != nil {
		return promptrun.Result{}, err
	}
	built, local, err := hooks.In(in.Repo, dir).Build()
	if err != nil {
		return promptrun.Result{}, err
	}
	spec := in.Resolved.Spec
	spec.SetCwd(dir)
	spec.Sandbox = &api.SandboxRef{Mode: api.SandboxOff}
	if err := started(spec); err != nil {
		return promptrun.Result{}, err
	}
	var agent captainai.Provider = planAgent{}
	if s.agent != nil {
		agent = s.agent
	}
	config := captainai.Config{Model: in.Config.Model, Tools: append(append([]api.ToolDefinition(nil), in.Config.Tools...), local...), OnApproval: in.Config.OnApproval}
	return promptrun.Run(ctx, promptrun.Input{
		Resolved: api.ResolvedSpec{Spec: spec}, Config: config, Provider: agent,
		Hooks: append(built, &capsetup.Plugin{BaseDir: dir}), CallerOwnsCommits: in.CallerOwnsCommits,
		Verify: in.Verify, Timeout: in.Timeout, Repo: dir, OnEvent: in.OnEvent,
	})
}

// inProcessExecutor runs each test or lint job's sandbox side in this process: the job's checkout
// of the pushed repository.
type inProcessExecutor struct {
	db       *gorm.DB
	bareDir  func(repo uuid.UUID) string
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
	dir := filepath.Join(os.TempDir(), "gavel-hosted-run-"+id.String())
	defer safefs.SafeRemove(dir, safefs.RemoveContext{AllowedRoot: os.TempDir()}) //nolint:errcheck
	if out, err := exec.Command("git", "clone", "-q", e.bareDir(record.Spec.RepoID), dir).CombinedOutput(); err != nil {
		return errors.New(string(out))
	}
	if out, err := exec.Command("git", "-C", dir, "checkout", "-q", "--detach", record.Spec.SHA).CombinedOutput(); err != nil {
		return errors.New(string(out))
	}
	// A test or lint job's work is the runner's, which these tests don't run; it ends checked out.
	return nil
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

// hostedRuns is a hosted dashboard whose agent runs execute in this process, serving one pushed
// repository, app, to two members, ada and grace.
type hostedRuns struct {
	db              *gorm.DB
	workspace, base string
	org, ada, grace uuid.UUID
	executor        *inProcessExecutor
	substrate       *inProcessSubstrate
	// logs is the store the dispatcher keeps the organization's run logs in.
	logs blob.Store
	// dataDir is serve's data directory, which keeps the runs' transcripts.
	dataDir string
}

// newHostedRuns serves a hosted dashboard whose runs' turns agent answers; nil is planAgent.
func newHostedRuns(t *testing.T, agent captainai.Provider) hostedRuns {
	t.Helper()
	substrate := &inProcessSubstrate{agent: agent}
	api.RegisterSandbox(api.SandboxSubstrate, substrate.selected)
	t.Cleanup(func() { api.RegisterSandbox(api.SandboxSubstrate, substraterun.Substrate) })
	h := serveHostedRuns(t, map[string]any{"envAPI": "env.example:443", "control": "control.example:443", "image": "runner@sha256:" + strings.Repeat("0", 64)}, nil)
	h.substrate = substrate
	return h
}

// serveHostedRuns serves a hosted dashboard whose agent runs execute on the substrate backend, as
// its options spell it; configure, when set, adjusts the dispatcher before serving.
func serveHostedRuns(t *testing.T, backend map[string]any, configure func(*sandbox.Dispatcher)) hostedRuns {
	t.Helper()
	db, _, org := openOrgDB(t)
	t.Setenv("HOME", t.TempDir())
	h := hostedRuns{db: db, org: org, ada: uuid.New(), grace: uuid.New()}
	for name, id := range map[string]uuid.UUID{"ada": h.ada, "grace": h.grace} {
		require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, "user_"+name).Error)
	}
	dataDir := t.TempDir()
	h.dataDir = dataDir
	h.workspace = "repo:" + pushedRepo(t, db, org, dataDir, "app").String()
	logs := blob.Local(t.TempDir())
	h.logs = logs
	h.executor = &inProcessExecutor{db: db, bareDir: func(id uuid.UUID) string {
		return filepath.Join(dataDir, "repos", org.String(), id.String()+".git")
	}}
	dispatcher := &sandbox.Dispatcher{
		DB: db, Image: "runner", Server: "https://gavel.example", TokenKey: []byte(strings.Repeat("k", 32)),
		Logs: func(uuid.UUID) blob.Store { return logs }, Limits: sandbox.Limits{Timeout: time.Minute}, Executor: h.executor,
	}
	if configure != nil {
		configure(dispatcher)
	}
	srv := &Server{}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)
	srv.SetHostedWorkspaces(HostedWorkspaces{DB: db, Sources: &sandbox.Sources{DataDir: dataDir}})
	sources := &sandbox.Sources{DataDir: dataDir}
	srv.SetAgentRuns(AgentRuns{
		Dispatcher: dispatcher, Sources: sources, Transcripts: RunTranscripts{DB: db, DataDir: dataDir},
		Substrate: backend,
	})
	h.base = serveActingAs(t, srv, org)
	return h
}

func TestHostedTodoRunsInASandboxJob(t *testing.T) {
	runsPlanStep(t, newHostedRuns(t, nil), "Fix the build", 30*time.Second)
}

// runsPlanStep runs the plan step of a new todo titled title on h, as its creator, and checks
// within wait that its sandbox run and Captain's run succeeded and that the plan was accepted as
// the todo's outcome.
func runsPlanStep(t *testing.T, h hostedRuns, title string, wait time.Duration) {
	t.Helper()
	db, base, workspace, creator := h.db, h.base, h.workspace, h.ada

	status, created := hostedCall(t, base, http.MethodPost, "/api/todos", creator, "org:member", map[string]any{"dir": workspace, "title": title})
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
	if !assert.Eventually(t, func() bool {
		require.NoError(t, db.Raw(`SELECT status, kind, user_id, sha FROM sandbox_runs WHERE todo_id = ?`, todo).Scan(&run).Error)
		var finished bool
		require.NoError(t, db.Raw(`SELECT count(*) > 0 FROM captain_prompt_runs WHERE id = ? AND state IN ('failed', 'cancelled', 'waiting')`, started["promptRunId"]).Scan(&finished).Error)
		return run.Status == "succeeded" || run.Status == "failed" || finished
	}, wait, 50*time.Millisecond, "the sandbox job ran") || run.Status != "succeeded" {
		var captain string
		require.NoError(t, db.Raw(`SELECT row_to_json(r)::text FROM captain_prompt_runs r WHERE id = ?`, started["promptRunId"]).Scan(&captain).Error)
		t.Fatalf("the sandbox job didn't succeed: %+v; Captain's run: %s", run, captain)
	}
	assert.Equal(t, "agent", run.Kind)
	assert.Equal(t, creator, run.UserID)
	assert.NotEmpty(t, run.SHA)

	var promptRun struct{ State, Error string }
	require.Eventually(t, func() bool {
		require.NoError(t, db.Raw(`SELECT state, coalesce(error, '') AS error FROM captain_prompt_runs WHERE id = ?`, started["promptRunId"]).Scan(&promptRun).Error)
		return promptRun.State == "succeeded" || promptRun.State == "failed" || promptRun.State == "cancelled"
	}, wait, 50*time.Millisecond, "Captain recorded the run's outcome")
	require.Equal(t, "succeeded", promptRun.State, "Captain's run: %s", promptRun.Error)

	// The outcome event is the run's last write: once it is there, the todo has its outcome.
	var outcomes int64
	require.Eventually(t, func() bool {
		require.NoError(t, db.Raw(`SELECT count(*) FROM todo_issue_events WHERE issue_id = ? AND kind = 'lifecycle_outcome'`, todo).Scan(&outcomes).Error)
		return outcomes == 1
	}, wait, 50*time.Millisecond, "the run's outcome reached the todo")
	var outcome struct {
		RunState string
		Success  bool
		Event    string
	}
	require.NoError(t, db.Raw(`SELECT payload->'run'->>'state' AS run_state, (payload->'envelope'->>'success')::boolean AS success,
		payload->'envelope'->>'event' AS event
		FROM todo_issue_events WHERE issue_id = ? AND kind = 'lifecycle_outcome'`, todo).Scan(&outcome).Error)
	assert.Equal(t, "succeeded", outcome.RunState, "the plan the sandboxed agent wrote was accepted")
	assert.True(t, outcome.Success)
	assert.Equal(t, "new", outcome.Event)
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
	assert.Equal(t, "agent runs need a substrate sandbox: start gavel serve with --sandbox substrate", refused["error"])
}

func TestEveryHostedRunGoesToASandbox(t *testing.T) {
	hosted := todoSource{Dir: t.TempDir(), Hosted: &HostedWorkspace{Repo: sandbox.Repo{ID: uuid.New()}, SHA: "abc123"}}
	todo := &types.TODO{ID: uuid.NewString()}
	override, err := continuationOverride(nil)
	require.NoError(t, err)
	continuation := run.Continuation{Dir: hosted.Dir, Todo: todo, Step: "run", Override: override}

	withSandbox := &Server{}
	withSandbox.SetAgentRuns(AgentRuns{Dispatcher: &sandbox.Dispatcher{}, Substrate: map[string]any{}})
	req, err := withSandbox.continuationRequest(context.Background(), hosted, continuation)
	require.NoError(t, err)
	assert.NotNil(t, req.Remote, "a hosted continuation runs in a sandbox")
	require.NotNil(t, req.PRs, "and reads linked PRs with the organization's GitHub App")
	_, _, err = req.PRs(context.Background(), "https://dev.azure.com/acme/app/_git/app/pullrequest/7")
	require.EqualError(t, err, "hosted todos work on GitHub PRs; https://dev.azure.com/acme/app/_git/app/pullrequest/7 isn't one")
	assert.True(t, req.Approvals, "a sandboxed run asks serve for approvals")

	_, err = (&Server{}).continuationRequest(context.Background(), hosted, continuation)
	require.ErrorIs(t, err, errAgentRunsNeedSubstrate)
	withDocker := &Server{}
	withDocker.SetAgentRuns(AgentRuns{Dispatcher: &sandbox.Dispatcher{}})
	_, err = withDocker.continuationRequest(context.Background(), hosted, continuation)
	require.ErrorIs(t, err, errAgentRunsNeedSubstrate, "a sandbox other than substrate runs no agent runs")

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
	h := newHostedRuns(t, workingAgent{})
	db, base, workspace, creator, member, substrate := h.db, h.base, h.workspace, h.ada, h.grace, h.substrate

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
	substrate.mu.Lock()
	assert.Equal(t, 1, substrate.canceled, "stopping the run stops it on substrate")
	substrate.mu.Unlock()
	var outcomes int64
	require.Eventually(t, func() bool {
		require.NoError(t, db.Raw(`SELECT count(*) FROM todo_issue_events WHERE issue_id = ? AND kind = 'lifecycle_outcome'`, todo).Scan(&outcomes).Error)
		return outcomes == 1
	}, 30*time.Second, 50*time.Millisecond, "the stopped run's outcome reached the todo")
}

func TestHostedTriageNewRunsWithItsToolsOnServe(t *testing.T) {
	h := newHostedRuns(t, workingAgent{})
	db, base, workspace, creator := h.db, h.base, h.workspace, h.ada
	_, created := hostedCall(t, base, http.MethodPost, "/api/todos", creator, "org:member", map[string]any{"dir": workspace, "title": "Fix the build"})
	todo := created["id"].(string)

	status, started := hostedCall(t, base, http.MethodPost, "/api/todos/run", creator, "org:member", map[string]any{"dir": workspace, "ref": todo, "step": "triage.new"})
	require.Equal(t, http.StatusOK, status, started)

	var handed promptrun.Input
	require.Eventually(t, func() bool {
		var ok bool
		handed, _, ok = h.substrate.handed()
		return ok
	}, 30*time.Second, 50*time.Millisecond)
	var tools []string
	for _, tool := range handed.Config.Tools {
		tools = append(tools, tool.Name)
	}
	assert.ElementsMatch(t, []string{
		"todo_search", "todo_get", "todo_update", "todo_labels", "todo_link", "todo_link_pr",
		"todo_set_parent", "todo_comment", "todo_plan", "todo_merge", "todo_close",
	}, tools, "the run calls serve for triage.new's todo tools, and only those")
	assert.NotNil(t, handed.Config.OnApproval, "and asks serve for the person's approval of a merge")

	stop := map[string]any{"ref": todo, "promptRunId": started["promptRunId"]}
	status, stopped := hostedCall(t, base, http.MethodPost, "/api/todos/session/stop?dir="+workspace, creator, "org:member", stop)
	require.Equal(t, http.StatusAccepted, status, stopped)
	var outcomes int64
	require.Eventually(t, func() bool {
		require.NoError(t, db.Raw(`SELECT count(*) FROM todo_issue_events WHERE issue_id = ? AND kind = 'lifecycle_outcome'`, todo).Scan(&outcomes).Error)
		return outcomes == 1
	}, 30*time.Second, 50*time.Millisecond, "the stopped run's outcome reached the todo")
}

func TestHostedRunToolsSeeOnlyWhatTheRunsUserSees(t *testing.T) {
	h := newHostedRuns(t, workingAgent{})
	db, base, workspace, creator, member := h.db, h.base, h.workspace, h.ada, h.grace
	_, private := hostedCall(t, base, http.MethodPost, "/api/todos", member, "org:member", map[string]any{"dir": workspace, "title": "Grace's private todo"})
	privateRef := private["id"].(string)
	_, created := hostedCall(t, base, http.MethodPost, "/api/todos", creator, "org:member", map[string]any{"dir": workspace, "title": "Fix the build"})
	todo := created["id"].(string)

	status, started := hostedCall(t, base, http.MethodPost, "/api/todos/run", creator, "org:member", map[string]any{"dir": workspace, "ref": todo, "step": "triage.new"})
	require.Equal(t, http.StatusOK, status, started)
	var handed promptrun.Input
	var runCtx context.Context
	require.Eventually(t, func() bool {
		var ok bool
		handed, runCtx, ok = h.substrate.handed()
		return ok
	}, 30*time.Second, 50*time.Millisecond)
	// The run's program calls its tools back on serve, which runs them here under the run's
	// context, as captain's relay does.
	call := func(tool string, input map[string]any) (string, error) {
		for _, definition := range handed.Config.Tools {
			if definition.Name == tool {
				output, err := definition.Handler(runCtx, input)
				data, _ := json.Marshal(output)
				return string(data), err
			}
		}
		t.Fatalf("the run has no tool %s", tool)
		return "", nil
	}

	searched, err := call("todo_search", map[string]any{"query": ""})
	require.NoError(t, err)
	assert.NotContains(t, searched, "Grace's private todo", "the run's search doesn't list another member's private todo")
	_, err = call("todo_get", map[string]any{"ref": privateRef})
	assert.Error(t, err, "the run can't read another member's private todo")
	_, err = call("todo_update", map[string]any{"ref": privateRef, "title": "Renamed by ada's run"})
	assert.Error(t, err, "nor change it")
	var title string
	require.NoError(t, db.Raw(`SELECT title FROM todo_issues WHERE id = ?`, privateRef).Scan(&title).Error)
	assert.Equal(t, "Grace's private todo", title)

	stop := map[string]any{"ref": todo, "promptRunId": started["promptRunId"]}
	status, stopped := hostedCall(t, base, http.MethodPost, "/api/todos/session/stop?dir="+workspace, creator, "org:member", stop)
	require.Equal(t, http.StatusAccepted, status, stopped)
}

func TestHostedApprovalsBelongToTheTodosOwner(t *testing.T) {
	h := newHostedRuns(t, workingAgent{})
	db, base, workspace, creator, member := h.db, h.base, h.workspace, h.ada, h.grace
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

func TestHostedRunSessionsAreSeenByThoseWhoSeeTheTodo(t *testing.T) {
	h := newHostedRuns(t, nil)
	_, created := hostedCall(t, h.base, http.MethodPost, "/api/todos", h.ada, "org:member", map[string]any{"dir": h.workspace, "title": "Fix the build"})
	todo := created["id"].(string)
	status, started := hostedCall(t, h.base, http.MethodPost, "/api/todos/run", h.ada, "org:member", map[string]any{"dir": h.workspace, "ref": todo, "step": "plan"})
	require.Equal(t, http.StatusOK, status, started)
	var outcomes int64
	require.Eventually(t, func() bool {
		require.NoError(t, h.db.Raw(`SELECT count(*) FROM todo_issue_events WHERE issue_id = ? AND kind = 'lifecycle_outcome'`, todo).Scan(&outcomes).Error)
		return outcomes == 1
	}, 30*time.Second, 50*time.Millisecond, "the plan run finished")

	detailPath := "/api/todos/session/detail?dir=" + h.workspace + "&ref=" + todo
	status, detail := hostedCall(t, h.base, http.MethodGet, detailPath, h.ada, "org:member", nil)
	require.Equal(t, http.StatusOK, status, detail)
	attempts := detail["attempts"].([]any)
	require.Len(t, attempts, 1)
	attempt := attempts[0].(map[string]any)
	assert.Equal(t, started["promptRunId"], attempt["promptRunId"])
	admission := attempt["admissionSessionId"].(string)

	planPath := "/api/todos/session/plan?dir=" + h.workspace + "&ref=" + todo
	status, plan := hostedCall(t, h.base, http.MethodGet, planPath, h.ada, "org:member", nil)
	require.Equal(t, http.StatusOK, status, plan)
	assert.Contains(t, plan["content"], "Fix the build.", "the plan the sandboxed agent wrote")

	statsPath := "/api/todos/session/stats?dir=" + h.workspace + "&sessionId=" + admission
	status, stats := hostedCall(t, h.base, http.MethodGet, statsPath, h.ada, "org:member", nil)
	require.Equal(t, http.StatusOK, status, stats)
	agentStatsPath := "/api/todos/session/stats?dir=" + h.workspace + "&sessionId=" + attempt["providerSessionId"].(string)
	status, stats = hostedCall(t, h.base, http.MethodGet, agentStatsPath, h.ada, "org:member", nil)
	require.Equal(t, http.StatusOK, status, "the agent's own session, as the dashboard follows it: %v", stats)

	transcriptPath := "/api/captain/sessions/" + admission
	status, transcript := hostedCall(t, h.base, http.MethodGet, transcriptPath, h.ada, "org:member", nil)
	require.Equal(t, http.StatusOK, status, "the run's transcript: %v", transcript)

	for _, path := range []string{detailPath, planPath, statsPath, agentStatsPath, transcriptPath} {
		status, refused := hostedCall(t, h.base, http.MethodGet, path, h.grace, "org:member", nil)
		assert.Equal(t, http.StatusNotFound, status, "another member doesn't see %s of a private todo: %v", path, refused)
	}
}

func TestHostedRunsUseTheRepositorysRuntimeForEachStep(t *testing.T) {
	h := newHostedRuns(t, nil)
	status, context := hostedCall(t, h.base, http.MethodGet, "/api/todos/run/context?dir="+h.workspace, h.ada, "org:member", nil)
	require.Equal(t, http.StatusOK, status, context)

	var steps []string
	for _, step := range context["lifecycle"].(map[string]any)["steps"].([]any) {
		steps = append(steps, step.(map[string]any)["name"].(string))
	}
	assert.Contains(t, steps, "plan")
	assert.Contains(t, steps, "run")
	assert.Empty(t, context["runtimePresets"], "a hosted run picks no presets")
	assert.NotEmpty(t, context["runtimes"])

	defaults := map[string]bool{}
	for step, raw := range context["promptDefaults"].(map[string]any) {
		resolved := raw.(map[string]any)
		if step == "verify" {
			continue // the definition of done runs checks, not a model
		}
		require.NotEmpty(t, resolved["model"], "step %s resolves a model", step)
		require.NotEmpty(t, resolved["mode"], "step %s resolves a mode", step)
		defaults[resolved["mode"].(string)+"/"+resolved["model"].(string)] = true
	}
	modes := context["modes"].([]any)
	require.NotEmpty(t, modes)
	for _, raw := range modes {
		mode := raw.(map[string]any)
		models := mode["models"].([]any)
		require.NotEmpty(t, models, "mode %s offers its default model", mode["id"])
		for _, model := range models {
			assert.True(t, defaults[mode["id"].(string)+"/"+model.(map[string]any)["id"].(string)],
				"mode %s offers only a step default, not %v", mode["id"], model.(map[string]any)["id"])
		}
	}

	_, created := hostedCall(t, h.base, http.MethodPost, "/api/todos", h.ada, "org:member", map[string]any{"dir": h.workspace, "title": "Fix the build"})
	status, refused := hostedCall(t, h.base, http.MethodPost, "/api/todos/run", h.ada, "org:member", map[string]any{
		"dir": h.workspace, "ref": created["id"], "step": "plan", "spec": map[string]any{"model": "gpt-5", "mode": "agent"},
	})
	assert.Equal(t, http.StatusBadRequest, status, refused)
	assert.Equal(t, "a hosted run uses the repository's runtime for each step; this one asks for model gpt-5", refused["error"])
}

func TestHostedRunPreviewsResolveAsTheRunWould(t *testing.T) {
	h := newHostedRuns(t, nil)
	_, created := hostedCall(t, h.base, http.MethodPost, "/api/todos", h.ada, "org:member", map[string]any{"dir": h.workspace, "title": "Fix the build", "visibility": "public"})
	preview := map[string]any{"dir": h.workspace, "ref": created["id"], "step": "plan"}

	status, previewed := hostedCall(t, h.base, http.MethodPost, "/api/todos/run/preview", h.ada, "org:member", preview)
	require.Equal(t, http.StatusOK, status, previewed)
	assert.Equal(t, "plan", previewed["step"])
	assert.Equal(t, "claude-opus-5-5", previewed["model"], "the repository's plan runtime")

	status, refused := hostedCall(t, h.base, http.MethodPost, "/api/todos/run/preview", h.grace, "org:member", preview)
	assert.Equal(t, http.StatusForbidden, status, "only those who can run it preview a run: %v", refused)

	preview["spec"] = map[string]any{"model": "gpt-5"}
	status, refused = hostedCall(t, h.base, http.MethodPost, "/api/todos/run/preview", h.ada, "org:member", preview)
	assert.Equal(t, http.StatusBadRequest, status, refused)
	assert.Equal(t, "a hosted run uses the repository's runtime for each step; this one asks for model gpt-5", refused["error"])
}

func TestHostedPlansAreEditedByTheirOwner(t *testing.T) {
	h := newHostedRuns(t, nil)
	_, created := hostedCall(t, h.base, http.MethodPost, "/api/todos", h.ada, "org:member", map[string]any{"dir": h.workspace, "title": "Fix the build", "visibility": "public"})
	todo := created["id"].(string)
	status, started := hostedCall(t, h.base, http.MethodPost, "/api/todos/run", h.ada, "org:member", map[string]any{"dir": h.workspace, "ref": todo, "step": "plan"})
	require.Equal(t, http.StatusOK, status, started)
	var outcomes int64
	require.Eventually(t, func() bool {
		require.NoError(t, h.db.Raw(`SELECT count(*) FROM todo_issue_events WHERE issue_id = ? AND kind = 'lifecycle_outcome'`, todo).Scan(&outcomes).Error)
		return outcomes == 1
	}, 30*time.Second, 50*time.Millisecond, "the plan run finished")

	edit := map[string]any{"dir": h.workspace, "ref": todo, "content": "# Plan\n\n1. Fix the build properly."}
	status, refused := hostedCall(t, h.base, http.MethodPost, "/api/todos/session/plan", h.grace, "org:member", edit)
	assert.Equal(t, http.StatusForbidden, status, "another member can't edit the plan: %v", refused)
	status, saved := hostedCall(t, h.base, http.MethodPost, "/api/todos/session/plan", h.ada, "org:member", edit)
	require.Equal(t, http.StatusOK, status, saved)
	_, plan := hostedCall(t, h.base, http.MethodGet, "/api/todos/session/plan?dir="+h.workspace+"&ref="+todo, h.grace, "org:member", nil)
	assert.Contains(t, plan["content"], "Fix the build properly.")
}
