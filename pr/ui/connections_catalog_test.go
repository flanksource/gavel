// Tests the connection catalog the Settings connections page reads and edits: a hosted org's
// form offers only git connections with literal values, members see them, and admins change them.
package ui

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// connectionsPageCall sends body to path as user in role, with headers, and returns the status and body.
func connectionsPageCall(t *testing.T, base, method, path string, user uuid.UUID, role string, body any, headers map[string]string) (int, string) {
	t.Helper()
	var payload bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&payload).Encode(body))
	}
	req, err := http.NewRequest(method, base+path, &payload)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Test-User", user.String())
	req.Header.Set("X-Test-Role", role)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(data)
}

func TestAHostedOrgsConnectionsPageKeepsOnlyItsGitConnections(t *testing.T) {
	db, _, org := openOrgDB(t)
	t.Setenv("GAVEL_SECRET_KEEPER_CONNECTION", "")
	t.Setenv("GAVEL_SECRET_KEY", base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("h"), 32)))
	admin, member := uuid.New(), uuid.New()
	for name, id := range map[string]uuid.UUID{"admin": admin, "member": member} {
		require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, "user_"+name).Error)
	}
	srv := &Server{}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)
	changed := 0
	srv.SetConnectionsChanged(func() { changed++ })
	base := serveActingAs(t, srv, org)

	status, schema := connectionsPageCall(t, base, http.MethodGet, "/api/catalog/connection", member, "org:member", nil, map[string]string{"Accept": "application/schema+json"})
	require.Equal(t, http.StatusOK, status, schema)
	assert.Contains(t, schema, `"azure_devops"`)
	assert.Contains(t, schema, `"github"`)
	assert.NotContains(t, schema, `"postgres"`, "the form offers only git connections")

	acme := map[string]any{"name": "acme", "type": "github", "password": "gh_hosted", "properties": map[string]any{"organization": "acme"}}
	status, body := connectionsPageCall(t, base, http.MethodPost, "/api/catalog/connection", member, "org:member", acme, nil)
	assert.Equal(t, http.StatusForbidden, status, "a member can't add a connection: %s", body)
	status, body = connectionsPageCall(t, base, http.MethodPost, "/api/catalog/connection", admin, "org:admin", acme, nil)
	require.Equal(t, http.StatusOK, status, body)
	assert.NotContains(t, body, "gh_hosted")
	assert.Equal(t, 1, changed, "saving a connection syncs the repositories it brings")
	var created struct{ ID string }
	require.NoError(t, json.Unmarshal([]byte(body), &created))
	require.NotEmpty(t, created.ID)

	status, body = connectionsPageCall(t, base, http.MethodPost, "/api/catalog/connection", admin, "org:admin",
		map[string]any{"name": "warehouse", "type": "postgres", "url": "postgres://10.0.0.5/app"}, nil)
	assert.Equal(t, http.StatusBadRequest, status, body)
	assert.Contains(t, body, "not postgres")

	status, dashboard := connectionsPageCall(t, base, http.MethodGet, "/api/catalog/connections/dashboard", member, "org:member", nil, nil)
	require.Equal(t, http.StatusOK, status, dashboard)
	assert.Contains(t, dashboard, `"acme"`, "a member sees the org's connections")
	assert.NotContains(t, dashboard, "gh_hosted")
	assert.NotContains(t, dashboard, `"readOnly":true`, "an admin can change them from the page")

	status, body = connectionsPageCall(t, base, http.MethodGet, "/api/catalog/connection/"+created.ID, member, "org:member", nil, nil)
	require.Equal(t, http.StatusOK, status, body)
	assert.Contains(t, body, `"acme"`)
	assert.NotContains(t, body, "gh_hosted")

	acme["properties"] = map[string]any{"organization": "acme-labs"}
	status, body = connectionsPageCall(t, base, http.MethodPut, "/api/catalog/connection/"+created.ID, admin, "org:admin", acme, nil)
	require.Equal(t, http.StatusOK, status, body)
	assert.Contains(t, body, "acme-labs")
	assert.Equal(t, 2, changed)

	require.NoError(t, db.Exec(`INSERT INTO repos (org_id, name, source, github_repo_id, connection_id, provider_repo_id) VALUES (?, 'acme/api', 'github', 101, ?, '101')`,
		org, created.ID).Error)
	status, body = connectionsPageCall(t, base, http.MethodDelete, "/api/catalog/connection/"+created.ID, admin, "org:admin", nil, nil)
	require.Less(t, status, 300, body)
	var repos int64
	require.NoError(t, db.Raw(`SELECT count(*) FROM repos WHERE org_id = ?`, org).Scan(&repos).Error)
	assert.Zero(t, repos, "deleting the connection removes the repositories only it brings")
}

func TestALocalDashboardsConnectionsPageOffersEveryConnectionType(t *testing.T) {
	db, _, _ := openOrgDB(t)
	t.Setenv("GAVEL_SECRET_KEEPER_CONNECTION", "")
	t.Setenv("GAVEL_SECRET_KEY", base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("l"), 32)))
	srv := &Server{}
	srv.SetConnectionCatalog(db)
	handler := srv.Handler()
	req, err := http.NewRequest(http.MethodGet, "/api/catalog/connection", nil)
	require.NoError(t, err)
	req.Header.Set("Accept", "application/schema+json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	assert.Contains(t, recorder.Body.String(), `"postgres"`, "a local dashboard keeps every type")
}
