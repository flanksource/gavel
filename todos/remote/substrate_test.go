// Tests SubstrateRunner: a hosted step's run goes to captain's substrate sandbox as an agent run of
// the dispatcher, from the todo's commit, and its log, transcripts and branch come back to serve.
package remote

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
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/captain/pkg/sandbox/substraterun"
	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/captain/pkg/substrate/substratetest"
	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/budget"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/sandbox/runclient"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// exitingExecutor runs nothing: each job exits with exitCode, and a cancelled job is recorded.
type exitingExecutor struct {
	exitCode  int
	submitted []sandbox.Job
	cancelled int
}

func (e *exitingExecutor) Name() string { return "exiting" }
func (e *exitingExecutor) Submit(_ context.Context, job sandbox.Job) (sandbox.Handle, error) {
	e.submitted = append(e.submitted, job)
	return exitingHandle{executor: e}, nil
}

type exitingHandle struct{ executor *exitingExecutor }

func (h exitingHandle) Logs(context.Context) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}
func (h exitingHandle) Wait(context.Context) (sandbox.Result, error) {
	return sandbox.Result{ExitCode: h.executor.exitCode}, nil
}
func (h exitingHandle) Cancel(context.Context) error {
	h.executor.cancelled++
	return nil
}
func (h exitingHandle) Ref() string { return "exiting" }

// environment stands in for captain's substrate sandbox: it keeps the selection and input it was
// handed, writes the environment's output, sends a transcript, and commits on the branch the
// selection names in the run's repository, as the sandbox brings a run's commits back.
type environment struct {
	t         *testing.T
	selection api.SandboxConfig
	in        promptrun.Input
	fail      error
	prStatus  []byte
	// deadline is when the run's context ends, as the sandbox was handed it.
	deadline time.Time
	// holds, when set, keeps the run going after its transcript until its context ends.
	holds bool
}

func (e *environment) Kind() api.SandboxKind { return api.SandboxSubstrate }
func (e *environment) Prepare(context.Context, *api.Spec) (*api.SandboxSession, error) {
	return &api.SandboxSession{}, nil
}
func (e *environment) Close() error              { return nil }
func (e *environment) IsolatesWorkspace() bool   { return true }
func (e *environment) ProvidesEgressProxy() bool { return true }
func (e *environment) RunRemote(ctx context.Context, in promptrun.Input, started func(api.Spec) error) (promptrun.Result, error) {
	e.in = in
	e.deadline, _ = ctx.Deadline()
	require.NoError(e.t, started(in.Resolved.Spec))
	_, err := io.WriteString(in.EnvironmentOutput, "the environment's output\n")
	require.NoError(e.t, err)
	in.OnTranscript("claude", "-tmp-captain-run-src/session-1.jsonl", []byte(`{"type":"user"}`+"\n"))
	if e.holds {
		select {
		case <-ctx.Done():
			return promptrun.Result{}, ctx.Err()
		case <-time.After(10 * time.Second):
			return promptrun.Result{}, errors.New("the run was never stopped")
		}
	}
	// The step's PR hook reads its PR through the caller, as gavel's program does.
	status, err := runclient.OverHostCall(func(name string, data json.RawMessage) (json.RawMessage, error) {
		return in.HostCalls[name](context.Background(), data)
	}).Do(context.Background(), "GET", "pr/status?number=7", nil)
	require.NoError(e.t, err)
	e.prStatus, _ = io.ReadAll(status.Body)
	if e.fail != nil {
		return promptrun.Result{}, e.fail
	}
	branch := e.selection.Options["branch"].(string)
	work := e.t.TempDir()
	gitRun(e.t, work, "clone", "-q", in.Repo, ".")
	gitRun(e.t, work, "checkout", "-q", "-b", "run", e.selection.Options["ref"].(string))
	require.NoError(e.t, os.WriteFile(filepath.Join(work, "fix.txt"), []byte("fixed\n"), 0o644))
	gitRun(e.t, work, "add", "fix.txt")
	gitRun(e.t, work, "commit", "-q", "-m", "fix the build")
	gitRun(e.t, work, "push", "-q", in.Repo, "run:refs/heads/"+branch)
	head := gitRun(e.t, work, "rev-parse", "HEAD")
	return promptrun.Result{Passed: true, Response: &api.Response{Text: "done", Workspace: &api.Workspace{
		Worktree: &api.WorktreeState{Repo: in.Repo, Branch: branch, Head: head},
	}}}, nil
}

