// Hosted mode's route allowlist: an organization's dashboard serves only routes known
// to be org-scoped, answers run endpoints without running anything, and 404s the rest.
package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

// hostedPassThrough lists the dashboard routes a hosted organization is served, as
// the patterns Handler registers them under, optionally narrowed to a method
// ("GET /api/x" for an any-method "/api/x"). Any other route is 404 in hosted mode,
// so a route added later stays off until it is known to be org-scoped. Projects are
// the organization's repositories and todos live in them (see hosted_workspaces.go);
// runtime presets still resolve from the host's directories, so they stay off.
var hostedPassThrough = []string{
	"/_assets/", "/favicon.svg", "/manifest.webmanifest", "/brand/gavel-logo.svg", "/brand/menubar.png",
	"/brand/menubar-unread.png", "/brand/apple-touch-icon.png", "/brand/icon-192.png", "/brand/icon-512.png",
	"GET /api/events", "POST /api/events/{conn}/subs", "DELETE /api/events/{conn}/subs/{id}",
	"/api/status", "GET /api/auth/config", "GET /api/me",
	"GET /api/ssh-keys", "POST /api/ssh-keys", "DELETE /api/ssh-keys/{id}",

	"/api/prs", "/api/prs/stream", "/api/prs/refresh", "/api/prs/bots", "/api/prs/closed",
	"/api/prs/pause", "/api/prs/detail", "GET /api/prs/commits/diff", "GET /api/prs/files/diff",
	"POST /api/prs/merge", "POST /api/prs/approve", "POST /api/prs/update-branch",
	"/api/prs/job-logs", "/api/prs/seen",
	"/api/config", "/api/repos", "/api/repos/favicon", "/api/orgs",

	"/api/settings/schema",
	// Answered by answerHostedProjectAction before the host's projects are read.
	"POST /api/projects/{name}/actions",

	"GET /api/projects", "/api/todos", "/api/todos/item", "/api/todos/labels", "/api/todos/links",
	"/api/todos/pr-links", "POST /api/todos/criteria", "/api/todos/attachments", "GET /api/todos/attachments/{id}",
}

// hostedAnswer is a fixed response a hosted dashboard gives instead of the handler.
type hostedAnswer struct {
	status  int
	message string
}

var (
	agentRunsUnavailable = hostedAnswer{http.StatusNotImplemented, "agent runs are not available in hosted mode yet"}
	projectsAreRepos     = hostedAnswer{http.StatusBadRequest, "projects are repos in hosted mode"}
	commitsFromAgentRuns = hostedAnswer{http.StatusBadRequest, "commits come from agent runs in hosted mode"}
	actionsUnavailable   = hostedAnswer{http.StatusNotImplemented, "project actions are not available in hosted mode yet"}
)

// hostedAnswers lists the routes a hosted dashboard answers itself, keyed like
// hostedPassThrough. Agent runs wait for sandboxed execution (M10); projects come
// from the org's repositories.
var hostedAnswers = map[string]hostedAnswer{
	"/api/todos/run":                         agentRunsUnavailable,
	"POST /api/todos/plan/approve":           agentRunsUnavailable,
	"POST /api/todos/plan/revise":            agentRunsUnavailable,
	"POST /api/todos/answer":                 agentRunsUnavailable,
	"POST /api/v1/todo/{id}/run":             agentRunsUnavailable,
	"POST /api/v1/todo/{id}/plan":            agentRunsUnavailable,
	"POST /api/v1/todo/{id}/triage":          agentRunsUnavailable,
	"POST /api/projects":                     projectsAreRepos,
	"PUT /api/projects/{name}":               projectsAreRepos,
	"DELETE /api/projects/{name}":            projectsAreRepos,
	"POST /api/projects/{name}/commit-queue": commitsFromAgentRuns,
	"POST /api/projects/{name}/branch/pr":    commitsFromAgentRuns,
}

// answerHostedProjectAction answers a project action in hosted mode: commits only
// come from agent runs, and test and lint wait for sandboxed execution (M9.5).
func answerHostedProjectAction(w http.ResponseWriter, r *http.Request) {
	var request projectActionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	answer := actionsUnavailable
	if request.Action == projectActionCommit {
		answer = commitsFromAgentRuns
	}
	writeJSONError(w, answer.status, errors.New(answer.message))
}

// SetHostedMode restricts the dashboard to the hosted route allowlist. Call it
// before Handler.
func (s *Server) SetHostedMode(hosted bool) {
	s.hosted = hosted
}

// hostedRoutes gates mux by the pattern mux itself resolves each request to, so the
// allowlist can never disagree with the handler that would run. Unmatched page loads
// reach mux's "/" route, which serves only the dashboard's own pages.
func (s *Server) hostedRoutes(mux *http.ServeMux) http.Handler {
	allowed := map[string]bool{}
	for _, pattern := range hostedPassThrough {
		allowed[pattern] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := mux.Handler(r)
		narrowed := r.Method + " " + pattern
		if answer, ok := hostedAnswers[pattern]; ok {
			writeJSONError(w, answer.status, errors.New(answer.message))
			return
		}
		if answer, ok := hostedAnswers[narrowed]; ok {
			writeJSONError(w, answer.status, errors.New(answer.message))
			return
		}
		page := pattern == "/" && (r.Method == http.MethodGet || r.Method == http.MethodHead) && !strings.HasPrefix(r.URL.Path, "/api/")
		if allowed[pattern] || allowed[narrowed] || page {
			mux.ServeHTTP(w, r)
			return
		}
		writeJSONError(w, http.StatusNotFound, fmt.Errorf("%s %s is not available in hosted mode", r.Method, r.URL.Path))
	})
}

// mustRegisterHostedRoutes panics unless every allowlist entry names a pattern mux
// registers. A mistyped entry would otherwise turn its route off without a trace.
func mustRegisterHostedRoutes(mux *http.ServeMux) {
	var missing []string
	entries := append([]string{}, hostedPassThrough...)
	for pattern := range hostedAnswers {
		entries = append(entries, pattern)
	}
	placeholder := regexp.MustCompile(`\{[^}]+\}`)
	for _, entry := range entries {
		method, path, narrowed := strings.Cut(entry, " ")
		if !narrowed {
			method, path = http.MethodGet, entry
		}
		sample := placeholder.ReplaceAllString(path, "x")
		req, err := http.NewRequest(method, sample, nil)
		if err != nil {
			panic(fmt.Sprintf("ui: hosted route %q: %v", entry, err))
		}
		_, got := mux.Handler(req)
		if got != entry && got != path {
			missing = append(missing, fmt.Sprintf("%q (resolves to %q)", entry, got))
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		panic("ui: hosted routes that the dashboard doesn't register: " + strings.Join(missing, ", "))
	}
}
