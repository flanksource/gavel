// Tests linking a GitHub App installation to an organization: the install redirect, the
// setup callback's checks (state, user, access, ownership), and repository sync.
package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/api/clerktest"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/github/githubtest"
	"github.com/flanksource/gavel/github/installations"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubLink(t *testing.T) {
	db, _, org := openOrgDB(t)
	ada, otherOrg := uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', 'user_123')`, ada).Error)
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', 'org_other', 'Elsewhere')`, otherOrg).Error)

	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_123"})
	clerk.GitHubToken = "ghu_ada"
	gh := githubtest.New(t)
	acme := github.InstallationAccount{Login: "acme", Type: "Organization"}
	gh.Install(42, acme, github.InstallationRepo{ID: 1, FullName: "acme/api", DefaultBranch: "main"})
	gh.Install(43, acme)
	gh.Install(99, github.InstallationAccount{Login: "stranger", Type: "User"})
	gh.GrantUser("ghu_ada", 42, 43)
	app := gh.App(t)
	store := installations.Store{DB: db, App: app}
	key := []byte(strings.Repeat("k", 32))
	link := &GitHubLink{App: app, Store: store, Clerk: api.ClerkOrganizations{SecretKey: clerktest.SecretKey, HTTPClient: clerk.BackendClient()}, SecretKey: key}
	handler := link.Handler()

	serve := func(method, target string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, target, nil)
		req = req.WithContext(tenant.WithScope(req.Context(), tenant.Scope{OrgID: org, UserID: &ada}))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}
	setup := func(installation, state string) *httptest.ResponseRecorder {
		t.Helper()
		return serve(http.MethodGet, "/api/github/setup?setup_action=install&installation_id="+installation+"&state="+url.QueryEscape(state))
	}
	state := installations.SignState(key, org, ada, time.Now())

	t.Run("install sends the user to GitHub with a signed state", func(t *testing.T) {
		rec := serve(http.MethodGet, "/api/github/install")
		require.Equal(t, http.StatusFound, rec.Code)
		target, err := url.Parse(rec.Header().Get("Location"))
		require.NoError(t, err)
		assert.Equal(t, "https://github.com/apps/gavel-test/installations/new", target.Scheme+"://"+target.Host+target.Path)
		got, err := installations.VerifyState(key, target.Query().Get("state"), time.Now())
		require.NoError(t, err)
		assert.Equal(t, org, got.Org)
		assert.Equal(t, ada, got.User)
	})

	t.Run("setup refuses what it can't trust", func(t *testing.T) {
		for name, test := range map[string]struct {
			rec    *httptest.ResponseRecorder
			status int
			want   string
		}{
			"bad state":       {setup("42", "forged"), http.StatusBadRequest, "isn't valid"},
			"another user":    {setup("42", installations.SignState(key, org, uuid.New(), time.Now())), http.StatusForbidden, "started by another user or organization"},
			"another org":     {setup("42", installations.SignState(key, otherOrg, ada, time.Now())), http.StatusForbidden, "started by another user or organization"},
			"no installation": {serve(http.MethodGet, "/api/github/setup?setup_action=request&state="+url.QueryEscape(state)), http.StatusBadRequest, "approve"},
			"not the user's":  {setup("99", state), http.StatusForbidden, "your GitHub account can't access installation 99"},
			"not a number":    {setup("abc", state), http.StatusBadRequest, `abc`},
		} {
			assert.Equal(t, test.status, test.rec.Code, name)
			assert.Contains(t, test.rec.Body.String(), test.want, name)
		}
		ids, err := store.Installations(t.Context(), org)
		require.NoError(t, err)
		assert.Empty(t, ids, "nothing was linked")
	})

	t.Run("setup links the installation, syncs its repositories and returns to the dashboard", func(t *testing.T) {
		rec := setup("42", state)
		require.Equal(t, http.StatusFound, rec.Code, rec.Body.String())
		assert.Equal(t, "/", rec.Header().Get("Location"))
		repos, err := store.Repos(t.Context(), org)
		require.NoError(t, err)
		assert.Equal(t, []string{"acme/api"}, repos)
	})

	t.Run("setup refuses an installation another org linked", func(t *testing.T) {
		require.NoError(t, store.Link(t.Context(), otherOrg, ada, github.Installation{ID: 43, Account: acme}))
		rec := setup("43", state)
		assert.Equal(t, http.StatusConflict, rec.Code)
		assert.Contains(t, rec.Body.String(), "linked to another organization")
	})

	t.Run("sync re-reads every installation's repositories", func(t *testing.T) {
		gh.Install(42, acme, github.InstallationRepo{ID: 1, FullName: "acme/api", DefaultBranch: "main"}, github.InstallationRepo{ID: 2, FullName: "acme/web", DefaultBranch: "main"})
		rec := serve(http.MethodPost, "/api/github/sync")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var body struct{ Repositories []string }
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, []string{"acme/api", "acme/web"}, body.Repositories)
	})
}
