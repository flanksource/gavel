// Tests /api/auth/config and /api/me: none mode reports the local login without a
// network call, and an organization's dashboard reports the member, their org and orgs.
package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/api/clerktest"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func getJSON(t *testing.T, handler http.HandlerFunc, req *http.Request) map[string]any {
	t.Helper()
	rec := httptest.NewRecorder()
	handler(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body
}

var localOrgJSON = map[string]any{"id": tenant.LocalOrgID.String(), "slug": "local", "name": "Local", "role": ""}

func TestAuthConfig(t *testing.T) {
	srv := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/api/auth/config", nil)
	assert.Equal(t, map[string]any{"mode": "none"}, getJSON(t, srv.handleAuthConfig, req))

	srv.SetAuthConfig(AuthConfigResponse{Mode: "clerk", PublishableKey: "pk_test_abc", Issuer: "https://clerk.example.com", ClientID: "cli"})
	assert.Equal(t, map[string]any{"mode": "clerk", "publishableKey": "pk_test_abc", "issuer": "https://clerk.example.com", "clientId": "cli"},
		getJSON(t, srv.handleAuthConfig, req))
}

func TestMeWithoutALocalLogin(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	body := getJSON(t, (&Server{}).handleMe, httptest.NewRequest(http.MethodGet, "/api/me", nil))
	assert.Equal(t, map[string]any{"user": nil, "org": localOrgJSON, "orgs": []any{localOrgJSON}}, body)
}

func TestMeWithALocalLogin(t *testing.T) {
	config := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", config)
	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_123", Name: "Ada Lovelace", Email: "ada@example.com"})
	session := &api.Session{Config: api.AuthConfig{Issuer: clerk.URL, ClientID: "cli"}, IDToken: clerk.IDToken(time.Now().Add(time.Hour)),
		Token: &oauth2.Token{AccessToken: "access", RefreshToken: "refresh"}}
	require.NoError(t, api.SaveSession(filepath.Join(config, "gavel", "clerk-session.json"), session))
	clerk.Close()

	body := getJSON(t, (&Server{}).handleMe, httptest.NewRequest(http.MethodGet, "/api/me", nil))
	assert.Equal(t, map[string]any{"id": "user_123", "name": "Ada Lovelace", "email": "ada@example.com", "githubLogin": ""}, body["user"])
	assert.Equal(t, localOrgJSON, body["org"])
}

func TestMeForAnOrgMember(t *testing.T) {
	db, _, org := openOrgDB(t)
	var user struct{ ID uuid.UUID }
	require.NoError(t, db.Raw(`INSERT INTO users (issuer, clerk_user_id, name, email, github_login)
		VALUES ('https://clerk.example', 'user_ada', 'Ada Lovelace', 'ada@example.com', 'ada') RETURNING id`).Scan(&user).Error)
	require.NoError(t, db.Exec(`UPDATE orgs SET slug = 'other' WHERE id = ?`, org).Error)
	require.NoError(t, db.Exec(`INSERT INTO org_memberships (user_id, org_id, role) VALUES (?, ?, 'org:admin')`, user.ID, org).Error)

	srv := &Server{}
	srv.SetOrg(org, db)
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req = req.WithContext(tenant.WithScope(req.Context(), tenant.Scope{OrgID: org, UserID: &user.ID, Role: "org:admin", GitHubLogin: "ada"}))
	body := getJSON(t, srv.handleMe, req)

	assert.Equal(t, map[string]any{"id": user.ID.String(), "name": "Ada Lovelace", "email": "ada@example.com", "githubLogin": "ada"}, body["user"])
	current := map[string]any{"id": org.String(), "slug": "other", "name": "Other", "role": "org:admin"}
	assert.Equal(t, current, body["org"])
	assert.Equal(t, []any{current}, body["orgs"])
}
