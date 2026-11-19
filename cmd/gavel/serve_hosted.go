// Runs `gavel serve --auth clerk`: Clerk sign-in in front of one dashboard per
// organization, with no host-local features (sessions, git, processes) exposed.
package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/flanksource/commons-db/git"
	commonssecret "github.com/flanksource/commons-db/secret"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/commons/properties"
	"github.com/flanksource/gavel/api"
	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/forge"
	"github.com/flanksource/gavel/github"
	ghcache "github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/github/installations"
	"github.com/flanksource/gavel/github/sharedaccount"
	shared "github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/modelkeys"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/serve"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/verify"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// clerkBackendHTTPClient reaches Clerk's Backend API; nil is the default client. It is a
// package variable so tests can route it to a fake Clerk.
var clerkBackendHTTPClient *http.Client

const (
	// orgIdleTTL is how long an organization's dashboard stays built without requests.
	orgIdleTTL = 30 * time.Minute
	// membershipMaxAge is how stale gavel's copy of a user's Clerk memberships may get.
	membershipMaxAge = 5 * time.Minute
)

// githubNotConnected is every hosted org's GitHub credential until M5 mints App
// installation tokens; the host's own token is never used for an organization.
func githubNotConnected() (string, error) {
	return "", errors.New("connect GitHub to this organization to see its pull requests")
}

