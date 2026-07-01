package cache

import (
	"context"
	"os"
	"testing"

	"github.com/flanksource/gavel/todos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGriteCacheDisabledNoOp(t *testing.T) {
	s := &Store{disabled: true}
	assert.False(t, s.Enabled())

	cursor, syncedAt, err := s.LoadCursor(context.Background(), "/repo")
	require.NoError(t, err)
	assert.Zero(t, cursor)
	assert.True(t, syncedAt.IsZero())

	issues, err := s.ListIssues(context.Background(), "/repo")
	require.NoError(t, err)
	assert.Empty(t, issues)

	got, err := s.GetIssue(context.Background(), "/repo", "x")
	require.NoError(t, err)
	assert.Nil(t, got)

	assert.NoError(t, s.UpsertSync(context.Background(), "/repo", []todos.CachedIssue{{IssueID: "x"}}, 1))
}

func TestGriteCacheRoundtripIntegration(t *testing.T) {
	dsn := os.Getenv(EnvDSN)
	if dsn == "" {
		t.Skipf("set %s=postgres://... to run integration tests", EnvDSN)
	}
	t.Setenv(EnvDSN, dsn)
	t.Setenv(EnvDisable, "")

	store, err := Open()
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	const repo = "/tmp/grite-cache-test"
	t.Cleanup(func() {
		_ = store.gorm().Where("repo = ?", repo).Delete(&GriteIssueCache{})
		_ = store.gorm().Where("repo = ?", repo).Delete(&GriteSyncCursor{})
	})

	ctx := context.Background()
	issues := []todos.CachedIssue{
		{Repo: repo, IssueID: "i1", Title: "First", State: "open", Labels: []string{"priority:high"}, UpdatedTS: 100, EventsJSON: []byte(`[{"event_id":"e1"}]`)},
		{Repo: repo, IssueID: "i2", Title: "Second", State: "closed", UpdatedTS: 200},
	}
	require.NoError(t, store.UpsertSync(ctx, repo, issues, 200))

	cursor, syncedAt, err := store.LoadCursor(ctx, repo)
	require.NoError(t, err)
	assert.Equal(t, int64(200), cursor)
	assert.False(t, syncedAt.IsZero())

	listed, err := store.ListIssues(ctx, repo)
	require.NoError(t, err)
	assert.Len(t, listed, 2)

	got, err := store.GetIssue(ctx, repo, "i1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "First", got.Title)
	assert.Equal(t, []string{"priority:high"}, got.Labels)
	assert.JSONEq(t, `[{"event_id":"e1"}]`, string(got.EventsJSON))

	// Re-upsert must overwrite, not duplicate.
	require.NoError(t, store.UpsertSync(ctx, repo, []todos.CachedIssue{{Repo: repo, IssueID: "i1", Title: "First (edited)", State: "open"}}, 250))
	got, err = store.GetIssue(ctx, repo, "i1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "First (edited)", got.Title)

	cursor, _, err = store.LoadCursor(ctx, repo)
	require.NoError(t, err)
	assert.Equal(t, int64(250), cursor)
}
