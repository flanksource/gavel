// Tests the platform dashboard API: only platform admins read it, and the tenant list
// merges gavel's organizations with Clerk's, still answering when Clerk can't.
package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/captain/pkg/substrate/substratetest"
	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/api/clerktest"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/github/githubtest"
	"github.com/flanksource/gavel/missioncontrol"
	"github.com/flanksource/gavel/missioncontrol/missioncontroltest"
	"github.com/flanksource/gavel/platform"
	"github.com/flanksource/gavel/sandbox"
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

func TestPlatformWindowsStartAtUTCMidnight(t *testing.T) {
	// The panels draw whole UTC days, so the oldest day's bar must count that whole day.
	start, err := windowStart(httptest.NewRequest(http.MethodGet, "/api/platform/sandbox-runs?days=7", nil), 30)
	require.NoError(t, err)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	assert.Equal(t, today.AddDate(0, 0, -7), start)

	start, err = windowStart(httptest.NewRequest(http.MethodGet, "/api/platform/sandbox-runs", nil), 30)
	require.NoError(t, err)
	assert.Equal(t, today.AddDate(0, 0, -30), start, "the fallback window")
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

// clerkOrganizationLists counts the Backend API's organization list requests clerk served.
func clerkOrganizationLists(clerk *clerktest.Server) int {
	n := 0
	for _, request := range clerk.BackendRequests() {
		if strings.HasPrefix(request, "GET /v1/organizations?") {
			n++
		}
	}
	return n
}

func TestPlatformTenantsReadClerksOrganizationsOncePerMinute(t *testing.T) {
	db, _, org := openOrgDB(t)
	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_ada"})
	clerk.Organizations = []clerktest.Organization{{ID: "org_beta", Slug: "beta", Name: "Beta", Members: 1}}
	now := time.Now()
	p := &PlatformAPI{
		Store: platform.Store{DB: db, Issuer: "https://clerk.example"},
		Clerk: api.ClerkOrganizations{SecretKey: clerktest.SecretKey, HTTPClient: clerk.BackendClient()},
		now:   func() time.Time { return now },
	}
	base := servePlatform(t, p, org, true)
	load := func() PlatformTenants {
		t.Helper()
		var got PlatformTenants
		require.Equal(t, http.StatusOK, platformGet(t, base+"/api/platform/tenants", &got))
		require.Empty(t, got.ClerkError)
		return got
	}

	load()
	now = now.Add(59 * time.Second)
	assert.Len(t, load().Tenants, 2, "the cached list still merges in Clerk's organizations")
	assert.Equal(t, 1, clerkOrganizationLists(clerk), "a load within a minute reuses Clerk's list")

	now = now.Add(2 * time.Second)
	load()
	assert.Equal(t, 2, clerkOrganizationLists(clerk), "a load after a minute reads Clerk again")
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

	require.Equal(t, http.StatusOK, platformGet(t, servePlatform(t, p, org, true)+"/api/platform/tenants", &got))
	assert.Contains(t, got.ClerkError, "check the Clerk secret key")
	assert.Equal(t, 2, clerkOrganizationLists(clerk), "a failed list isn't kept: the next load asks Clerk again")
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

func TestPlatformModelUsage(t *testing.T) {
	db, _, org := openOrgDB(t)
	session, turn := uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO captain_sessions (id, source) VALUES (?, 'test')`, session).Error)
	require.NoError(t, db.Exec(`INSERT INTO captain_turns (id, session_id, turn_index, status) VALUES (?, ?, 0, 'ended')`, turn, session).Error)
	for i, age := range []string{"2 days", "20 days"} {
		require.NoError(t, db.Exec(`INSERT INTO captain_model_calls (id, turn_id, call_index, model, currency, input_tokens, input_cost, created_at)
			VALUES (?, ?, ?, 'claude-sonnet', 'USD', 10, 0.5, now() - ?::interval)`, uuid.New(), turn, i, age).Error)
	}
	p := &PlatformAPI{Store: platform.Store{DB: db, Issuer: "https://clerk.example"}}
	base := servePlatform(t, p, org, true)

	var usage PlatformModelUsage
	require.Equal(t, http.StatusOK, platformGet(t, base+"/api/platform/model-usage?days=7", &usage))
	require.Len(t, usage.Days, 1, "only the calls in the window: %+v", usage.Days)
	assert.Equal(t, 0.5, usage.Days[0].CostUSD)

	require.Equal(t, http.StatusOK, platformGet(t, base+"/api/platform/model-usage", &usage))
	assert.Len(t, usage.Days, 2, "the default window is 30 days")

	var failure map[string]string
	assert.Equal(t, http.StatusBadRequest, platformGet(t, base+"/api/platform/model-usage?days=61", &failure))
	assert.Contains(t, failure["error"], "days must be a whole number from 1 to 60")
}

func TestPlatformCaptainHistory(t *testing.T) {
	db, _, org := openOrgDB(t)
	session := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO captain_sessions (id, source) VALUES (?, 'test')`, session).Error)
	for _, age := range []string{"2 days", "20 days"} {
		require.NoError(t, db.Exec(`INSERT INTO captain_prompt_runs (id, session_id, root_session_id, state, error, queued_at, created_at)
			VALUES (?, ?, ?, 'failed', 'boom', now() - ?::interval, now() - ?::interval)`, uuid.New(), session, session, age, age).Error)
	}
	p := &PlatformAPI{Store: platform.Store{DB: db, Issuer: "https://clerk.example"}}
	base := servePlatform(t, p, org, true)

	var history platform.CaptainHistory
	require.Equal(t, http.StatusOK, platformGet(t, base+"/api/platform/captain-history?days=7", &history))
	require.Len(t, history.Days, 1, "only the runs in the window: %+v", history.Days)
	require.Len(t, history.Failures, 1)
	assert.Equal(t, "boom", history.Failures[0].Error)

	require.Equal(t, http.StatusOK, platformGet(t, base+"/api/platform/captain-history", &history))
	assert.Len(t, history.Failures, 2, "the default window is 30 days")

	var failure map[string]string
	assert.Equal(t, http.StatusBadRequest, platformGet(t, base+"/api/platform/captain-history?days=x", &failure))
	assert.Contains(t, failure["error"], "days must be a whole number from 1 to 60")
}

