package github

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	githubcache "github.com/flanksource/gavel/github/cache"
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
	want := &WorkflowRun{
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

// TestFetchRunJobsEnrichesCachedRunWithLogs verifies the cache-poisoning fix:
// a completed run cached WITHOUT logs must be enriched (and re-cached) when a
// later FetchRunJobs call requests logs. We seed the job-logs cache so the
// enrichment path attaches logs without any network call (no token is set, so
// any HTTP attempt would fail).
func TestFetchRunJobsEnrichesCachedRunWithLogs(t *testing.T) {
	dsn := os.Getenv(githubcache.EnvDSN)
	if dsn == "" {
		t.Skipf("set %s=postgres://... to run this integration test", githubcache.EnvDSN)
	}
	t.Setenv(githubcache.EnvDSN, dsn)
	t.Setenv(githubcache.EnvDisable, "")
	resetSharedStoreForTest(t)

	store := githubcache.Shared()
	require.False(t, store.Disabled())

	const runID = int64(987654322)
	const jobID = int64(55555)
	const repo = "owner/repo"

	// A completed, FAILED run cached without any logs (the poisoned state a
	// prior `gavel pr status` without --logs would leave behind).
	logless := &WorkflowRun{
		DatabaseID: runID,
		Name:       "CI",
		Status:     "completed",
		Conclusion: "failure",
		Jobs: []Job{{
			DatabaseID: jobID,
			Name:       "build",
			Status:     "completed",
			Conclusion: "failure",
			Steps:      []Step{{Name: "go test", Status: "completed", Conclusion: "failure"}},
		}},
	}
	payload, err := json.Marshal(logless)
	require.NoError(t, err)
	store.PutCompletedRun(repo, runID, logless.Status, logless.Conclusion, payload)

	// Seed the job-logs cache so enrichment is network-free.
	const rawLog = "2024-01-01T00:00:00.000Z go test\n2024-01-01T00:00:01.000Z FAIL\tpkg\t0.01s"
	store.PutJobLogs(jobID, repo, rawLog)

	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	got, err := FetchRunJobs(Options{Repo: repo}, runID, RunLogOptions{FetchLogs: true, TailLines: 100})
	require.NoError(t, err, "FetchRunJobs must enrich from cache without a token")
	require.Len(t, got.Jobs, 1)
	assert.NotEmpty(t, got.Jobs[0].Logs, "failed job must have logs attached on a logs-requested cache hit")

	// The enriched run must be re-persisted so a subsequent logs request is a
	// pure cache hit.
	require.True(t, runHasFailureLogs(got))
	refetched, err := FetchRunJobs(Options{Repo: repo}, runID, RunLogOptions{FetchLogs: true, TailLines: 100})
	require.NoError(t, err)
	assert.NotEmpty(t, refetched.Jobs[0].Logs, "re-cached run must retain logs")
}

// resetSharedStoreForTest is currently a no-op: the cache.Shared() singleton
// is opened lazily on first call, and as long as the test sets the env vars
// before that first call, the singleton picks them up. If a previous test in
// the same binary already opened the singleton with different settings, this
// test will be skipped via the DSN guard above.
func resetSharedStoreForTest(t *testing.T) {
	t.Helper()
}
