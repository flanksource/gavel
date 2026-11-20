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
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/commons-db/secret"
	"github.com/flanksource/gavel/api/clerktest"
	"github.com/flanksource/gavel/connections"
	"github.com/flanksource/gavel/github"
	ghcache "github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/github/githubtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/internal/logtest"
	"github.com/flanksource/gavel/modelkeys"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/sshkeys"
	"github.com/flanksource/gavel/tenant"
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

func newHostedFixture(t *testing.T, configure ...func(*DashboardOptions)) hostedFixture {
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
	for _, change := range configure {
		change(&opts)
	}
	require.NoError(t, opts.Validate())
	t.Cleanup(func() { useLocalConnections(true) })
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

	status, body = do(http.MethodGet, "/api/todos/verification/schema", "", cookie)
	assert.Equal(t, http.StatusOK, status, "the verification fixture editor's schema: %s", body)

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
	push, err := hostedPushServer(f.opts, f.db.Gorm(), f.stack.members, f.stack.dispatcher)
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
	assert.Contains(t, out, "push checks need a sandbox: start gavel serve with --sandbox")
	var alpha struct{ ID uuid.UUID }
	require.NoError(t, f.db.Gorm().Raw(`SELECT id FROM orgs WHERE slug = 'alpha'`).Scan(&alpha).Error)
	bares, err := filepath.Glob(filepath.Join(f.opts.DataDir, "repos", alpha.ID.String(), "*.git"))
	require.NoError(t, err)
	assert.Len(t, bares, 1, "the push landed in the org's repository directory")

	out, err = gitPush("gamma/app")
	require.Error(t, err)
	assert.Contains(t, out, "repository not found or access denied")
}

func TestHostedGitHub(t *testing.T) {
	gh := githubtest.New(t)
	f := newHostedFixture(t, func(o *DashboardOptions) {
		o.GitHubAppID, o.GitHubAppSlug, o.GitHubAppPrivateKey = strconv.Itoa(githubtest.AppID), "gavel-test", gh.KeyPath
	})
	f.clerk.GitHubToken = "ghu_ada"
	gh.Install(42, github.InstallationAccount{Login: "acme", Type: "Organization"}, github.InstallationRepo{ID: 1, FullName: "acme/api", DefaultBranch: "main"})
	gh.GrantUser("ghu_ada", 42)
	gh.AddPullRequest(githubtest.PullRequest{Repo: "acme/api", Number: 7, Title: "Add login", Author: "ada"})

	noRedirects := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	cookie := "__session=" + f.clerk.SessionToken(hostedPublicURL, time.Now().Add(time.Hour))
	call := func(method, path, body string) *http.Response {
		t.Helper()
		req, err := http.NewRequest(method, f.server.URL+path, strings.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Cookie", cookie)
		req.Header.Set("Origin", hostedPublicURL)
		resp, err := noRedirects.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { resp.Body.Close() })
		return resp
	}

	install := call(http.MethodPost, "/api/github/install", "")
	installJSON := readAll(t, install)
	require.Equal(t, http.StatusOK, install.StatusCode, installJSON)
	var installBody struct{ URL string }
	require.NoError(t, json.Unmarshal([]byte(installJSON), &installBody))
	target, err := url.Parse(installBody.URL)
	require.NoError(t, err)
	assert.Equal(t, "/apps/gavel-test/installations/new", target.Path)

	// GitHub returns the browser to this page, which the SPA serves without a session:
	// the page then sends what GitHub reported with a fresh session token.
	page, err := http.Get(f.server.URL + "/github/setup?setup_action=install&installation_id=42&state=x")
	require.NoError(t, err)
	t.Cleanup(func() { page.Body.Close() })
	assert.Equal(t, http.StatusOK, page.StatusCode)
	assert.Contains(t, readAll(t, page), "<html")

	listed := call(http.MethodGet, "/api/github/installations", "")
	listedJSON := readAll(t, listed)
	require.Equal(t, http.StatusOK, listed.StatusCode, listedJSON)
	assert.JSONEq(t, `[{"id":42,"account":"acme","accountType":"Organization","status":"available"}]`, listedJSON)

	setupBody, err := json.Marshal(map[string]string{"installation_id": "42", "state": target.Query().Get("state")})
	require.NoError(t, err)
	setup := call(http.MethodPost, "/api/github/setup", string(setupBody))
	require.Equal(t, http.StatusOK, setup.StatusCode, readAll(t, setup))

	require.Equal(t, http.StatusAccepted, call(http.MethodPost, "/api/prs/refresh", "").StatusCode)
	var snapshot struct {
		Viewer string
		PRs    []struct {
			Repo   string
			Number int
		} `json:"prs"`
	}
	require.Eventually(t, func() bool {
		resp := call(http.MethodGet, "/api/prs", "")
		return json.NewDecoder(resp.Body).Decode(&snapshot) == nil && len(snapshot.PRs) == 1
	}, 10*time.Second, 50*time.Millisecond, "the org's poller fetches its installation's PRs")
	assert.Equal(t, "acme/api", snapshot.PRs[0].Repo)
	assert.Equal(t, 7, snapshot.PRs[0].Number)
	assert.Equal(t, "ada", snapshot.Viewer, "@me is the signed-in member")

	searched := false
	for _, search := range gh.Searches() {
		if search.Token == "ghs_42" && strings.Contains(search.Query, "repo:acme/api") {
			searched = true
		}
	}
	assert.True(t, searched, "searched with the installation's token: %+v", gh.Searches())
}

