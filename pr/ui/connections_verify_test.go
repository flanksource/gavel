// Tests the connections page's Test for a hosted organization: it checks a git connection's
// credentials against its provider, only for an admin, and only at public addresses by default.
package ui

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flanksource/commons-db/git"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// connectionTestResult is what the connections page's Test answers.
type connectionTestResult struct {
	OK      bool
	Message string
}

func TestAHostedOrgsConnectionTestChecksItsCredentials(t *testing.T) {
	db, _, org := openOrgDB(t)
	t.Setenv("GAVEL_SECRET_KEEPER_CONNECTION", "")
	t.Setenv("GAVEL_SECRET_KEY", base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("v"), 32)))
	admin, member := uuid.New(), uuid.New()
	for name, id := range map[string]uuid.UUID{"admin": admin, "member": member} {
		require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, "user_"+name).Error)
	}
	gitlab := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v4/user" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer glpat-good" {
			http.Error(w, `{"message":"401 Unauthorized"}`, http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"id":1,"username":"batman"}`))
	}))
	t.Cleanup(gitlab.Close)

	srv := &Server{}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)
	srv.SetConnectionClientOptions(git.ClientOptions{})
	base := serveActingAs(t, srv, org)
	test := func(user uuid.UUID, role string, draft map[string]any) (int, connectionTestResult, string) {
		t.Helper()
		status, body := connectionsPageCall(t, base, http.MethodPost, "/api/catalog/connection/test", user, role, draft, nil)
		var result connectionTestResult
		if status == http.StatusOK {
			require.NoError(t, json.Unmarshal([]byte(body), &result), body)
		}
		return status, result, body
	}
	draft := func(password string) map[string]any {
		return map[string]any{"name": "acme-gitlab", "type": "gitlab", "url": gitlab.URL, "password": password, "insecure_tls": true}
	}

	t.Run("a token the provider accepts passes", func(t *testing.T) {
		status, result, body := test(admin, "org:admin", draft("glpat-good"))
		require.Equal(t, http.StatusOK, status, body)
		assert.True(t, result.OK, body)
		assert.Equal(t, "gitlab accepted the connection's credentials", result.Message)
		assert.NotContains(t, body, "glpat-good")
	})

	t.Run("a token the provider refuses fails, saying why", func(t *testing.T) {
		status, result, body := test(admin, "org:admin", draft("glpat-stolen"))
		require.Equal(t, http.StatusOK, status, body)
		assert.False(t, result.OK, body)
		assert.Contains(t, result.Message, "gitlab didn't confirm the connection's credentials")
		assert.Contains(t, result.Message, "401")
		assert.NotContains(t, body, "glpat-stolen")
	})

	t.Run("an edited connection whose form leaves the token blank is tested with its saved token", func(t *testing.T) {
		status, body := connectionsPageCall(t, base, http.MethodPost, "/api/catalog/connection", admin, "org:admin", draft("glpat-good"), nil)
		require.Equal(t, http.StatusOK, status, body)
		var saved struct{ ID string }
		require.NoError(t, json.Unmarshal([]byte(body), &saved))
		edited := draft("")
		edited["id"] = saved.ID
		status, result, body := test(admin, "org:admin", edited)
		require.Equal(t, http.StatusOK, status, body)
		assert.True(t, result.OK, body)
	})

	t.Run("a connection of another organization lends no saved token", func(t *testing.T) {
		edited := draft("")
		edited["id"] = uuid.NewString()
		status, _, body := test(admin, "org:admin", edited)
		assert.Equal(t, http.StatusNotFound, status, body)
	})

	t.Run("a member can't test", func(t *testing.T) {
		status, _, body := test(member, "org:member", draft("glpat-good"))
		assert.Equal(t, http.StatusForbidden, status, body)
	})

	t.Run("a type other than git is refused", func(t *testing.T) {
		status, _, body := test(admin, "org:admin", map[string]any{"name": "warehouse", "type": "postgres", "url": "postgres://10.0.0.5/app"})
		assert.Equal(t, http.StatusBadRequest, status, body)
		assert.Contains(t, body, "a hosted organization keeps only git connections")
	})

	t.Run("resolving a connection or checking their health is still refused", func(t *testing.T) {
		for _, path := range []string{"/api/catalog/connection/resolve", "/api/catalog/connections/health"} {
			status, body := connectionsPageCall(t, base, http.MethodPost, path, admin, "org:admin", draft("glpat-good"), nil)
			assert.Equal(t, http.StatusForbidden, status, "%s: %s", path, body)
			assert.Contains(t, body, "a hosted organization's connections are tested, but never resolved or health-checked, from the server")
		}
	})
}

func TestAHostedOrgsConnectionTestReachesOnlyPublicAddresses(t *testing.T) {
	db, _, org := openOrgDB(t)
	t.Setenv("GAVEL_SECRET_KEEPER_CONNECTION", "")
	t.Setenv("GAVEL_SECRET_KEY", base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("p"), 32)))
	admin := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', 'user_admin')`, admin).Error)
	asked := false
	internal := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { asked = true }))
	t.Cleanup(internal.Close)

	srv := &Server{}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)
	base := serveActingAs(t, srv, org)
	status, body := connectionsPageCall(t, base, http.MethodPost, "/api/catalog/connection/test", admin, "org:admin",
		map[string]any{"name": "internal", "type": "gitlab", "url": internal.URL, "password": "glpat-good"}, nil)
	require.Equal(t, http.StatusOK, status, body)
	var result connectionTestResult
	require.NoError(t, json.Unmarshal([]byte(body), &result), body)
	assert.False(t, result.OK, body)
	assert.Contains(t, result.Message, "gitlab didn't confirm the connection's credentials", "the test's failure names the provider")
	assert.Contains(t, result.Message, "defensive HTTP client refuses localhost address 127.0.0.1")
	assert.False(t, asked, "a hosted server never dials a private address")
}
