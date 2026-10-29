// Tests the dispatcher with an in-memory executor: a run's record moves from queued to running to
// its outcome, its log is stored and streamed, and the job gets the server, a run token and the
// runner command.
package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/commons-db/secret"
	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/modelkeys"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// memoryExecutor runs nothing: each job prints output and exits with exitCode.
type memoryExecutor struct {
	output    string
	exitCode  int
	submitErr error
	waitErr   error
	submitted []Job
	cancelled []string
}

func (e *memoryExecutor) Name() string { return "memory" }

func (e *memoryExecutor) Submit(_ context.Context, job Job) (Handle, error) {
	if e.submitErr != nil {
		return nil, e.submitErr
	}
	e.submitted = append(e.submitted, job)
	return &memoryHandle{executor: e, ref: "memory-" + job.RunID.String()}, nil
}

type memoryHandle struct {
	executor *memoryExecutor
	ref      string
}

func (h *memoryHandle) Logs(context.Context) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(h.executor.output)), nil
}
func (h *memoryHandle) Wait(context.Context) (Result, error) {
	return Result{ExitCode: h.executor.exitCode}, h.executor.waitErr
}
func (h *memoryHandle) Cancel(context.Context) error {
	h.executor.cancelled = append(h.executor.cancelled, h.ref)
	return nil
}
func (h *memoryHandle) Ref() string { return h.ref }

type sandboxRunRow struct {
	Status     string
	ExitCode   *int
	Backend    string
	BackendRef string
	LogKey     string
	Kind       string
	Ref        string
	SHA        string
	StartedAt  *time.Time
	FinishedAt *time.Time
}

func readRun(t *testing.T, db *gorm.DB, id uuid.UUID) sandboxRunRow {
	t.Helper()
	var row sandboxRunRow
	require.NoError(t, db.Raw(`SELECT status, exit_code, backend, backend_ref, log_key, kind, ref, sha, started_at, finished_at FROM sandbox_runs WHERE id = ?`, id).Scan(&row).Error)
	return row
}

func newDispatcher(t *testing.T, executor Executor) (*Dispatcher, *gorm.DB, blob.Store, uuid.UUID) {
	t.Helper()
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_sandbox_dispatch"}).DSN())
	t.Setenv(database.EnvDisable, "")
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	db := opened.Gorm()
	org := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)
	logs := blob.Local(t.TempDir())
	return &Dispatcher{
		DB:       db,
		Executor: executor,
		Logs:     func(o uuid.UUID) blob.Store { return blob.WithPrefix(logs, "orgs/"+o.String()+"/") },
		Image:    "ghcr.io/flanksource/gavel-runner:1",
		Server:   "https://gavel.example",
		TokenKey: []byte(strings.Repeat("k", 32)),
		Limits:   Limits{Timeout: 30 * time.Minute},
	}, db, logs, org
}

func TestDispatcherRunsAJobToItsOutcome(t *testing.T) {
	executor := &memoryExecutor{output: "go test ./...\nFAIL\n", exitCode: 1}
	dispatcher, db, logs, org := newDispatcher(t, executor)
	ctx := context.Background()
	var repoRow struct{ ID uuid.UUID }
	require.NoError(t, db.Raw(`INSERT INTO repos (org_id, name, source, github_repo_id) VALUES (?, 'acme/app', 'github', 1) RETURNING id`, org).Scan(&repoRow).Error)
	repo := repoRow.ID
	var streamed bytes.Buffer

	run, err := dispatcher.Start(ctx, tenant.Scope{OrgID: org}, RunSpec{Kind: "test", Args: []string{"./..."}, RepoID: repo, Ref: "refs/heads/main", SHA: "abc123"}, &streamed)
	require.NoError(t, err)
	result, err := run.Wait(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, result.ExitCode)

	row := readRun(t, db, run.ID)
	assert.Equal(t, "failed", row.Status)
	require.NotNil(t, row.ExitCode)
	assert.Equal(t, 1, *row.ExitCode)
	assert.Equal(t, "memory", row.Backend)
	assert.Equal(t, "memory-"+run.ID.String(), row.BackendRef)
	assert.Equal(t, "test", row.Kind)
	assert.Equal(t, "refs/heads/main", row.Ref)
	assert.Equal(t, "abc123", row.SHA)
	assert.NotNil(t, row.StartedAt)
	assert.NotNil(t, row.FinishedAt)
	assert.Equal(t, "runs/"+run.ID.String()+"/log.txt", row.LogKey)

	reader, err := logs.Read(ctx, "orgs/"+org.String()+"/"+row.LogKey)
	require.NoError(t, err)
	stored, err := io.ReadAll(reader)
	require.NoError(t, reader.Close())
	require.NoError(t, err)
	assert.Equal(t, "go test ./...\nFAIL\n", string(stored))
	assert.Equal(t, "go test ./...\nFAIL\n", streamed.String(), "the log is also streamed to the caller")

	record, err := LoadRun(ctx, db, run.ID)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), record.CreatedAt, time.Minute)
	record.CreatedAt = time.Time{}
	assert.Equal(t, RunRecord{ID: run.ID, Org: org, Status: "failed", Spec: RunSpec{Kind: "test", Args: []string{"./..."}, RepoID: repo, Ref: "refs/heads/main", SHA: "abc123"}}, record)
	_, err = LoadRun(ctx, db, uuid.New())
	assert.ErrorIs(t, err, ErrRunNotFound)

	require.Len(t, executor.submitted, 1)
	job := executor.submitted[0]
	assert.Equal(t, run.ID, job.RunID)
	assert.Equal(t, org, job.Scope.OrgID)
	assert.Equal(t, "ghcr.io/flanksource/gavel-runner:1", job.Image)
	assert.Equal(t, []string{"gavel", "runner", "exec", "--run", run.ID.String()}, job.Command)
	assert.Equal(t, map[string]string{"GAVEL_SERVER": "https://gavel.example", "GAVEL_RUN_ID": run.ID.String(), "GAVEL_DB": "off", "CI": "1"}, job.Env)
	tokenRun, tokenOrg, err := VerifyRunToken(dispatcher.TokenKey, job.SecretEnv["GAVEL_RUN_TOKEN"], time.Now().Add(44*time.Minute))
	require.NoError(t, err, "the token outlives the timeout by 15 minutes")
	assert.Equal(t, run.ID, tokenRun)
	assert.Equal(t, org, tokenOrg)
	_, _, err = VerifyRunToken(dispatcher.TokenKey, job.SecretEnv["GAVEL_RUN_TOKEN"], time.Now().Add(46*time.Minute))
	assert.Error(t, err, "and no longer")
}