func TestHostedGitHubThroughTheGHCLI(t *testing.T) {
	gh := githubtest.New(t)
	bin := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(bin, "gh"), []byte("#!/bin/sh\n[ \"$1 $2\" = 'auth token' ] && echo gho_cli\n"), 0o755))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	gh.GrantUserRepos("gho_cli", github.InstallationRepo{ID: 1, FullName: "acme/api", DefaultBranch: "main"},
		github.InstallationRepo{ID: 2, FullName: "acme/web", DefaultBranch: "main"})
	gh.AddPullRequest(githubtest.PullRequest{Repo: "acme/api", Number: 7, Title: "Add login", Author: "ada"})
	gh.AddPullRequest(githubtest.PullRequest{Repo: "acme/web", Number: 8, Title: "Not served", Author: "ada"})
	f := newHostedFixture(t, func(o *DashboardOptions) { o.GitHubAuth, o.GitHubRepos = githubAuthGH, []string{"acme/api"} })

	cookie := "__session=" + f.clerk.SessionToken(hostedPublicURL, time.Now().Add(time.Hour))
	call := func(method, path string) *http.Response {
		t.Helper()
		req, err := http.NewRequest(method, f.server.URL+path, nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", cookie)
		req.Header.Set("Origin", hostedPublicURL)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { resp.Body.Close() })
		return resp
	}

	var snapshot struct {
		PRs []struct {
			Repo   string
			Number int
		} `json:"prs"`
	}
	require.Eventually(t, func() bool {
		resp := call(http.MethodGet, "/api/prs")
		return json.NewDecoder(resp.Body).Decode(&snapshot) == nil && len(snapshot.PRs) == 1
	}, 10*time.Second, 50*time.Millisecond, "the org's poller fetches the named repository's PRs without any installation")
	assert.Equal(t, "acme/api", snapshot.PRs[0].Repo)
	assert.Equal(t, 7, snapshot.PRs[0].Number)

	for _, search := range gh.Searches() {
		assert.Equal(t, "gho_cli", search.Token, "searched with the gh CLI's login")
		assert.NotContains(t, search.Query, "acme/web", "only the named repositories are searched")
	}
	var held []string
	require.NoError(t, f.db.Gorm().Raw(`SELECT name FROM repos WHERE source = 'github' AND installation_id IS NULL`).Scan(&held).Error)
	assert.Equal(t, []string{"acme/api"}, held, "the org holds the named repository for sandbox runs and todos")

	assert.Equal(t, http.StatusNotFound, call(http.MethodPost, "/api/github/install").StatusCode, "there is no App to install")
}

