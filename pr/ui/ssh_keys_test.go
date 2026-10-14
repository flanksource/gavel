// Tests the /api/ssh-keys routes: a signed-in user adds, lists and deletes their own
// keys, with descriptive refusals, and none mode explains the keys are hosted-only.
package ui

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flanksource/gavel/sshkeys"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

func newEd25519Line(t *testing.T) string {
	t.Helper()
	public, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	key, err := gossh.NewPublicKey(public)
	require.NoError(t, err)
	return strings.TrimSpace(string(gossh.MarshalAuthorizedKey(key)))
}

func TestSSHKeyRoutes(t *testing.T) {
	db, _, org := openOrgDB(t)
	users := map[string]uuid.UUID{}
	for _, name := range []string{"ada", "bob"} {
		id := uuid.New()
		require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, "user_"+name).Error)
		users[name] = id
	}
	srv := &Server{}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)
	handler := srv.Handler()
	call := func(user, method, path, body string) (int, string) {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		id := users[user]
		req = req.WithContext(tenant.WithScope(req.Context(), tenant.Scope{OrgID: org, UserID: &id}))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code, rec.Body.String()
	}
	line := newEd25519Line(t)

	status, body := call("ada", http.MethodPost, "/api/ssh-keys", `{"name":"laptop","publicKey":"`+line+`"}`)
	require.Equal(t, http.StatusCreated, status, body)
	var added sshkeys.Key
	require.NoError(t, json.Unmarshal([]byte(body), &added))
	assert.Equal(t, "laptop", added.Name)
	assert.True(t, strings.HasPrefix(added.Fingerprint, "SHA256:"))

	status, body = call("bob", http.MethodPost, "/api/ssh-keys", `{"name":"mine","publicKey":"`+line+`"}`)
	assert.Equal(t, http.StatusConflict, status)
	assert.Contains(t, body, "already registered")

	status, body = call("ada", http.MethodPost, "/api/ssh-keys", `{"name":"bad","publicKey":"ssh-rsa AAAA"}`)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, body, "not an SSH public key")
	status, body = call("ada", http.MethodPost, "/api/ssh-keys", `{`)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, body, "invalid JSON")

	status, body = call("ada", http.MethodGet, "/api/ssh-keys", "")
	require.Equal(t, http.StatusOK, status)
	var keys []sshkeys.Key
	require.NoError(t, json.Unmarshal([]byte(body), &keys))
	require.Len(t, keys, 1)
	status, body = call("bob", http.MethodGet, "/api/ssh-keys", "")
	require.Equal(t, http.StatusOK, status)
	assert.JSONEq(t, `[]`, body, "keys belong to a user, not the org")

	status, _ = call("bob", http.MethodDelete, "/api/ssh-keys/"+added.ID.String(), "")
	assert.Equal(t, http.StatusNotFound, status, "another user's key")
	status, body = call("ada", http.MethodDelete, "/api/ssh-keys/not-a-uuid", "")
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, body, "not-a-uuid")
	status, _ = call("ada", http.MethodDelete, "/api/ssh-keys/"+added.ID.String(), "")
	assert.Equal(t, http.StatusNoContent, status)
}

func TestSSHKeysAreHostedOnly(t *testing.T) {
	rec := httptest.NewRecorder()
	(&Server{}).Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/ssh-keys", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "SSH keys are only used by the hosted push server")
}
