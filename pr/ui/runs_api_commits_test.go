// Tests the run API's commits endpoint: an agent run's job uploads the commits its run made, and
// serve puts them on the run's own branch of the repository.
package ui

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/sandbox"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunsAPIPushesAnAgentRunsCommits(t *testing.T) {
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_runs_api_commits"}).DSN())
	t.Setenv(database.EnvDisable, "")
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	db := opened.Gorm()
	org := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)
	dataDir := t.TempDir()
	repo := pushedRepo(t, db, org, dataDir, "app")
	bare := filepath.Join(dataDir, "repos", org.String(), repo.String()+".git")
	base := gitOutput(t, bare, "rev-parse", "HEAD")

	todo, agentRun, testRun := uuid.New(), uuid.New(), uuid.New()
	for id, kind := range map[uuid.UUID]string{agentRun: "agent", testRun: "test"} {
		spec, err := json.Marshal(sandbox.RunSpec{Kind: kind, RepoID: repo, Ref: "HEAD", SHA: base, TodoID: &todo})
		require.NoError(t, err)
		require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, repo_id, kind, status, backend, spec) VALUES (?, ?, ?, ?, 'running', 'memory', ?)`, id, org, repo, kind, string(spec)).Error)
	}

	// The run's commits, on the branch its worktree made.
	work := t.TempDir()
	runGit(t, work, "clone", "-q", bare, ".")
	runGit(t, work, "checkout", "-q", "-b", "worktree-run")
	require.NoError(t, os.WriteFile(filepath.Join(work, "fix.txt"), []byte("fixed"), 0o644))
	runGit(t, work, "add", "fix.txt")
	runGit(t, work, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "fix")
	tip := gitOutput(t, work, "rev-parse", "HEAD")
	runGit(t, work, "bundle", "create", filepath.Join(work, "run.bundle"), "worktree-run", "^"+base)
	bundle, err := os.ReadFile(filepath.Join(work, "run.bundle"))
	require.NoError(t, err)

	key := []byte(strings.Repeat("k", 32))
	server := httptest.NewServer((&RunsAPI{DB: db, TokenKey: key, Sources: &sandbox.Sources{DataDir: dataDir}}).Handler())
	t.Cleanup(server.Close)
	post := func(run uuid.UUID, body []byte) (int, string) {
		t.Helper()
		req, err := http.NewRequest(http.MethodPost, server.URL+"/api/runs/"+run.String()+"/commits", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+sandbox.SignRunToken(key, run, org, time.Now().Add(time.Hour)))
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		return resp.StatusCode, string(data)
	}

	status, body := post(agentRun, bundle)
	require.Equal(t, http.StatusCreated, status, body)
	branch := "gavel/" + todo.String()[:8] + "-" + agentRun.String()[:8]
	assert.JSONEq(t, `{"branch":"`+branch+`","sha":"`+tip+`"}`, body)
	assert.Equal(t, tip, gitOutput(t, bare, "rev-parse", "refs/heads/"+branch))

	status, body = post(agentRun, bundle)
	assert.Equal(t, http.StatusConflict, status, "a run's branch is created once")
	assert.Contains(t, body, "create branch "+branch+" of app")

	status, body = post(testRun, bundle)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, body, "run "+testRun.String()+" is a test run; only agent runs push commits")
}

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	require.NoError(t, err, "git %v", args)
	return strings.TrimSpace(string(out))
}