func TestAHostedOrgSyncsASavedConnectionThroughPublicAddressesOnly(t *testing.T) {
	logged := logtest.Capture(t)
	f := newHostedFixture(t)
	t.Setenv("GAVEL_SECRET_KEEPER_CONNECTION", "")
	t.Setenv("GAVEL_SECRET_KEY", f.opts.secretKey)
	var reached atomic.Int64
	gitlab := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached.Add(1)
		_, _ = w.Write([]byte("[]"))
	}))
	t.Cleanup(gitlab.Close)

	body := `{"name":"acme-gitlab","type":"gitlab","url":"` + gitlab.URL + `","password":"glpat","properties":{"group":"acme"}}`
	req, err := http.NewRequest(http.MethodPost, f.server.URL+"/api/connections", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Cookie", "__session="+f.clerk.SessionToken(hostedPublicURL, time.Now().Add(time.Hour)))
	req.Header.Set("Origin", hostedPublicURL)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, readAll(t, resp))
	require.NoError(t, resp.Body.Close())

	assert.Eventually(t, func() bool {
		lines := logged.String()
		return strings.Contains(lines, `list the repositories of connection "acme-gitlab"`) && strings.Contains(lines, "refuses localhost address 127.0.0.1")
	}, 10*time.Second, 20*time.Millisecond, "saving the connection syncs it, refusing an address that isn't public: %s", logged.String())
	assert.Zero(t, reached.Load(), "nothing reached the loopback API")
}

func TestAHostedOrgWithoutTheGitHubAppListsItsAzurePullRequestsThroughPublicAddressesOnly(t *testing.T) {
	logged := logtest.Capture(t)
	f := newHostedFixture(t)
	t.Setenv("GAVEL_SECRET_KEEPER_CONNECTION", "")
	t.Setenv("GAVEL_SECRET_KEY", f.opts.secretKey)
	var reached atomic.Int64
	azure := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached.Add(1)
		_, _ = w.Write([]byte(`{"value":[]}`))
	}))
	t.Cleanup(azure.Close)
	call := func(method, path, body string) string {
		t.Helper()
		req, err := http.NewRequest(method, f.server.URL+path, strings.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Cookie", "__session="+f.clerk.SessionToken(hostedPublicURL, time.Now().Add(time.Hour)))
		req.Header.Set("Origin", hostedPublicURL)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		data := readAll(t, resp)
		require.NoError(t, resp.Body.Close())
		require.Less(t, resp.StatusCode, 300, data)
		return data
	}

	call(http.MethodPost, "/api/connections", `{"name":"acme-azure","type":"azure_devops","url":"`+azure.URL+`/acme","password":"azure-pat"}`)
	var org, connection struct{ ID uuid.UUID }
	require.NoError(t, f.db.Gorm().Raw(`SELECT id FROM orgs WHERE clerk_org_id = 'org_a'`).Scan(&org).Error)
	require.NoError(t, f.db.Gorm().Raw(`SELECT id FROM connections WHERE org_id = ? AND name = 'acme-azure'`, org.ID).Scan(&connection).Error)
	require.NoError(t, f.db.Gorm().Exec(`INSERT INTO repos (org_id, name, source, connection_id, remote_url, provider_repo_id, default_branch)
		VALUES (?, 'Platform/rules', 'azure_devops', ?, ?, 'rules-id', 'main')`, org.ID, connection.ID, azure.URL+"/acme/Platform/_git/rules").Error)
	call(http.MethodPost, "/api/prs/refresh", "")

	assert.Eventually(t, func() bool {
		var listed struct{ Error string }
		require.NoError(t, json.Unmarshal([]byte(call(http.MethodGet, "/api/prs", "")), &listed))
		return strings.Contains(listed.Error, "list the pull requests of Platform/rules") && strings.Contains(listed.Error, "refuses localhost address 127.0.0.1")
	}, 10*time.Second, 50*time.Millisecond, "the org's poller lists the repository's pull requests, refusing an address that isn't public: %s", logged.String())
	assert.Zero(t, reached.Load(), "nothing reached the loopback API")
}

