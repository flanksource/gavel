// Runs `gavel serve --auth clerk`: Clerk sign-in in front of one dashboard per
// organization, with no host-local features (sessions, git, processes) exposed.
package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/commons/properties"
	"github.com/flanksource/gavel/api"
	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/github"
	ghcache "github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/github/installations"
	shared "github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/serve"
	"github.com/flanksource/gavel/tenant"
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
	ui.Build = ui.BuildInfo{Version: version, Commit: commit, Date: date}

	_, listener, err := bindUIListener(opts.Addr, opts.Port)
	if err != nil {
		return err
	}
	go http.Serve(listener, stack.handler) //nolint:errcheck
	logger.Infof("Hosted dashboard at %s (listening on %s)", opts.PublicURL, listener.Addr())

	if opts.SSHAddr != "" {
		push, err := hostedPushServer(opts, db, stack.members)
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
	// stop releases every organization's Server.
	stop func()
}

// hostedDashboard builds the handler --auth clerk serves: the Authenticator in front of
// one dashboard Server per organization, and an org-less Server for page loads.
func hostedDashboard(ctx context.Context, opts DashboardOptions, db *gorm.DB, cache *ghcache.Store) (hostedStack, error) {
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
	if opts.GitHubAppID != 0 {
		if app, err = github.LoadApp(int64(opts.GitHubAppID), opts.GitHubAppSlug, opts.GitHubAppPrivateKey); err != nil {
			return hostedStack{}, err
		}
	}
	store := installations.Store{DB: db, App: app}

	orgs := ui.NewOrgServers(func(ctx context.Context, scope tenant.Scope) (*ui.Server, error) {
		cfg, err := ui.LoadOrgSearchConfig(ctx, db, scope.OrgID)
		if err != nil {
			return nil, err
		}
		ghOpts := github.Options{TokenSource: githubNotConnected, Cache: cache.ForOrg(scope.OrgID)}
		if app != nil {
			ghOpts.TokenSource = store.TokenSource(scope.OrgID)
		}
		srv := ui.NewServer(gavelctx.New(ctx), interval, ghOpts, cfg)
		srv.SetOrg(scope.OrgID, db)
		srv.SetHostedMode(true)
		srv.SetAuthConfig(authConfig)
		if app != nil {
			ui.NewPoller(srv, orgPRSearch(srv, store, cache, scope.OrgID), interval).Start(ctx)
			go syncReposEvery(ctx, store, scope.OrgID, repoSyncInterval)
		}
		return srv, nil
	}, orgIdleTTL)
	var next http.Handler = orgs.Handler()
	if app != nil {
		secret, err := base64.StdEncoding.DecodeString(opts.secretKey)
		if err != nil {
			return hostedStack{}, fmt.Errorf("decode GAVEL_SECRET_KEY: %w", err)
		}
		link := &ui.GitHubLink{App: app, Store: store, Clerk: backend, SecretKey: secret}
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
	}
	return hostedStack{handler: auth.Middleware(next, shell.Handler()), members: auth.Members, stop: stop}, nil
}

// repoSyncInterval is how often an active organization's repositories are re-read from
// its GitHub App installations.
const repoSyncInterval = 10 * time.Minute

// syncReposEvery re-reads org's installation repositories until ctx ends.
func syncReposEvery(ctx context.Context, store installations.Store, org uuid.UUID, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := store.SyncAll(ctx, org); err != nil && ctx.Err() == nil {
				logger.Warnf("sync the repositories of org %s: %v", org, err)
			}
		}
	}
}

// orgPRSearch is an org's PR poller search: each installation's repositories, searched
// with that installation's token, merged. An org without repositories has no PRs.
func orgPRSearch(srv *ui.Server, store installations.Store, cache *ghcache.Store, org uuid.UUID) func(time.Time) (github.PRSearchResults, *github.RateLimit, error) {
	return func(since time.Time) (github.PRSearchResults, *github.RateLimit, error) {
		groups, err := store.ReposByInstallation(context.Background(), org)
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
			opts := github.Options{TokenSource: store.InstallationTokenSource(group.Installation), Cache: cache.ForOrg(org)}
			results, rl, err := github.SearchPRs(opts, search)
			if err != nil {
				return all, rl, fmt.Errorf("search the pull requests of installation %d: %w", group.Installation, err)
			}
			all = append(all, results...)
			rateLimit = rl
		}
		return all, rateLimit, nil
	}
}

// hostedPushServer is the git push server for --auth clerk: registered SSH keys only,
// pushes to <org>/<repo> for the org's members, and no hook ever runs on the host.
func hostedPushServer(opts DashboardOptions, db *gorm.DB, members *api.Members) (*serve.Server, error) {
	hostKey := opts.SSHHostKey
	if hostKey == "" {
		hostKey = filepath.Join(opts.DataDir, "ssh_host_key")
	}
	return serve.NewServer(serve.Options{
		HostKeyPath: hostKey,
		RepoDir:     filepath.Join(opts.DataDir, "repos"),
		Authorizer:  serve.DBAuthorizer{DB: db, Members: members, Issuer: strings.TrimRight(opts.ClerkIssuer, "/"), DataDir: opts.DataDir},
		// Push checks need sandboxed execution (M9.5); until then a push is only stored.
		OnPushed: func(_ context.Context, push serve.Push) {
			fmt.Fprintln(push.Out, "push received; push checks are not available yet")
		},
	})
}
