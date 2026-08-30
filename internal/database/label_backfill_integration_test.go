package database

import (
	"os"
	"path/filepath"
	"testing"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flanksource/gavel/todos/labels"
)

// TestBackfillTodoLabelsAdoptsLabelsInUse exercises the shipped backfill script
// against a workspace whose TODOs were labelled before per-project labels
// existed.
//
// It re-runs the script rather than relying on the one the migration already
// performed: migrations apply to an empty database, so the interesting case —
// a backlog that already has labels — only exists after seeding. The script is
// written to be safe to repeat (it adopts, never overrides), which is exactly
// what makes running it here a fair test of what shipped.
func TestBackfillTodoLabelsAdoptsLabelsInUse(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres label backfill tests")
	}

	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir:  filepath.Join(t.TempDir(), "postgres"),
		Database: "gavel_label_backfill",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stop()) })

	t.Setenv(EnvDSN, dsn)
	t.Setenv(EnvDisable, "")
	t.Setenv(LegacyEnvDSN, "")
	t.Setenv(LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())

	db, err := Open(t.Context(), WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	gorm := db.Gorm()

	const (
		workspaceOne = "d1000000-0000-0000-0000-000000000001"
		workspaceTwo = "d1000000-0000-0000-0000-000000000002"
	)
	require.NoError(t, gorm.Exec(`
		INSERT INTO todo_workspaces (id, repo_key) VALUES
			(?, 'github.com/example/backfill-one'),
			(?, 'github.com/example/backfill-two')`, workspaceOne, workspaceTwo).Error)

	require.NoError(t, gorm.Exec(`
		INSERT INTO todo_issues (workspace_id, title, labels) VALUES
			(?, 'namespaced', ARRAY['area/ui', 'area/api']),
			(?, 'plain and lifecycle', ARRAY['flaky', 'status:open', 'priority:high', 'session:abc']),
			(?, 'a built-in name', ARRAY['bug']),
			(?, 'already defined', ARRAY['curated']),
			(?, 'another workspace', ARRAY['flaky', 'two-only'])`,
		workspaceOne, workspaceOne, workspaceOne, workspaceOne, workspaceTwo).Error)

	// A label someone already coloured by hand must survive untouched — the
	// backfill adopts what is undefined, it does not re-decide what is defined.
	require.NoError(t, gorm.Exec(`
		INSERT INTO todo_labels (workspace_id, name, color) VALUES (?, 'curated', 'rose')`,
		workspaceOne).Error)

	require.NoError(t, gorm.Exec(readEmbeddedSchemaFile(t, "schema/120_backfill_todo_labels.sql")).Error)

	adopted := map[string]string{}
	var rows []struct {
		Name  string `gorm:"column:name"`
		Color string `gorm:"column:color"`
	}
	require.NoError(t, gorm.Raw(
		`SELECT name, color FROM todo_labels WHERE workspace_id = ?`, workspaceOne).Scan(&rows).Error)
	for _, row := range rows {
		adopted[row.Name] = row.Color
	}

	t.Run("labels in use become editable project labels", func(t *testing.T) {
		assert.Contains(t, adopted, "flaky")
		assert.Contains(t, adopted, "area/ui")
		assert.Contains(t, adopted, "area/api")
	})

	t.Run("adoption keeps the colour each label already rendered with", func(t *testing.T) {
		// The SQL hue function mirrors labels.Hash; if the two ever disagree,
		// adoption silently repaints every backlog it touches.
		for _, name := range []string{"flaky", "area/ui", "area/api"} {
			assert.Equal(t, string(labels.Hash(name)), adopted[name],
				"%s: SQL hue and labels.Hash have drifted", name)
		}
	})

	t.Run("a namespace shares one hue", func(t *testing.T) {
		assert.Equal(t, adopted["area/ui"], adopted["area/api"],
			"namespaced labels colour as a family")
	})

	t.Run("lifecycle labels are not vocabulary", func(t *testing.T) {
		for _, reserved := range []string{"status:open", "priority:high", "session:abc"} {
			assert.NotContains(t, adopted, reserved)
		}
	})

	t.Run("built-ins keep their hand-picked colours", func(t *testing.T) {
		assert.NotContains(t, adopted, "bug",
			"a built-in already resolves and is already editable; adopting it would repaint it")
	})

	t.Run("an existing definition is left alone", func(t *testing.T) {
		assert.Equal(t, "rose", adopted["curated"])
	})

	t.Run("adoption is per workspace", func(t *testing.T) {
		assert.NotContains(t, adopted, "two-only")

		var names []string
		require.NoError(t, gorm.Raw(
			`SELECT name FROM todo_labels WHERE workspace_id = ? ORDER BY name`, workspaceTwo).Scan(&names).Error)
		assert.Equal(t, []string{"flaky", "two-only"}, names)
	})

	t.Run("the helper function does not outlive the migration", func(t *testing.T) {
		var remaining int
		require.NoError(t, gorm.Raw(
			`SELECT COUNT(*) FROM pg_proc WHERE proname = 'gavel_backfill_label_hue'`).Scan(&remaining).Error)
		assert.Zero(t, remaining, "the backfill helper must be dropped, not left in the schema")
	})
}
