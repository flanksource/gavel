// Tests that a todo's visibility and creator reach its TODO, and that create and edit
// requests set its visibility.
package runtime_test

import (
	"testing"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/todos"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTodoVisibilityRoundTrips(t *testing.T) {
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_runtime_visibility"}).DSN())
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	migrated, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, migrated.Close()) })
	db := migrated.Gorm()
	org, user := uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', 'user_ada')`, user).Error)
	scope := tenant.Scope{OrgID: org, UserID: &user, Role: "org:member"}
	ctx := tenant.WithScope(t.Context(), scope)

	provider, err := todoruntime.NewForOrg(ctx, db, todoruntime.WorkspaceOptions{Name: "app", RootPath: t.TempDir()}, scope)
	require.NoError(t, err)

	private, err := provider.Create(ctx, todos.CreateRequest{Title: "Private by default"})
	require.NoError(t, err)
	assert.Equal(t, "private", private.Visibility)
	assert.Equal(t, user.String(), private.CreatedBy)

	public, err := provider.Create(ctx, todos.CreateRequest{Title: "Shared", Visibility: "public"})
	require.NoError(t, err)
	assert.Equal(t, "public", public.Visibility)

	visibility := "private"
	require.NoError(t, provider.Edit(ctx, public, todos.EditRequest{Visibility: &visibility}))
	reloaded, err := provider.Get(ctx, public.ID)
	require.NoError(t, err)
	assert.Equal(t, "private", reloaded.Visibility)

	_, err = provider.Create(ctx, todos.CreateRequest{Title: "Odd", Visibility: "secret"})
	require.EqualError(t, err, `invalid native todo input: visibility "secret" is not private or public`)
}
