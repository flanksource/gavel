// Tests CI results from GitLab CI and Azure Pipelines: a GitLab ID token from gitlab.com or a GitLab
// a connection syncs repositories from, and an Azure job's access token for its running build.
package ui

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/api/clerktest"
	"github.com/flanksource/gavel/connections"
	"github.com/flanksource/gavel/tenant"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ciRepo records a repository a connection of org's brings, answering its ID.
func ciRepo(t *testing.T, db *gorm.DB, org uuid.UUID, source, providerID, remoteURL string, connection uuid.UUID) uuid.UUID {
	t.Helper()
	var repo struct{ ID uuid.UUID }
	require.NoError(t, db.Raw(`INSERT INTO repos (org_id, name, source, connection_id, remote_url, provider_repo_id, default_branch)
		VALUES (?, ?, ?, ?, ?, ?, 'main') RETURNING id`, org, remoteURL, source, connection, remoteURL, providerID).Scan(&repo).Error)
	return repo.ID
}

// ciUpload posts a snapshot to server with bearer and headers, answering the status and body.
func ciUpload(t *testing.T, server, bearer string, headers map[string]string) (int, string) {
	t.Helper()
	snapshot := []byte(`{"metadata":{"started":"2026-11-22T09:00:00Z"},"tests":[{"name":"TestA","failed":true}]}`)
	req, err := http.NewRequest(http.MethodPost, server+"/api/ci/results", bytes.NewReader(snapshot))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+bearer)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

// ciError is the reason a refused upload's body gives.
func ciError(t *testing.T, body string) string {
	t.Helper()
	var refused struct{ Error string }
	require.NoError(t, json.Unmarshal([]byte(body), &refused), body)
	return refused.Error
}

// ciStoredSHA is the commit of the run an upload answering body stored for repo of org.
func ciStoredSHA(t *testing.T, results *CIResults, org, repo uuid.UUID, body string) string {
	t.Helper()
	var created struct{ RunID string }
	require.NoError(t, json.Unmarshal([]byte(body), &created))
	row, err := results.Cache.ForOrg(org).GetTestRun(t.Context(), RepoWorkspace(repo), created.RunID)
	require.NoError(t, err)
	require.NotNil(t, row, "org %s has run %s", org, created.RunID)
	return row.SHA
}

func TestCIResultsFromGitLab(t *testing.T) {
	db, cache := openCIResults(t)
	org, other := ciOrg(t, db, "Acme"), ciOrg(t, db, "Initech")
	connection := func(org uuid.UUID) uuid.UUID {
		id := uuid.New()
		require.NoError(t, db.Exec(`INSERT INTO connections (id, org_id, namespace, name, type, key_id, payload) VALUES (?, ?, '', ?, 'gitlab', 'k', '\x00')`,
			id, org, id.String()).Error)
		return id
	}
	selfManaged := ciRepo(t, db, org, "gitlab", "42", "https://gitlab.example.com/acme/api.git", connection(org))
	sameIDElsewhere := ciRepo(t, db, other, "gitlab", "42", "https://gitlab.other.example/acme/api.git", connection(other))

	keys := clerktest.New(t, "cli", clerktest.User{ID: "user_123"})
	var mu sync.Mutex
	var askedFor []string
	results := &CIResults{DB: db, Cache: cache, GitLab: api.GitLabOIDCVerifier{Audience: ciAudience, JWKS: func(issuer string) *api.JWKSCache {
		mu.Lock()
		defer mu.Unlock()
		askedFor = append(askedFor, issuer)
		return api.NewJWKSCache(keys.URL+"/.well-known/jwks.json", keys.Client(), 10*time.Minute)
	}}}
	server := httptest.NewServer(results.Handler())
	t.Cleanup(server.Close)
	token := func(issuer, project string) string {
		return keys.Sign(jwt.MapClaims{"iss": issuer, "aud": ciAudience, "exp": time.Now().Add(5 * time.Minute).Unix(), "iat": time.Now().Unix(),
			"project_id": project, "project_path": "acme/api", "sha": "abc123", "ref": "main", "pipeline_id": "1001", "job_id": "5005"})
	}

	t.Run("a self-managed GitLab's job stores results for the project synced from it alone", func(t *testing.T) {
		status, body := ciUpload(t, server.URL, token("https://gitlab.example.com", "42"), nil)
		require.Equal(t, http.StatusCreated, status, body)
		assert.Equal(t, "abc123", ciStoredSHA(t, results, org, selfManaged, body), "the token's commit when the snapshot names none")
		rows, err := cache.ForOrg(other).ListTestRunsForDir(t.Context(), RepoWorkspace(sameIDElsewhere))
		require.NoError(t, err)
		assert.Empty(t, rows, "project 42 on another GitLab is another project")
	})

	t.Run("gitlab.com is trusted without a repository synced from it", func(t *testing.T) {
		status, body := ciUpload(t, server.URL, token("https://gitlab.com", "99"), nil)
		assert.Equal(t, http.StatusForbidden, status, body)
		assert.Equal(t, "GitLab project acme/api (99) on gitlab.com is not synced by a gavel organization; add a gitlab connection that brings it", ciError(t, body))
	})

	for name, issuer := range map[string]string{"a GitLab no connection syncs from": "https://gitlab.evil.example", "a GitLab over http": "http://gitlab.example.com"} {
		t.Run(name+" is not trusted, and its keys are never fetched", func(t *testing.T) {
			mu.Lock()
			askedFor = nil
			mu.Unlock()
			status, body := ciUpload(t, server.URL, token(issuer, "42"), nil)
			assert.Equal(t, http.StatusUnauthorized, status, body)
			assert.Equal(t, "token issuer "+issuer+" is neither GitHub Actions, gitlab.com nor a GitLab a gavel organization's git connection syncs repositories from", ciError(t, body))
			mu.Lock()
			defer mu.Unlock()
			assert.Empty(t, askedFor)
		})
	}
}