func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, out)
	return strings.TrimSpace(string(out))
}

// hostedRun is a hosted todo's repository on serve, pushed, with one commit, and a runner for its
// steps whose substrate sandbox is env.
type hostedRun struct {
	runner      *SubstrateRunner
	logs        blob.Store
	gitDir      string
	base        string
	todo        uuid.UUID
	org         uuid.UUID
	transcripts []string
}

func newHostedRun(t *testing.T, env *environment) *hostedRun {
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_remote_substrate"}).DSN())
	t.Setenv(database.EnvDisable, "")
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	db := opened.Gorm()
	org := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)
	var repo sandbox.Repo
	require.NoError(t, db.Raw(`INSERT INTO repos (org_id, name, source) VALUES (?, 'app', 'push') RETURNING id, org_id AS org, source, name`, org).Scan(&repo).Error)
	sources := &sandbox.Sources{DataDir: t.TempDir()}
	gitDir, err := sources.GitDir(t.Context(), repo)
	require.NoError(t, err)
	work := t.TempDir()
	gitRun(t, work, "init", "-q", "-b", "main")
	gitRun(t, work, "commit", "-q", "--allow-empty", "-m", "first")
	gitRun(t, work, "clone", "-q", "--bare", work, gitDir)
	logs := blob.Local(t.TempDir())
	hosted := &hostedRun{logs: logs, gitDir: gitDir, base: gitRun(t, work, "rev-parse", "HEAD"), todo: uuid.New(), org: org}
	hosted.runner = &SubstrateRunner{
		Dispatcher: &sandbox.Dispatcher{
			DB: db, Executor: &exitingExecutor{}, Server: "http://gavel.gavel.svc:8080", TokenKey: []byte(strings.Repeat("k", 32)),
			Logs: func(uuid.UUID) blob.Store { return logs }, Limits: sandbox.Limits{Timeout: time.Hour},
		},
		Sources: sources,
		Backend: map[string]any{"envAPI": "env.example:443", "control": "control.example:443", "image": "ghcr.io/flanksource/gavel-runner@sha256:" + strings.Repeat("0", 64)},
		Scope:   tenant.Scope{OrgID: org}, Repo: repo, SHA: hosted.base, TodoID: hosted.todo,
		RunAPI: runAPIFor(t, []byte(strings.Repeat("k", 32))),
		Transcripts: func(_ context.Context, run sandbox.RunRecord, source, name string, data io.Reader) error {
			content, err := io.ReadAll(data)
			require.NoError(t, err)
			hosted.transcripts = append(hosted.transcripts, run.Org.String()+" "+source+"/"+name+" "+string(content))
			return nil
		},
	}
	api.RegisterSandbox(api.SandboxSubstrate, func(config api.SandboxConfig) (api.Sandbox, error) {
		env.selection = config
		return env, nil
	})
	t.Cleanup(func() { api.RegisterSandbox(api.SandboxSubstrate, substraterun.Substrate) })
	return hosted
}

// step is a hosted step's run, as the lifecycle hands it to a remote runner.
func step() lifecycle.RemoteRun {
	spec := api.Spec{Model: api.Model{Name: "claude-sonnet-5", Provider: api.Anthropic, Mode: api.ModeAgent}, Prompt: api.Prompt{User: "fix the build"}}
	spec.Sandbox = &api.SandboxRef{Mode: api.SandboxOff}
	hooks := lifecycle.RunHooks{Env: map[string]string{"GAVEL_ISSUE_ID": "todo-1"}}
	data, _ := json.Marshal(hooks)
	return lifecycle.RemoteRun{
		Input: promptrun.Input{Resolved: api.ResolvedSpec{Spec: spec}, Config: captainai.Config{Model: spec.Model}, Timeout: time.Minute, HostData: data},
		Hooks: hooks,
	}
}

