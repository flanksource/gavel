// Serves the platform dashboard's API: every organization's data, to platform admins only.
// Each source answers on its own route, so one source's outage leaves the others readable.
package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/missioncontrol"
	"github.com/flanksource/gavel/platform"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
)

// PlatformAPI serves /api/platform/ to platform admins.
type PlatformAPI struct {
	Store platform.Store
	// Clerk lists the instance's organizations, including those that never used gavel.
	Clerk api.ClerkOrganizations
	// App lists the GitHub App's installations; nil when serve runs without one.
	App *github.App
	// MissionControl reads gavel's infrastructure; nil when serve runs without one.
	MissionControl *missioncontrol.Source
	// SubstrateView is the namespace/name of the Mission Control View graphing substrate's
	// metrics; empty when serve runs without one.
	SubstrateView string
}

// platformSourceTimeout bounds how long one route waits for its source.
const platformSourceTimeout = 10 * time.Second

// maxWindowDays bounds the dashboard's windows to the default retention of sandbox runs.
const maxWindowDays = 60

// Handler serves the routes; it must sit behind the Authenticator, which sets the scope.
func (p *PlatformAPI) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/platform/tenants", p.tenants)
	mux.HandleFunc("GET /api/platform/sandbox-runs", p.sandboxRuns)
	mux.HandleFunc("GET /api/platform/model-usage", p.modelUsage)
	mux.HandleFunc("GET /api/platform/budgets", p.budgets)
	mux.HandleFunc("GET /api/platform/github", p.gitHubApp)
	mux.HandleFunc("GET /api/platform/infrastructure", p.infrastructure)
	mux.HandleFunc("GET /api/platform/substrate/metrics", p.substrateMetrics)
	return platformAdminsOnly(mux)
}

// platformAdminsOnly answers 403, saying how to become one, to a caller who isn't a platform admin.
func platformAdminsOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !tenant.FromContext(r.Context()).PlatformAdmin {
			writeAuthFailure(w, authFailure{status: http.StatusForbidden, code: "not_platform_admin",
				message: `the platform dashboard is for platform admins: set "gavel_role": "platform_admin" in your Clerk user's public metadata; gavel picks it up when it next reads your Clerk profile`})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// windowStart is the start of the ?days= window that ends now, fallback days long without one.
func windowStart(r *http.Request, fallback int) (time.Time, error) {
	days := fallback
	if raw := r.URL.Query().Get("days"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > maxWindowDays {
			return time.Time{}, fmt.Errorf("days must be a whole number from 1 to %d, got %q", maxWindowDays, raw)
		}
		days = n
	}
	return time.Now().AddDate(0, 0, -days), nil
}

// PlatformTenants is every organization, from gavel and from Clerk. ClerkError says why
// Clerk's organizations are missing, when they are.
type PlatformTenants struct {
	Tenants    []PlatformTenant `json:"tenants"`
	ClerkError string           `json:"clerkError,omitempty"`
}

// PlatformTenant is an organization. InGavel is false for one that exists only in Clerk;
// ClerkMembers is nil for one Clerk doesn't list.
type PlatformTenant struct {
	platform.Tenant
	InGavel      bool `json:"inGavel"`
	ClerkMembers *int `json:"clerkMembers,omitempty"`
}

func (p *PlatformAPI) tenants(w http.ResponseWriter, r *http.Request) {
	since, err := windowStart(r, 30)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), platformSourceTimeout)
	defer cancel()
	tenants, err := p.Store.Tenants(ctx, since)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	result := PlatformTenants{Tenants: make([]PlatformTenant, 0, len(tenants))}
	byClerkID := map[string]int{}
	for i, t := range tenants {
		result.Tenants = append(result.Tenants, PlatformTenant{Tenant: t, InGavel: true})
		byClerkID[t.ClerkOrgID] = i
	}
	orgs, err := p.Clerk.Organizations(ctx)
	if err != nil {
		result.ClerkError = fmt.Sprintf("list the organizations in Clerk: %v", err)
		writeJSON(w, result)
		return
	}
	for _, org := range orgs {
		members := org.Members
		if i, ok := byClerkID[org.ID]; ok {
			result.Tenants[i].ClerkMembers = &members
			continue
		}
		result.Tenants = append(result.Tenants, PlatformTenant{
			Tenant:       platform.Tenant{ClerkOrgID: org.ID, Slug: org.Slug, Name: org.Name, CreatedAt: org.CreatedAt},
			ClerkMembers: &members,
		})
	}
	slices.SortStableFunc(result.Tenants, func(a, b PlatformTenant) int { return strings.Compare(a.Name, b.Name) })
	writeJSON(w, result)
}

func (p *PlatformAPI) sandboxRuns(w http.ResponseWriter, r *http.Request) {
	since, err := windowStart(r, 30)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), platformSourceTimeout)
	defer cancel()
	usage, err := p.Store.SandboxUsage(ctx, since)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, usage)
}

// PlatformModelUsage is the platform's model calls per day, organization, model and provider.
type PlatformModelUsage struct {
	Days []platform.ModelUsageDay `json:"days"`
}

func (p *PlatformAPI) modelUsage(w http.ResponseWriter, r *http.Request) {
	since, err := windowStart(r, 30)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), platformSourceTimeout)
	defer cancel()
	days, err := p.Store.ModelUsage(ctx, since)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, PlatformModelUsage{Days: days})
}

// PlatformBudgets is every live budget rule with its groups' spend in the current window.
type PlatformBudgets struct {
	Budgets []platform.Budget `json:"budgets"`
}

