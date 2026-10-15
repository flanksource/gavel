// Tests the run API that sandbox jobs call: only a valid run token for the run in the path is
// accepted, the spec comes back, and results are stored as a run of the job's repository.
package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	ghcache "github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/sandbox"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunsAPI(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_runs_api"}).DSN())
	t.Setenv(database.EnvDisable, "")
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	db := opened.Gorm()
	cache, err := ghcache.Open()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, cache.Close()) })

	org := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)
	var repo struct{ ID uuid.UUID }
	require.NoError(t, db.Raw(`INSERT INTO repos (org_id, name, source, github_repo_id) VALUES (?, 'acme/app', 'github', 1) RETURNING id`, org).Scan(&repo).Error)
	run, otherRun := uuid.New(), uuid.New()
	spec := sandbox.RunSpec{Kind: "test", Args: []string{"./..."}, RepoID: repo.ID, Ref: "refs/heads/main", SHA: "abc123"}
	encoded, err := json.Marshal(spec)
	require.NoError(t, err)
	for _, id := range []uuid.UUID{run, otherRun} {
		require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, repo_id, kind, status, backend, spec) VALUES (?, ?, ?, 'test', 'running', 'memory', ?)`,
			id, org, repo.ID, string(encoded)).Error)
	}

	key := []byte(strings.Repeat("k", 32))
	api := &RunsAPI{DB: db, TokenKey: key, Cache: cache}
	server := httptest.NewServer(api.Handler())
	t.Cleanup(server.Close)
	token := func(id uuid.UUID, exp time.Time) string { return sandbox.SignRunToken(key, id, org, exp) }
	valid := token(run, time.Now().Add(time.Hour))
	call := func(method, path, bearer string, body []byte) (int, string) {
		t.Helper()
		req, err := http.NewRequest(method, server.URL+path, bytes.NewReader(body))
		require.NoError(t, err)
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(data)
	}

	t.Run("spec", func(t *testing.T) {
		status, body := call(http.MethodGet, "/api/runs/"+run.String()+"/spec", valid, nil)
		require.Equal(t, http.StatusOK, status, body)
		var got sandbox.RunSpec
		require.NoError(t, json.Unmarshal([]byte(body), &got))
		assert.Equal(t, spec, got)
	})

	t.Run("results", func(t *testing.T) {
		snapshot := []byte(`{"metadata":{"started":"2026-10-15T09:00:00Z"},"tests":[{"name":"TestA","failed":true}]}`)
		status, body := call(http.MethodPost, "/api/runs/"+run.String()+"/results", valid, snapshot)
		require.Equal(t, http.StatusCreated, status, body)
		assert.Contains(t, body, "run-2026-10-15T09-00-00Z")
		row, err := cache.ForOrg(org).GetTestRun(context.Background(), RepoWorkspace(repo.ID), "run-2026-10-15T09-00-00Z")
		require.NoError(t, err)
		require.NotNil(t, row, "stored as a run of the job's repository")
		assert.Equal(t, "abc123", row.SHA)
		assert.Equal(t, 1, row.Failed)
	})

	for name, tc := range map[string]struct {
		bearer string
		path   string
		status int
		want   string
	}{
		"no token":            {"", "/api/runs/" + run.String() + "/spec", http.StatusUnauthorized, "run token"},
		"expired token":       {token(run, time.Now().Add(-time.Minute)), "/api/runs/" + run.String() + "/spec", http.StatusUnauthorized, "run token expired"},
		"another run's token": {token(otherRun, time.Now().Add(time.Hour)), "/api/runs/" + run.String() + "/spec", http.StatusForbidden, "this run token is for run " + otherRun.String()},
		"unknown run":         {token(uuid.Nil, time.Now().Add(time.Hour)), "/api/runs/" + uuid.Nil.String() + "/spec", http.StatusNotFound, "sandbox run not found"},
		"not a run id":        {valid, "/api/runs/abc/spec", http.StatusBadRequest, "is not a run id"},
	} {
		t.Run(name, func(t *testing.T) {
			status, body := call(http.MethodGet, tc.path, tc.bearer, nil)
			assert.Equal(t, tc.status, status, body)
			assert.Contains(t, body, tc.want)
		})
	}
}