func TestHostedSourcesDropGitHubIDMirrorsAndReachConnectionsDefensively(t *testing.T) {
	dataDir := t.TempDir()
	byGitHubID := filepath.Join(dataDir, "mirrors", uuid.NewString(), "42.git")
	require.NoError(t, os.MkdirAll(filepath.Join(byGitHubID, "objects"), 0o755))
	f := newHostedFixture(t, func(o *DashboardOptions) { o.DataDir = dataDir })
	assert.NoDirExists(t, byGitHubID, "a mirror named by GitHub repository id is mirrored again under its repository")

	t.Setenv("GAVEL_SECRET_KEEPER_CONNECTION", "")
	t.Setenv("GAVEL_SECRET_KEY", f.opts.secretKey)
	org := uuid.New()
	require.NoError(t, f.db.Gorm().Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)
	store, closeStore, err := connections.OpenStore(t.Context(), connections.Options{DB: f.db.Gorm(), Scope: tenant.Scope{OrgID: org}})
	require.NoError(t, err)
	remote := "https://127.0.0.1:9/acme/app.git"
	saved, err := store.SaveConnection(t.Context(), &models.Connection{Name: "app", Type: models.ConnectionTypeGit, URL: remote, Password: "secret"})
	require.NoError(t, err)
	require.NoError(t, closeStore())

	sources, err := hostedSources(dataDir, f.db.Gorm(), nil)
	require.NoError(t, err)
	_, err = sources.Bundle(t.Context(), sandbox.Repo{ID: uuid.New(), Org: org, Source: "git", Name: "acme/app", ConnectionID: &saved.ID, RemoteURL: remote},
		"refs/heads/main", io.Discard)
	require.ErrorContains(t, err, "refuses localhost address 127.0.0.1", "a connection's remote is reached only at a public address")
}

func TestStoppingTheHostedStackLetsARepositorySyncFinish(t *testing.T) {
	gh := githubtest.New(t)
	f := newHostedFixture(t, func(o *DashboardOptions) {
		o.GitHubAppID, o.GitHubAppSlug, o.GitHubAppPrivateKey = strconv.Itoa(githubtest.AppID), "gavel-test", gh.KeyPath
	})
	// The org's first repository sync waits on this lock.
	lock := f.db.Gorm().Begin()
	require.NoError(t, lock.Exec(`LOCK TABLE github_installations IN ACCESS EXCLUSIVE MODE`).Error)
	t.Cleanup(func() { lock.Rollback() })

	req, err := http.NewRequest(http.MethodGet, f.server.URL+"/api/prs", nil)
	require.NoError(t, err)
	req.Header.Set("Cookie", "__session="+f.clerk.SessionToken(hostedPublicURL, time.Now().Add(time.Hour)))
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, readAll(t, resp))
	require.NoError(t, resp.Body.Close())
	require.Eventually(t, func() bool {
		var waiting int
		require.NoError(t, f.db.Gorm().Raw(`SELECT count(*) FROM pg_stat_activity WHERE pid <> pg_backend_pid()
			AND wait_event_type = 'Lock' AND query LIKE '%SELECT id FROM github_installations WHERE org_id%'`).Scan(&waiting).Error)
		return waiting > 0
	}, 10*time.Second, 20*time.Millisecond, "the org's repository sync reads its installations")

	stopped := make(chan struct{})
	go func() {
		f.stack.stop()
		close(stopped)
	}()
	select {
	case <-stopped:
		t.Fatal("stopping returned while the org's repository sync was still in flight")
	case <-time.After(300 * time.Millisecond):
	}
	require.NoError(t, lock.Commit().Error)
	select {
	case <-stopped:
	case <-time.After(10 * time.Second):
		t.Fatal("stopping never returned after the org's repository sync could finish")
	}
}

