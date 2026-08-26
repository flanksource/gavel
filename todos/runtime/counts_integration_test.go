package runtime

import (
	"os"
	"path/filepath"
	"testing"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCountByStatusMatchesListIntegration pins the SQL aggregate to the
// materializing path it replaces: List decodes every issue and derives
// types.Status per item, so bucketing that output is an independent oracle for
// CountByStatus. Any drift between the GROUP BY projection and
// todoStatusWithPlan shows up here as a bucket mismatch.
func TestCountByStatusMatchesListIntegration(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres native runtime tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir: filepath.Join(t.TempDir(), "postgres"), Database: "gavel_todo_counts",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stop()) })

	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })

	root := filepath.Join(t.TempDir(), "counts-workspace")
	require.NoError(t, os.MkdirAll(root, 0o755))
	provider, err := New(t.Context(), opened.Gorm(), WorkspaceOptions{
		Name: "Counts", RootPath: root, Repositories: []string{"example/counts"},
	})
	require.NoError(t, err)

	// An empty workspace must report no buckets rather than erroring.
	empty, err := provider.CountByStatus(t.Context())
	require.NoError(t, err)
	assert.Equal(t, map[types.Status]int{}, empty)

	// Two of each durable status so a bucket that silently collapses to 1 (or
	// borrows another bucket's rows) cannot pass.
	seed := []struct {
		status types.Status
		count  int
	}{
		{types.StatusPending, 3},
		{types.StatusDraft, 2},
		{types.StatusCompleted, 4},
		{types.StatusVerified, 2},
	}
	for _, spec := range seed {
		for i := 0; i < spec.count; i++ {
			created, err := provider.Create(t.Context(), todos.CreateRequest{
				Title:  string(spec.status) + "-" + string(rune('a'+i)),
				Body:   "Body for " + string(spec.status),
				Status: spec.status,
			})
			require.NoError(t, err)
			if created.Status == spec.status {
				continue
			}
			// Create normalizes transient statuses onto durable ones; drive the
			// remainder through the same UpdateState path the UI uses.
			status := spec.status
			require.NoError(t, provider.UpdateState(t.Context(), created, todos.StateUpdate{Status: &status}))
		}
	}

	// The three GROUP BY columns only earn their keep on rows where execution
	// state and plan approval actually differ from the durable status: an
	// admitted prompt run projects a non-idle execution, and an unreviewed plan
	// turns an otherwise pending issue into StatusReview.
	running, err := provider.Create(t.Context(), todos.CreateRequest{
		Title: "Has an admitted prompt run", Body: "Body", Status: types.StatusPending,
	})
	require.NoError(t, err)
	preparation, err := provider.PrepareRun(t.Context(), running, todos.RunPreparation{
		Mode: types.ModeRun, ExecutorName: "codex",
	})
	require.NoError(t, err)
	require.NotEmpty(t, preparation.SessionID)

	awaitingReview, err := provider.Create(t.Context(), todos.CreateRequest{
		Title: "Has a plan awaiting review", Body: "Body", Status: types.StatusPending,
		Plan: &todos.CreatePlanRequest{Markdown: "# Plan\n\nDo the thing.\n"},
	})
	require.NoError(t, err)
	assert.Equal(t, types.StatusReview, awaitingReview.Status, "an unreviewed plan must derive StatusReview")

	items, err := provider.List(t.Context(), todos.DiscoveryFilters{})
	require.NoError(t, err)
	oracle := map[types.Status]int{}
	for _, item := range items {
		oracle[item.Status]++
	}

	counts, err := provider.CountByStatus(t.Context())
	require.NoError(t, err)
	assert.Equal(t, oracle, counts, "aggregate counts must match the per-item derivation from List")

	total := 0
	for _, n := range counts {
		total += n
	}
	assert.Equal(t, len(items), total, "aggregate must cover every issue exactly once")
}
