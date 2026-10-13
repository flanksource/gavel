// Tests that a Server scoped to an organization reads and writes only that
// organization's rows in the GitHub cache, such as which PRs have been seen.
package ui

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/github"
	ghcache "github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// openOrgDB migrates a fresh database and returns it, its GitHub cache, and a second org's ID.
func openOrgDB(t *testing.T) (*gorm.DB, *ghcache.Store, uuid.UUID) {
	t.Helper()
	dsn := commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_prui_org"}).DSN()
	t.Setenv(database.EnvDSN, dsn)
	db, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	org := uuid.New()
	require.NoError(t, db.Gorm().Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Other')`, org, org.String()).Error)
	store, err := ghcache.Open()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, store.Close()) })
	return db.Gorm(), store, org
}

func TestServerScopesSeenPRsToItsOrg(t *testing.T) {
	db, store, org := openOrgDB(t)
	local := &Server{githubCache: store}
	other := &Server{githubCache: store}
	other.SetOrg(org, db)
	prs := github.PRSearchResults{{Repo: "acme/app", Number: 1, UpdatedAt: time.Now().Add(-time.Hour)}}

	require.NoError(t, local.MarkSeen("acme/app", 1))
	assert.Empty(t, local.UnreadMap(prs), "the org that read the PR")
	assert.Equal(t, map[string]bool{"acme/app#1": true}, other.UnreadMap(prs), "another org hasn't read it")
}

func TestOrgConfigPersistsToOrgSettings(t *testing.T) {
	db, _, org := openOrgDB(t)
	original := settingsPath
	settingsPath = filepath.Join(t.TempDir(), "pr-ui-settings.json")
	t.Cleanup(func() { settingsPath = original })

	got, err := LoadOrgSearchConfig(t.Context(), db, org)
	require.NoError(t, err)
	assert.Equal(t, SearchConfig{}, got, "an org without settings gets the default")

	srv := &Server{refreshCh: make(chan struct{}, 1)}
	srv.SetOrg(org, db)
	rec := httptest.NewRecorder()
	srv.handleConfig(rec, httptest.NewRequest(http.MethodPost, "/api/config", strings.NewReader(`{"repos":["acme/app"],"org":"acme"}`)))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	got, err = LoadOrgSearchConfig(t.Context(), db, org)
	require.NoError(t, err)
	assert.Equal(t, SearchConfig{Repos: []string{"acme/app"}, Org: "acme"}, got)
	local, err := LoadOrgSearchConfig(t.Context(), db, tenant.LocalOrgID)
	require.NoError(t, err)
	assert.Equal(t, SearchConfig{}, local, "another org's settings are untouched")
	assert.NoFileExists(t, settingsPath, "an org's settings never reach the host's settings file")
}
