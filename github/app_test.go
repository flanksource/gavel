// Tests GitHub App authentication against a fake GitHub API: the App JWT's claims,
// installation token caching, installation lookup, and repository pagination.
package github

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeGitHubApp struct {
	*httptest.Server
	key          *rsa.PrivateKey
	tokenCalls   atomic.Int64
	tokenExpires time.Time
	repos        int
	lastClaims   atomic.Value
}

func newFakeGitHubApp(t *testing.T, repos int) *fakeGitHubApp {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	f := &fakeGitHubApp{key: key, tokenExpires: time.Now().Add(time.Hour), repos: repos}
	f.Server = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.Close)
	t.Setenv("GITHUB_API_URL", f.URL)
	return f
}

// appJWT checks the request carries an App JWT signed by the App's key, issued by App
// 123 and valid for 10 minutes from its iat. Times are checked against the App's clock,
// not this one, so the test can move the App's clock; the test checks them itself.
func (f *fakeGitHubApp) appJWT(r *http.Request) bool {
	raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok {
		return false
	}
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) { return &f.key.PublicKey, nil },
		jwt.WithValidMethods([]string{"RS256"}), jwt.WithoutClaimsValidation())
	if err != nil || claims["iss"] != "123" {
		return false
	}
	iat, _ := claims.GetIssuedAt()
	exp, _ := claims.GetExpirationTime()
	if iat == nil || exp == nil || exp.Sub(iat.Time) != 10*time.Minute {
		return false
	}
	f.lastClaims.Store(claims)
	return true
}

func (f *fakeGitHubApp) serve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/app/installations/42/access_tokens":
		if !f.appJWT(r) {
			http.Error(w, `{"message":"Bad credentials"}`, http.StatusUnauthorized)
			return
		}
		n := f.tokenCalls.Add(1)
		json.NewEncoder(w).Encode(map[string]any{"token": fmt.Sprintf("ghs_%d", n), "expires_at": f.tokenExpires.UTC().Format(time.RFC3339)}) //nolint:errcheck
	case r.Method == http.MethodGet && r.URL.Path == "/app/installations/42":
		if !f.appJWT(r) {
			http.Error(w, `{"message":"Bad credentials"}`, http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"id": 42, "account": map[string]any{"login": "acme", "id": 7, "type": "Organization"}}) //nolint:errcheck
	case r.Method == http.MethodGet && r.URL.Path == "/user/installations":
		if r.Header.Get("Authorization") != "Bearer ghu_ada" {
			http.Error(w, `{"message":"Bad credentials"}`, http.StatusUnauthorized)
			return
		}
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		all := []map[string]any{}
		for i := 0; i < 120; i++ {
			all = append(all, map[string]any{"id": 5000 + i, "app_id": 123, "account": map[string]any{"login": fmt.Sprintf("org%d", i), "id": i, "type": "Organization"}})
		}
		all = append(all, map[string]any{"id": 9999, "app_id": 456})
		var installs []map[string]any
		for i := (page - 1) * 100; i < len(all) && i < page*100; i++ {
			installs = append(installs, all[i])
		}
		json.NewEncoder(w).Encode(map[string]any{"total_count": len(all), "installations": installs}) //nolint:errcheck
	case r.Method == http.MethodGet && r.URL.Path == "/installation/repositories":
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ghs_") || r.URL.Query().Get("per_page") != "100" {
			http.Error(w, `{"message":"Bad credentials"}`, http.StatusUnauthorized)
			return
		}
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		repos := []map[string]any{}
		for i := (page - 1) * 100; i < f.repos && i < page*100; i++ {
			repos = append(repos, map[string]any{"id": 1000 + i, "full_name": fmt.Sprintf("acme/repo-%d", i), "private": i%2 == 0, "default_branch": "main"})
		}
		json.NewEncoder(w).Encode(map[string]any{"total_count": f.repos, "repositories": repos}) //nolint:errcheck
	default:
		http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
	}
}

func writeAppKey(t *testing.T, key *rsa.PrivateKey) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "app.pem")
	der := x509.MarshalPKCS1PrivateKey(key)
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der}), 0o600))
	return path
}

func TestAppInstallationToken(t *testing.T) {
	fake := newFakeGitHubApp(t, 0)
	app, err := LoadApp(123, "gavel", writeAppKey(t, fake.key))
	require.NoError(t, err)
	now := time.Now()
	app.now = func() time.Time { return now }
	ctx := context.Background()

	token, err := app.InstallationToken(ctx, 42)
	require.NoError(t, err)
	assert.Equal(t, "ghs_1", token)
	iat, _ := fake.lastClaims.Load().(jwt.MapClaims).GetIssuedAt()
	assert.WithinDuration(t, now.Add(-time.Minute), iat.Time, time.Second, "issued a minute early for clock skew")
	token, err = app.InstallationToken(ctx, 42)
	require.NoError(t, err)
	assert.Equal(t, "ghs_1", token, "cached until 5 minutes before it expires")
	assert.Equal(t, int64(1), fake.tokenCalls.Load())

	now = fake.tokenExpires.Add(-4 * time.Minute)
	token, err = app.InstallationToken(ctx, 42)
	require.NoError(t, err)
	assert.Equal(t, "ghs_2", token, "refreshed inside the 5 minute margin")

	_, err = app.InstallationToken(ctx, 99)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "installation 99")
	assert.Contains(t, err.Error(), "HTTP 404")
}

func TestAppInstallationAndRepos(t *testing.T) {
	fake := newFakeGitHubApp(t, 250)
	app, err := LoadApp(123, "gavel", writeAppKey(t, fake.key))
	require.NoError(t, err)
	ctx := context.Background()

	installation, err := app.Installation(ctx, 42)
	require.NoError(t, err)
	assert.Equal(t, Installation{ID: 42, Account: InstallationAccount{Login: "acme", ID: 7, Type: "Organization"}}, installation)

	repos, err := app.InstallationRepos(ctx, 42)
	require.NoError(t, err)
	require.Len(t, repos, 250, "every page is read")
	assert.Equal(t, InstallationRepo{ID: 1000, FullName: "acme/repo-0", Private: true, DefaultBranch: "main"}, repos[0])
	assert.Equal(t, "acme/repo-249", repos[249].FullName)
}

func TestLoadAppRejectsBadKeys(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.pem")
	_, err := LoadApp(123, "gavel", missing)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read the GitHub App private key")

	notPEM := filepath.Join(t.TempDir(), "bad.pem")
	require.NoError(t, os.WriteFile(notPEM, []byte("not a key"), 0o600))
	_, err = LoadApp(123, "gavel", notPEM)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse the GitHub App private key")

	_, err = LoadApp(0, "gavel", notPEM)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GitHub App ID")
}

func TestAppUserInstallations(t *testing.T) {
	fake := newFakeGitHubApp(t, 0)
	app, err := LoadApp(123, "gavel", writeAppKey(t, fake.key))
	require.NoError(t, err)

	installations, err := app.UserInstallations(context.Background(), "ghu_ada")
	require.NoError(t, err)
	require.Len(t, installations, 120, "every page, and only this App's installations")
	assert.Equal(t, Installation{ID: 5000, Account: InstallationAccount{Login: "org0", ID: 0, Type: "Organization"}}, installations[0])
	for _, installation := range installations {
		assert.NotEqual(t, int64(9999), installation.ID)
	}

	_, err = app.UserInstallations(context.Background(), "ghu_wrong")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Bad credentials")
	assert.NotContains(t, err.Error(), "ghu_wrong")
}
