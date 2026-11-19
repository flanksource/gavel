// Tests a hosted organization's connection routes: members list its git connections, admins keep
// them in its catalog encrypted under GAVEL_SECRET_KEY, and nothing but git types is stored.
package ui

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"strings"
	"testing"

	"github.com/flanksource/commons-db/connection/catalog"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAHostedOrgKeepsItsGitConnections(t *testing.T) {
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
	base := serveActingAs(t, srv, org)
	acme := map[string]any{"name": "acme", "type": "github", "password": "gh_hosted", "properties": map[string]any{"organization": "acme"}}

	status, body := hostedCall(t, base, http.MethodPost, "/api/connections", member, "org:member", acme)
	assert.Equal(t, http.StatusForbidden, status, "a member can't add a connection: %v", body)

	status, body = hostedCall(t, base, http.MethodPost, "/api/connections", admin, "org:admin", acme)
	require.Equal(t, http.StatusOK, status, body)
	var row catalog.Record
	require.NoError(t, db.Where("org_id = ?", org).First(&row).Error)
	assert.Equal(t, "github", row.Type)
	assert.True(t, strings.HasPrefix(row.KeyID, "local-"), "encrypted with the key GAVEL_SECRET_KEY derives: %s", row.KeyID)
	assert.NotContains(t, string(row.Payload), "gh_hosted")

	status, listed := hostedRaw(t, base, http.MethodGet, "/api/connections", member)
	require.Equal(t, http.StatusOK, status, listed)
	assert.Contains(t, listed, `"acme"`, "a member lists the org's connections")
	assert.NotContains(t, listed, "gh_hosted")

	status, body = hostedCall(t, base, http.MethodPost, "/api/connections", admin, "org:admin",
		map[string]any{"name": "warehouse", "type": "postgres", "url": "postgres://10.0.0.5/app"})
	assert.Equal(t, http.StatusBadRequest, status, body)
	assert.Contains(t, body["error"], "a hosted organization keeps only git connections (git, github, gitlab, azure_devops), not postgres")
	status, body = hostedCall(t, base, http.MethodPatch, "/api/connections/"+row.ID.String(), admin, "org:admin", map[string]any{"type": "http"})
	assert.Equal(t, http.StatusBadRequest, status, body)
	assert.Contains(t, body["error"], "not http")

	status, body = hostedCall(t, base, http.MethodPost, "/api/connections/"+row.ID.String()+"/validate", admin, "org:admin", nil)
	assert.Equal(t, http.StatusNotFound, status, "validating dials the connection, which hosted mode doesn't yet guard: %v", body)
}
