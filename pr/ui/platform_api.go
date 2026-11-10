// Serves the platform dashboard's API: every organization's data, to platform admins only.
// Each source answers on its own route, so one source's outage leaves the others readable.
package ui

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/platform"
	"github.com/flanksource/gavel/tenant"
)

// PlatformAPI serves /api/platform/ to platform admins.
type PlatformAPI struct {
	Store platform.Store
	// Clerk lists the instance's organizations, including those that never used gavel.
	Clerk api.ClerkOrganizations
}

// platformSourceTimeout bounds how long one route waits for its source.
const platformSourceTimeout = 10 * time.Second

// maxWindowDays bounds the dashboard's windows to the default retention of sandbox runs.
const maxWindowDays = 60

// Handler serves the routes; it must sit behind the Authenticator, which sets the scope.
func (p *PlatformAPI) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/platform/tenants", p.tenants)
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