func TestDispatcherRunLifetimeIsTheTimeoutPlusTheRunTokensGrace(t *testing.T) {
	assert.Equal(t, 45*time.Minute, (&Dispatcher{Limits: Limits{Timeout: 30 * time.Minute}}).RunLifetime())
	assert.Equal(t, 75*time.Minute, (&Dispatcher{}).RunLifetime(), "the default 60m timeout")
}

// withModelKeys gives the dispatcher org keys and per-run credential URIs, as --sandbox
// substrate does, with commons-db's local secret keeper.
func withModelKeys(t *testing.T, dispatcher *Dispatcher) {
	t.Helper()
	previous := secret.KMSConnection
	secret.KMSConnection = modelkeys.KeeperConnection([]byte(strings.Repeat("s", 32)))
	secret.InvalidateKeeper()
	t.Cleanup(func() { secret.KMSConnection = previous; secret.InvalidateKeeper() })
	dispatcher.Keys = modelkeys.Store{DB: dispatcher.DB}
	dispatcher.CredentialURI = func(run uuid.UUID, provider string) string {
		return "ate-secret://gavel/beta/" + run.String() + "/" + provider
	}
}

func TestDispatcherGivesEachRunItsOrganizationsKeys(t *testing.T) {
	executor := &memoryExecutor{output: "ok\n"}
	dispatcher, db, _, org := newDispatcher(t, executor)
	withModelKeys(t, dispatcher)
	require.NoError(t, dispatcher.Keys.Set(t.Context(), org, nil, "claude-code-oauth-token", "sk-ant-oat01-dispatch-test"))

	run, err := dispatcher.Start(t.Context(), tenant.Scope{OrgID: org}, RunSpec{Kind: "agent"}, nil)
	require.NoError(t, err)
	_, err = run.Wait(t.Context())
	require.NoError(t, err)
	require.Len(t, executor.submitted, 1)
	assert.Equal(t, []InjectedCredential{{
		Host: "api.anthropic.com", Header: "Authorization", Prefix: "Bearer ", Env: "CLAUDE_CODE_OAUTH_TOKEN",
		URI: "ate-secret://gavel/beta/" + run.ID.String() + "/anthropic",
	}}, executor.submitted[0].Credentials)
	assert.NotContains(t, executor.submitted[0].SecretEnv, "CLAUDE_CODE_OAUTH_TOKEN", "the key itself never reaches the job")

	other := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Other')`, other, other.String()).Error)
	run, err = dispatcher.Start(t.Context(), tenant.Scope{OrgID: other}, RunSpec{Kind: "test"}, nil)
	require.NoError(t, err, "a test run starts without a model key")
	_, err = run.Wait(t.Context())
	require.NoError(t, err)
	require.Len(t, executor.submitted, 2)
	assert.Empty(t, executor.submitted[1].Credentials, "another org's keys aren't used")
}

