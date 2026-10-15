// Tests the CI results endpoint: a GitHub Actions OIDC token for a linked repository stores the
// snapshot in that organization's blob store and indexes it; anything else is refused.
package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/api/clerktest"
	ghcache "github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/internal/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ciAudience = "https://gavel.example"

func TestCIResults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_ci_results"}).DSN())
	t.Setenv(database.EnvDisable, "")
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	db := opened.Gorm()
	cache, err := ghcache.Open()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, cache.Close()) })

	org := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)
	var repo struct{ ID uuid.UUID }
	require.NoError(t, db.Raw(`INSERT INTO repos (org_id, name, source, github_repo_id) VALUES (?, 'acme/app', 'github', 123456) RETURNING id`, org).Scan(&repo).Error)
	repoID := repo.ID

	keys := clerktest.New(t, "cli", clerktest.User{ID: "user_123"})
	results := &CIResults{
		DB:    db,
		OIDC:  api.GitHubOIDCVerifier{Audience: ciAudience, JWKS: api.NewJWKSCache(keys.URL+"/.well-known/jwks.json", keys.Client(), 10*time.Minute)},
		Cache: cache,
	}
	server := httptest.NewServer(results.Handler())
	t.Cleanup(server.Close)
	token := func(repositoryID string) string {
		return keys.Sign(jwt.MapClaims{
			"iss": api.GitHubOIDCIssuer, "aud": ciAudience, "sub": "repo:acme/app:ref:refs/heads/main",
			"exp": time.Now().Add(5 * time.Minute).Unix(), "iat": time.Now().Unix(),
			"repository": "acme/app", "repository_id": repositoryID, "repository_owner_id": "789",
			"sha": "abc123", "ref": "refs/heads/main", "run_id": "42", "workflow": "CI",
		})
	}
	post := func(bearer string, body []byte) *http.Response {
		t.Helper()
		req, err := http.NewRequest(http.MethodPost, server.URL+"/api/ci/results", bytes.NewReader(body))
		require.NoError(t, err)
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { _ = resp.Body.Close() })
		return resp
	}
	snapshot := []byte(`{"metadata":{"started":"2026-10-15T09:00:00Z"},"tests":[{"name":"TestA","failed":true},{"name":"TestB","passed":true}]}`)

	t.Run("a linked repository's results are stored", func(t *testing.T) {
		resp := post(token("123456"), snapshot)
		body, _ := io.ReadAll(resp.Body)
		require.Equal(t, http.StatusCreated, resp.StatusCode, string(body))
		var created struct{ RunID string }
		require.NoError(t, json.Unmarshal(body, &created))
		assert.Equal(t, "run-2026-10-15T09-00-00Z", created.RunID)

		reader, err := cache.OrgBlobs(org).Read(context.Background(), "repos/"+repoID.String()+"/runs/"+created.RunID+".json")
		require.NoError(t, err)
		stored, err := io.ReadAll(reader)
		require.NoError(t, reader.Close())
		require.NoError(t, err)
		assert.Contains(t, string(stored), "TestA")

		row, err := cache.ForOrg(org).GetTestRun(context.Background(), "repo:"+repoID.String(), created.RunID)
		require.NoError(t, err)
		require.NotNil(t, row)
		assert.Equal(t, 1, row.Failed)
		assert.Equal(t, 1, row.Passed)
		assert.Equal(t, "abc123", row.SHA, "the token's commit when the snapshot names none")
	})

	t.Run("a second upload in the same second gets its own run", func(t *testing.T) {
		resp := post(token("123456"), snapshot)
		body, _ := io.ReadAll(resp.Body)
		require.Equal(t, http.StatusCreated, resp.StatusCode, string(body))
		assert.Contains(t, string(body), "run-2026-10-15T09-00-00Z-2")
	})

	for name, tc := range map[string]struct {
		bearer string
		body   []byte
		status int
		want   string
	}{
		"no token":          {"", snapshot, http.StatusUnauthorized, "GitHub Actions OIDC token"},
		"invalid token":     {"not-a-token", snapshot, http.StatusUnauthorized, "token is malformed"},
		"unlinked repo":     {token("999"), snapshot, http.StatusForbidden, "acme/app (999) is not linked to a gavel organization"},
		"not a snapshot":    {token("123456"), []byte(`{"tests":`), http.StatusBadRequest, "decode the results"},
		"larger than 50 MB": {token("123456"), []byte(`{"tests":[],"x":"` + strings.Repeat("a", maxCIResultsBytes) + `"}`), http.StatusRequestEntityTooLarge, "50 MB"},
	} {
		t.Run(name, func(t *testing.T) {
			resp := post(tc.bearer, tc.body)
			body, _ := io.ReadAll(resp.Body)
			assert.Equal(t, tc.status, resp.StatusCode, string(body))
			assert.Contains(t, string(body), tc.want)
		})
	}
}
