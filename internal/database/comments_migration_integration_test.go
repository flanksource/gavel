package database_test

import (
	"context"
	"os"
	"slices"
	"testing"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/commons-db/migrate"
	"github.com/flanksource/gavel/comments"
	commentpostgres "github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/internal/database"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	editorTenant = "gavel-local-editor"
	localTenant  = "00000000-0000-0000-0000-000000000001"
)

type tenantAccess struct{ tenant string }

func (a tenantAccess) Principal(context.Context) (comments.Principal, error) {
	return comments.Principal{TenantID: a.tenant, Actor: comments.Actor{Kind: "user", ID: "reviewer"}}, nil
}
func (tenantAccess) Authorize(context.Context, comments.Principal, string, comments.Target) error {
	return nil
}
func (tenantAccess) Resolve(_ context.Context, _ comments.Principal, t comments.Target) (comments.Target, error) {
	return t, nil
}

// seedThread writes a thread with a reply and a read marker for tenant under scope.
func seedThread(t *testing.T, db *gorm.DB, tenant, scope string) {
	t.Helper()
	store, err := commentpostgres.New(db)
	require.NoError(t, err)
	access := tenantAccess{tenant: tenant}
	svc, err := comments.New(comments.Options{Store: store, Access: access, Targets: access, Workflows: []comments.Workflow{comments.HumanReview()}})
	require.NoError(t, err)
	c, err := svc.Create(t.Context(), comments.CreateRequest{Target: comments.Target{URI: scope + "/document", ScopeURI: scope, Kind: "document"}, Body: "Review", WorkflowID: "human-review", RequestID: "seed-" + tenant})
	require.NoError(t, err)
	c, err = svc.Act(t.Context(), comments.ActionRequest{ID: c.ID, ExpectedVersion: c.Version, Action: "reply", Body: "Reviewed"})
	require.NoError(t, err)
	_, err = svc.MarkRead(t.Context(), c.ID, c.Version)
	require.NoError(t, err)
}

func commentColumns(t *testing.T, db *gorm.DB) []string {
	t.Helper()
	var columns []string
	require.NoError(t, db.Raw(`SELECT table_name || '.' || column_name || ':' || data_type || ':' || is_nullable
		FROM information_schema.columns WHERE table_schema = 'public' AND table_name LIKE 'comment%'
		ORDER BY table_name, column_name`).Scan(&columns).Error)
	return columns
}

func tenantRows(t *testing.T, db *gorm.DB) map[string]int64 {
	t.Helper()
	rows := map[string]int64{}
	for _, table := range []string{"comments", "comment_messages", "comment_events", "comment_reads", "comment_sequences"} {
		for name, tenant := range map[string]string{"editor": editorTenant, "local": localTenant} {
			var count int64
			require.NoError(t, db.Table(table).Where("tenant_id = ?", tenant).Count(&count).Error)
			rows[table+" "+name] = count
		}
	}
	return rows
}

// A database provisioned by the retired standalone comments module (Atlas name
// gavel-comments) preserves the existing comment columns while adding delivery
// tables, and the pre-project editor tenant is dropped while project threads stay.
func TestCommentsMigrateIntoGavelSchemaAndDropEditorTenant(t *testing.T) {
	dsn := dbtest.ForT(t, dbtest.Options{Name: "gavel_comments_migrate"}).DSN()
	require.NoError(t, migrate.Apply(t.Context(), dsn, os.DirFS("testdata/legacy_comments"), migrate.WithDir("schema"), migrate.WithName("gavel-comments")))
	db, err := commonsdb.NewGorm(dsn, commonsdb.DefaultGormConfig())
	require.NoError(t, err)
	seedThread(t, db, editorTenant, "gavel://edit/file/0123")
	seedThread(t, db, localTenant, "gavel://project/demo")
	before := commentColumns(t, db)

	t.Setenv("HOME", t.TempDir())
	require.NoError(t, database.Migrate(t.Context(), dsn))
	require.NoError(t, database.Migrate(t.Context(), dsn), "migrating twice is a no-op")

	expected := append(before,
		"comment_agent_sessions.group_key:text:NO",
		"comment_agent_sessions.recorded_at:timestamp with time zone:NO",
		"comment_agent_sessions.scope_uri:text:NO",
		"comment_agent_sessions.session_id:text:NO",
		"comment_agent_sessions.tenant_id:text:NO",
		"comment_routes.id:text:NO",
		"comment_routes.payload:jsonb:NO",
		"comment_routes.scope_uri:text:NO",
		"comment_routes.target_prefix:text:NO",
		"comment_routes.tenant_id:text:NO",
	)
	slices.Sort(expected)
	require.Equal(t, expected, commentColumns(t, db))
	require.Equal(t, map[string]int64{
		"comments editor": 0, "comments local": 1,
		"comment_messages editor": 0, "comment_messages local": 1,
		"comment_events editor": 0, "comment_events local": 2,
		"comment_reads editor": 0, "comment_reads local": 1,
		"comment_sequences editor": 0, "comment_sequences local": 1,
	}, tenantRows(t, db))
}

// A fresh database gets the comment tables from gavel's own migration.
func TestCommentsTablesComeFromGavelMigration(t *testing.T) {
	dsn := dbtest.ForT(t, dbtest.Options{Name: "gavel_comments_fresh"}).DSN()
	t.Setenv("HOME", t.TempDir())
	require.NoError(t, database.Migrate(t.Context(), dsn))
	db, err := commonsdb.NewGorm(dsn, commonsdb.DefaultGormConfig())
	require.NoError(t, err)
	seedThread(t, db, localTenant, "gavel://project/demo")
	require.Equal(t, int64(1), tenantRows(t, db)["comments local"])
}
