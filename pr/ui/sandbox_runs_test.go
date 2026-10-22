// Tests /api/sandbox-runs: a hosted organization's sandbox runs as its members may see them, and
// each run's stored log.
package ui

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hostedGet fetches path as user, returning the status and the body.
func hostedGet(t *testing.T, base, path string, user uuid.UUID) (int, string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, base+path, nil)
	require.NoError(t, err)
	req.Header.Set("X-Test-User", user.String())
	req.Header.Set("X-Test-Role", "org:member")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(body)
}

func TestSandboxRunsListTheRunsAMemberMaySee(t *testing.T) {
	h := newHostedRuns(t, nil)
	db := h.db
	require.NoError(t, db.Exec(`UPDATE users SET name = 'Ada Lovelace' WHERE id = ?`, h.ada).Error)
	require.NoError(t, db.Exec(`UPDATE users SET name = 'Grace Hopper' WHERE id = ?`, h.grace).Error)
	repo := strings.TrimPrefix(h.workspace, "repo:")
	_, private := hostedCall(t, h.base, http.MethodPost, "/api/todos", h.grace, "org:member", map[string]any{"dir": h.workspace, "title": "Grace's private todo"})
	_, public := hostedCall(t, h.base, http.MethodPost, "/api/todos", h.grace, "org:member", map[string]any{"dir": h.workspace, "title": "Shared", "visibility": "public"})
	run := func(user uuid.UUID, kind, todo, status string, minutesAgo int) uuid.UUID {
		id := uuid.New()
		var todoID any
		if todo != "" {
			todoID = todo
		}
		require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, user_id, repo_id, kind, ref, sha, todo_id, status, exit_code, backend, log_key, spec, created_at, started_at, finished_at)
			VALUES (?, ?, ?, ?, ?, 'refs/heads/main', 'abc1234', ?, ?, 3, 'substrate', ?, '{}', now() - make_interval(mins => ?), now() - make_interval(mins => ?), now())`,
			id, h.org, user, repo, kind, todoID, status, "runs/"+id.String()+"/log.txt", minutesAgo, minutesAgo).Error)
		return id
	}
	pushCheck := run(h.ada, "push-check", "", "failed", 3)
	shared := run(h.grace, "agent", public["id"].(string), "succeeded", 2)
	hidden := run(h.grace, "agent", private["id"].(string), "succeeded", 1)
	require.NoError(t, h.logs.Write(context.Background(), "runs/"+pushCheck.String()+"/log.txt", strings.NewReader("checking\nexit 3\n")))
	require.NoError(t, h.logs.Write(context.Background(), "runs/"+hidden.String()+"/log.txt", strings.NewReader("grace's run\n")))

	otherOrg := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Other')`, otherOrg, otherOrg.String()).Error)
	elsewhere := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, spec) VALUES (?, ?, 'push-check', 'failed', 'substrate', '{}')`, elsewhere, otherOrg).Error)

	status, listed := hostedCall(t, h.base, http.MethodGet, "/api/sandbox-runs", h.ada, "org:member", nil)
	require.Equal(t, http.StatusOK, status, listed)
	items, _ := listed["items"].([]any)
	var ids []string
	for _, item := range items {
		ids = append(ids, item.(map[string]any)["id"].(string))
	}
	assert.Equal(t, []string{shared.String(), pushCheck.String()}, ids, "newest first, without another member's private todo's run or another org's")
	first := items[1].(map[string]any)
	assert.Equal(t, "push-check", first["kind"])
	assert.Equal(t, "refs/heads/main", first["ref"])
	assert.Equal(t, "abc1234", first["sha"])
	assert.Equal(t, "failed", first["status"])
	assert.EqualValues(t, 3, first["exitCode"])
	assert.Equal(t, "app", first["repo"])
	assert.Equal(t, "Ada Lovelace", first["user"])
	assert.Equal(t, "substrate", first["backend"])
	assert.Equal(t, true, first["hasLog"])
	assert.NotEmpty(t, first["createdAt"])
	assert.NotEmpty(t, first["finishedAt"])
	second := items[0].(map[string]any)
	assert.Equal(t, public["id"], second["todoId"])
	assert.Equal(t, "Shared", second["todoTitle"])

	status, log := hostedGet(t, h.base, "/api/sandbox-runs/"+pushCheck.String()+"/log", h.ada)
	require.Equal(t, http.StatusOK, status, log)
	assert.Equal(t, "checking\nexit 3\n", log)
	status, _ = hostedGet(t, h.base, "/api/sandbox-runs/"+shared.String()+"/log", h.ada)
	assert.Equal(t, http.StatusNotFound, status, "a run with no stored log")
	status, refused := hostedGet(t, h.base, "/api/sandbox-runs/"+hidden.String()+"/log", h.ada)
	assert.Equal(t, http.StatusNotFound, status, "another member's private todo's run: %s", refused)
	status, _ = hostedGet(t, h.base, "/api/sandbox-runs/"+elsewhere.String()+"/log", h.ada)
	assert.Equal(t, http.StatusNotFound, status, "another org's run")
	status, log = hostedGet(t, h.base, "/api/sandbox-runs/"+hidden.String()+"/log", h.grace)
	require.Equal(t, http.StatusOK, status, log)
	assert.Equal(t, "grace's run\n", log, "its creator reads it")
}
