// Tests the connections page's Test and health check for a hosted organization: each checks a git
// connection's credentials against its provider, at public addresses only by default.
package ui

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

	t.Run("resolving a connection is still refused", func(t *testing.T) {
		status, body := connectionsPageCall(t, base, http.MethodPost, "/api/catalog/connection/resolve", admin, "org:admin", draft("glpat-good"), nil)
		assert.Equal(t, http.StatusForbidden, status, body)
		assert.Contains(t, body, "a hosted organization's connections are tested and health-checked, but never resolved, from the server")
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

// connectionHealthReply is what the connections page's health check answers.
type connectionHealthReply struct {
	Results []struct {
		ID, State, Detail string
		CheckedAt         time.Time
	}
}

func TestAHostedOrgsConnectionHealthChecksEachConnectionsCredentials(t *testing.T) {
	db, _, org := openOrgDB(t)
	t.Setenv("GAVEL_SECRET_KEEPER_CONNECTION", "")
	t.Setenv("GAVEL_SECRET_KEY", base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("h"), 32)))
	admin, member := uuid.New(), uuid.New()
	for name, id := range map[string]uuid.UUID{"admin": admin, "member": member} {
		require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, "user_"+name).Error)
	}
	gitlab := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v4/user" && r.Header.Get("Authorization") == "Bearer glpat-good" {
			_, _ = w.Write([]byte(`{"id":1,"username":"batman"}`))
			return
		}
		http.Error(w, `{"message":"401 Unauthorized"}`, http.StatusUnauthorized)
	}))
	t.Cleanup(gitlab.Close)
	down := httptest.NewTLSServer(http.NotFoundHandler())
	down.Close()

	srv := &Server{}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)
	srv.SetConnectionClientOptions(git.ClientOptions{})
	base := serveActingAs(t, srv, org)
	save := func(name, url, token string) string {
		t.Helper()
		status, body := connectionsPageCall(t, base, http.MethodPost, "/api/catalog/connection", admin, "org:admin",
			map[string]any{"name": name, "type": "gitlab", "url": url, "password": token, "insecure_tls": true}, nil)
		require.Equal(t, http.StatusOK, status, body)
		var saved struct{ ID string }
		require.NoError(t, json.Unmarshal([]byte(body), &saved))
		return saved.ID
	}
	good, stolen, gone := save("good", gitlab.URL, "glpat-good"), save("stolen", gitlab.URL, "glpat-stolen"), save("gone", down.URL, "glpat-good")

	t.Run("a member sees each connection's state: healthy, refused credentials, or unreachable", func(t *testing.T) {
		before := time.Now()
		status, body := connectionsPageCall(t, base, http.MethodPost, "/api/catalog/connections/health", member, "org:member",
			map[string]any{"ids": []string{good, stolen, gone}, "force": true}, nil)
		require.Equal(t, http.StatusOK, status, body)
		assert.NotContains(t, body, "glpat-")
		var reply connectionHealthReply
		require.NoError(t, json.Unmarshal([]byte(body), &reply), body)
		require.Len(t, reply.Results, 3, body)
		states := map[string]string{}
		for _, result := range reply.Results {
			states[result.ID] = result.State
			assert.False(t, result.CheckedAt.Before(before.Truncate(time.Second)), "checked now: %s", body)
		}
		assert.Equal(t, map[string]string{good: "healthy", stolen: "credentials", gone: "unreachable"}, states, body)
		for _, result := range reply.Results {
			switch result.ID {
			case good:
				assert.Equal(t, "gitlab accepted the connection's credentials", result.Detail)
			case stolen:
				assert.Contains(t, result.Detail, "gitlab didn't confirm the connection's credentials")
				assert.Contains(t, result.Detail, "401")
			case gone:
				assert.Contains(t, result.Detail, "gitlab didn't confirm the connection's credentials")
			}
		}
	})

	t.Run("a connection the organization doesn't have is refused", func(t *testing.T) {
		other := uuid.NewString()
		status, body := connectionsPageCall(t, base, http.MethodPost, "/api/catalog/connections/health", admin, "org:admin",
			map[string]any{"ids": []string{good, other}}, nil)
		assert.Equal(t, http.StatusBadRequest, status, body)
		assert.Contains(t, body, "unknown connection ids: "+other)
	})

	t.Run("a check names at least one connection", func(t *testing.T) {
		status, body := connectionsPageCall(t, base, http.MethodPost, "/api/catalog/connections/health", admin, "org:admin",
			map[string]any{"ids": []string{}}, nil)
		assert.Equal(t, http.StatusBadRequest, status, body)
		assert.Contains(t, body, "name the connections to check in ids")
	})
}