func TestPlatformBudgets(t *testing.T) {
	db, _, org := openOrgDB(t)
	require.NoError(t, db.Exec(`INSERT INTO captain_budget_rules (id, name, amount, "window") VALUES (?, 'Monthly', 100, 'now/M')`, uuid.New()).Error)
	p := &PlatformAPI{Store: platform.Store{DB: db, Issuer: "https://clerk.example"}}

	var got PlatformBudgets
	require.Equal(t, http.StatusOK, platformGet(t, servePlatform(t, p, org, true)+"/api/platform/budgets", &got))
	require.Len(t, got.Budgets, 1)
	assert.Equal(t, "Monthly", got.Budgets[0].Name)
	assert.NotNil(t, got.Budgets[0].WindowStart)
}

func TestPlatformSubstrateInventory(t *testing.T) {
	db, _, org := openOrgDB(t)
	run := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, created_at) VALUES (?, ?, 'agent', 'running', 'substrate', now())`, run, org).Error)
	server := substratetest.Start(t)
	worker := substrate.Worker{Name: "pool-a-0", Pool: "pool-a", Node: "node-1", State: "active",
		Capacity: substrate.Resources{CPUMillicores: 4000, MemoryBytes: 8 << 30, Actors: 10}}
	server.AddWorker(worker)
	for atespace, name := range map[string]string{"ate-env": sandbox.JobName(run), "other": "default-template-0"} {
		options := server.Options()
		options.Atespace = atespace
		client, err := substrate.Dial(options)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, client.Close()) })
		_, err = client.Create(t.Context(), substrate.Environment{ID: name})
		require.NoError(t, err)
	}
	client, err := substrate.Dial(server.Options())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	p := &PlatformAPI{Store: platform.Store{DB: db, Issuer: "https://clerk.example"}, Substrate: client}

	var got PlatformSubstrateInventory
	require.Equal(t, http.StatusOK, platformGet(t, servePlatform(t, p, org, true)+"/api/platform/substrate/inventory", &got))
	assert.Equal(t, []substrate.Worker{worker}, got.Workers)
	require.Len(t, got.Actors, 2)
	assert.Equal(t, sandbox.JobName(run), got.Actors[0].Name)
	assert.Equal(t, &run, got.Actors[0].Run)
	assert.Equal(t, &platform.RunOrg{Org: org, OrgName: "Other", Kind: "agent"}, got.Actors[0].Org)
	assert.Equal(t, "default-template-0", got.Actors[1].Name)
	assert.Nil(t, got.Actors[1].Run, "an actor that isn't a run's environment names no run")
	assert.Empty(t, got.WorkersError)
	assert.Empty(t, got.ActorsError)
}

func TestPlatformSubstrateInventoryListsActorsByAtespaceThenName(t *testing.T) {
	db, _, org := openOrgDB(t)
	server := substratetest.Start(t)
	for _, atespace := range []string{"a-b", "a"} {
		options := server.Options()
		options.Atespace = atespace
		client, err := substrate.Dial(options)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, client.Close()) })
		_, err = client.Create(t.Context(), substrate.Environment{ID: "env"})
		require.NoError(t, err)
	}
	client, err := substrate.Dial(server.Options())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	p := &PlatformAPI{Store: platform.Store{DB: db, Issuer: "https://clerk.example"}, Substrate: client}

	var got PlatformSubstrateInventory
	require.Equal(t, http.StatusOK, platformGet(t, servePlatform(t, p, org, true)+"/api/platform/substrate/inventory", &got))
	var order []string
	for _, actor := range got.Actors {
		order = append(order, actor.Atespace)
	}
	assert.Equal(t, []string{"a", "a-b"}, order)
}

func TestPlatformSubstrateInventoryWhenSubstrateIsDown(t *testing.T) {
	db, _, org := openOrgDB(t)
	options := substratetest.Start(t).Options()
	options.Control = "127.0.0.1:1"
	client, err := substrate.Dial(options)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	p := &PlatformAPI{Store: platform.Store{DB: db, Issuer: "https://clerk.example"}, Substrate: client}

	var failure map[string]string
	assert.Equal(t, http.StatusBadGateway, platformGet(t, servePlatform(t, p, org, true)+"/api/platform/substrate/inventory", &failure))
	assert.Contains(t, failure["error"], "list substrate's workers")
	assert.Contains(t, failure["error"], "list the actors of every atespace")
}

func TestPlatformSubstrateInventoryWithoutSubstrate(t *testing.T) {
	db, _, org := openOrgDB(t)
	p := &PlatformAPI{Store: platform.Store{DB: db, Issuer: "https://clerk.example"}}

	var failure map[string]string
	assert.Equal(t, http.StatusNotFound, platformGet(t, servePlatform(t, p, org, true)+"/api/platform/substrate/inventory", &failure))
	assert.Contains(t, failure["error"], "--sandbox substrate")
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
