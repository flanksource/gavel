package github

import (
	"context"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"os"
	"strings"
	"time"
)

// AuthState classifies the outcome of ProbeToken — callers (notably pr/ui's
// /api/status handler) use it to decide dot color + user-facing message.
type AuthState string

const (
	// AuthStateOK: token resolved and GitHub accepted it.
	AuthStateOK AuthState = "ok"
	// AuthStateNoToken: Options.token() found nothing — user has not set
	// GITHUB_TOKEN / GH_TOKEN and no auth.json. The install command can
	// fix this.
	AuthStateNoToken AuthState = "no-token"
	// AuthStateInvalid: token was presented but GitHub returned 401. The
	// most common cause is an expired PAT or a revoked fine-grained token.
	AuthStateInvalid AuthState = "invalid"
	// AuthStateRateLimited: token works but we're out of requests. Degraded
	// rather than down — the next reset window will fix it.
	AuthStateRateLimited AuthState = "rate-limited"
	// AuthStateUnreachable: HTTP call failed before GitHub could classify —
	// DNS / TCP / TLS errors. Degraded: probably transient.
	AuthStateUnreachable AuthState = "unreachable"
)

// AuthProbeResult captures the outcome of a ProbeToken call. Message is
// human-readable and safe to show in the UI tooltip / CLI status.
type AuthProbeResult struct {
	State     AuthState
	Message   string
	RateLimit *RateLimit
	// Login is the GitHub username the token authenticates as, when known.
	// Surfaced in the UI so operators can spot "wrong account" mistakes.
	Login string
}

// ProbeToken resolves a token via Options.token() then calls GitHub's
// /rate_limit endpoint — a cheap authenticated request that doesn't burn
// a real API credit and returns the user's current limits. The endpoint
// URL can be overridden via the GITHUB_API_URL env var (internal use only,
// tests).
//
// The call is capped to a 5s timeout so a network stall can't block the
// background probe loop.
func ProbeToken(opts Options) AuthProbeResult {
	return probeToken(opts, githubAPIBase())
}

func probeToken(opts Options, baseURL string) AuthProbeResult {
	token, err := opts.token()
	if err != nil {
		if strings.Contains(err.Error(), ErrNoTokenMarker) {
			return AuthProbeResult{State: AuthStateNoToken, Message: "no GitHub token — run `gavel system install` to persist one"}
		}
		return AuthProbeResult{State: AuthStateUnreachable, Message: err.Error()}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := nethttp.NewRequestWithContext(ctx, "GET", baseURL+"/rate_limit", nil)
	if err != nil {
		return AuthProbeResult{State: AuthStateUnreachable, Message: fmt.Sprintf("build request: %v", err)}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := nethttp.DefaultClient.Do(req)
	if err != nil {
		return AuthProbeResult{State: AuthStateUnreachable, Message: fmt.Sprintf("GitHub unreachable: %v", err)}
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case nethttp.StatusOK:
		// Pull the core-resource rate limit out of the structured JSON
		// rather than the headers — GitHub returns `resources.core` in the
		// body which is what matters for REST calls.
		var body struct {
			Resources struct {
				Core RateLimit `json:"core"`
			} `json:"resources"`
		}
		rl := &body.Resources.Core
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			// 200 with undecodable body is odd but still a valid auth OK —
			// don't punish the user for a parse glitch.
			rl = nil
		} else {
			rl.Resource = "core"
		}
		login := userLoginFromRequest(ctx, baseURL, token)
		if rl != nil && rl.Remaining <= 0 {
			return AuthProbeResult{State: AuthStateRateLimited, Message: "rate limit exhausted — resets at " + rateLimitResetStr(rl), RateLimit: rl, Login: login}
		}
		return AuthProbeResult{State: AuthStateOK, Message: "authenticated as " + loginOrToken(login), RateLimit: rl, Login: login}
	case nethttp.StatusUnauthorized:
		return AuthProbeResult{State: AuthStateInvalid, Message: "GitHub rejected the token (401) — it's probably expired or revoked"}
	case nethttp.StatusForbidden:
		// 403 from /rate_limit with X-RateLimit-Remaining:0 means we hit
		// the secondary rate limit; otherwise it's a permissions/SSO issue.
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			rl := ParseRateLimit(resp.Header)
			return AuthProbeResult{State: AuthStateRateLimited, Message: "rate limit exhausted", RateLimit: rl}
		}
		return AuthProbeResult{State: AuthStateInvalid, Message: "GitHub rejected the token (403) — missing scopes or SSO not authorized"}
	default:
		return AuthProbeResult{State: AuthStateUnreachable, Message: fmt.Sprintf("unexpected GitHub status %d", resp.StatusCode)}
	}
}

// userLoginFromRequest is best-effort — we fall back silently if /user
// errors so a working auth probe isn't downgraded by a transient /user
// glitch. Uses a fresh 3s context so we don't spend the remaining deadline
// from the /rate_limit call.
func userLoginFromRequest(_ context.Context, baseURL, token string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := nethttp.NewRequestWithContext(ctx, "GET", baseURL+"/user", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := nethttp.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != nethttp.StatusOK {
		return ""
	}
	var body struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return ""
	}
	return body.Login
}

func loginOrToken(login string) string {
	if login == "" {
		return "GitHub"
	}
	return login
}

func rateLimitResetStr(rl *RateLimit) string {
	if rl == nil || rl.Reset == 0 {
		return "unknown time"
	}
	return time.Unix(rl.Reset, 0).Format(time.RFC3339)
}

// githubAPIBase returns the API root. Overridable via GITHUB_API_URL for
// GitHub Enterprise or tests (probeToken takes an explicit baseURL).
func githubAPIBase() string {
	if v := os.Getenv("GITHUB_API_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "https://api.github.com"
}
