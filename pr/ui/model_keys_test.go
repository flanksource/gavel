// Tests the /api/model-keys routes: every member sees which keys the organization has, only
// admins change them, keys never come back out, and none mode explains the keys are hosted-only.
package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flanksource/commons-db/secret"
	"github.com/flanksource/gavel/modelkeys"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const modelKeysTestAPIKey = "sk-ant-api03-route-test-key-WXYZ"

func useModelKeyKeeper(t *testing.T) {
	t.Helper()
	previous := secret.KMSConnection
	secret.KMSConnection = modelkeys.KeeperConnection([]byte("a secret of at least thirty-two bytes"))
	secret.InvalidateKeeper()
	t.Cleanup(func() {
		secret.KMSConnection = previous
		secret.InvalidateKeeper()
	})
}

func TestModelKeyRoutes(t *testing.T) {
	useModelKeyKeeper(t)
	db, _, org := openOrgDB(t)
	admin, member := uuid.New(), uuid.New()
	for id, clerk := range map[uuid.UUID]string{admin: "user_admin", member: "user_member"} {
		require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, clerk).Error)
	}
	srv := &Server{}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)
	handler := srv.Handler()
	call := func(user uuid.UUID, role, method, path, body string) (int, string) {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req = req.WithContext(tenant.WithScope(req.Context(), tenant.Scope{OrgID: org, UserID: &user, Role: role}))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code, rec.Body.String()
	}

	status, body := call(member, "org:member", http.MethodGet, "/api/model-keys", "")
	require.Equal(t, http.StatusOK, status, body)
	assert.JSONEq(t, `{"kinds":[
		{"name":"anthropic-api-key","provider":"anthropic","label":"Anthropic API key"},
		{"name":"claude-code-oauth-token","provider":"anthropic","label":"Claude Code OAuth token"}],
		"keys":[]}`, body)

	status, body = call(member, "org:member", http.MethodPut, "/api/model-keys/anthropic", `{"kind":"anthropic-api-key","key":"`+modelKeysTestAPIKey+`"}`)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Contains(t, body, "only an organization admin can change its model keys")

	status, body = call(admin, "org:admin", http.MethodPut, "/api/model-keys/anthropic", `{"kind":"anthropic-api-key","key":"`+modelKeysTestAPIKey+`"}`)
	require.Equal(t, http.StatusOK, status, body)
	assert.NotContains(t, body, "route-test-key")
	var saved modelkeys.Summary
	require.NoError(t, json.Unmarshal([]byte(body), &saved))
	assert.Equal(t, "anthropic", saved.Provider)
	assert.Equal(t, "anthropic-api-key", saved.Kind)
	assert.Equal(t, "WXYZ", saved.Hint)
	assert.Equal(t, &admin, saved.UpdatedBy)

	value, err := modelkeys.Store{DB: db}.Value(t.Context(), org, "anthropic")
	require.NoError(t, err)
	assert.Equal(t, modelKeysTestAPIKey, value)

	status, body = call(member, "org:member", http.MethodGet, "/api/model-keys", "")
	require.Equal(t, http.StatusOK, status, body)
	assert.NotContains(t, body, "route-test-key")
	var listed struct{ Keys []modelkeys.Summary }
	require.NoError(t, json.Unmarshal([]byte(body), &listed))
	require.Len(t, listed.Keys, 1)
	assert.Equal(t, "WXYZ", listed.Keys[0].Hint)

	for name, test := range map[string]struct {
		path, body string
		status     int
		want       string
	}{
		"invalid key":    {"/api/model-keys/anthropic", `{"kind":"claude-code-oauth-token","key":"` + modelKeysTestAPIKey + `"}`, http.StatusBadRequest, `a Claude Code OAuth token starts with \"sk-ant-oat\"`},
		"unknown kind":   {"/api/model-keys/anthropic", `{"kind":"nope","key":"x"}`, http.StatusBadRequest, `unknown kind \"nope\"`},
		"other provider": {"/api/model-keys/openai", `{"kind":"anthropic-api-key","key":"` + modelKeysTestAPIKey + `"}`, http.StatusBadRequest, `kind anthropic-api-key is a key for anthropic, not openai`},
		"invalid JSON":   {"/api/model-keys/anthropic", `{`, http.StatusBadRequest, "invalid JSON"},
	} {
		status, body := call(admin, "org:admin", http.MethodPut, test.path, test.body)
		assert.Equal(t, test.status, status, name)
		assert.Contains(t, body, test.want, name)
		assert.NotContains(t, body, "route-test-key", name)
	}

	status, body = call(member, "org:member", http.MethodDelete, "/api/model-keys/anthropic", "")
	assert.Equal(t, http.StatusForbidden, status)
	assert.Contains(t, body, "only an organization admin can change its model keys")
	status, _ = call(admin, "org:admin", http.MethodDelete, "/api/model-keys/anthropic", "")
	assert.Equal(t, http.StatusNoContent, status)
	status, body = call(admin, "org:admin", http.MethodDelete, "/api/model-keys/anthropic", "")
	assert.Equal(t, http.StatusNotFound, status)
	assert.Contains(t, body, "has no anthropic key")
}

func TestModelKeysAreHostedOnly(t *testing.T) {
	rec := httptest.NewRecorder()
	(&Server{}).Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/model-keys", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "model keys are only used by hosted sandbox runs")
}
