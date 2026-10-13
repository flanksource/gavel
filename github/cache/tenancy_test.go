// Tests that every cache table is isolated per organization: a write in one org is invisible
// to another, the same key coexists across orgs, and upserts never reassign owner columns.
package cache

import (
	"context"
	nethttp "net/http"
	"testing"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openTenancyStores returns stores for the local org and a second org over one migrated database.
func openTenancyStores(t *testing.T) (local, other *Store) {
	t.Helper()
	dsn := commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_cache_tenancy"}).DSN()
	resetSharedStore(t)
	t.Setenv(EnvDatabaseDSN, dsn)
	t.Setenv(EnvDSN, "")
	t.Setenv(EnvDisable, "")
	migrated, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	require.NoError(t, migrated.Close())

	local, err = Open()
	require.NoError(t, err)
	require.False(t, local.Disabled())
	t.Cleanup(func() { require.NoError(t, local.Close()) })
	otherOrg := uuid.New()
	require.NoError(t, local.gorm().Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Other')`, otherOrg, otherOrg.String()).Error)
	return local, local.ForOrg(otherOrg)
}

func TestCacheIsolatesOrganizations(t *testing.T) {
	local, other := openTenancyStores(t)
	ctx := context.Background()
	require.Equal(t, tenant.LocalOrgID, local.org)

	t.Run("http", func(t *testing.T) {
		const url = "https://api.github.com/repos/a/b"
		headers := nethttp.Header{}
		headers.Set("ETag", `"local"`)
		local.StoreHTTP(url, "GET", 200, []byte("local"), headers, nil)
		assert.Nil(t, other.LookupHTTP(url, "GET"))
		other.TouchHTTP(url, "GET")
		headers.Set("ETag", `"other"`)
		other.StoreHTTP(url, "GET", 200, []byte("other"), headers, nil)
		assert.Equal(t, []byte("other"), other.LookupHTTP(url, "GET").Body)
		assert.Equal(t, []byte("local"), local.LookupHTTP(url, "GET").Body)

		// Upserts refresh the payload but never reassign the row's owner.
		user := uuid.New()
		require.NoError(t, local.gorm().Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, user, user.String()).Error)
		require.NoError(t, local.gorm().Exec(`UPDATE http_cache_entries SET user_id = ? WHERE org_id = ?`, user, tenant.LocalOrgID).Error)
		local.StoreHTTP(url, "GET", 200, []byte("refreshed"), headers, nil)
		var owner struct {
			UserID *uuid.UUID
			OrgID  uuid.UUID
		}
		require.NoError(t, local.gorm().Raw(`SELECT user_id, org_id FROM http_cache_entries WHERE org_id = ? AND url = ?`, tenant.LocalOrgID, url).Scan(&owner).Error)
		require.NotNil(t, owner.UserID)
		assert.Equal(t, user, *owner.UserID)
		assert.Equal(t, []byte("refreshed"), local.LookupHTTP(url, "GET").Body)
	})

	t.Run("workflow runs", func(t *testing.T) {
		local.PutCompletedRun("a/b", 7, "completed", "success", []byte(`{"org":"local"}`))
		assert.Nil(t, other.GetCompletedRunPayload(7))
		other.PutCompletedRun("a/b", 7, "completed", "failure", []byte(`{"org":"other"}`))
		assert.Equal(t, []byte(`{"org":"other"}`), other.GetCompletedRunPayload(7))
		assert.Equal(t, []byte(`{"org":"local"}`), local.GetCompletedRunPayload(7))
	})

	t.Run("job logs", func(t *testing.T) {
		local.PutJobLogs(9, "a/b", "local logs")
		_, ok := other.GetJobLogs(9)
		assert.False(t, ok)
		other.PutJobLogs(9, "a/b", "other logs")
		logs, _ := other.GetJobLogs(9)
		assert.Equal(t, "other logs", logs)
		logs, _ = local.GetJobLogs(9)
		assert.Equal(t, "local logs", logs)
	})

	t.Run("workflow definitions", func(t *testing.T) {
		local.PutWorkflowDef("a/b", 3, "sha1", ".github/workflows/ci.yml", "local: true")
		_, _, ok := other.GetWorkflowDef("a/b", 3, "sha1")
		assert.False(t, ok)
		other.PutWorkflowDef("a/b", 3, "sha1", ".github/workflows/ci.yml", "other: true")
		yaml, _, _ := other.GetWorkflowDef("a/b", 3, "sha1")
		assert.Equal(t, "other: true", yaml)
		yaml, _, _ = local.GetWorkflowDef("a/b", 3, "sha1")
		assert.Equal(t, "local: true", yaml)
	})

	t.Run("seen pull requests", func(t *testing.T) {
		keys := []SeenKey{{Repo: "a/b", Number: 1}, {Repo: "a/b", Number: 2}}
		require.NoError(t, local.MarkSeen("a/b", 1))
		seen, err := other.LoadSeenMap(keys)
		require.NoError(t, err)
		assert.Empty(t, seen)
		require.NoError(t, other.MarkManySeen(keys))
		require.NoError(t, other.MarkSeen("a/b", 1))
		seen, err = other.LoadSeenMap(keys)
		require.NoError(t, err)
		assert.Len(t, seen, 2)
		seen, err = local.LoadSeenMap(keys)
		require.NoError(t, err)
		assert.Len(t, seen, 1)
	})

	t.Run("pull request index", func(t *testing.T) {
		key := PRKey{Provider: "github", Repo: "a/b", Number: 3}
		row := func(title string) PullRequest {
			return PullRequest{Provider: "github", Repo: "a/b", Number: 3, URL: "https://github.com/a/b/pull/3", Title: title}
		}
		require.NoError(t, local.UpsertPRs([]PullRequest{row("local title")}))
		found, err := other.GetPRs([]PRKey{key})
		require.NoError(t, err)
		assert.Empty(t, found)
		searched, err := other.SearchPRs(PRQuery{Limit: 10})
		require.NoError(t, err)
		assert.Empty(t, searched)

		require.NoError(t, other.UpsertPRs([]PullRequest{row("other title")}))
		found, err = other.GetPRs([]PRKey{key})
		require.NoError(t, err)
		assert.Equal(t, "other title", found[key].Title)
		searched, err = local.SearchPRs(PRQuery{Text: "title", Limit: 10})
		require.NoError(t, err)
		require.Len(t, searched, 1)
		assert.Equal(t, "local title", searched[0].Title)
	})

	t.Run("test runs", func(t *testing.T) {
		require.NoError(t, local.UpsertTestRuns(ctx, "/ws", []TestRunCache{{RunID: "run-1", Kind: "test", StartedTS: 100, Passed: 1}}, 100))
		cursor, _, err := other.LoadTestRunCursor(ctx, "/ws")
		require.NoError(t, err)
		assert.Zero(t, cursor)
		runs, err := other.ListTestRuns(ctx)
		require.NoError(t, err)
		assert.Empty(t, runs)
		runs, err = other.ListTestRunsForDir(ctx, "/ws")
		require.NoError(t, err)
		assert.Empty(t, runs)
		run, err := other.GetTestRun(ctx, "/ws", "run-1")
		require.NoError(t, err)
		assert.Nil(t, run)

		require.NoError(t, other.UpsertTestRuns(ctx, "/ws", []TestRunCache{{RunID: "run-1", Kind: "lint", StartedTS: 200, Passed: 9}}, 200))
		run, err = other.GetTestRun(ctx, "/ws", "run-1")
		require.NoError(t, err)
		assert.Equal(t, 9, run.Passed)
		run, err = local.GetTestRun(ctx, "/ws", "run-1")
		require.NoError(t, err)
		assert.Equal(t, 1, run.Passed)
		cursor, _, err = local.LoadTestRunCursor(ctx, "/ws")
		require.NoError(t, err)
		assert.EqualValues(t, 100, cursor)
		runs, err = local.ListTestRuns(ctx)
		require.NoError(t, err)
		assert.Len(t, runs, 1)
	})
}

func TestStatusShowsHostDetailsOnlyToTheLocalOrg(t *testing.T) {
	local, other := openTenancyStores(t)

	hostView := local.Status()
	assert.True(t, hostView.Enabled)
	assert.NotEmpty(t, hostView.DSNSource)
	assert.NotEmpty(t, hostView.Counts)

	orgView := other.Status()
	assert.True(t, orgView.Enabled)
	assert.Empty(t, orgView.Error)
	assert.Empty(t, orgView.DSNSource, "an organization never sees how the host's database is configured")
	assert.Empty(t, orgView.DSNMasked)
	assert.Empty(t, orgView.Counts, "table counts span every organization")
}
