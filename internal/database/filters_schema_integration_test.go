package database_test

import (
	"testing"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/internal/database/dbtest"
	"github.com/flanksource/gavel/tenant"
	"github.com/stretchr/testify/require"
)

// The local org's filters have a null owner; the unique key must still allow only
// one row per name for it, and a re-run migration must leave the table alone.
func TestUserFiltersSchemaTreatsNullOwnerAsOne(t *testing.T) {
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_user_filters"}).DSN())
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	first, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	require.NoError(t, first.Close())
	db, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	insert := `INSERT INTO user_filters (org_id, name, kind, "values") VALUES (?, 'hidden-projects', ?, '{}')`
	require.NoError(t, db.Gorm().Exec(insert, tenant.LocalOrgID, "workspace").Error)
	dbtest.RequireConstraintViolation(t, dbtest.Quiet(db.Gorm()).Exec(insert, tenant.LocalOrgID, "workspace").Error, "user_filters_owner_name_key")
	dbtest.RequireConstraintViolation(t, dbtest.Quiet(db.Gorm()).Exec(
		`INSERT INTO user_filters (org_id, name, kind) VALUES (?, 'hidden-labels', 'label')`, tenant.LocalOrgID).Error, "user_filters_kind_check")
}
