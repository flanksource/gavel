// Runs `gavel serve --auth clerk`: Clerk sign-in in front of one dashboard per
// organization, with no host-local features (sessions, git, processes) exposed.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/commons/properties"
	"github.com/flanksource/gavel/api"
	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/github"
	ghcache "github.com/flanksource/gavel/github/cache"
	shared "github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/tenant"
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
	handler, stop, err := hostedDashboard(ctx, opts, db, ghcache.Shared())
	if err != nil {
		return err
	}
	defer stop()
	ui.Build = ui.BuildInfo{Version: version, Commit: commit, Date: date}

	_, listener, err := bindUIListener(opts.Addr, opts.Port)
	if err != nil {
		return err
	}
	go http.Serve(listener, handler) //nolint:errcheck
	logger.Infof("Hosted dashboard at %s (listening on %s)", opts.PublicURL, listener.Addr())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	return nil
}

// hostedDashboard builds the handler --auth clerk serves: the Authenticator in front of
// one dashboard Server per organization, and an org-less Server for page loads.
// stop releases every organization's Server.
func hostedDashboard(ctx context.Context, opts DashboardOptions, db *gorm.DB, cache *ghcache.Store) (http.Handler, func(), error) {
	interval, err := time.ParseDuration(opts.Interval)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid interval %q: %w", opts.Interval, err)
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

	orgs := ui.NewOrgServers(func(ctx context.Context, scope tenant.Scope) (*ui.Server, error) {
		cfg, err := ui.LoadOrgSearchConfig(ctx, db, scope.OrgID)
		if err != nil {
			return nil, err
		}
		ghOpts := github.Options{TokenSource: githubNotConnected, Cache: cache.ForOrg(scope.OrgID)}
		srv := ui.NewServer(gavelctx.New(ctx), interval, ghOpts, cfg)
		srv.SetOrg(scope.OrgID, db)
		srv.SetHostedMode(true)
		srv.SetAuthConfig(authConfig)
		return srv, nil
	}, orgIdleTTL)

	shellCtx, stopShell := context.WithCancel(ctx)
	shell := ui.NewServer(gavelctx.New(shellCtx), interval, github.Options{TokenSource: githubNotConnected}, ui.SearchConfig{})
	shell.SetHostedMode(true)
	shell.SetAuthConfig(authConfig)

	stop := func() {
		orgs.Close()
		stopShell()
	}
	return auth.Middleware(orgs.Handler(), shell.Handler()), stop, nil
}
