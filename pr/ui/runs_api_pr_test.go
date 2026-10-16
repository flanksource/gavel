// Tests the run API's PR endpoints: an agent run on a linked PR reads its status, resolves its
// threads and pushes its source branch through serve, for that PR alone.
package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/sandbox"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// scriptedPR is GitHub for one PR: its status and the threads resolved on it.
type scriptedPR struct {
	provider.Client
	snapshot model.StatusSnapshot
	resolved []string
}

func (p *scriptedPR) Status(_ context.Context, number int, _ model.StatusOptions) (*model.StatusSnapshot, error) {
	if number != p.snapshot.PR.Number {
		return nil, errors.New("no such PR")
	}
	snapshot := p.snapshot
	return &snapshot, nil
}

func (p *scriptedPR) ResolveThread(_ context.Context, _ int, thread string) error {
	p.resolved = append(p.resolved, thread)
	return nil
}

func TestRunsAPIProxiesARunsPR(t *testing.T) {
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_runs_api_pr"}).DSN())
	t.Setenv(database.EnvDisable, "")
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	db := opened.Gorm()
	org := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)

	// The PR's base repository, mirrored, and its source branch.
	dataDir, remotes := t.TempDir(), t.TempDir()
	work := t.TempDir()
	runGit(t, work, "init", "-q", "-b", "main")
	runGit(t, work, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "--allow-empty", "-m", "first")
	remote := filepath.Join(remotes, "acme", "app.git")
	runGit(t, work, "clone", "-q", "--bare", work, remote)
	runGit(t, remote, "branch", "topic", "main")
	base := gitOutput(t, remote, "rev-parse", "refs/heads/topic")
	require.NoError(t, db.Exec(`INSERT INTO github_installations (id, org_id, account_login, account_type) VALUES (7, ?, 'acme', 'Organization')`, org).Error)
	var repo struct{ ID uuid.UUID }
	require.NoError(t, db.Raw(`INSERT INTO repos (org_id, name, source, github_repo_id, installation_id) VALUES (?, 'acme/app', 'github', 42, 7) RETURNING id`, org).Scan(&repo).Error)

	const prURL = "https://github.com/acme/app/pull/7"
	prRun, otherRun := uuid.New(), uuid.New()
	todo := uuid.New()
	for id, pr := range map[uuid.UUID]string{prRun: prURL, otherRun: ""} {
		spec, err := json.Marshal(sandbox.RunSpec{Kind: "agent", RepoID: repo.ID, TodoID: &todo, PR: pr})
		require.NoError(t, err)
		require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, repo_id, kind, status, backend, spec) VALUES (?, ?, ?, 'agent', 'running', 'memory', ?)`, id, org, repo.ID, string(spec)).Error)
	}
	github := &scriptedPR{snapshot: model.StatusSnapshot{PR: &model.PRInfo{
		Number: 7, State: "OPEN", HeadRefOID: base, HeadRefName: "topic", HeadRepositoryURL: "file://" + remote,
	}}}
	key := []byte(strings.Repeat("k", 32))
	api := &RunsAPI{
		DB: db, TokenKey: key,
		Sources: &sandbox.Sources{DataDir: dataDir, GitHubURL: "file://" + remotes, Token: func(context.Context, int64) (string, error) { return "installation-token", nil }},
		PRs: func(_ context.Context, gotOrg uuid.UUID, url string) (provider.Client, int, error) {
			require.Equal(t, org, gotOrg)
			require.Equal(t, prURL, url)
			return github, 7, nil
		},
	}
	server := httptest.NewServer(api.Handler())
	t.Cleanup(server.Close)
	call := func(run uuid.UUID, method, path string, body []byte) (int, string) {
		t.Helper()
		req, err := http.NewRequest(method, server.URL+"/api/runs/"+run.String()+"/"+path, bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+sandbox.SignRunToken(key, run, org, time.Now().Add(time.Hour)))
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		return resp.StatusCode, string(data)
	}

	status, body := call(prRun, http.MethodGet, "pr/status?number=7&threads=true", nil)
	require.Equal(t, http.StatusOK, status, body)
	assert.Contains(t, body, base)

	status, body = call(prRun, http.MethodGet, "pr/status?number=8", nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Contains(t, body, "run "+prRun.String()+" works on PR #7, not #8")

	status, body = call(prRun, http.MethodGet, "pr/head", nil)
	require.Equal(t, http.StatusOK, status, body)
	assert.JSONEq(t, `{"sha":"`+base+`","branch":"topic"}`, body)

	status, body = call(prRun, http.MethodPost, "pr/threads/thread-1/resolve?number=7", nil)
	assert.Equal(t, http.StatusAccepted, status, body)
	assert.Equal(t, []string{"thread-1"}, github.resolved)

	// A turn's commits on the source branch.
	turn := t.TempDir()
	runGit(t, turn, "clone", "-q", "--branch", "topic", remote, ".")
	require.NoError(t, os.WriteFile(filepath.Join(turn, "fix.txt"), []byte("fixed"), 0o644))
	runGit(t, turn, "add", "fix.txt")
	runGit(t, turn, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "fix")
	tip := gitOutput(t, turn, "rev-parse", "HEAD")
	runGit(t, turn, "bundle", "create", filepath.Join(turn, "turn.bundle"), "HEAD", "^"+base)
	bundle, err := os.ReadFile(filepath.Join(turn, "turn.bundle"))
	require.NoError(t, err)

	status, body = call(prRun, http.MethodPost, "pr/push?base="+base, bundle)
	require.Equal(t, http.StatusOK, status, body)
	assert.JSONEq(t, `{"sha":"`+tip+`"}`, body)
	assert.Equal(t, tip, gitOutput(t, remote, "rev-parse", "refs/heads/topic"))

	status, body = call(prRun, http.MethodPost, "pr/push?base="+strings.Repeat("a", 40), bundle)
	assert.Equal(t, http.StatusConflict, status)
	assert.Contains(t, body, "the PR's source branch is at "+base+", not "+strings.Repeat("a", 40))

	status, body = call(otherRun, http.MethodGet, "pr/head", nil)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, body, "run "+otherRun.String()+" works on no PR")
}
