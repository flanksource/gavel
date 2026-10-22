// Tests that a todo's visibility and creator reach its TODO, that create and edit requests set
// its visibility, and that a link to another member's private todo names it without revealing it.
package runtime_test

import (
	"context"
	"testing"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/todos"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/flanksource/gavel/todos/types"
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

func TestLinksNameAnotherMembersPrivateTodoWithoutRevealingIt(t *testing.T) {
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_runtime_private_links"}).DSN())
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	migrated, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, migrated.Close()) })
	db := migrated.Gorm()
	org, ada, grace := uuid.New(), uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id, name) VALUES (?, 'https://clerk.example', 'user_ada', 'Ada Lovelace'), (?, 'https://clerk.example', 'user_grace', 'Grace Hopper')`, ada, grace).Error)
	root := t.TempDir()
	open := func(user uuid.UUID) (*todoruntime.Provider, context.Context) {
		scope := tenant.Scope{OrgID: org, UserID: &user, Role: "org:member"}
		ctx := tenant.WithScope(t.Context(), scope)
		provider, err := todoruntime.NewForOrg(ctx, db, todoruntime.WorkspaceOptions{Name: "app", RootPath: root}, scope)
		require.NoError(t, err)
		return provider, ctx
	}

	graces, graceCtx := open(grace)
	blocker, err := graces.Create(graceCtx, todos.CreateRequest{Title: "Rotate the signing key"})
	require.NoError(t, err)
	blocked, err := graces.Create(graceCtx, todos.CreateRequest{Title: "Ship the release", Visibility: "public"})
	require.NoError(t, err)
	_, err = graces.Link(graceCtx, blocked, blocker.ID, types.RelationDependsOn)
	require.NoError(t, err)

	adas, adaCtx := open(ada)
	shared, err := adas.Get(adaCtx, blocked.ID)
	require.NoError(t, err)
	links, err := adas.Links(adaCtx, shared)
	require.NoError(t, err, "a link to a private todo doesn't stop the public one's links from loading")
	require.Len(t, links, 1)
	assert.Equal(t, types.RelationDependsOn, links[0].Relation)
	assert.True(t, links[0].TargetPrivate)
	assert.Equal(t, "Grace Hopper", links[0].TargetOwner)
	assert.Equal(t, blocker.ID[:8], links[0].TargetShortID)
	assert.Empty(t, links[0].TargetTitle, "the private todo's title stays hidden")
	assert.Empty(t, links[0].TargetStatus)

	own, err := graces.Links(graceCtx, blocked)
	require.NoError(t, err)
	require.Len(t, own, 1)
	assert.False(t, own[0].TargetPrivate, "its creator sees the todo itself")
	assert.Equal(t, "Rotate the signing key", own[0].TargetTitle)
}
