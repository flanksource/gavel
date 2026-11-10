// Tests the platform dashboard API: only platform admins read it, and the tenant list
// merges gavel's organizations with Clerk's, still answering when Clerk can't.
package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/api/clerktest"
	"github.com/flanksource/gavel/platform"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// servePlatform serves p acting for a user of org who is a platform admin when admin is set.
func servePlatform(t *testing.T, p *PlatformAPI, org uuid.UUID, admin bool) string {
	t.Helper()
	handler := p.Handler()
	user := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := tenant.Scope{OrgID: org, UserID: &user, Role: "org:member", PlatformAdmin: admin}
		handler.ServeHTTP(w, r.WithContext(tenant.WithScope(r.Context(), scope)))
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func platformGet(t *testing.T, url string, into any) int {
	t.Helper()
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.NoError(t, json.NewDecoder(resp.Body).Decode(into))
	return resp.StatusCode
}

func TestPlatformAPIIsForPlatformAdmins(t *testing.T) {
	db, _, org := openOrgDB(t)
	p := &PlatformAPI{Store: platform.Store{DB: db, Issuer: "https://clerk.example"}}

	var failure map[string]string
	status := platformGet(t, servePlatform(t, p, org, false)+"/api/platform/tenants", &failure)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "not_platform_admin", failure["error"])
	assert.Contains(t, failure["message"], `gavel_role`, "the message says how to become a platform admin")
}

func TestPlatformTenants(t *testing.T) {
	db, _, gavelOnly := openOrgDB(t)
	alpha := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, slug, name) VALUES (?, 'https://clerk.example', 'org_alpha', 'alpha', 'Alpha')`, alpha).Error)
	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_ada"})
	created := time.UnixMilli(1760000000000).UTC()
	clerk.Organizations = []clerktest.Organization{
		{ID: "org_alpha", Slug: "alpha", Name: "Alpha", Members: 3, CreatedAt: created},
		{ID: "org_beta", Slug: "beta", Name: "Beta", Members: 1, CreatedAt: created},
	}
	p := &PlatformAPI{
		Store: platform.Store{DB: db, Issuer: "https://clerk.example"},
		Clerk: api.ClerkOrganizations{SecretKey: clerktest.SecretKey, HTTPClient: clerk.BackendClient()},
	}
	base := servePlatform(t, p, alpha, true)

	var got PlatformTenants
	require.Equal(t, http.StatusOK, platformGet(t, base+"/api/platform/tenants?days=7", &got))
	assert.Empty(t, got.ClerkError)
	require.Len(t, got.Tenants, 3, "%+v", got.Tenants)
	byClerkID := map[string]PlatformTenant{}
	for _, row := range got.Tenants {
		byClerkID[row.ClerkOrgID] = row
	}
	assert.True(t, byClerkID["org_alpha"].InGavel)
	assert.Equal(t, alpha, byClerkID["org_alpha"].ID)
	require.NotNil(t, byClerkID["org_alpha"].ClerkMembers)
	assert.Equal(t, 3, *byClerkID["org_alpha"].ClerkMembers, "Clerk's member count, besides those who used gavel")
	assert.Equal(t, PlatformTenant{Tenant: platform.Tenant{ClerkOrgID: "org_beta", Slug: "beta", Name: "Beta", CreatedAt: created}, ClerkMembers: ptr(1)},
		byClerkID["org_beta"], "an organization that never used gavel")
	assert.True(t, byClerkID[gavelOnly.String()].InGavel, "an organization Clerk no longer lists")
	assert.Nil(t, byClerkID[gavelOnly.String()].ClerkMembers)

	var failure map[string]string
	assert.Equal(t, http.StatusBadRequest, platformGet(t, base+"/api/platform/tenants?days=90", &failure))
	assert.Equal(t, `days must be a whole number from 1 to 60, got "90"`, failure["error"])
}

func TestPlatformTenantsWithoutClerk(t *testing.T) {
	db, _, org := openOrgDB(t)
	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_ada"})
	p := &PlatformAPI{
		Store: platform.Store{DB: db, Issuer: "https://clerk.example"},
		Clerk: api.ClerkOrganizations{SecretKey: "sk_test_wrong", HTTPClient: clerk.BackendClient()},
	}

	var got PlatformTenants
	require.Equal(t, http.StatusOK, platformGet(t, servePlatform(t, p, org, true)+"/api/platform/tenants", &got))
	require.Len(t, got.Tenants, 1, "gavel's organizations, without Clerk's")
	assert.Equal(t, org, got.Tenants[0].ID)
	assert.Contains(t, got.ClerkError, "check the Clerk secret key")
}

func TestPlatformSandboxRuns(t *testing.T) {
	db, _, org := openOrgDB(t)
	require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, created_at) VALUES (?, ?, 'agent', 'running', 'substrate', now() - interval '2 days')`, uuid.New(), org).Error)
	require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, created_at) VALUES (?, ?, 'test', 'failed', 'substrate', now() - interval '20 days')`, uuid.New(), org).Error)
	p := &PlatformAPI{Store: platform.Store{DB: db, Issuer: "https://clerk.example"}}
	base := servePlatform(t, p, org, true)

	var usage platform.SandboxUsage
	require.Equal(t, http.StatusOK, platformGet(t, base+"/api/platform/sandbox-runs?days=7", &usage))
	require.Len(t, usage.Days, 1, "only the runs in the window: %+v", usage.Days)
	assert.Equal(t, "agent", usage.Days[0].Kind)
	assert.Len(t, usage.Active, 1)

	require.Equal(t, http.StatusOK, platformGet(t, base+"/api/platform/sandbox-runs", &usage))
	assert.Len(t, usage.Days, 2, "the default window is 30 days")

	var failure map[string]string
	assert.Equal(t, http.StatusBadRequest, platformGet(t, base+"/api/platform/sandbox-runs?days=0", &failure))
	assert.Contains(t, failure["error"], "days must be a whole number from 1 to 60")
}