func (p *PlatformAPI) budgets(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), platformSourceTimeout)
	defer cancel()
	budgets, err := p.Store.Budgets(ctx, time.Now())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, PlatformBudgets{Budgets: budgets})
}

// PlatformGitHub is the GitHub App's installations, from GitHub and from gavel. GitHubError
// says why GitHub's list is missing, when it is.
type PlatformGitHub struct {
	App           PlatformGitHubApp      `json:"app"`
	Installations []PlatformInstallation `json:"installations"`
	GitHubError   string                 `json:"githubError,omitempty"`
}

// PlatformGitHubApp names the GitHub App serve runs as.
type PlatformGitHubApp struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
}

// PlatformInstallation is one installation. Org is nil for one no organization linked;
// Uninstalled marks one gavel has linked that GitHub no longer lists.
type PlatformInstallation struct {
	ID             int64                    `json:"id"`
	Account        string                   `json:"account"`
	AccountType    string                   `json:"accountType,omitempty"`
	CreatedAt      *time.Time               `json:"createdAt,omitempty"`
	SuspendedAt    *time.Time               `json:"suspendedAt,omitempty"`
	Org            *PlatformInstallationOrg `json:"org,omitempty"`
	Repos          int                      `json:"repos"`
	Uninstalled    bool                     `json:"uninstalled,omitempty"`
	RateLimit      *github.RateLimit        `json:"rateLimit,omitempty"`
	RateLimitError string                   `json:"rateLimitError,omitempty"`
}

// PlatformInstallationOrg is the organization an installation is linked to.
type PlatformInstallationOrg struct {
	ID   uuid.UUID `json:"id"`
	Slug string    `json:"slug"`
	Name string    `json:"name"`
}

// linkedRow is a linked installation as gavel knows it, without GitHub's details.
func linkedRow(linked platform.LinkedInstallation) PlatformInstallation {
	return PlatformInstallation{
		ID: linked.ID, Account: linked.AccountLogin, Repos: linked.Repos,
		Org: &PlatformInstallationOrg{ID: linked.OrgID, Slug: linked.OrgSlug, Name: linked.OrgName},
	}
}

// gitHubApp answers GitHub's installations of the App merged with those gavel's
// organizations linked, with each active installation's rate limit.
func (p *PlatformAPI) gitHubApp(w http.ResponseWriter, r *http.Request) {
	if p.App == nil {
		writeJSONError(w, http.StatusNotFound, errors.New("the GitHub App isn't configured: start gavel serve with --github-app-id, --github-app-slug and --github-app-private-key"))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), platformSourceTimeout)
	defer cancel()
	linked, err := p.Store.LinkedInstallations(ctx)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	result := PlatformGitHub{App: PlatformGitHubApp{ID: p.App.ID, Slug: p.App.Slug}, Installations: []PlatformInstallation{}}
	installs, err := p.App.Installations(ctx)
	if err != nil {
		result.GitHubError = err.Error()
		for _, l := range linked {
			result.Installations = append(result.Installations, linkedRow(l))
		}
		writeJSON(w, result)
		return
	}
	unlisted := map[int64]platform.LinkedInstallation{}
	for _, l := range linked {
		unlisted[l.ID] = l
	}
	for _, install := range installs {
		row := PlatformInstallation{ID: install.ID, Account: install.Account.Login, AccountType: install.Account.Type, CreatedAt: &install.CreatedAt, SuspendedAt: install.SuspendedAt}
		if l, ok := unlisted[install.ID]; ok {
			row.Org, row.Repos = linkedRow(l).Org, l.Repos
			delete(unlisted, install.ID)
		}
		if install.SuspendedAt == nil {
			if limit, err := p.App.InstallationRateLimit(ctx, install.ID); err != nil {
				row.RateLimitError = err.Error()
			} else {
				row.RateLimit = &limit
			}
		}
		result.Installations = append(result.Installations, row)
	}
	for _, l := range linked {
		if _, ok := unlisted[l.ID]; ok {
			row := linkedRow(l)
			row.Uninstalled = true
			result.Installations = append(result.Installations, row)
		}
	}
	writeJSON(w, result)
}

// errNoMissionControl answers the Mission Control routes when serve runs without one.
var errNoMissionControl = errors.New("no Mission Control is configured: start gavel serve with --mission-control-url, --mission-control-infra and GAVEL_MISSION_CONTROL_TOKEN")

// infrastructure answers what Mission Control knows about gavel's infrastructure.
func (p *PlatformAPI) infrastructure(w http.ResponseWriter, r *http.Request) {
	if p.MissionControl == nil {
		writeJSONError(w, http.StatusNotFound, errNoMissionControl)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), platformSourceTimeout)
	defer cancel()
	infrastructure, err := p.MissionControl.Infrastructure(ctx)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, infrastructure)
}

// errNoSubstrateView answers the substrate metrics route when serve runs without its View.
var errNoSubstrateView = errors.New("no View graphs substrate's metrics: start gavel serve with --mission-control-substrate-view namespace/name")

// substrateMetrics answers the panels of the Mission Control View graphing substrate's metrics.
func (p *PlatformAPI) substrateMetrics(w http.ResponseWriter, r *http.Request) {
	if p.MissionControl == nil {
		writeJSONError(w, http.StatusNotFound, errNoMissionControl)
		return
	}
	if p.SubstrateView == "" {
		writeJSONError(w, http.StatusNotFound, errNoSubstrateView)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), platformSourceTimeout)
	defer cancel()
	view, err := p.MissionControl.View(ctx, p.SubstrateView)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, view)
}
