// Tests /api/commit-identity: the name and email a signed-in user's hosted runs commit as, over
// their GitHub identity, and that a local dashboard has none.
package ui

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostedUsersSetTheNameAndEmailTheirRunsCommitAs(t *testing.T) {
	db, _, org := openOrgDB(t)
	ada := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id, name, email, github_login, github_id)
		VALUES (?, 'https://clerk.example', 'user_ada', 'Ada Lovelace', 'ada@example.com', 'ada', 42)`, ada).Error)
	base := hostedTodosServer(t, db, org, t.TempDir())
	githubDefault := map[string]any{"name": "ada", "email": "42+ada@users.noreply.github.com"}

	status, identity := hostedCall(t, base, http.MethodGet, "/api/commit-identity", ada, "org:member", nil)
	require.Equal(t, http.StatusOK, status, identity)
	assert.Equal(t, map[string]any{"name": "", "email": "", "default": githubDefault}, identity)

	status, saved := hostedCall(t, base, http.MethodPut, "/api/commit-identity", ada, "org:member", map[string]any{"name": "Ada Lovelace", "email": ""})
	require.Equal(t, http.StatusOK, status, saved)
	assert.Equal(t, map[string]any{"name": "Ada Lovelace", "email": "", "default": githubDefault}, saved)

	status, refused := hostedCall(t, base, http.MethodPut, "/api/commit-identity", ada, "org:member", map[string]any{"name": "Ada", "email": "ada"})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid commit identity: the email must be one address, like ada@example.com", refused["error"])
}

func TestALocalDashboardHasNoCommitIdentity(t *testing.T) {
	server := httptest.NewServer((&Server{}).Handler())
	t.Cleanup(server.Close)
	resp, err := http.Get(server.URL + "/api/commit-identity")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "local commits are by the machine's own git identity")
}