func TestSubstrateRunnerRunsAHostedStepOnSubstrate(t *testing.T) {
	env := &environment{t: t}
	hosted := newHostedRun(t, env)

	result, err := hosted.runner.Run(t.Context(), step())
	require.NoError(t, err)
	assert.Equal(t, "done", result.Response.Text)

	var row struct{ ID uuid.UUID }
	require.NoError(t, hosted.runner.Dispatcher.DB.Raw(`SELECT id FROM sandbox_runs WHERE org_id = ?`, hosted.org).Scan(&row).Error)
	id := row.ID
	record, err := sandbox.LoadRun(t.Context(), hosted.runner.Dispatcher.DB, id)
	require.NoError(t, err)
	assert.Equal(t, "succeeded", record.Status)
	assert.Equal(t, "agent", record.Spec.Kind)
	assert.Equal(t, hosted.base, record.Spec.SHA)

	options := env.selection.Options
	assert.Equal(t, sandbox.JobName(id), options["environment"], "the environment is named as serve's key checks expect")
	assert.Equal(t, hosted.base, options["ref"])
	assert.Equal(t, sandbox.RunBranch(hosted.todo, id), options["branch"])
	assert.Equal(t, []string{"gavel", "runner", "run-task"}, options["command"])
	assert.Equal(t, "env.example:443", options["envAPI"], "the backend's own options stay")
	runEnv := options["env"].(map[string]string)
	assert.NotContains(t, runEnv, "GAVEL_RUN_TOKEN", "the environment holds no token: the caller acts on the run's PR for it")
	assert.NotContains(t, runEnv, "GAVEL_SERVER")
	assert.Equal(t, `{"run":"`+id.String()+`"}`, string(env.prStatus), "the PR call went to serve's run API, as the run")

	assert.Equal(t, hosted.gitDir, env.in.Repo, "the run starts from serve's git directory of the repository")
	assert.Equal(t, &api.SandboxRef{Mode: api.SandboxSubstrate}, env.in.Resolved.Spec.Sandbox)
	assert.Equal(t, []string{hosted.org.String() + " claude/-tmp-captain-run-src/session-1.jsonl " + `{"type":"user"}` + "\n"}, hosted.transcripts)
	stored, err := hosted.logs.Read(t.Context(), "runs/"+id.String()+"/log.txt")
	require.NoError(t, err)
	defer stored.Close()
	data, err := io.ReadAll(stored)
	require.NoError(t, err)
	assert.Equal(t, "the environment's output\n", string(data))

	branch := sandbox.RunBranch(hosted.todo, id)
	assert.Equal(t, result.Response.Workspace.Worktree.Head, gitRun(t, hosted.gitDir, "rev-parse", "refs/heads/"+branch), "the run's branch is the repository's")
}

func TestSubstrateRunnerCreatesTheRunsEnvironmentFromItsOrganizationsTemplate(t *testing.T) {
	env := &environment{t: t}
	hosted := newHostedRun(t, env)
	server := substratetest.Start(t)
	server.AddTemplate("ate-env", "gavel-runner", substrate.Size{CPUMillicores: 2000, MemoryBytes: 4 << 30})
	client, err := substrate.Dial(server.Options())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	templates := &sandbox.OrgTemplates{DB: hosted.runner.Dispatcher.DB, Client: client, Atespace: "ate-env", Base: "gavel-runner"}
	hosted.runner.Dispatcher.Templates = templates
	hosted.runner.Backend["template"] = "gavel-runner"
	want, err := templates.Resolve(t.Context(), hosted.org)
	require.NoError(t, err)

	_, err = hosted.runner.Run(t.Context(), step())
	require.NoError(t, err)
	assert.Equal(t, want.Name, env.selection.Options["template"], "the organization's template replaces the backend's")
	assert.Equal(t, "gavel-runner", hosted.runner.Backend["template"], "the backend shared by every run keeps its own")
}

// budgetedRun is a hosted run whose organization's runs come from its own template and are
// admitted within its budgets, with a model cost budget of 5 USD this month.
func budgetedRun(t *testing.T, env *environment) *hostedRun {
	t.Helper()
	hosted := newHostedRun(t, env)
	server := substratetest.Start(t)
	server.AddTemplate("ate-env", "gavel-runner", substrate.Size{CPUMillicores: 2000, MemoryBytes: 4 << 30})
	client, err := substrate.Dial(server.Options())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	d := hosted.runner.Dispatcher
	d.Templates = &sandbox.OrgTemplates{DB: d.DB, Client: client, Atespace: "ate-env", Base: "gavel-runner"}
	d.Admission = &sandbox.Admission{Budgets: budget.Store{DB: d.DB}, RunLifetime: d.RunLifetime(), Wait: time.Minute, Recheck: time.Minute}
	_, err = d.Admission.Budgets.Set(t.Context(), hosted.org, nil, budget.Budget{Resource: budget.USD, Kind: budget.Cumulative, Amount: 5, Window: "now/M"})
	require.NoError(t, err)
	return hosted
}

