// Tests Members against a migrated database and the fake Clerk: browser sessions act in
// their active org, CLI calls pick an org they belong to, and memberships follow Clerk.
package api_test

import (
	"context"
	"errors"
	"testing"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/api/clerktest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type membersFixture struct {
	db      *gorm.DB
	clerk   *clerktest.Server
	members *api.Members
}

func newMembersFixture(t *testing.T, maxAge time.Duration) membersFixture {
	t.Helper()
	dsn := commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_members"}).DSN()
	t.Setenv(database.EnvDSN, dsn)
	db, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_ada", Name: "Ada Lovelace", Email: "ada@example.com", GitHubLogin: "ada"})
	clerk.Memberships = []clerktest.Membership{
		{OrgID: "org_a", OrgSlug: "alpha", OrgName: "Alpha", Role: "org:admin"},
		{OrgID: "org_b", OrgSlug: "beta", OrgName: "Beta", Role: "org:member"},
	}
	backend := api.ClerkOrganizations{SecretKey: clerktest.SecretKey, HTTPClient: clerk.BackendClient()}
	return membersFixture{db: db.Gorm(), clerk: clerk, members: api.NewMembers(db.Gorm(), backend, clerk.URL, maxAge)}
}

func (f membersFixture) orgID(t *testing.T, clerkOrgID string) uuid.UUID {
	t.Helper()
	var org struct{ ID uuid.UUID }
	require.NoError(t, f.db.Raw(`SELECT id FROM orgs WHERE clerk_org_id = ?`, clerkOrgID).Scan(&org).Error)
	require.NotEqual(t, uuid.Nil, org.ID)
	return org.ID
}

func (f membersFixture) user(t *testing.T) (id uuid.UUID, name, email, githubLogin string) {
	t.Helper()
	var row struct {
		ID          uuid.UUID
		Name, Email string
		GitHubLogin *string `gorm:"column:github_login"`
	}
	require.NoError(t, f.db.Raw(`SELECT id, name, email, github_login FROM users WHERE clerk_user_id = 'user_ada'`).Scan(&row).Error)
	login := ""
	if row.GitHubLogin != nil {
		login = *row.GitHubLogin
	}
	return row.ID, row.Name, row.Email, login
}

func TestMembersForSession(t *testing.T) {
	f := newMembersFixture(t, 5*time.Minute)
	ctx := context.Background()

	scope, err := f.members.ForSession(ctx, api.SessionClaims{UserID: "user_ada", OrgID: "org_a", OrgSlug: "alpha", OrgRole: "org:admin"})
	require.NoError(t, err)
	userID, name, email, login := f.user(t)
	assert.Equal(t, "Ada Lovelace", name, "a new user's profile comes from Clerk")
	assert.Equal(t, "ada@example.com", email)
	assert.Equal(t, "ada", login)
	assert.Equal(t, tenant.Scope{OrgID: f.orgID(t, "org_a"), UserID: &userID, Role: "org:admin", GitHubLogin: "ada"}, scope)
	requests := len(f.clerk.BackendRequests())

	scope, err = f.members.ForSession(ctx, api.SessionClaims{UserID: "user_ada", OrgID: "org_a", OrgSlug: "alpha", OrgRole: "org:member"})
	require.NoError(t, err)
	assert.Equal(t, "org:member", scope.Role, "the session's org role is authoritative")
	var role string
	require.NoError(t, f.db.Raw(`SELECT role FROM org_memberships WHERE user_id = ? AND org_id = ?`, userID, scope.OrgID).Scan(&role).Error)
	assert.Equal(t, "org:member", role)
	assert.Len(t, f.clerk.BackendRequests(), requests, "a known user's session needs no Clerk call")

	_, err = f.members.ForSession(ctx, api.SessionClaims{UserID: "user_ada"})
	require.ErrorIs(t, err, api.ErrNoActiveOrg)
	assert.Contains(t, err.Error(), "choose an organization")
}

func TestMembersForAccessToken(t *testing.T) {
	f := newMembersFixture(t, 5*time.Minute)
	ctx := context.Background()

	scope, err := f.members.ForAccessToken(ctx, "user_ada", "beta")
	require.NoError(t, err)
	userID, _, _, _ := f.user(t)
	assert.Equal(t, tenant.Scope{OrgID: f.orgID(t, "org_b"), UserID: &userID, Role: "org:member", GitHubLogin: "ada"}, scope)

	byClerkID, err := f.members.ForAccessToken(ctx, "user_ada", "org_a")
	require.NoError(t, err)
	assert.Equal(t, f.orgID(t, "org_a"), byClerkID.OrgID, "the org can be named by its Clerk ID")

	_, err = f.members.ForAccessToken(ctx, "user_ada", "")
	require.ErrorIs(t, err, api.ErrNoActiveOrg)
	assert.Contains(t, err.Error(), "set X-Gavel-Org to one of: alpha, beta")

	_, err = f.members.ForAccessToken(ctx, "user_ada", "gamma")
	require.ErrorIs(t, err, api.ErrNotMember)
	assert.Contains(t, err.Error(), `"gamma"`)

	f.clerk.Memberships = f.clerk.Memberships[:1]
	only, err := f.members.ForAccessToken(ctx, "user_ada", "")
	require.ErrorIs(t, err, api.ErrNoActiveOrg, "memberships are not re-read from Clerk before MaxAge: %+v", only)
}

func TestMembersFollowClerkAfterMaxAge(t *testing.T) {
	f := newMembersFixture(t, 300*time.Millisecond)
	ctx := context.Background()
	_, err := f.members.ForAccessToken(ctx, "user_ada", "beta")
	require.NoError(t, err)

	f.clerk.Memberships = f.clerk.Memberships[:1]
	_, err = f.members.ForAccessToken(ctx, "user_ada", "beta")
	require.NoError(t, err, "still a member until the copy is stale")

	time.Sleep(400 * time.Millisecond)
	_, err = f.members.ForAccessToken(ctx, "user_ada", "beta")
	require.ErrorIs(t, err, api.ErrNotMember, "removed in Clerk")
	scope, err := f.members.ForAccessToken(ctx, "user_ada", "")
	require.NoError(t, err, "a single membership needs no X-Gavel-Org")
	assert.Equal(t, f.orgID(t, "org_a"), scope.OrgID)
}

func TestMembersReportsClerkFailures(t *testing.T) {
	f := newMembersFixture(t, 5*time.Minute)
	broken := api.NewMembers(f.db, api.ClerkOrganizations{SecretKey: "sk_test_wrong", HTTPClient: f.clerk.BackendClient()}, f.clerk.URL, time.Minute)
	_, err := broken.ForAccessToken(context.Background(), "user_ada", "alpha")
	require.Error(t, err)
	assert.False(t, errors.Is(err, api.ErrNotMember) || errors.Is(err, api.ErrNoActiveOrg), "a Clerk outage is not a membership answer: %v", err)
	assert.Contains(t, err.Error(), "check the Clerk secret key")
}

func TestMembersOrgs(t *testing.T) {
	f := newMembersFixture(t, 5*time.Minute)
	orgs, err := f.members.Orgs(context.Background(), "user_ada")
	require.NoError(t, err)
	assert.Equal(t, []api.MemberOrg{
		{ID: f.orgID(t, "org_a"), ClerkOrgID: "org_a", Slug: "alpha", Name: "Alpha", Role: "org:admin"},
		{ID: f.orgID(t, "org_b"), ClerkOrgID: "org_b", Slug: "beta", Name: "Beta", Role: "org:member"},
	}, orgs, "every membership, without picking one")
}
