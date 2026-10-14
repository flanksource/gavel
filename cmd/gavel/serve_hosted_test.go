// Tests `gavel serve --auth clerk` end to end: the real handler stack against a migrated
// database and the fake Clerk, from the public shell to an org's settings.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/api/clerktest"
	ghcache "github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/sshkeys"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

const hostedPublicURL = "https://gavel.example.com"

// hostedFixture is `gavel serve --auth clerk` running against a migrated database
// and the fake Clerk, whose user belongs to orgs "alpha" and "beta".
type hostedFixture struct {
	server *httptest.Server
	clerk  *clerktest.Server
	db     *database.DB
	opts   DashboardOptions
	stack  hostedStack
}

func newHostedFixture(t *testing.T) hostedFixture {
	t.Helper()
	dsn := commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_serve_hosted"}).DSN()
	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(ghcache.EnvDisable, "")
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	cache, err := ghcache.Open()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, cache.Close()) })

	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_ada", Name: "Ada Lovelace", Email: "ada@example.com",
		OrgID: "org_a", OrgSlug: "alpha", OrgRole: "org:admin", GitHubLogin: "ada"})
	clerk.Memberships = []clerktest.Membership{
		{OrgID: "org_a", OrgSlug: "alpha", OrgName: "Alpha", Role: "org:admin"},
		{OrgID: "org_b", OrgSlug: "beta", OrgName: "Beta", Role: "org:member"},
	}
	originalAuth, originalBackend := authHTTPClient, clerkBackendHTTPClient
	authHTTPClient, clerkBackendHTTPClient = clerk.Client(), clerk.BackendClient()
	t.Cleanup(func() { authHTTPClient, clerkBackendHTTPClient = originalAuth, originalBackend })

	opts := DashboardOptions{
		PRListOptions:       PRListOptions{Interval: "60s"},
		Auth:                authClerk,
		PublicURL:           hostedPublicURL,
		ClerkIssuer:         clerk.URL,
		ClerkPublishableKey: "pk_test_abc",
		ClerkClientID:       "cli",
		clerkSecretKey:      clerktest.SecretKey,
		secretKey:           base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))),
		DataDir:             t.TempDir(),
	}
	require.NoError(t, opts.Validate())
	stack, err := hostedDashboard(t.Context(), opts, opened.Gorm(), cache)
	require.NoError(t, err)
	t.Cleanup(stack.stop)
	server := httptest.NewServer(stack.handler)
	t.Cleanup(server.Close)
	return hostedFixture{server: server, clerk: clerk, db: opened, opts: opts, stack: stack}
}

func TestHostedDashboard(t *testing.T) {
	f := newHostedFixture(t)
	server, clerk, opened := f.server, f.clerk, f.db

	session := clerk.SessionToken(hostedPublicURL, time.Now().Add(time.Hour))
	do := func(method, path, body string, header http.Header) (int, string) {
		t.Helper()
		req, err := http.NewRequest(method, server.URL+path, strings.NewReader(body))
		require.NoError(t, err)
		for name, values := range header {
			req.Header[name] = values
		}
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		return resp.StatusCode, string(data)
	}
	cookie := http.Header{"Cookie": {"__session=" + session}}

	status, body := do(http.MethodGet, "/prs", "", nil)
	assert.Equal(t, http.StatusOK, status, "the page shell is public")
	assert.Contains(t, body, "<html")

	status, body = do(http.MethodGet, "/api/auth/config", "", nil)
	require.Equal(t, http.StatusOK, status)
	assert.JSONEq(t, `{"mode":"clerk","publishableKey":"pk_test_abc","issuer":"`+clerk.URL+`","clientId":"cli"}`, body)

	status, _ = do(http.MethodGet, "/api/me", "", nil)
	assert.Equal(t, http.StatusUnauthorized, status)

	status, body = do(http.MethodGet, "/api/me", "", cookie)
	require.Equal(t, http.StatusOK, status, body)
	var me ui.MeResponse
	require.NoError(t, json.Unmarshal([]byte(body), &me))
	require.NotNil(t, me.User)
	assert.Equal(t, "Ada Lovelace", me.User.Name)
	assert.Equal(t, "alpha", me.Org.Slug)
	assert.Len(t, me.Orgs, 2)

	status, body = do(http.MethodGet, "/api/me", "", http.Header{
		"Authorization": {"Bearer " + clerk.AccessToken(time.Now().Add(time.Hour))}, "X-Gavel-Org": {"beta"}})
	require.Equal(t, http.StatusOK, status, body)
	require.NoError(t, json.Unmarshal([]byte(body), &me))
	assert.Equal(t, "beta", me.Org.Slug, "the CLI picks its org with X-Gavel-Org")

	status, _ = do(http.MethodGet, "/api/activity", "", cookie)
	assert.Equal(t, http.StatusNotFound, status, "routes off the hosted allowlist")

	status, body = do(http.MethodPost, "/api/config", `{"repos":["acme/app"]}`,
		http.Header{"Cookie": cookie["Cookie"], "Origin": {hostedPublicURL}, "Content-Type": {"application/json"}})
	require.Equal(t, http.StatusOK, status, body)
	var alpha struct{ ID uuid.UUID }
	require.NoError(t, opened.Gorm().Raw(`SELECT id FROM orgs WHERE clerk_org_id = 'org_a'`).Scan(&alpha).Error)
	saved, err := ui.LoadOrgSearchConfig(t.Context(), opened.Gorm(), alpha.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"acme/app"}, saved.Repos)
}