func TestStoppingTheHostedStackLetsAPRPollFinish(t *testing.T) {
	gh := githubtest.New(t)
	bin := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(bin, "gh"), []byte("#!/bin/sh\n[ \"$1 $2\" = 'auth token' ] && echo gho_cli\n"), 0o755))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	gh.GrantUserRepos("gho_cli", github.InstallationRepo{ID: 1, FullName: "acme/api", DefaultBranch: "main"})
	gh.AddPullRequest(githubtest.PullRequest{Repo: "acme/api", Number: 7, Title: "Add login", Author: "ada"})
	f := newHostedFixture(t, func(o *DashboardOptions) { o.GitHubAuth, o.GitHubRepos = githubAuthGH, []string{"acme/api"} })
	// The org's poller indexes the PR it finds; indexing it waits on this lock.
	lock := f.db.Gorm().Begin()
	require.NoError(t, lock.Exec(`LOCK TABLE pull_requests IN SHARE MODE`).Error)
	t.Cleanup(func() { lock.Rollback() })

	req, err := http.NewRequest(http.MethodGet, f.server.URL+"/api/prs", nil)
	require.NoError(t, err)
	req.Header.Set("Cookie", "__session="+f.clerk.SessionToken(hostedPublicURL, time.Now().Add(time.Hour)))
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, readAll(t, resp))
	require.NoError(t, resp.Body.Close())
	require.Eventually(t, func() bool {
		var waiting int
		require.NoError(t, f.db.Gorm().Raw(`SELECT count(*) FROM pg_stat_activity WHERE pid <> pg_backend_pid()
			AND wait_event_type = 'Lock' AND query LIKE '%INSERT INTO "pull_requests"%'`).Scan(&waiting).Error)
		return waiting > 0
	}, 10*time.Second, 20*time.Millisecond, "the org's poller indexes the PR it found")

	stopped := make(chan struct{})
	go func() {
		f.stack.stop()
		close(stopped)
	}()
	select {
	case <-stopped:
		t.Fatal("stopping returned while the org's PR poll was still in flight")
	case <-time.After(300 * time.Millisecond):
	}
	require.NoError(t, lock.Commit().Error)
	select {
	case <-stopped:
	case <-time.After(10 * time.Second):
		t.Fatal("stopping never returned after the org's PR poll could finish")
	}
}

func readAll(t *testing.T, resp *http.Response) string {
	t.Helper()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(data)
}

func TestHostedDashboardRefusesAnIncompleteStorageConfig(t *testing.T) {
	f := newHostedFixture(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte("storage:\n  type: gcs\n  bucket: blobs\n"), 0o644))

	_, err := hostedDashboard(t.Context(), f.opts, f.db.Gorm(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "storage.gcs.credentials is required")
}

func TestHostedDashboardAcceptsCIResultsWithoutClerk(t *testing.T) {
	f := newHostedFixture(t)
	resp, err := http.Post(f.server.URL+"/api/ci/results", "application/json", strings.NewReader(`{}`))
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, string(body), "GitHub Actions OIDC token", "answered by the CI endpoint, not the Clerk middleware")
}

func TestHostedDashboardServesTheRunAPIWithoutClerk(t *testing.T) {
	f := newHostedFixture(t)
	resp, err := http.Get(f.server.URL + "/api/runs/" + uuid.NewString() + "/spec")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, string(body), "run token", "answered by the run API, not the Clerk middleware")
}

func TestHostedDashboardServesTheCredentialFetchWithoutClerk(t *testing.T) {
	fetch := func(f hostedFixture) (int, string) {
		t.Helper()
		resp, err := http.Post(f.server.URL+"/api/credentials/fetch", "application/json", strings.NewReader(`{}`))
		require.NoError(t, err)
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(body)
	}
	logged := logtest.Capture(t)
	substrate := newHostedFixture(t, func(o *DashboardOptions) {
		o.Sandbox, o.SubstrateTemplate = "substrate", "gavel-runner"
		o.SubstrateEnvAPI, o.SubstrateControl = "ate-env-api.example:443", "substrate-api.example:443"
		o.SubstrateGavelName, o.credentialProviderToken = "beta", "provider-token"
	})
	status, body := fetch(substrate)
	assert.Equal(t, http.StatusUnauthorized, status)
	assert.Contains(t, body, "credential provider's token", "answered by the credential fetch, not the Clerk middleware")
	// The substrate the fixture names does not exist, so the sweep the server starts with fails.
	assert.Eventually(t, func() bool {
		return strings.Contains(logged.String(), "sweep the sandbox environments runs left behind: list the actors of atespace")
	}, 10*time.Second, 20*time.Millisecond, "the failed sweep is a warning")

	status, body = fetch(newHostedFixture(t))
	assert.NotContains(t, body, "credential provider", "not served without substrate and a provider token (%d)", status)
}

func TestCredentialProviderTokenComesFromTheEnvironment(t *testing.T) {
	opts := DashboardOptions{}.withEnv(func(name string) string {
		return map[string]string{"GAVEL_CREDENTIAL_PROVIDER_TOKEN": "provider-token"}[name]
	})
	assert.Equal(t, "provider-token", opts.credentialProviderToken)
}

