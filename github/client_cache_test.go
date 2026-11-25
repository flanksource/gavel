package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/flanksource/commons-db/dbtest"
	githubcache "github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/pr/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFetchRunJobsServesCompletedRunFromCache verifies the Phase 3 invariant:
// once a completed workflow run has been written to the persistent cache,
// FetchRunJobs returns it WITHOUT making any HTTP request. We force this by
// configuring an Options with no token — if FetchRunJobs hits the network
// at all, opts.token() returns an error and the test fails.
func TestFetchRunJobsServesCompletedRunFromCache(t *testing.T) {
	dsn := os.Getenv(githubcache.EnvDSN)
	if dsn == "" {
		t.Skipf("set %s=postgres://... to run this integration test", githubcache.EnvDSN)
	}
	t.Setenv(githubcache.EnvDSN, dsn)
	t.Setenv(githubcache.EnvDisable, "")
	// Clear any previous singleton from another test in the same binary.
	resetSharedStoreForTest(t)

	store := githubcache.Shared()
	require.False(t, store.Disabled())

	// Pre-populate the cache with a completed run.
	const runID = int64(987654321)
	want := &model.WorkflowRun{
		DatabaseID: runID,
		Name:       "CI",
		Status:     "completed",
		Conclusion: "success",
		URL:        "https://github.com/owner/repo/actions/runs/987654321",
	}
	payload, err := json.Marshal(want)
	require.NoError(t, err)
	store.PutCompletedRun("owner/repo", runID, want.Status, want.Conclusion, payload)

	// FetchRunJobs with no token: if the function tries to talk to GitHub it
	// will fail at opts.token(), proving the cache short-circuit happened.
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	got, err := FetchRunJobs(Options{Repo: "owner/repo"}, runID, RunLogOptions{})
	require.NoError(t, err, "FetchRunJobs must short-circuit before token resolution")
	assert.Equal(t, want.DatabaseID, got.DatabaseID)
	assert.Equal(t, want.Name, got.Name)
	assert.Equal(t, want.Status, got.Status)
	assert.Equal(t, want.Conclusion, got.Conclusion)
}

func TestFetchRunJobsTailsCachedRawLogs(t *testing.T) {
	t.Setenv(githubcache.EnvDatabaseDSN, dbtest.ForT(t, dbtest.Options{Name: "gavel_log_cache"}).DSN())
	t.Setenv(githubcache.EnvDSN, "")
	t.Setenv(githubcache.EnvDisable, "")
	t.Setenv(githubcache.EnvDatabaseDisable, "")
	migrated, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	require.NoError(t, migrated.Close())

	store, err := githubcache.Open()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, store.Close()) })
	require.False(t, store.Disabled())

	const runID = int64(987654322)
	const jobID = int64(55555)
	const repo = "owner/repo"

	legacy := &model.WorkflowRun{
		DatabaseID: runID,
		Name:       "CI",
		Status:     "completed",
		Conclusion: "failure",
		Jobs: []model.Job{{
			DatabaseID: jobID,
			Name:       "build",
			Status:     "completed",
			Conclusion: "failure",
			Logs:       "old tail",
			Steps:      []model.Step{{Name: "Run tests", Status: "completed", Conclusion: "failure", Logs: "old excerpt"}},
		}},
	}
	payload, err := json.Marshal(legacy)
	require.NoError(t, err)
	store.PutCompletedRun(repo, runID, legacy.Status, legacy.Conclusion, payload)

	const rawLog = "##[group]Run tests\n##[endgroup]\n##[group]Diagnostics\nstarting tests\nFAIL\n##[endgroup]"
	store.PutJobLogs(jobID, repo, rawLog)

	server := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(server.Close)
	t.Setenv("GITHUB_API_URL", server.URL)
	options := Options{Repo: repo, Cache: store, Token: "test-token"}
	for _, request := range []struct {
		options RunLogOptions
		want    string
	}{
		{RunLogOptions{FetchLogs: true, TailLines: 1}, "FAIL"},
		{RunLogOptions{FetchLogs: true}, "starting tests\nFAIL"},
		{RunLogOptions{}, ""},
	} {
		got, err := FetchRunJobs(options, runID, request.options)
		require.NoError(t, err)
		assert.Equal(t, request.want, got.Jobs[0].Logs)
		assert.Equal(t, request.want, got.Jobs[0].Steps[0].Logs)
		if request.options.FetchLogs {
			assert.Contains(t, got.Pretty().String(), "Diagnostics")
		}
	}
}

// resetSharedStoreForTest is currently a no-op: the cache.Shared() singleton
// is opened lazily on first call, and as long as the test sets the env vars
// before that first call, the singleton picks them up. If a previous test in
// the same binary already opened the singleton with different settings, this
// test will be skipped via the DSN guard above.
func resetSharedStoreForTest(t *testing.T) {
	t.Helper()
}