func TestDispatcherRefusesAnAgentRunWithoutAModelKey(t *testing.T) {
	executor := &memoryExecutor{output: "ok\n"}
	dispatcher, db, _, org := newDispatcher(t, executor)
	withModelKeys(t, dispatcher)

	_, err := dispatcher.Start(t.Context(), tenant.Scope{OrgID: org}, RunSpec{Kind: "agent"}, nil)
	require.ErrorIs(t, err, ErrNoModelKey)
	assert.EqualError(t, err, "this organization has no model key: an organization admin adds one under Model keys in the organization's settings")
	assert.Empty(t, executor.submitted)
	var runs int64
	require.NoError(t, db.Raw(`SELECT count(*) FROM sandbox_runs WHERE org_id = ?`, org).Scan(&runs).Error)
	assert.Zero(t, runs, "a refused run isn't recorded")
}

func TestDispatcherWithoutCredentialURIsUsesNoOrganizationKeys(t *testing.T) {
	executor := &memoryExecutor{output: "ok\n"}
	dispatcher, _, _, org := newDispatcher(t, executor)
	run, err := dispatcher.Start(t.Context(), tenant.Scope{OrgID: org}, RunSpec{Kind: "agent"}, nil)
	require.NoError(t, err, "only substrate needs organization keys")
	_, err = run.Wait(t.Context())
	require.NoError(t, err)
	assert.Empty(t, executor.submitted[0].Credentials)
}

func TestDispatcherRecordsASuccess(t *testing.T) {
	dispatcher, db, _, org := newDispatcher(t, &memoryExecutor{output: "ok\n"})
	run, err := dispatcher.Start(context.Background(), tenant.Scope{OrgID: org}, RunSpec{Kind: "lint"}, nil)
	require.NoError(t, err)
	_, err = run.Wait(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "succeeded", readRun(t, db, run.ID).Status)
}

func TestDispatcherRecordsAJobThatCouldNotStart(t *testing.T) {
	dispatcher, db, _, org := newDispatcher(t, &memoryExecutor{submitErr: errors.New("no capacity")})
	run, err := dispatcher.Start(context.Background(), tenant.Scope{OrgID: org}, RunSpec{Kind: "test"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no capacity")
	require.NotNil(t, run, "the run's ID names its record")
	row := readRun(t, db, run.ID)
	assert.Equal(t, "failed", row.Status)
	assert.NotNil(t, row.FinishedAt)
}

func TestDispatcherRecordsAJobThatDidNotFinish(t *testing.T) {
	dispatcher, db, _, org := newDispatcher(t, &memoryExecutor{waitErr: errors.New("sandbox job timed out after 30m0s")})
	run, err := dispatcher.Start(context.Background(), tenant.Scope{OrgID: org}, RunSpec{Kind: "test"}, nil)
	require.NoError(t, err)
	_, err = run.Wait(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
	row := readRun(t, db, run.ID)
	assert.Equal(t, "failed", row.Status)
	assert.Nil(t, row.ExitCode)
}

func TestDispatcherRejectsAnUnknownKind(t *testing.T) {
	dispatcher, _, _, org := newDispatcher(t, &memoryExecutor{})
	_, err := dispatcher.Start(context.Background(), tenant.Scope{OrgID: org}, RunSpec{Kind: "deploy"}, nil)
	require.EqualError(t, err, `sandbox run kind "deploy" is not one of test, lint, push-check, agent`)
}

func TestDispatcherStartsAnAgentRunUnderTheCallersID(t *testing.T) {
	executor := &memoryExecutor{output: "ok\n"}
	dispatcher, _, _, org := newDispatcher(t, executor)
	dispatcher.SecretEnv = map[string]string{"ANTHROPIC_API_KEY": "model-key", "GAVEL_RUN_TOKEN": "forged"}
	id, todo := uuid.New(), uuid.New()
	job := json.RawMessage(`{"root":"/srv/repo"}`)

	run, err := dispatcher.StartRun(context.Background(), id, tenant.Scope{OrgID: org}, RunSpec{Kind: "agent", TodoID: &todo, Step: "run", Agent: job}, nil)
	require.NoError(t, err)
	assert.Equal(t, id, run.ID)
	require.Len(t, executor.submitted, 1)
	assert.Equal(t, "model-key", executor.submitted[0].SecretEnv["ANTHROPIC_API_KEY"])
	assert.NotContains(t, []string{"", "forged"}, executor.submitted[0].SecretEnv["GAVEL_RUN_TOKEN"], "a configured secret never replaces the run's token")

	record, err := LoadRun(context.Background(), dispatcher.DB, id)
	require.NoError(t, err)
	assert.JSONEq(t, string(job), string(record.Spec.Agent))
	assert.Equal(t, "run", record.Spec.Step)

	require.NoError(t, run.Cancel(context.Background()))
	assert.Equal(t, []string{"memory-" + id.String()}, executor.cancelled)
	_, err = run.Wait(context.Background())
	require.NoError(t, err)
}
