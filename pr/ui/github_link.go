// Connects an organization to the gavel GitHub App: /api/github/install sends the user to
// GitHub, /api/github/setup links the installation they made, /api/github/sync re-reads repos.
package ui

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/github/installations"
	"github.com/flanksource/gavel/tenant"
)

// GitHubLink serves the GitHub App install routes for the request's organization.
type GitHubLink struct {
	App   *github.App
	Store installations.Store
	// Clerk reads the user's GitHub token, to check they can reach the installation.
	Clerk     api.ClerkOrganizations
	SecretKey []byte
}

// Handler serves the routes; it must sit behind the Authenticator, which sets the scope.
func (l *GitHubLink) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/github/install", l.install)
	mux.HandleFunc("GET /api/github/setup", l.setup)
	mux.HandleFunc("POST /api/github/sync", l.sync)
	return mux
}

func (l *GitHubLink) install(w http.ResponseWriter, r *http.Request) {
	scope := tenant.FromContext(r.Context())
	if scope.UserID == nil {
		writeJSONError(w, http.StatusUnauthorized, errors.New("sign in to connect GitHub"))
		return
	}
	state := installations.SignState(l.SecretKey, scope.OrgID, *scope.UserID, time.Now())
	http.Redirect(w, r, "https://github.com/apps/"+url.PathEscape(l.App.Slug)+"/installations/new?state="+url.QueryEscape(state), http.StatusFound)
}

// setup links the installation GitHub reports, after checking the state was signed for
// this org and user and that the user's GitHub account can reach the installation.
func (l *GitHubLink) setup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	scope := tenant.FromContext(ctx)
	query := r.URL.Query()
	state, err := installations.VerifyState(l.SecretKey, query.Get("state"), time.Now())
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	if scope.UserID == nil || state.User != *scope.UserID || state.Org != scope.OrgID {
		writeJSONError(w, http.StatusForbidden, errors.New("this GitHub install was started by another user or organization; start it again from gavel"))
		return
	}
	raw := query.Get("installation_id")
	if raw == "" {
		writeJSONError(w, http.StatusBadRequest, errors.New("GitHub didn't send an installation; if an organization owner must approve the App, connect GitHub again once they do"))
		return
	}
	installationID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("installation_id %q isn't a number", raw))
		return
	}
	var user struct{ ClerkUserID string }
	if err := l.Store.DB.WithContext(ctx).Raw(`SELECT clerk_user_id FROM users WHERE id = ?`, *scope.UserID).Scan(&user).Error; err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Errorf("read the signed-in user: %w", err))
		return
	}
	token, err := l.Clerk.GitHubToken(ctx, user.ClerkUserID)
	if err != nil {
		writeJSONError(w, http.StatusForbidden, fmt.Errorf("read your GitHub connection from Clerk: %w", err))
		return
	}
	reachable, err := l.App.UserInstallations(ctx, token)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err)
		return
	}
	if !slices.Contains(reachable, installationID) {
		writeJSONError(w, http.StatusForbidden, fmt.Errorf("your GitHub account can't access installation %d", installationID))
		return
	}
	installation, err := l.App.Installation(ctx, installationID)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err)
		return
	}
	err = l.Store.Link(ctx, scope.OrgID, *scope.UserID, installation)
	if errors.Is(err, installations.ErrLinkedToAnotherOrg) {
		writeJSONError(w, http.StatusConflict, err)
		return
	}
	if err == nil {
		err = l.Store.SyncRepos(ctx, scope.OrgID, installationID)
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (l *GitHubLink) sync(w http.ResponseWriter, r *http.Request) {
	org := tenant.FromContext(r.Context()).OrgID
	if err := l.Store.SyncAll(r.Context(), org); err != nil {
		writeJSONError(w, http.StatusBadGateway, err)
		return
	}
	repos, err := l.Store.Repos(r.Context(), org)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, map[string][]string{"repositories": repos})
}