func TestSubstrateRunnerCapsARunsSpendAtWhatItsModelCostBudgetLeaves(t *testing.T) {
	env := &environment{t: t}
	hosted := budgetedRun(t, env)

	run := step()
	_, err := hosted.runner.Run(t.Context(), run)
	require.NoError(t, err)
	assert.Equal(t, 5.0, env.in.Resolved.Spec.Budget.Cost, "a run's spend stops at what the organization's budget leaves")

	run = step()
	run.Input.Resolved.Spec.Budget.Cost = 2
	_, err = hosted.runner.Run(t.Context(), run)
	require.NoError(t, err)
	assert.Equal(t, 2.0, env.in.Resolved.Spec.Budget.Cost, "a step's own lower ceiling stays")
}

func TestSubstrateRunnerStopsARunOnceItsModelCostBudgetIsUsedUp(t *testing.T) {
	env := &environment{t: t, holds: true}
	hosted := budgetedRun(t, env)
	db := hosted.runner.Dispatcher.DB
	keep := hosted.runner.Transcripts
	// Ingesting the run's transcript files the model call that uses the budget up.
	hosted.runner.Transcripts = func(ctx context.Context, run sandbox.RunRecord, source, name string, data io.Reader) error {
		session, turn := uuid.New(), uuid.New()
		require.NoError(t, db.Exec(`INSERT INTO captain_sessions (id, source) VALUES (?, 'test')`, session).Error)
		require.NoError(t, db.Exec(`INSERT INTO session_owners (session_id, org_id) VALUES (?, ?)`, session, run.Org).Error)
		require.NoError(t, db.Exec(`INSERT INTO captain_turns (id, session_id, turn_index, status) VALUES (?, ?, 0, 'ended')`, turn, session).Error)
		require.NoError(t, db.Exec(`INSERT INTO captain_model_calls (id, turn_id, call_index, model, currency, input_tokens, input_cost, created_at)
			VALUES (?, ?, 0, 'claude-sonnet', 'USD', 1000, 6, now())`, uuid.New(), turn).Error)
		return keep(ctx, run, source, name, data)
	}

	_, err := hosted.runner.Run(t.Context(), step())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "budget used up: this organization has used $6.00 of its $5.00 model cost budget")
	var row struct{ Status, StatusMessage string }
	require.NoError(t, db.Raw(`SELECT status, status_message FROM sandbox_runs WHERE org_id = ?`, hosted.org).Scan(&row).Error)
	assert.Equal(t, "cancelled", row.Status)
	assert.Contains(t, row.StatusMessage, "stopped: budget used up")
}

func TestSubstrateRunnerBoundsARunByTheDispatchersTimeout(t *testing.T) {
	env := &environment{t: t}
	hosted := newHostedRun(t, env)
	hosted.runner.Dispatcher.Limits.Timeout = 10 * time.Minute
	run := step()
	run.Input.Timeout = 30 * time.Minute

	_, err := hosted.runner.Run(t.Context(), run)
	require.NoError(t, err)
	require.False(t, env.deadline.IsZero(), "the run has a deadline")
	assert.LessOrEqual(t, time.Until(env.deadline), 10*time.Minute, "the run ends within the dispatcher's timeout, not its step's longer budget")
	assert.Equal(t, hosted.runner.Dispatcher.RunLifetime().String(), env.selection.Options["maxLifetime"], "its environment outlives it")
}

func TestSubstrateRunnerRecordsAFailedRun(t *testing.T) {
	env := &environment{t: t, fail: errors.New("the agent gave up")}
	hosted := newHostedRun(t, env)

	_, err := hosted.runner.Run(t.Context(), step())
	require.EqualError(t, err, "the agent gave up")
	var status string
	require.NoError(t, hosted.runner.Dispatcher.DB.Raw(`SELECT status FROM sandbox_runs WHERE org_id = ?`, hosted.org).Scan(&status).Error)
	assert.Equal(t, "failed", status)
}

// runAPIFor stands in for serve's run API: it answers a run's PR status with the run its token is
// for.
func runAPIFor(t *testing.T, key []byte) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/runs/{id}/pr/status", func(w http.ResponseWriter, r *http.Request) {
		run, _, err := sandbox.VerifyRunToken(key, strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), time.Now())
		require.NoError(t, err)
		assert.Equal(t, r.PathValue("id"), run.String())
		_, _ = w.Write([]byte(`{"run":"` + run.String() + `"}`))
	})
	return mux
}