func runHostedDashboard(opts DashboardOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opened, err := shared.Shared(ctx, shared.WithMigrations())
	if err != nil {
		return fmt.Errorf("open the database: %w", err)
	}
	db, err := shared.Require(ctx, "gavel serve --auth clerk")
	if err != nil {
		return err
	}
	logger.Infof("%s", serveDatabaseStartupMessage(opened, 0))
	stack, err := hostedDashboard(ctx, opts, db, ghcache.Shared())
	if err != nil {
		return err
	}
	defer stack.stop()
	if err := startRetention(ctx, db, ghcache.Shared(), ui.TranscriptsDir(opts.DataDir)); err != nil {
		return err
	}
	ui.Build = ui.BuildInfo{Version: version, Commit: commit, Date: date}

	_, listener, err := bindUIListener(opts.Addr, opts.Port)
	if err != nil {
		return err
	}
	go http.Serve(listener, stack.handler) //nolint:errcheck
	logger.Infof("Hosted dashboard at %s (listening on %s)", opts.PublicURL, listener.Addr())

	if opts.SSHAddr != "" {
		push, err := hostedPushServer(opts, db, stack.members, stack.dispatcher)
		if err != nil {
			return err
		}
		defer push.Close() //nolint:errcheck
		pushListener, err := net.Listen("tcp", opts.SSHAddr)
		if err != nil {
			return fmt.Errorf("listen for pushes on %s: %w", opts.SSHAddr, err)
		}
		go push.Serve(pushListener) //nolint:errcheck
		logger.Infof("Hosted push server listening on %s", pushListener.Addr())
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	return nil
}

// hostedStack is what --auth clerk serves over HTTP, and the membership resolver the
// push server shares with it.
type hostedStack struct {
	handler http.Handler
	members *api.Members
	// dispatcher runs push checks; nil without --sandbox.
	dispatcher *sandbox.Dispatcher
	// stop releases every organization's Server.
	stop func()
}

// hostedDashboard builds the handler --auth clerk serves: the Authenticator in front of
// one dashboard Server per organization, and an org-less Server for page loads.
func hostedDashboard(ctx context.Context, opts DashboardOptions, db *gorm.DB, cache *ghcache.Store) (hostedStack, error) {
	// Each organization reaches its repositories with its own credentials, never the host's
	// connection files or local catalog.
	useLocalConnections(false)
	interval, err := time.ParseDuration(opts.Interval)
	if err != nil {
		return hostedStack{}, fmt.Errorf("invalid interval %q: %w", opts.Interval, err)
	}
	issuer := strings.TrimRight(opts.ClerkIssuer, "/")
	authConfig := ui.AuthConfigResponse{Mode: authClerk, PublishableKey: opts.ClerkPublishableKey, Issuer: issuer, ClientID: opts.ClerkClientID}
	backend := api.ClerkOrganizations{SecretKey: opts.clerkSecretKey, HTTPClient: clerkBackendHTTPClient}
	jwks := api.NewJWKSCache(issuer+"/.well-known/jwks.json", authHTTPClient, 10*time.Minute)
	auth := &ui.Authenticator{
		Session:   api.SessionVerifier{Issuer: issuer, AuthorizedParties: []string{strings.TrimRight(opts.PublicURL, "/")}, JWKS: jwks},
		Access:    api.NewAccessTokenVerifier(issuer, opts.ClerkClientID, jwks, backend, properties.Duration(10*time.Minute, "auth.token.cache.ttl")),
		Members:   api.NewMembers(db, backend, issuer, membershipMaxAge),
		PublicURL: opts.PublicURL,
	}

	var app *github.App
	var account hostedGitHub
	switch {
	case opts.GitHubAuth == githubAuthGH:
		if _, err := sharedaccount.GHToken(); err != nil {
			return hostedStack{}, fmt.Errorf("--github-auth gh: %w", err)
		}
		account = sharedaccount.New(db, opts.GitHubRepos, sharedaccount.GHToken)
	case opts.GitHubAppID != "":
		id, err := opts.gitHubAppID()
		if err != nil {
			return hostedStack{}, fmt.Errorf("--github-app-id %q: %w", opts.GitHubAppID, err)
		}
		if app, err = github.LoadApp(id, opts.GitHubAppSlug, opts.GitHubAppPrivateKey); err != nil {
			return hostedStack{}, err
		}
		account = installations.Store{DB: db, App: app}
	}
	// Sandbox jobs and hosted todo workspaces read repositories from the pushed ones and
	// from GitHub mirrors.
	sources := &sandbox.Sources{DataDir: opts.DataDir}
	if account != nil {
		sources.Token = account.RepoToken
	}
	blobs, err := verify.OpenGlobalStorage(ctx)
	if err != nil {
		return hostedStack{}, err
	}
	secret, err := base64.StdEncoding.DecodeString(opts.secretKey)
	if err != nil {
		return hostedStack{}, fmt.Errorf("decode GAVEL_SECRET_KEY: %w", err)
	}
	// Organizations' model keys are encrypted with a local key derived from the secret.
	commonssecret.KMSConnection = modelkeys.KeeperConnection(secret)
	commonssecret.InvalidateKeeper()
	// Sandbox jobs authenticate with their run token; runs on substrate reach it through serve,
	// as their run.
	runs := &ui.RunsAPI{DB: db, TokenKey: runTokenKey(secret), Cache: cache, Sources: sources}
	if account != nil {
		runs.PRs = func(ctx context.Context, org uuid.UUID, url string) (provider.Client, int, error) {
			return ui.OrganizationPRs(account.TokenSource(org), cache.ForOrg(org))(ctx, url)
		}
	}
	dispatcher, err := newSandboxDispatcher(ctx, opts, db, cache, secret, sources, runs.Handler())
	if err != nil {
		return hostedStack{}, err
	}
	substrateBackend, err := substrateAgentBackend(opts)
	if err != nil {
		return hostedStack{}, err
	}
	if dispatcher != nil {
		if sweeper, ok := dispatcher.Executor.(sandbox.Sweeper); ok {
			go sweepSandboxEvery(ctx, sweeper, dispatcher.RunLifetime())
		}
	}

	// stop waits for the organizations' PR polls and repository syncs, which it never cancels
	// partway through.
	var orgWork sync.WaitGroup
	orgs := ui.NewOrgServers(func(ctx context.Context, scope tenant.Scope) (*ui.Server, error) {
		cfg, err := ui.LoadOrgSearchConfig(ctx, db, scope.OrgID)
		if err != nil {
			return nil, err
		}
		ghOpts := github.Options{TokenSource: githubNotConnected, Cache: cache.ForOrg(scope.OrgID)}
		if account != nil {
			ghOpts.TokenSource = account.TokenSource(scope.OrgID)
		}
		srv := ui.NewServer(gavelctx.New(ctx), interval, ghOpts, cfg)
		srv.SetOrg(scope.OrgID, db)
		srv.SetBlobStore(blobs)
		srv.SetHostedMode(true)
		srv.SetHostedWorkspaces(ui.HostedWorkspaces{DB: db, Sources: sources})
		srv.SetFixtureSchemaProvider(func() (any, error) { return fixturesSchemaDocument() })
		if dispatcher != nil {
			srv.SetAgentRuns(ui.AgentRuns{
				Dispatcher: dispatcher, Substrate: substrateBackend, Sources: sources,
				Transcripts: ui.RunTranscripts{DB: db, DataDir: sources.DataDir}, RunAPI: runs.Handler(),
			})
		}
		srv.SetAuthConfig(authConfig)
		if err := srv.StartWorkspaceData(ctx, db); err != nil {
			return nil, err
		}
		if account != nil {
			poller := ui.NewPoller(srv, orgPRSearch(srv, account, cache, scope.OrgID), interval)
			orgWork.Go(func() { poller.Run(ctx) })
		}
		changed := make(chan struct{}, 1)
		srv.SetConnectionsChanged(func() {
			select {
			case changed <- struct{}{}:
			default:
			}
		})
		sync := orgRepoSync(db, account, scope.OrgID)
		orgWork.Go(func() { syncReposEvery(ctx, sync, srv, scope.OrgID, changed, repoSyncInterval) })
		return srv, nil
	}, orgIdleTTL)
	next := orgs.Handler()
	if app != nil {
		link := &ui.GitHubLink{App: app, Store: installations.Store{DB: db, App: app}, Clerk: backend, SecretKey: secret}
		mux := http.NewServeMux()
		mux.Handle("/api/github/", link.Handler())
		mux.Handle("/", next)
		next = mux
	}

	shellCtx, stopShell := context.WithCancel(ctx)
	shell := ui.NewServer(gavelctx.New(shellCtx), interval, github.Options{TokenSource: githubNotConnected}, ui.SearchConfig{})
	shell.SetHostedMode(true)
	shell.SetAuthConfig(authConfig)

	stop := func() {
		orgs.Close()
		stopShell()
		orgWork.Wait()
	}
	// CI jobs authenticate with their GitHub Actions OIDC token, not a Clerk session.
	ciResults := &ui.CIResults{
		DB:    db,
		OIDC:  api.GitHubOIDCVerifier{Audience: opts.PublicURL, JWKS: api.NewJWKSCache(api.GitHubOIDCJWKSURL, &http.Client{Timeout: 10 * time.Second}, 10*time.Minute)},
		Cache: cache,
	}
	root := http.NewServeMux()
	root.Handle("/api/ci/", ciResults.Handler())
	root.Handle("/api/runs/", runs.Handler())
	// Substrate's credential provider fetches organizations' model keys for their runs.
	if dispatcher != nil && opts.Sandbox == "substrate" && opts.credentialProviderToken != "" {
		credentials := &ui.CredentialsAPI{
			DB: db, Keys: modelkeys.Store{DB: db}, Atespace: opts.SubstrateAtespace,
			RunLifetime: dispatcher.RunLifetime(), Token: opts.credentialProviderToken,
		}
		root.Handle("/api/credentials/", credentials.Handler())
	}
	root.Handle("/", auth.Middleware(next, shell.Handler()))
	return hostedStack{handler: root, members: auth.Members, stop: stop, dispatcher: dispatcher}, nil
}

// hostedGitHub is how hosted orgs reach GitHub: their GitHub App installations
// (installations.Store), or one gh CLI login every org shares (sharedaccount.Account).
type hostedGitHub interface {
	// TokenSource is org's token for its PR dashboard and runs.
	TokenSource(org uuid.UUID) func() (string, error)
	// RepoToken is a token for a repository recorded with installationID.
	RepoToken(ctx context.Context, installationID int64) (string, error)
	// SearchGroups is org's repositories, grouped by the token they are searched with.
	SearchGroups(ctx context.Context, org uuid.UUID) ([]github.RepoGroup, error)
	// SyncAll re-reads org's repositories from GitHub.
	SyncAll(ctx context.Context, org uuid.UUID) error
}

// repoSyncInterval is how often an active organization's repositories are re-read from
// GitHub.
const repoSyncInterval = 10 * time.Minute

// repoSyncTimeout bounds one sync of an organization's repositories, which is also how long
// stopping the organization's dashboard can wait for it.
const repoSyncTimeout = 2 * time.Minute

// orgRepoSync re-reads org's repositories: from the server's GitHub account, when it has one,
// and from org's git connections, whose clients reach only public addresses since the
// organization, not the operator, chose them.
func orgRepoSync(db *gorm.DB, account hostedGitHub, org uuid.UUID) func(context.Context) error {
	return func(ctx context.Context) error {
		var errs []error
		if account != nil {
			errs = append(errs, account.SyncAll(ctx, org))
		}
		errs = append(errs, forge.SyncRepos(ctx, db, tenant.Scope{OrgID: org}, git.ClientOptions{Defensive: true}))
		return errors.Join(errs...)
	}
}

// syncReposEvery re-reads org's repositories with sync now, refreshing srv's PRs from them,
// then whenever changed signals and every interval until ctx ends. A sync in flight when ctx
// ends runs to completion rather than being cancelled partway through its queries.
func syncReposEvery(ctx context.Context, sync func(context.Context) error, srv *ui.Server, org uuid.UUID, changed <-chan struct{}, every time.Duration) {
	syncRepos := func() {
		syncCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), repoSyncTimeout)
		defer cancel()
		if err := sync(syncCtx); err != nil {
			logger.Warnf("sync the repositories of org %s: %v", org, err)
		}
	}
	syncRepos()
	select {
	case srv.RefreshCh() <- struct{}{}:
	default:
	}
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			syncRepos()
		case <-changed:
			syncRepos()
		}
	}
}

