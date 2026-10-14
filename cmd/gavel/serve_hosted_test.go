// Tests `gavel serve --auth clerk` end to end: the real handler stack against a migrated
// database and the fake Clerk, from the public shell to an org's settings.
package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/api/clerktest"
	ghcache "github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const hostedPublicURL = "https://gavel.example.com"

func TestHostedDashboard(t *testing.T) {
	dsn := commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_serve_hosted"}).DSN()
	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(ghcache.EnvDisable, "")
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	cache, err := ghcache.Open()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, cache.Close()) })

	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_ada", Name: "Ada Lovelace", Email: "ada@example.com",
		OrgID: "org_a", OrgSlug: "alpha", OrgRole: "org:admin", GitHubLogin: "ada"})
	clerk.Memberships = []clerktest.Membership{
		{OrgID: "org_a", OrgSlug: "alpha", OrgName: "Alpha", Role: "org:admin"},
		{OrgID: "org_b", OrgSlug: "beta", OrgName: "Beta", Role: "org:member"},
	}
	originalAuth, originalBackend := authHTTPClient, clerkBackendHTTPClient
	authHTTPClient, clerkBackendHTTPClient = clerk.Client(), clerk.BackendClient()
	t.Cleanup(func() { authHTTPClient, clerkBackendHTTPClient = originalAuth, originalBackend })

	opts := DashboardOptions{
		PRListOptions:       PRListOptions{Interval: "60s"},
		Auth:                authClerk,
		PublicURL:           hostedPublicURL,
		ClerkIssuer:         clerk.URL,
		ClerkPublishableKey: "pk_test_abc",
		ClerkClientID:       "cli",
		clerkSecretKey:      clerktest.SecretKey,
		secretKey:           base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))),
	}
	require.NoError(t, opts.Validate())
	handler, stop, err := hostedDashboard(t.Context(), opts, opened.Gorm(), cache)
	require.NoError(t, err)
	t.Cleanup(stop)
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	session := clerk.SessionToken(hostedPublicURL, time.Now().Add(time.Hour))
	do := func(method, path, body string, header http.Header) (int, string) {
		t.Helper()
		req, err := http.NewRequest(method, server.URL+path, strings.NewReader(body))
		require.NoError(t, err)
		for name, values := range header {
			req.Header[name] = values
		}
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		return resp.StatusCode, string(data)
	}
	cookie := http.Header{"Cookie": {"__session=" + session}}

	status, body := do(http.MethodGet, "/prs", "", nil)
	assert.Equal(t, http.StatusOK, status, "the page shell is public")
	assert.Contains(t, body, "<html")

	status, body = do(http.MethodGet, "/api/auth/config", "", nil)
	require.Equal(t, http.StatusOK, status)
	assert.JSONEq(t, `{"mode":"clerk","publishableKey":"pk_test_abc","issuer":"`+clerk.URL+`","clientId":"cli"}`, body)

	status, _ = do(http.MethodGet, "/api/me", "", nil)
	assert.Equal(t, http.StatusUnauthorized, status)

	status, body = do(http.MethodGet, "/api/me", "", cookie)
	require.Equal(t, http.StatusOK, status, body)
	var me ui.MeResponse
	require.NoError(t, json.Unmarshal([]byte(body), &me))
	require.NotNil(t, me.User)
	assert.Equal(t, "Ada Lovelace", me.User.Name)
	assert.Equal(t, "alpha", me.Org.Slug)
	assert.Len(t, me.Orgs, 2)

	status, body = do(http.MethodGet, "/api/me", "", http.Header{
		"Authorization": {"Bearer " + clerk.AccessToken(time.Now().Add(time.Hour))}, "X-Gavel-Org": {"beta"}})
	require.Equal(t, http.StatusOK, status, body)
	require.NoError(t, json.Unmarshal([]byte(body), &me))
	assert.Equal(t, "beta", me.Org.Slug, "the CLI picks its org with X-Gavel-Org")

	status, _ = do(http.MethodGet, "/api/activity", "", cookie)
	assert.Equal(t, http.StatusNotFound, status, "routes off the hosted allowlist")

	status, body = do(http.MethodPost, "/api/config", `{"repos":["acme/app"]}`,
		http.Header{"Cookie": cookie["Cookie"], "Origin": {hostedPublicURL}, "Content-Type": {"application/json"}})
	require.Equal(t, http.StatusOK, status, body)
	var alpha struct{ ID uuid.UUID }
	require.NoError(t, opened.Gorm().Raw(`SELECT id FROM orgs WHERE clerk_org_id = 'org_a'`).Scan(&alpha).Error)
	saved, err := ui.LoadOrgSearchConfig(t.Context(), opened.Gorm(), alpha.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"acme/app"}, saved.Repos)
}