func TestRunTokenKeyIsDerivedFromTheSecret(t *testing.T) {
	secret := []byte(strings.Repeat("k", 32))
	key := runTokenKey(secret)
	assert.Len(t, key, 32)
	assert.NotEqual(t, secret, key, "not the secret itself")
	assert.Equal(t, key, runTokenKey(secret))
}

func TestHostedDashboardEncryptsModelKeysWithTheSecret(t *testing.T) {
	previous := secret.KMSConnection
	t.Cleanup(func() { secret.KMSConnection = previous; secret.InvalidateKeeper() })
	secret.KMSConnection = ""
	newHostedFixture(t)
	assert.Equal(t, modelkeys.KeeperConnection([]byte(strings.Repeat("k", 32))), secret.KMSConnection)
}

func TestHostedDashboardWithoutASandboxHasNoDispatcher(t *testing.T) {
	f := newHostedFixture(t)
	assert.Nil(t, f.stack.dispatcher, "pushes are only stored")
}

func TestHostedDashboardRefusesAnUnpreparedContainerSandbox(t *testing.T) {
	f := newHostedFixture(t)
	opts := f.opts
	opts.Sandbox, opts.SandboxImage, opts.SandboxServeAddr = "docker", "runner:1", "10.0.0.5:9092"
	opts.SandboxNetwork = "gavel-sandbox-missing-" + uuid.NewString()[:8]
	require.NoError(t, opts.Validate())

	logged := logtest.Capture(t)
	_, err := hostedDashboard(t.Context(), opts, f.db.Gorm(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "the sandbox network "+opts.SandboxNetwork+" is missing")
	assert.Contains(t, err.Error(), "docker network create "+opts.SandboxNetwork)
	assert.Contains(t, logged.String(), "--sandbox docker: runs get no model keys")
}

func TestHostedDashboardServesTodosInTheOrgsRepositories(t *testing.T) {
	f := newHostedFixture(t)
	cookie := "__session=" + f.clerk.SessionToken(hostedPublicURL, time.Now().Add(time.Hour))
	do := func(method, path, body string) (int, string) {
		t.Helper()
		req, err := http.NewRequest(method, f.server.URL+path, strings.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Cookie", cookie)
		req.Header.Set("Origin", hostedPublicURL)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		return resp.StatusCode, string(data)
	}
	status, body := do(http.MethodGet, "/api/me", "")
	require.Equal(t, http.StatusOK, status, body)
	var org struct{ ID uuid.UUID }
	require.NoError(t, f.db.Gorm().Raw(`SELECT id FROM orgs WHERE clerk_org_id = 'org_a'`).Scan(&org).Error)
	var repo struct{ ID uuid.UUID }
	require.NoError(t, f.db.Gorm().Raw(`INSERT INTO repos (org_id, name, source) VALUES (?, 'app', 'push') RETURNING id`, org.ID).Scan(&repo).Error)
	work := t.TempDir()
	gitIn(t, work, "init", "-q", "-b", "main")
	gitIn(t, work, "commit", "-q", "--allow-empty", "-m", "init")
	gitIn(t, work, "clone", "-q", "--bare", work, filepath.Join(f.opts.DataDir, "repos", org.ID.String(), repo.ID.String()+".git"))
	workspace := "repo:" + repo.ID.String()

	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		status, body = do(http.MethodGet, "/api/projects", "")
		assert.Equal(collect, http.StatusOK, status, body)
		assert.Contains(collect, body, `"dir":"`+workspace+`"`)
	}, 5*time.Second, 20*time.Millisecond)

	status, body = do(http.MethodPost, "/api/todos", `{"dir":"`+workspace+`","title":"Fix the build"}`)
	require.Equal(t, http.StatusCreated, status, body)
	var created struct{ ID string }
	require.NoError(t, json.Unmarshal([]byte(body), &created))

	status, body = do(http.MethodPost, "/api/todos/run", `{"dir":"`+workspace+`","ref":"`+created.ID+`","step":"plan"}`)
	assert.Equal(t, http.StatusNotImplemented, status)
	assert.Contains(t, body, "agent runs need a substrate sandbox: start gavel serve with --sandbox substrate")
}
