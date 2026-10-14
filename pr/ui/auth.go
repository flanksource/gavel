// Authenticates hosted dashboard requests: verifies the Clerk session or CLI access
// token, resolves the organization it acts in, and puts that tenant.Scope on the request.
package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/tenant"
	"github.com/golang-jwt/jwt/v5"
)

// sessionCookie is the cookie Clerk's browser SDK keeps the session token in.
const sessionCookie = "__session"

// Authenticator resolves the caller's identity and active organization for each request
// and rejects requests that have neither.
type Authenticator struct {
	Session api.SessionVerifier
	Access  *api.AccessTokenVerifier
	Members *api.Members
	// PublicURL is the dashboard's origin; cookie-authenticated writes must come from it.
	PublicURL string
}

// authFailure is a rejected request: its status, a stable code clients branch on, and why.
type authFailure struct {
	status  int
	code    string
	message string
}

// Middleware serves page loads, assets and the auth config with public, which needs no
// identity, and every other request with next, scoped to the caller's organization.
// It answers /api/auth/orgs itself: listing your organizations needs none picked.
func (a *Authenticator) Middleware(next, public http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if publicRequest(r) {
			public.ServeHTTP(w, r)
			return
		}
		caller, failure := a.verify(r)
		if failure == nil && r.Method == http.MethodGet && r.URL.Path == "/api/auth/orgs" {
			orgs, err := a.Members.Orgs(r.Context(), caller.clerkUserID)
			if err != nil {
				failure = classifyAuthError(r, err)
			} else {
				writeJSON(w, orgs)
				return
			}
		}
		var scope tenant.Scope
		if failure == nil {
			scope, failure = a.scope(r, caller)
		}
		if failure != nil {
			writeAuthFailure(w, *failure)
			return
		}
		next.ServeHTTP(w, r.WithContext(tenant.WithScope(r.Context(), scope)))
	})
}

// publicRequest reports a request that needs no identity: a page load or an asset (the
// SPA gates itself), or the auth config the SPA signs in with.
func publicRequest(r *http.Request) bool {
	if r.URL.Path == "/api/auth/config" {
		return true
	}
	return (r.Method == http.MethodGet || r.Method == http.MethodHead) && !strings.HasPrefix(r.URL.Path, "/api/")
}

// caller is a verified Clerk user, and their browser session's claims when they came
// with a session rather than a CLI access token.
type caller struct {
	clerkUserID string
	session     *api.SessionClaims
}

// verify checks the request's token: a bearer token, or else the session cookie.
func (a *Authenticator) verify(r *http.Request) (caller, *authFailure) {
	raw, fromCookie := "", false
	if token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		raw = strings.TrimSpace(token)
	} else if cookie, err := r.Cookie(sessionCookie); err == nil {
		raw, fromCookie = cookie.Value, true
	}
	if raw == "" {
		return caller{}, &authFailure{http.StatusUnauthorized, "unauthenticated",
			"sign in to the dashboard, or send a Clerk session or access token as Authorization: Bearer <token>"}
	}
	// A browser sends its cookie on every request, so a write carried by the cookie
	// must come from the dashboard's own pages.
	if fromCookie && unsafeMethod(r.Method) && !a.sameOrigin(r) {
		return caller{}, &authFailure{http.StatusForbidden, "cross_site_request",
			fmt.Sprintf("a %s signed in by cookie must come from %s, not %q", r.Method, a.PublicURL, r.Header.Get("Origin"))}
	}
	if tokenType(raw) == "at+jwt" {
		user, err := a.Access.Verify(r.Context(), raw)
		if err != nil {
			return caller{}, classifyAuthError(r, err)
		}
		return caller{clerkUserID: user}, nil
	}
	claims, err := a.Session.Verify(r.Context(), raw)
	if err != nil {
		return caller{}, classifyAuthError(r, err)
	}
	return caller{clerkUserID: claims.UserID, session: &claims}, nil
}

// scope resolves the organization the caller acts in: a session's active org, or the
// org a CLI request names with X-Gavel-Org.
func (a *Authenticator) scope(r *http.Request, c caller) (tenant.Scope, *authFailure) {
	var scope tenant.Scope
	var err error
	if c.session != nil {
		scope, err = a.Members.ForSession(r.Context(), *c.session)
	} else {
		scope, err = a.Members.ForAccessToken(r.Context(), c.clerkUserID, r.Header.Get("X-Gavel-Org"))
	}
	if err != nil {
		return tenant.Scope{}, classifyAuthError(r, err)
	}
	return scope, nil
}

// classifyAuthError answers a verification or membership error with its status and code.
func classifyAuthError(r *http.Request, err error) *authFailure {
	switch {
	case errors.Is(err, api.ErrInvalidToken):
		return &authFailure{http.StatusUnauthorized, "unauthenticated", err.Error()}
	case errors.Is(err, api.ErrNoActiveOrg):
		return &authFailure{http.StatusConflict, "no_active_org", err.Error()}
	case errors.Is(err, api.ErrNotMember):
		return &authFailure{http.StatusForbidden, "not_a_member", err.Error()}
	default:
		logger.Warnf("authenticate %s %s: %v", r.Method, r.URL.Path, err)
		return &authFailure{http.StatusServiceUnavailable, "auth_unavailable",
			"could not check the request's identity: " + err.Error()}
	}
}

// tokenType reads a JWT's header typ without verifying it, only to pick the verifier
// that then checks the token in full.
func tokenType(raw string) string {
	token, _, err := jwt.NewParser().ParseUnverified(raw, jwt.MapClaims{})
	if err != nil {
		return ""
	}
	typ, _ := token.Header["typ"].(string)
	return typ
}

func unsafeMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}

// sameOrigin reports a request from the dashboard's own pages: an Origin equal to
// PublicURL, or, without an Origin, the browser's Sec-Fetch-Site: same-origin.
func (a *Authenticator) sameOrigin(r *http.Request) bool {
	if origin := r.Header.Get("Origin"); origin != "" {
		return origin == strings.TrimSuffix(a.PublicURL, "/")
	}
	return r.Header.Get("Sec-Fetch-Site") == "same-origin"
}

func writeAuthFailure(w http.ResponseWriter, failure authFailure) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(failure.status)
	json.NewEncoder(w).Encode(map[string]string{"error": failure.code, "message": failure.message}) //nolint:errcheck
}