func TestHostedPush(t *testing.T) {
	f := newHostedFixture(t)
	ctx := t.Context()
	_, err := f.stack.members.Orgs(ctx, "user_ada")
	require.NoError(t, err)
	var ada struct{ ID uuid.UUID }
	require.NoError(t, f.db.Gorm().Raw(`SELECT id FROM users WHERE clerk_user_id = 'user_ada'`).Scan(&ada).Error)

	_, private, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	signer, err := gossh.NewSignerFromKey(private)
	require.NoError(t, err)
	_, err = sshkeys.Store{DB: f.db.Gorm()}.Add(ctx, ada.ID, "laptop", string(gossh.MarshalAuthorizedKey(signer.PublicKey())))
	require.NoError(t, err)
	block, err := gossh.MarshalPrivateKey(private, "")
	require.NoError(t, err)
	keyPath := filepath.Join(t.TempDir(), "id_ed25519")
	require.NoError(t, os.WriteFile(keyPath, pem.EncodeToMemory(block), 0o600))

	f.opts.SSHHostKey = filepath.Join(t.TempDir(), "ssh_host_key")
	push, err := hostedPushServer(f.opts, f.db.Gorm(), f.stack.members)
	require.NoError(t, err)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go push.Serve(listener) //nolint:errcheck
	t.Cleanup(func() { _ = push.Close() })

	repo := t.TempDir()
	for _, args := range [][]string{{"init", "-b", "main"}, {"-c", "user.email=a@b.c", "-c", "user.name=A", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "first"}} {
		out, err := exec.Command("git", append([]string{"-C", repo}, args...)...).CombinedOutput()
		require.NoError(t, err, string(out))
	}
	gitPush := func(path string) (string, error) {
		cmd := exec.Command("git", "-C", repo, "push", "ssh://127.0.0.1/"+path, "main")
		cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p %d",
			keyPath, listener.Addr().(*net.TCPAddr).Port))
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	out, err := gitPush("alpha/app")
	require.NoError(t, err, out)
	assert.Contains(t, out, "push checks are not available yet")
	var alpha struct{ ID uuid.UUID }
	require.NoError(t, f.db.Gorm().Raw(`SELECT id FROM orgs WHERE slug = 'alpha'`).Scan(&alpha).Error)
	bares, err := filepath.Glob(filepath.Join(f.opts.DataDir, "repos", alpha.ID.String(), "*.git"))
	require.NoError(t, err)
	assert.Len(t, bares, 1, "the push landed in the org's repository directory")

	out, err = gitPush("gamma/app")
	require.Error(t, err)
	assert.Contains(t, out, "repository not found or access denied")
}
