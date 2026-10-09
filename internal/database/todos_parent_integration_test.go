package database_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type todoParentFixture struct {
	issue          string
	sibling        string
	otherWorkspace string
}

// assertTodoParentConstraints covers the two parent rules the database enforces
// on its own, so a writer that bypasses native.SetIssueParent still cannot store
// them. It leaves fixture.sibling parented to fixture.issue, which is the row a
// repeated apply of the bundle then has to preserve.
func assertTodoParentConstraints(t *testing.T, db *gorm.DB, fixture todoParentFixture) {
	t.Helper()
	setParent := func(issue, parent string) error {
		return db.Exec(`UPDATE todo_issues SET parent_issue_id = ? WHERE id = ?`, parent, issue).Error
	}

	assertParentRejected(t, setParent(fixture.issue, fixture.issue),
		"23514", "todo_issues_parent_not_self")
	assertParentRejected(t, setParent(fixture.issue, fixture.otherWorkspace),
		"23503", "todo_issues_parent_fkey")
	require.NoError(t, setParent(fixture.sibling, fixture.issue))

	var parent string
	require.NoError(t, db.Raw(
		`SELECT parent_issue_id FROM todo_issues WHERE id = ?`, fixture.sibling).Scan(&parent).Error)
	assert.Equal(t, fixture.issue, parent)

	var indexed bool
	require.NoError(t, db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes
			WHERE schemaname = 'public' AND indexname = 'todo_issues_parent_idx'
			  AND indexdef LIKE '%WHERE (parent_issue_id IS NOT NULL)'
		)`).Scan(&indexed).Error)
	assert.True(t, indexed, "todo_issues_parent_idx must be partial on parent_issue_id IS NOT NULL")
}

// TestTodoParentColumnArrivesOnAnExistingDatabase covers the upgrade every real
// database takes. A fresh install declares (id, workspace_id) as a unique
// constraint inside CREATE TABLE; a database provisioned before the parent
// column has it as a unique index and no column. The bundle must add the column
// and its foreign key against that index, and must not rebuild the index the
// alias and relationship foreign keys already depend on.
func TestTodoParentColumnArrivesOnAnExistingDatabase(t *testing.T) {
	dsn := dbtest.ForT(t, dbtest.Options{Name: "gavel_todo_parent_upgrade"}).DSN()

	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())

	db, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	fresh := todoIssueKeyDependents(t, db.Gorm())

	// Postgres cannot turn a unique constraint back into a bare index, so the
	// rewind drops it, which takes the foreign keys that reference it, and then
	// puts back the index and every foreign key a pre-parent database has.
	require.NoError(t, db.Gorm().Exec(`
		ALTER TABLE public.todo_issues DROP COLUMN parent_issue_id;
		ALTER TABLE public.todo_issues DROP CONSTRAINT todo_issues_id_workspace_id_key CASCADE;
		CREATE UNIQUE INDEX todo_issues_id_workspace_id_key ON public.todo_issues (id, workspace_id);
		ALTER TABLE public.todo_issue_aliases ADD CONSTRAINT todo_issue_aliases_issue_workspace_fkey
			FOREIGN KEY (issue_id, workspace_id) REFERENCES public.todo_issues (id, workspace_id) ON DELETE CASCADE;
		ALTER TABLE public.todo_issue_relationships ADD CONSTRAINT todo_issue_relationships_issue_workspace_fkey
			FOREIGN KEY (issue_id, workspace_id) REFERENCES public.todo_issues (id, workspace_id);
		ALTER TABLE public.todo_issue_relationships ADD CONSTRAINT todo_issue_relationships_target_workspace_fkey
			FOREIGN KEY (target_issue_id, workspace_id) REFERENCES public.todo_issues (id, workspace_id)`).Error)

	existing := todoIssueKeyDependents(t, db.Gorm())
	delete(fresh, "todo_issues_parent_fkey")
	require.ElementsMatch(t, slices.Collect(maps.Keys(fresh)), slices.Collect(maps.Keys(existing)),
		"the rewound database must hold every foreign key a pre-parent database has on (id, workspace_id)")

	workspaceID, otherWorkspaceID := uuid.NewString(), uuid.NewString()
	fixture := todoParentFixture{issue: uuid.NewString(), sibling: uuid.NewString(), otherWorkspace: uuid.NewString()}
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO todo_workspaces (id, repo_key) VALUES
			(?, 'github.com/acme/parent-upgrade'),
			(?, 'github.com/acme/parent-upgrade-other')`, workspaceID, otherWorkspaceID).Error)
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO todo_issues (id, workspace_id, title) VALUES
			(?, ?, 'existing issue'), (?, ?, 'existing sibling'), (?, ?, 'other workspace issue')`,
		fixture.issue, workspaceID, fixture.sibling, workspaceID, fixture.otherWorkspace, otherWorkspaceID).Error)

	upgraded, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err, "migrating a pre-parent database must add the column against the existing unique index")
	require.NoError(t, upgraded.Close())

	assertTodoParentConstraints(t, db.Gorm(), fixture)
	assertParentRejected(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_aliases (workspace_id, alias, issue_id)
		VALUES (?, 'cross-workspace', ?)`, otherWorkspaceID, fixture.issue).Error,
		"23503", "todo_issue_aliases_issue_workspace_fkey")

	var keyKinds []string
	require.NoError(t, db.Gorm().Raw(`
		SELECT CASE WHEN constraint_row.oid IS NULL THEN 'index' ELSE 'constraint' END
		FROM pg_class AS index_row
		LEFT JOIN pg_constraint AS constraint_row ON constraint_row.conindid = index_row.oid
			AND constraint_row.contype = 'u'
		WHERE index_row.relname = 'todo_issues_id_workspace_id_key'`).Scan(&keyKinds).Error)
	assert.Equal(t, []string{"index"}, keyKinds, "the existing unique index must be kept, not rebuilt as a constraint")

	upgradedKeys := todoIssueKeyDependents(t, db.Gorm())
	assert.Contains(t, upgradedKeys, "todo_issues_parent_fkey")
	delete(upgradedKeys, "todo_issues_parent_fkey")
	assert.Equal(t, existing, upgradedKeys, "the foreign keys that were already there must be left as they were, not recreated")

	again, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err, "reapplying the bundle to an upgraded database should be idempotent")
	require.NoError(t, again.Close())
}

// todoIssueKeyDependents maps each foreign key resting on the (id, workspace_id)
// key of todo_issues to its oid, which changes when a constraint is recreated.
func todoIssueKeyDependents(t *testing.T, db *gorm.DB) map[string]int64 {
	t.Helper()
	var rows []struct {
		Name string
		ID   int64
	}
	require.NoError(t, db.Raw(`
		SELECT conname AS name, oid::bigint AS id FROM pg_constraint
		WHERE contype = 'f' AND conindid = 'public.todo_issues_id_workspace_id_key'::regclass`).Scan(&rows).Error)
	dependents := make(map[string]int64, len(rows))
	for _, row := range rows {
		dependents[row.Name] = row.ID
	}
	return dependents
}

// assertParentRejected demands the named constraint, so a write that fails for
// an unrelated reason cannot pass as parent enforcement.
func assertParentRejected(t *testing.T, err error, code, constraint string) {
	t.Helper()
	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr, constraint)
	assert.Equal(t, code, pgErr.Code, constraint)
	assert.Equal(t, constraint, pgErr.ConstraintName, constraint)
}
