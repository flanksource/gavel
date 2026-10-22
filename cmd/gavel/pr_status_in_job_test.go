// Tests gavel pr status inside a sandbox job: with no GitHub credentials there, the PR's status
// comes from serve's run API, as a job's verification reads it.
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/prwatch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPRStatusInsideASandboxJobAsksServe(t *testing.T) {
	asked := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/runs/run-1/pr/status", r.URL.Path)
		require.Equal(t, "Bearer run-token", r.Header.Get("Authorization"))
		asked++
		_ = json.NewEncoder(w).Encode(model.StatusSnapshot{PR: &model.PRInfo{Number: 7, State: "OPEN", Title: "Fix the build"}})
	}))
	t.Cleanup(server.Close)
	t.Setenv("GAVEL_SERVER", server.URL)
	t.Setenv("GAVEL_RUN_TOKEN", "run-token")
	t.Setenv("GAVEL_RUN_ID", "run-1")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	t.Setenv("HOME", t.TempDir())
	t.Chdir(t.TempDir())

	result, err := runPRStatus(PRStatusOptions{Args: []string{"https://github.com/acme/app/pull/7"}, Interval: "1s", TailLogs: 100, Context: context.Background()})
	require.NoError(t, err)
	watched, ok := result.(*prwatch.PRWatchResult)
	require.True(t, ok, "%T", result)
	assert.Equal(t, "Fix the build", watched.PR.Title)
	assert.Equal(t, 1, asked)
}
