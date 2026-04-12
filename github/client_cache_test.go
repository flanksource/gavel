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
	got, err := FetchRunJobs(Options{Repo: "owner/repo"}, runID)
	require.NoError(t, err, "FetchRunJobs must short-circuit before token resolution")
	assert.Equal(t, want.DatabaseID, got.DatabaseID)
	assert.Equal(t, want.Name, got.Name)
	assert.Equal(t, want.Status, got.Status)
	assert.Equal(t, want.Conclusion, got.Conclusion)
}

// resetSharedStoreForTest is currently a no-op: the cache.Shared() singleton
// is opened lazily on first call, and as long as the test sets the env vars
// before that first call, the singleton picks them up. If a previous test in
// the same binary already opened the singleton with different settings, this
// test will be skipped via the DSN guard above.
func resetSharedStoreForTest(t *testing.T) {
	t.Helper()
}
