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
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/github/githubtest"
	"github.com/flanksource/gavel/missioncontrol"
	"github.com/flanksource/gavel/missioncontrol/missioncontroltest"
	"github.com/flanksource/gavel/platform"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/incident-commander/clientapi"
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

func TestPlatformGitHub(t *testing.T) {
	db, _, org := openOrgDB(t)
	gh := githubtest.New(t)
	acme := github.InstallationAccount{Login: "acme", Type: "Organization"}
	gh.Install(7, acme, github.InstallationRepo{ID: 1, FullName: "acme/api"})
	gh.Install(8, github.InstallationAccount{Login: "solo", Type: "User"})
	gh.Install(9, github.InstallationAccount{Login: "flaky", Type: "Organization"})
	suspended := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	gh.Suspend(8, suspended)
	gh.FailRateLimit(9)
	require.NoError(t, db.Exec(`INSERT INTO github_installations (id, org_id, account_login, account_type) VALUES (7, ?, 'acme', 'Organization'), (11, ?, 'gone', 'Organization')`, org, org).Error)
	require.NoError(t, db.Exec(`INSERT INTO repos (org_id, name, source, github_repo_id, installation_id) VALUES (?, 'acme/api', 'github', 1, 7)`, org).Error)
	p := &PlatformAPI{Store: platform.Store{DB: db, Issuer: "https://clerk.example"}, App: gh.App(t)}
	base := servePlatform(t, p, org, true)

	var got PlatformGitHub
	require.Equal(t, http.StatusOK, platformGet(t, base+"/api/platform/github", &got))
	assert.Equal(t, PlatformGitHubApp{ID: githubtest.AppID, Slug: "gavel-test"}, got.App)
	assert.Empty(t, got.GitHubError)
	require.Len(t, got.Installations, 4, "%+v", got.Installations)
	linked, solo, flaky, gone := got.Installations[0], got.Installations[1], got.Installations[2], got.Installations[3]

	assert.Equal(t, "acme", linked.Account)
	require.NotNil(t, linked.Org)
	assert.Equal(t, org, linked.Org.ID)
	assert.Equal(t, 1, linked.Repos)
	require.NotNil(t, linked.RateLimit)
	assert.Equal(t, 5000, linked.RateLimit.Limit)

	assert.Nil(t, solo.Org, "installed on GitHub but linked to no organization")
	require.NotNil(t, solo.SuspendedAt)
	assert.Equal(t, suspended, *solo.SuspendedAt)
	assert.Nil(t, solo.RateLimit, "a suspended installation gets no token")
	assert.Empty(t, solo.RateLimitError)

	assert.Nil(t, flaky.RateLimit)
	assert.Contains(t, flaky.RateLimitError, "installation 9", "one installation's failure leaves the rest")

	assert.Equal(t, int64(11), gone.ID)
	assert.True(t, gone.Uninstalled, "linked in gavel, but GitHub no longer lists it")
	require.NotNil(t, gone.Org)

	t.Setenv("GITHUB_API_URL", "http://127.0.0.1:1")
	require.Equal(t, http.StatusOK, platformGet(t, base+"/api/platform/github", &got))
	assert.Contains(t, got.GitHubError, "cannot reach GitHub")
	require.Len(t, got.Installations, 2, "gavel's linked installations, without GitHub's")
	assert.Equal(t, int64(7), got.Installations[0].ID)
}

func TestPlatformGitHubWithoutTheApp(t *testing.T) {
	db, _, org := openOrgDB(t)
	p := &PlatformAPI{Store: platform.Store{DB: db, Issuer: "https://clerk.example"}}

	var failure map[string]string
	assert.Equal(t, http.StatusNotFound, platformGet(t, servePlatform(t, p, org, true)+"/api/platform/github", &failure))
	assert.Contains(t, failure["error"], "--github-app-id")
}

func TestPlatformSubstrateMetrics(t *testing.T) {
	db, _, org := openOrgDB(t)
	mc := missioncontroltest.New(t)
	mc.SetView("substrate", "capacity", map[string]any{
		"panels": []map[string]any{{"name": "Crashes", "type": "number", "rows": []map[string]any{{"value": 3}}}},
	})
	store := platform.Store{DB: db, Issuer: "https://clerk.example"}
	source := missioncontrol.New(mc.URL, missioncontroltest.Token, "namespace=gavel")

	var failure map[string]string
	unconfigured := servePlatform(t, &PlatformAPI{Store: store}, org, true)
	assert.Equal(t, http.StatusNotFound, platformGet(t, unconfigured+"/api/platform/substrate/metrics", &failure))
	assert.Contains(t, failure["error"], "--mission-control-url")

	noView := servePlatform(t, &PlatformAPI{Store: store, MissionControl: source}, org, true)
	assert.Equal(t, http.StatusNotFound, platformGet(t, noView+"/api/platform/substrate/metrics", &failure))
	assert.Equal(t, "no View graphs substrate's metrics: start gavel serve with --mission-control-substrate-view namespace/name", failure["error"])

	var got missioncontrol.View
	base := servePlatform(t, &PlatformAPI{Store: store, MissionControl: source, SubstrateView: "substrate/capacity"}, org, true)
	require.Equal(t, http.StatusOK, platformGet(t, base+"/api/platform/substrate/metrics", &got))
	require.Len(t, got.Panels, 1)
	assert.Equal(t, 3.0, *got.Panels[0].Value)

	missing := servePlatform(t, &PlatformAPI{Store: store, MissionControl: source, SubstrateView: "substrate/missing"}, org, true)
	assert.Equal(t, http.StatusBadGateway, platformGet(t, missing+"/api/platform/substrate/metrics", &failure))
	assert.Contains(t, failure["error"], "view substrate/missing not found in Mission Control")
}

func TestPlatformInfrastructure(t *testing.T) {
	db, _, org := openOrgDB(t)
	mc := missioncontroltest.New(t)
	mc.SetResources(clientapi.SearchResourcesResponse{Configs: []clientapi.SelectedResource{{ID: "c1", Name: "gavel", Type: "Kubernetes::Deployment", Health: "healthy"}}})
	store := platform.Store{DB: db, Issuer: "https://clerk.example"}

	var failure map[string]string
	unconfigured := servePlatform(t, &PlatformAPI{Store: store}, org, true)
	assert.Equal(t, http.StatusNotFound, platformGet(t, unconfigured+"/api/platform/infrastructure", &failure))
	assert.Contains(t, failure["error"], "--mission-control-url")

	var got missioncontrol.Infrastructure
	base := servePlatform(t, &PlatformAPI{Store: store, MissionControl: missioncontrol.New(mc.URL, missioncontroltest.Token, "namespace=gavel")}, org, true)
	require.Equal(t, http.StatusOK, platformGet(t, base+"/api/platform/infrastructure", &got))
	require.Len(t, got.Configs, 1)
	assert.Equal(t, "gavel", got.Configs[0].Name)

	refused := servePlatform(t, &PlatformAPI{Store: store, MissionControl: missioncontrol.New(mc.URL, "wrong-token", "namespace=gavel")}, org, true)
	assert.Equal(t, http.StatusBadGateway, platformGet(t, refused+"/api/platform/infrastructure", &failure))
	assert.Contains(t, failure["error"], "check GAVEL_MISSION_CONTROL_TOKEN")
}
