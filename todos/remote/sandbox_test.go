// Tests SandboxJobs: each remote run's job starts as an agent run of the todo's repository, at the
// commit serve resolved it from, under the run's ID, and can be waited for and cancelled.
package remote

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/sandbox"
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

func TestSandboxJobsStartsAnAgentRunOfTheTodosRepository(t *testing.T) {
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_remote_sandbox"}).DSN())
	t.Setenv(database.EnvDisable, "")
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	db := opened.Gorm()
	org := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)
	var repo struct{ ID uuid.UUID }
	require.NoError(t, db.Raw(`INSERT INTO repos (org_id, name, source) VALUES (?, 'app', 'push') RETURNING id`, org).Scan(&repo).Error)
	executor := &exitingExecutor{exitCode: 3}
	logs := blob.Local(t.TempDir())
	dispatcher := &sandbox.Dispatcher{
		DB: db, Executor: executor, Image: "runner", Server: "https://gavel.example", TokenKey: []byte(strings.Repeat("k", 32)),
		Logs: func(uuid.UUID) blob.Store { return logs }, Limits: sandbox.Limits{Timeout: time.Minute},
	}
	todo, id := uuid.New(), uuid.New()
	jobs := SandboxJobs{Dispatcher: dispatcher, Scope: tenant.Scope{OrgID: org}, RepoID: repo.ID, SHA: "abc123", TodoID: todo}

	started, err := jobs.Start(context.Background(), id, Job{Root: "/srv/repo"})
	require.NoError(t, err)

	record, err := sandbox.LoadRun(context.Background(), db, id)
	require.NoError(t, err)
	assert.Equal(t, "agent", record.Spec.Kind)
	assert.Equal(t, repo.ID, record.Spec.RepoID)
	assert.Equal(t, "HEAD", record.Spec.Ref)
	assert.Equal(t, "abc123", record.Spec.SHA)
	require.NotNil(t, record.Spec.TodoID)
	assert.Equal(t, todo, *record.Spec.TodoID)
	var job Job
	require.NoError(t, json.Unmarshal(record.Spec.Agent, &job))
	assert.Equal(t, "/srv/repo", job.Root, "the run's spec carries its job")

	require.NoError(t, started.Cancel(context.Background()))
	assert.Equal(t, 1, executor.cancelled)
	code, err := started.Wait(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 3, code)
}

func TestSandboxJobsStartsAPRRunFromThePRsHead(t *testing.T) {
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_remote_sandbox_pr"}).DSN())
	t.Setenv(database.EnvDisable, "")
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	db := opened.Gorm()
	org := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)
	logs := blob.Local(t.TempDir())
	dispatcher := &sandbox.Dispatcher{
		DB: db, Executor: &exitingExecutor{}, Image: "runner", Server: "https://gavel.example", TokenKey: []byte(strings.Repeat("k", 32)),
		Logs: func(uuid.UUID) blob.Store { return logs }, Limits: sandbox.Limits{Timeout: time.Minute},
	}
	id := uuid.New()
	head := strings.Repeat("c", 40)
	job := Job{Hooks: lifecycle.RunHooks{PR: &lifecycle.PRHooks{URL: "https://github.com/acme/app/pull/7", Number: 7, Target: provider.RunTarget{HeadSHA: head}}}}

	_, err = SandboxJobs{Dispatcher: dispatcher, Scope: tenant.Scope{OrgID: org}, SHA: "default-branch", TodoID: uuid.New()}.Start(context.Background(), id, job)
	require.NoError(t, err)

	record, err := sandbox.LoadRun(context.Background(), db, id)
	require.NoError(t, err)
	assert.Equal(t, "refs/pull/7/head", record.Spec.Ref)
	assert.Equal(t, head, record.Spec.SHA, "the run starts from the PR's head, not the default branch")
	assert.Equal(t, "https://github.com/acme/app/pull/7", record.Spec.PR)
}
