// Connects an organization to the gavel GitHub App: linking an installation the user's GitHub
// account already reaches, or a new one made on GitHub's install page; and re-reading repos.
package ui

import (
	"encoding/json"
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
	"github.com/google/uuid"
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
// Install and setup are POSTs from the dashboard rather than browser navigations, so they
// carry a fresh session token: the session cookie expires within a minute, and a trip
// through GitHub's install page can take longer.
func (l *GitHubLink) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/github/install", l.install)
	mux.HandleFunc("POST /api/github/setup", l.setup)
	mux.HandleFunc("GET /api/github/installations", l.installations)
	mux.HandleFunc("POST /api/github/installations/{id}/link", l.linkInstallation)
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
	writeJSON(w, map[string]string{"url": "https://github.com/apps/" + url.PathEscape(l.App.Slug) + "/installations/new?state=" + url.QueryEscape(state)})
}

// setupRequest is what GitHub sent back to the dashboard's /github/setup page.
type setupRequest struct {
	InstallationID string `json:"installation_id"`
	State          string `json:"state"`
}

// setup links the installation GitHub reports, after checking the state was signed for
// this org and user and that the user's GitHub account can reach the installation, and
// answers with the organization's repositories.
func (l *GitHubLink) setup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	scope := tenant.FromContext(ctx)
	var body setupRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("the setup request must be JSON with installation_id and state: %w", err))
		return
	}
	state, err := installations.VerifyState(l.SecretKey, body.State, time.Now())
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	if scope.UserID == nil || state.User != *scope.UserID || state.Org != scope.OrgID {
		writeJSONError(w, http.StatusForbidden, errors.New("this GitHub install was started by another user or organization; start it again from gavel"))
		return
	}
	raw := body.InstallationID
	if raw == "" {
		writeJSONError(w, http.StatusBadRequest, errors.New("GitHub didn't send an installation; if an organization owner must approve the App, connect GitHub again once they do"))
		return
	}
	installationID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("installation_id %q isn't a number", raw))
		return
	}
	l.link(w, r, installationID)
}

// installationStatus says whether an installation is linked to the caller's organization
// ("linked"), to another organization ("linked_elsewhere"), or to none ("available").
type installationStatus struct {
	ID          int64  `json:"id"`
	Account     string `json:"account"`
	AccountType string `json:"accountType"`
	Status      string `json:"status"`
}

// installations lists the App installations the user's GitHub account can access, so
// one installed before (or by someone else) can be linked without visiting GitHub.
func (l *GitHubLink) installations(w http.ResponseWriter, r *http.Request) {
	org := tenant.FromContext(r.Context()).OrgID
	reachable, ok := l.reachable(w, r)
	if !ok {
		return
	}
	ids := make([]int64, len(reachable))
	for i, installation := range reachable {
		ids[i] = installation.ID
	}
	owners, err := l.Store.Owners(r.Context(), ids)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	list := make([]installationStatus, len(reachable))
	for i, installation := range reachable {
		status := "available"
		if owner, linked := owners[installation.ID]; linked {
			status = "linked_elsewhere"
			if owner == org {
				status = "linked"
			}
		}
		list[i] = installationStatus{ID: installation.ID, Account: installation.Account.Login, AccountType: installation.Account.Type, Status: status}
	}
	writeJSON(w, list)
}

// linkInstallation links an installation the user picked from installations. The click
// is the intent GitHub's state stands for in setup: the request carries a bearer token,
// which no other site can send.
func (l *GitHubLink) linkInstallation(w http.ResponseWriter, r *http.Request) {
	raw := r.PathValue("id")
	installationID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("installation ID %q isn't a number", raw))
		return
	}
	l.link(w, r, installationID)
}

// reachable lists the App installations the signed-in user's GitHub account can access,
// answering the request with the reason when they can't be read.
func (l *GitHubLink) reachable(w http.ResponseWriter, r *http.Request) ([]github.Installation, bool) {
	ctx := r.Context()
	scope := tenant.FromContext(ctx)
	if scope.UserID == nil {
		writeJSONError(w, http.StatusUnauthorized, errors.New("sign in to connect GitHub"))
		return nil, false
	}
	var user struct{ ClerkUserID string }
	if err := l.Store.DB.WithContext(ctx).Raw(`SELECT clerk_user_id FROM users WHERE id = ?`, *scope.UserID).Scan(&user).Error; err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Errorf("read the signed-in user: %w", err))
		return nil, false
	}
	token, err := l.Clerk.GitHubToken(ctx, user.ClerkUserID)
	if err != nil {
		writeJSONError(w, http.StatusForbidden, fmt.Errorf("read your GitHub connection from Clerk: %w", err))
		return nil, false
	}
	reachable, err := l.App.UserInstallations(ctx, token)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err)
		return nil, false
	}
	return reachable, true
}

// link links installationID to the caller's organization once the user's GitHub account
// is shown to reach it, syncs its repositories, and answers with the org's repositories.
func (l *GitHubLink) link(w http.ResponseWriter, r *http.Request, installationID int64) {
	ctx := r.Context()
	scope := tenant.FromContext(ctx)
	reachable, ok := l.reachable(w, r)
	if !ok {
		return
	}
	i := slices.IndexFunc(reachable, func(installation github.Installation) bool { return installation.ID == installationID })
	if i < 0 {
		writeJSONError(w, http.StatusForbidden, fmt.Errorf("your GitHub account can't access installation %d", installationID))
		return
	}
	err := l.Store.Link(ctx, scope.OrgID, *scope.UserID, reachable[i])
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
	l.writeRepos(w, r, scope.OrgID)
}

func (l *GitHubLink) sync(w http.ResponseWriter, r *http.Request) {
	org := tenant.FromContext(r.Context()).OrgID
	if err := l.Store.SyncAll(r.Context(), org); err != nil {
		writeJSONError(w, http.StatusBadGateway, err)
		return
	}
	l.writeRepos(w, r, org)
}

// writeRepos answers with the organization's repositories.
func (l *GitHubLink) writeRepos(w http.ResponseWriter, r *http.Request, org uuid.UUID) {
	repos, err := l.Store.Repos(r.Context(), org)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, map[string][]string{"repositories": repos})
}