// orgPRSearch is an org's PR poller search: each group of its repositories, searched with
// that group's token, merged. An org without repositories has no PRs.
func orgPRSearch(srv *ui.Server, account hostedGitHub, cache *ghcache.Store, org uuid.UUID) func(time.Time) (github.PRSearchResults, *github.RateLimit, error) {
	return func(since time.Time) (github.PRSearchResults, *github.RateLimit, error) {
		groups, err := account.SearchGroups(context.Background(), org)
		if err != nil {
			return nil, nil, err
		}
		search := github.PRSearchOptions{State: "open", Since: since, Status: true, ShowAuthor: true}
		if srv.ShowClosed() {
			search.State = "all"
			if since.IsZero() {
				search.Since = time.Now().Add(-closedPRLookback)
			}
		}
		if !srv.IncludeBots() {
			search.ExcludeAuthors = srv.KnownBots()
		}
		var all github.PRSearchResults
		var rateLimit *github.RateLimit
		for _, group := range groups {
			search.Repos = group.Repos
			opts := github.Options{TokenSource: group.Token, Cache: cache.ForOrg(org)}
			results, rl, err := github.SearchPRs(opts, search)
			if err != nil {
				return all, rl, fmt.Errorf("search the pull requests of %s: %w", group.Name, err)
			}
			all = append(all, results...)
			rateLimit = rl
		}
		return all, rateLimit, nil
	}
}

// hostedPushServer is the git push server for --auth clerk: registered SSH keys only,
// pushes to <org>/<repo> for the org's members, and no hook ever runs on the host.
func hostedPushServer(opts DashboardOptions, db *gorm.DB, members *api.Members, dispatcher *sandbox.Dispatcher) (*serve.Server, error) {
	hostKey := opts.SSHHostKey
	if hostKey == "" {
		hostKey = filepath.Join(opts.DataDir, "ssh_host_key")
	}
	return serve.NewServer(serve.Options{
		HostKeyPath: hostKey,
		RepoDir:     filepath.Join(opts.DataDir, "repos"),
		Authorizer:  serve.DBAuthorizer{DB: db, Members: members, Issuer: strings.TrimRight(opts.ClerkIssuer, "/"), DataDir: opts.DataDir},
		OnPushed:    pushChecks(dispatcher),
	})
}

// runTokenKey derives the key run tokens are signed with from GAVEL_SECRET_KEY, so a run token
// can't be mistaken for another value signed with the secret.
func runTokenKey(secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte("gavel-run-token"))
	return mac.Sum(nil)
}
