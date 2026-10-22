// Tests push checks: every updated branch of a push gets a push-check run whose log and verdict
// reach the pushing SSH session; deleted branches and tags get none.
package main

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/serve"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// scriptedExecutor plays back the same output and a zero exit code for every job.
type scriptedExecutor struct {
	jobs []sandbox.Job
}

func (e *scriptedExecutor) Name() string { return "scripted" }
func (e *scriptedExecutor) Submit(_ context.Context, job sandbox.Job) (sandbox.Handle, error) {
	e.jobs = append(e.jobs, job)
	return &scriptedHandle{ref: "scripted-" + job.RunID.String()}, nil
}

type scriptedHandle struct{ ref string }

func (h *scriptedHandle) Logs(context.Context) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("running checks\n")), nil
}
func (h *scriptedHandle) Wait(context.Context) (sandbox.Result, error) { return sandbox.Result{}, nil }
func (h *scriptedHandle) Cancel(context.Context) error                 { return nil }
func (h *scriptedHandle) Ref() string                                  { return h.ref }

func TestPushChecksRunEachUpdatedBranch(t *testing.T) {
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_push_checks"}).DSN())
	t.Setenv(database.EnvDisable, "")
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	db := opened.Gorm()
	org := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)
	var repo struct{ ID uuid.UUID }
	require.NoError(t, db.Raw(`INSERT INTO repos (org_id, name, source) VALUES (?, 'app', 'push') RETURNING id`, org).Scan(&repo).Error)

	user := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', 'user_ada')`, user).Error)

	executor := &scriptedExecutor{}
	logs := blob.Local(t.TempDir())
	dispatcher := &sandbox.Dispatcher{
		DB: db, Executor: executor, Image: "runner", Server: "https://gavel.example", TokenKey: []byte(strings.Repeat("k", 32)),
		Logs: func(o uuid.UUID) blob.Store { return blob.WithPrefix(logs, "orgs/"+o.String()+"/") },
	}
	var out bytes.Buffer
	pushChecks(dispatcher)(context.Background(), serve.Push{
		OrgID: org, RepoID: repo.ID, UserID: user, Out: &out,
		Refs: []serve.RefUpdate{
			{Name: "refs/heads/feature", Old: "", New: "aaa111"},
			{Name: "refs/heads/gone", Old: "bbb222", New: ""},
			{Name: "refs/heads/main", Old: "ccc333", New: "ddd444"},
			{Name: "refs/tags/v1", Old: "", New: "eee555"},
		},
	})

	require.Len(t, executor.jobs, 2, "one run per updated branch: %s", out.String())
	var kinds []string
	require.NoError(t, db.Raw(`SELECT ref || '@' || sha FROM sandbox_runs WHERE kind = 'push-check' ORDER BY ref`).Scan(&kinds).Error)
	assert.Equal(t, []string{"refs/heads/feature@aaa111", "refs/heads/main@ddd444"}, kinds)
	text := out.String()
	assert.Contains(t, text, "running checks")
	assert.Contains(t, text, "feature: push check PASSED")
	assert.Contains(t, text, "main: push check PASSED")
	assert.NotContains(t, text, "gone")
	assert.NotContains(t, text, "v1")
}

func TestPushChecksWithoutASandbox(t *testing.T) {
	var out bytes.Buffer
	pushChecks(nil)(context.Background(), serve.Push{Out: &out, Refs: []serve.RefUpdate{{Name: "refs/heads/main", New: "abc"}}})
	assert.Equal(t, "push received; push checks need a sandbox: start gavel serve with --sandbox\n", out.String())
}
