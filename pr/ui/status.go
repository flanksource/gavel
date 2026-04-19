package ui

import (
	"encoding/json"
	nethttp "net/http"
	"time"

	"github.com/flanksource/gavel/github"
	ghcache "github.com/flanksource/gavel/github/cache"
)

// Severity is the per-component health classification surfaced at /api/status
// and rendered as a colored dot in the UI header.
//   - ok       → green  (fully working)
//   - degraded → yellow (partial: rate-limited, stale rows, auth missing but
//     cache works)
//   - down     → red    (cache disconnected, no token, recent fetch error)
type Severity string

const (
	SeverityOK       Severity = "ok"
	SeverityDegraded Severity = "degraded"
	SeverityDown     Severity = "down"
)

// ComponentStatus is one line in the aggregated status response. The UI
// renders Message in a tooltip; Detail is the structured sub-status (e.g.
// cache row counts) that `gavel system status` displays inline.
type ComponentStatus struct {
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Detail   any      `json:"detail,omitempty"`
}

// StatusResponse aggregates db/github/cache health. Exposed at /api/status
// so both the CLI (`gavel system status`) and the PR UI's header indicator
// can drive off a single source of truth instead of probing components
// individually.
type StatusResponse struct {
	// Overall is the worst of the component severities — if any component
	// is `down` the overall is `down`; if any is `degraded` the overall is
	// `degraded`; otherwise `ok`.
	Overall   Severity        `json:"overall"`
	Database  ComponentStatus `json:"database"`
	GitHub    ComponentStatus `json:"github"`
	CheckedAt time.Time       `json:"checkedAt"`
}

func (s *Server) handleStatus(w nethttp.ResponseWriter, _ *nethttp.Request) {
	resp := s.buildStatus()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// buildStatus snapshots current state. Cheap: no new HTTP calls — it reads
// the already-cached rateLimit + fetch error on the Server and the cache
// Store's row counts. Safe to call on every /api/status hit.
func (s *Server) buildStatus() StatusResponse {
	return StatusResponse{
		Overall:   SeverityOK, // overwritten below
		Database:  statusFromCache(ghcache.Shared().Status()),
		GitHub:    s.statusFromGitHub(),
		CheckedAt: time.Now(),
	}.withOverall()
}

// withOverall computes the Overall severity from the component severities.
func (r StatusResponse) withOverall() StatusResponse {
	worst := SeverityOK
	for _, c := range []ComponentStatus{r.Database, r.GitHub} {
		if c.Severity == SeverityDown {
			worst = SeverityDown
			break
		}
		if c.Severity == SeverityDegraded && worst != SeverityDown {
			worst = SeverityDegraded
		}
	}
	r.Overall = worst
	return r
}

// statusFromCache translates a cache.Status into a ComponentStatus. Down
// when the store is disabled or the Status struct surfaced an error;
// degraded when enabled but empty (no rows yet); ok otherwise.
func statusFromCache(st ghcache.Status) ComponentStatus {
	switch {
	case !st.Enabled:
		msg := "cache disabled"
		if st.Error != "" {
			msg = st.Error
		}
		return ComponentStatus{Severity: SeverityDown, Message: msg, Detail: st}
	case st.Error != "":
		return ComponentStatus{Severity: SeverityDegraded, Message: st.Error, Detail: st}
	default:
		return ComponentStatus{Severity: SeverityOK, Message: "connected", Detail: st}
	}
}

// statusFromGitHub turns the cached auth-probe result into a component
// status. The probe is the authoritative signal for "is the token valid":
// it calls GET /rate_limit explicitly, so a 401 shows up as AuthStateInvalid
// regardless of whatever else the poller may have logged.
//
// We intentionally do NOT use the Server's `err` field here — that captures
// search-side failures like "cannot determine org" which are configuration
// issues, not auth problems. Surfacing them under the github health banner
// confused operators into thinking their token was bad.
func (s *Server) statusFromGitHub() ComponentStatus {
	s.mu.RLock()
	auth := s.auth
	checkedAt := s.authCheckedAt
	s.mu.RUnlock()

	detail := map[string]any{}
	if auth.RateLimit != nil {
		detail["rateLimit"] = auth.RateLimit
	}
	if auth.Login != "" {
		detail["login"] = auth.Login
	}
	if !checkedAt.IsZero() {
		detail["checkedAt"] = checkedAt
	}
	detail["state"] = string(auth.State)

	switch auth.State {
	case github.AuthStateOK:
		return ComponentStatus{Severity: SeverityOK, Message: auth.Message, Detail: detail}
	case github.AuthStateNoToken, github.AuthStateInvalid:
		return ComponentStatus{Severity: SeverityDown, Message: auth.Message, Detail: detail}
	case github.AuthStateRateLimited, github.AuthStateUnreachable:
		return ComponentStatus{Severity: SeverityDegraded, Message: auth.Message, Detail: detail}
	default:
		// State is "" when the background probe hasn't completed its first
		// run yet. Degraded rather than ok — we genuinely don't know.
		return ComponentStatus{Severity: SeverityDegraded, Message: "checking token…", Detail: detail}
	}
}
