// Tests that a client reads and writes only its organization's GitHub cache: cached job
// logs, completed runs and stale HTTP responses never reach another organization.
package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	githubcache "github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/pr/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openOrgCaches returns the local org's cache and a second org's, over one migrated database.
func openOrgCaches(t *testing.T) (local, other *githubcache.Store) {
	t.Helper()
	dsn := commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_github_org_cache"}).DSN()
	t.Setenv(githubcache.EnvDatabaseDSN, dsn)
	t.Setenv(githubcache.EnvDSN, "")
	t.Setenv(githubcache.EnvDisable, "")
	migrated, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	otherOrg := uuid.New()
	require.NoError(t, migrated.Gorm().Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Other')`, otherOrg, otherOrg.String()).Error)
	require.NoError(t, migrated.Close())
	local, err = githubcache.Open()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, local.Close()) })
	return local, local.ForOrg(otherOrg)
}

func TestClientUsesOnlyItsOrganizationsCache(t *testing.T) {
	local, other := openOrgCaches(t)
	var requests atomic.Int64
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		http.NotFound(w, nil)
	}))
	t.Cleanup(api.Close)
	t.Setenv("GITHUB_API_URL", api.URL)
	token := func() (string, error) { return "org-token", nil }

	t.Run("job logs", func(t *testing.T) {
		local.PutJobLogs(4242, "owner/repo", "private logs")
		job := &model.Job{DatabaseID: 4242, Conclusion: "success"}
		require.Error(t, FetchJobLogs(Options{Repo: "owner/repo", TokenSource: token, Cache: other}, job, 0))
		assert.Equal(t, int64(1), requests.Load(), "the other org must ask GitHub, not read the cache")

		require.NoError(t, FetchJobLogs(Options{Repo: "owner/repo", TokenSource: token, Cache: local}, job, 0))
		assert.Equal(t, int64(1), requests.Load(), "the owning org is served from its cache")
	})

	t.Run("completed runs", func(t *testing.T) {
		requests.Store(0)
		payload, err := json.Marshal(model.WorkflowRun{DatabaseID: 777, Name: "CI", Status: "completed", Conclusion: "success"})
		require.NoError(t, err)
		local.PutCompletedRun("owner/repo", 777, "completed", "success", payload)

		_, err = FetchRunJobs(Options{Repo: "owner/repo", TokenSource: token, Cache: other}, 777, RunLogOptions{})
		require.Error(t, err)
		assert.Equal(t, int64(1), requests.Load())

		run, err := FetchRunJobs(Options{Repo: "owner/repo", TokenSource: token, Cache: local}, 777, RunLogOptions{})
		require.NoError(t, err)
		assert.Equal(t, "CI", run.Name)
	})

	t.Run("stale HTTP responses on a network error", func(t *testing.T) {
		down := httptest.NewServer(http.NotFoundHandler())
		down.Close()
		t.Setenv("GITHUB_API_URL", down.URL)
		local.StoreHTTP(down.URL+"/repos/owner/repo", "GET", 200, []byte(`{"private":true}`), http.Header{"Etag": {`"v1"`}}, nil)

		_, err := cachedGet(context.Background(), other, "org-token", "/repos/owner/repo", nil)
		require.Error(t, err, "the other org gets the network error, not the owner's stale body")

		result, err := cachedGet(context.Background(), local, "org-token", "/repos/owner/repo", nil)
		require.NoError(t, err)
		assert.True(t, result.FromCache)
	})
}