func TestCIResultsFromAzurePipelines(t *testing.T) {
	db, cache := openCIResults(t)
	scope := tenant.Scope{OrgID: ciOrg(t, db, "Acme"), Role: "org:admin"}
	const sha = "2222222222222222222222222222222222222222"
	var mu sync.Mutex
	var authorizations []string
	azure := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		authorizations = append(authorizations, r.Header.Get("Authorization"))
		mu.Unlock()
		if r.Header.Get("Authorization") != "Bearer job-token" {
			http.Error(w, "the job's token is required", http.StatusUnauthorized)
			return
		}
		builds := map[string]string{
			"/acme/Platform/_apis/build/builds/91": `{"id":91,"status":"inProgress","sourceVersion":"` + sha + `","definition":{"id":23},"repository":{"id":"RULES-GUID","type":"TfsGit"}}`,
			"/acme/Platform/_apis/build/builds/92": `{"id":92,"status":"completed","result":"succeeded","sourceVersion":"` + sha + `","definition":{"id":23},"repository":{"id":"rules-guid","type":"TfsGit"}}`,
			"/acme/Platform/_apis/build/builds/93": `{"id":93,"status":"inProgress","sourceVersion":"` + sha + `","definition":{"id":24},"repository":{"id":"other-guid","type":"TfsGit"}}`,
		}
		build, ok := builds[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(build))
	}))
	t.Cleanup(azure.Close)
	store, closeStore, err := connections.OpenStore(t.Context(), connections.Options{DB: db, Scope: scope})
	require.NoError(t, err)
	saved, err := store.SaveConnection(t.Context(), &models.Connection{Name: "acme-azure", Type: models.ConnectionTypeAzureDevops, URL: azure.URL + "/acme",
		Password: "azure-pat", InsecureTLS: true})
	require.NoError(t, err)
	require.NoError(t, closeStore())
	repo := ciRepo(t, db, scope.OrgID, "azure_devops", "rules-guid", azure.URL+"/acme/Platform/_git/rules", saved.ID)

	results := &CIResults{DB: db, Cache: cache, ClientOptions: git.ClientOptions{}}
	server := httptest.NewServer(results.Handler())
	t.Cleanup(server.Close)
	build := func(id, repository string) map[string]string {
		return map[string]string{CIAzureBuildHeader: id, CIAzureRepositoryHeader: repository}
	}

	t.Run("a running build's job stores its results, read with the job's own token alone", func(t *testing.T) {
		status, body := ciUpload(t, server.URL, "job-token", build("91", "rules-guid"))
		require.Equal(t, http.StatusCreated, status, body)
		assert.Equal(t, sha, ciStoredSHA(t, results, scope.OrgID, repo, body), "the build's source version")
		mu.Lock()
		defer mu.Unlock()
		assert.Equal(t, []string{"Bearer job-token"}, authorizations, "never the connection's PAT")
	})

	for name, tc := range map[string]struct {
		bearer  string
		headers map[string]string
		status  int
		want    string
	}{
		"a completed build": {"job-token", build("92", "rules-guid"), http.StatusForbidden,
			"Azure DevOps build 92 has completed; only a running build uploads its results"},
		"another repository's build": {"job-token", build("93", "rules-guid"), http.StatusForbidden,
			"Azure DevOps build 93 built repository other-guid, not rules-guid"},
		"a token Azure DevOps refuses": {"stolen", build("91", "rules-guid"), http.StatusUnauthorized,
			"Azure DevOps refused the job's access token for build 91"},
		"a build the project doesn't have": {"job-token", build("94", "rules-guid"), http.StatusForbidden,
			"Azure DevOps has no build 94 the job's access token can read"},
		"an unsynced repository": {"job-token", build("91", "nope-guid"), http.StatusForbidden,
			"Azure DevOps repository nope-guid is not synced by a gavel organization; add an azure_devops connection that brings it"},
		"no repository": {"job-token", map[string]string{CIAzureBuildHeader: "91"}, http.StatusBadRequest,
			"send " + CIAzureRepositoryHeader + " with the build's repository ID (BUILD_REPOSITORY_ID)"},
		"a build ID that isn't a number": {"job-token", build("../91", "rules-guid"), http.StatusBadRequest,
			`Azure DevOps build ID "../91" is not a positive number`},
	} {
		t.Run(name+" is refused", func(t *testing.T) {
			status, body := ciUpload(t, server.URL, tc.bearer, tc.headers)
			assert.Equal(t, tc.status, status, body)
			assert.Contains(t, ciError(t, body), tc.want)
		})
	}
}
