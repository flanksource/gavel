// Tests hosted project actions: test and lint run as sandbox jobs of the project's repository on
// the branch a member picks, several at once, seen by the whole organization.
package ui

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pushBranch pushes a branch with one more commit than main to the bare repository and returns
// its commit.
func pushBranch(t *testing.T, bare, branch string) string {
	t.Helper()
	work := t.TempDir()
	git := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = work
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v: %s", args, out)
		return strings.TrimSpace(string(out))
	}
	git("clone", "-q", bare, ".")
	git("checkout", "-q", "-b", branch)
	require.NoError(t, os.WriteFile(filepath.Join(work, "feature.txt"), []byte(branch), 0o644))
	git("add", ".")
	git("commit", "-q", "-m", "feature")
	git("push", "-q", "origin", branch)
	return git("rev-parse", "HEAD")
}

func TestHostedProjectActionsRunOnTheBranchAMemberPicks(t *testing.T) {
	h := newHostedRuns(t, nil)
	repo := uuid.MustParse(strings.TrimPrefix(h.workspace, "repo:"))
	sha := pushBranch(t, h.executor.bareDir(repo), "feat/x")

	status, branches := hostedCall(t, h.base, http.MethodGet, "/api/projects/app/branches", h.ada, "org:member", nil)
	require.Equal(t, http.StatusOK, status, branches)
	assert.Equal(t, map[string]any{"default": "main", "names": []any{"feat/x", "main"}}, branches)

	started := map[string]string{}
	for _, action := range []string{"test", "lint", "test"} {
		status, body := hostedCall(t, h.base, http.MethodPost, "/api/projects/app/actions", h.ada, "org:member", map[string]any{"action": action, "branch": "feat/x"})
		require.Equal(t, http.StatusAccepted, status, body)
		id := body["runId"].(string)
		assert.Equal(t, action, body["action"])
		assert.Equal(t, true, body["running"])
		assert.Equal(t, "/runs/"+id, body["href"])
		started[id] = action
	}
	require.Len(t, started, 3, "a project runs several actions at once")

	type run struct {
		ID     uuid.UUID
		Kind   string
		Status string
		Ref    string
		SHA    string
		UserID uuid.UUID
		RepoID uuid.UUID
		TodoID *uuid.UUID
	}
	var runs []run
	require.Eventually(t, func() bool {
		runs = nil
		require.NoError(t, h.db.Raw(`SELECT id, kind, status, ref, sha, user_id, repo_id, todo_id FROM sandbox_runs WHERE org_id = ?`, h.org).Scan(&runs).Error)
		for _, r := range runs {
			if r.Status != "succeeded" {
				return false
			}
		}
		return len(runs) == 3
	}, 30*time.Second, 50*time.Millisecond, "every job checked out the branch's commit: %+v", runs)
	for _, r := range runs {
		assert.Equal(t, started[r.ID.String()], r.Kind)
		assert.Equal(t, "refs/heads/feat/x", r.Ref)
		assert.Equal(t, sha, r.SHA)
		assert.Equal(t, h.ada, r.UserID)
		assert.Equal(t, repo, r.RepoID)
		assert.Nil(t, r.TodoID)
	}

	status, listed := hostedCall(t, h.base, http.MethodGet, "/api/sandbox-runs", h.grace, "org:member", nil)
	require.Equal(t, http.StatusOK, status, listed)
	assert.Len(t, listed["items"], 3, "another member sees the project's runs")
}

func TestHostedProjectActionsRefuseWhatTheyCantRun(t *testing.T) {
	h := newHostedRuns(t, nil)
	for _, c := range []struct {
		name, project string
		body          map[string]any
		status        int
		message       string
	}{
		{"no branch", "app", map[string]any{"action": "test"}, http.StatusBadRequest, "pick the branch to run test on"},
		{"unknown branch", "app", map[string]any{"action": "lint", "branch": "nope"}, http.StatusBadRequest, `unknown branch "nope": app has no such branch`},
		{"options", "app", map[string]any{"action": "test", "branch": "main", "options": map[string]any{"paths": []string{"./..."}}}, http.StatusBadRequest,
			"hosted test and lint run on the whole branch; files and options aren't supported"},
		{"files", "app", map[string]any{"action": "lint", "branch": "main", "files": []string{"a.go"}}, http.StatusBadRequest,
			"hosted test and lint run on the whole branch; files and options aren't supported"},
		{"unknown action", "app", map[string]any{"action": "deploy", "branch": "main"}, http.StatusBadRequest, `unsupported project action "deploy"`},
		{"commit", "app", map[string]any{"action": "commit", "branch": "main"}, http.StatusBadRequest, "commits come from agent runs in hosted mode"},
		{"unknown project", "elsewhere", map[string]any{"action": "test", "branch": "main"}, http.StatusNotFound, `project "elsewhere" is not one of this organization's repositories`},
	} {
		t.Run(c.name, func(t *testing.T) {
			status, body := hostedCall(t, h.base, http.MethodPost, "/api/projects/"+c.project+"/actions", h.ada, "org:member", c.body)
			assert.Equal(t, c.status, status, body)
			assert.Equal(t, c.message, body["error"])
		})
	}
	status, body := hostedCall(t, h.base, http.MethodGet, "/api/projects/elsewhere/branches", h.ada, "org:member", nil)
	assert.Equal(t, http.StatusNotFound, status, body)
	assert.Equal(t, `project "elsewhere" is not one of this organization's repositories`, body["error"])

	var runs int64
	require.NoError(t, h.db.Raw(`SELECT count(*) FROM sandbox_runs`).Scan(&runs).Error)
	assert.Zero(t, runs, "nothing was dispatched")
}
