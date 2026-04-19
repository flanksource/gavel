package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/prwatch"
)

type SearchConfig struct {
	Repos       []string `json:"repos"`
	All         bool     `json:"all,omitempty"`
	Org         string   `json:"org,omitempty"`
	Author      string   `json:"author,omitempty"`
	Any         bool     `json:"any,omitempty"`
	Bots        bool     `json:"bots,omitempty"`
	IgnoredOrgs []string `json:"ignoredOrgs,omitempty"`
}

type repoInfo struct {
	Repo     string `json:"repo"`
	PRCount  int    `json:"prCount"`
	Selected bool   `json:"selected"`
}

type Server struct {
	mu          sync.RWMutex
	prs         github.PRSearchResults
	fetchedAt   time.Time
	interval    time.Duration
	err         error
	paused      bool
	rateLimit   *github.RateLimit
	updated     chan struct{}
	refreshCh   chan struct{}
	subscribers []chan github.PRSearchResults
	ghOpts      github.Options
	config      SearchConfig

	detailCache  *DetailCache
	detailSyncer *DetailSyncer

	RepoSearchFn func() (github.PRSearchResults, error)
	repoCache    []repoInfo
	repoCacheAt  time.Time

	// auth is the cached result of the most recent GitHub token probe.
	// Refreshed in the background every authProbeInterval so /api/status
	// responds instantly. First probe runs asynchronously from NewServer.
	auth          github.AuthProbeResult
	authCheckedAt time.Time

	// orgs is a short-TTL cache of GET /user/orgs for the header's org
	// chooser. Users with 30+ orgs are the exception not the rule, so one
	// page is enough; a 5-minute TTL keeps the dropdown snappy without
	// masking fresh memberships for long.
	orgs         []github.Org
	orgsCachedAt time.Time
}

const orgsCacheTTL = 5 * time.Minute

// authProbeInterval is how often the background goroutine re-probes the
// token. 5 minutes balances "catch token expiry quickly" vs "don't flood
// the user's rate-limit budget with self-checks".
const authProbeInterval = 5 * time.Minute

type snapshot struct {
	PRs         github.PRSearchResults `json:"prs"`
	FetchedAt   time.Time              `json:"fetchedAt"`
	NextFetchIn int                    `json:"nextFetchIn"`
	Incremental bool                   `json:"incremental"`
	Paused      bool                   `json:"paused"`
	Error       string                 `json:"error,omitempty"`
	Config      SearchConfig           `json:"config"`
	RateLimit   *github.RateLimit      `json:"rateLimit,omitempty"`
	// Unread maps prKey("repo#number") → true for PRs whose UpdatedAt is
	// newer than the recorded SeenAt (or that have never been seen). PRs
	// marked as read are omitted to keep the map sparse on the wire.
	Unread     map[string]bool         `json:"unread,omitempty"`
	SyncStatus map[string]PRSyncStatus `json:"syncStatus,omitempty"`
}

func NewServer(interval time.Duration, ghOpts github.Options, config SearchConfig) *Server {
	s := &Server{
		interval:    interval,
		ghOpts:      ghOpts,
		config:      config,
		updated:     make(chan struct{}, 1),
		refreshCh:   make(chan struct{}, 1),
		detailCache: NewDetailCache(),
	}
	// Probe runs in the background so NewServer stays fast. First /api/status
	// hit before the probe completes returns State="" which handleStatus
	// treats as "probing" (degraded, "checking token...").
	go s.refreshAuthProbe()
	go s.authProbeLoop()
	return s
}

// authProbeLoop refreshes the cached GitHub auth state every
// authProbeInterval. Running this in a goroutine means /api/status
// responses are always served from cache — no user request blocks on a
// GitHub round-trip.
func (s *Server) authProbeLoop() {
	t := time.NewTicker(authProbeInterval)
	defer t.Stop()
	for range t.C {
		s.refreshAuthProbe()
	}
}

// refreshAuthProbe performs the probe and stores the result. Called both
// from the loop above and whenever we want a fresh reading (e.g. after the
// daemon records a successful fetch, which implies the token still works).
func (s *Server) refreshAuthProbe() {
	res := github.ProbeToken(s.ghOpts)
	s.mu.Lock()
	s.auth = res
	s.authCheckedAt = time.Now()
	s.mu.Unlock()
}

func (s *Server) DetailCache() *DetailCache {
	return s.detailCache
}

func (s *Server) SetDetailSyncer(ds *DetailSyncer) {
	s.detailSyncer = ds
}

func (s *Server) notifyDetailSyncer() {
	if s.detailSyncer != nil {
		s.detailSyncer.Notify()
	}
}

func (s *Server) GetConfig() SearchConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *Server) SetConfig(cfg SearchConfig) {
	s.mu.Lock()
	s.config = cfg
	s.mu.Unlock()
}

func (s *Server) SetResults(prs github.PRSearchResults, incremental bool) {
	s.mu.Lock()
	s.prs = prs
	s.fetchedAt = time.Now()
	s.err = nil
	subs := s.subscribers
	s.mu.Unlock()
	s.notify()
	for _, ch := range subs {
		select {
		case ch <- prs:
		default:
		}
	}
}

func (s *Server) Subscribe() chan github.PRSearchResults {
	ch := make(chan github.PRSearchResults, 1)
	s.mu.Lock()
	s.subscribers = append(s.subscribers, ch)
	s.mu.Unlock()
	return ch
}

func (s *Server) SetError(err error) {
	s.mu.Lock()
	s.err = err
	s.mu.Unlock()
	s.notify()
}

func (s *Server) SetRateLimit(rl *github.RateLimit) {
	if rl == nil {
		return
	}
	s.mu.Lock()
	s.rateLimit = rl
	s.mu.Unlock()
}

func (s *Server) IsPaused() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.paused
}

func (s *Server) TogglePause() {
	s.mu.Lock()
	s.paused = !s.paused
	s.mu.Unlock()
	s.notify()
}

func (s *Server) notify() {
	select {
	case s.updated <- struct{}{}:
	default:
	}
}

func (s *Server) RefreshCh() chan struct{} {
	return s.refreshCh
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoute)
	mux.HandleFunc("/api/prs", s.handleJSON)
	mux.HandleFunc("/api/prs/stream", s.handleSSE)
	mux.HandleFunc("/api/prs/refresh", s.handleRefresh)
	mux.HandleFunc("/api/prs/pause", s.handlePause)
	mux.HandleFunc("/api/prs/detail", s.handleDetail)
	mux.HandleFunc("/api/prs/job-logs", s.handleJobLogs)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/repos", s.handleRepos)
	mux.HandleFunc("/api/orgs", s.handleOrgs)
	mux.HandleFunc("/api/repos/favicon", s.handleRepoFavicon)
	mux.HandleFunc("/api/activity", s.handleActivity)
	mux.HandleFunc("/api/activity/stream", s.handleActivityStream)
	mux.HandleFunc("/api/activity/reset", s.handleActivityReset)
	mux.HandleFunc("/api/activity/cache", s.handleActivityCache)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/favicon.svg", handleFavicon)
	mux.HandleFunc("/brand/gavel-logo.svg", handleLogo)
	mux.HandleFunc("/brand/menubar.png", handleMenubarIcon)
	mux.HandleFunc("/brand/menubar-unread.png", handleMenubarUnreadIcon)
	mux.HandleFunc("/api/prs/seen", s.handleSeen)
	mux.HandleFunc("/results/", s.handleGavelResults)
	return mux
}

func handleFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	fmt.Fprint(w, faviconSVG)
}

func handleLogo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	fmt.Fprint(w, logoSVG)
}

func handleMenubarIcon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	if _, err := w.Write(menubarPNG); err != nil {
		logger.Debugf("write menubar icon: %v", err)
	}
}

func handleMenubarUnreadIcon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	if _, err := w.Write(menubarUnreadPNG); err != nil {
		logger.Debugf("write menubar unread icon: %v", err)
	}
}

func (s *Server) handleRoute(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" && r.URL.RawQuery == "" {
		http.Redirect(w, r, "/prs", http.StatusFound)
		return
	}
	req, ok := parseRouteRequest(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if req.IsExport {
		s.handleExport(w, r, req)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, pageHTML())
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request, req routeRequest) {
	report, err := s.buildExportReport(req)
	if err != nil {
		if err == errRouteNodeNotFound {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeExportResponse(w, r, report, req.Format)
}

func pageHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>gavel · PR Dashboard</title>
    <link rel="icon" type="image/svg+xml" href="/favicon.svg">
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://code.iconify.design/iconify-icon/2.0.0/iconify-icon.min.js"></script>
    <style>
        @keyframes gavel-progress-slide {
            0%   { left: -35%; }
            100% { left: 100%; }
        }
        .gavel-progress-bar {
            animation: gavel-progress-slide 1.1s ease-in-out infinite;
        }
    </style>
</head>
<body>
    <div id="root"></div>
    <script>` + bundleJS + `</script>
</body>
</html>`
}

// snapshotLocked builds a snapshot using the already-held RLock. It does
// NOT compute the unread map — that requires a database round-trip and is
// populated by withUnread() outside the lock.
func (s *Server) snapshotLocked() snapshot {
	snap := snapshot{
		PRs:         s.prs,
		FetchedAt:   s.fetchedAt,
		NextFetchIn: int(s.interval.Seconds()),
		Paused:      s.paused,
		Config:      s.config,
		RateLimit:   s.rateLimit,
	}
	if s.err != nil {
		snap.Error = s.err.Error()
	}
	return snap
}

// withUnread attaches the unread map to a snapshot. Runs the cache lookup
// with no server lock held so a slow postgres round-trip never blocks the
// poller or other handlers. Errors are logged and the snap is returned
// without an unread map, so a cache outage degrades to "everything unread"
// rather than failing the request.
func (s *Server) withUnread(snap snapshot) snapshot {
	if len(snap.PRs) == 0 {
		return snap
	}
	snap.Unread = s.UnreadMap(snap.PRs)
	return snap
}

// UnreadMap computes which PRs in the given slice are unread. A PR is
// unread iff its UpdatedAt is newer than the recorded SeenAt, or no SeenPR
// row exists. Only unread keys are present in the returned map (sparse).
// Safe to call with no server lock held.
func (s *Server) UnreadMap(prs github.PRSearchResults) map[string]bool {
	if len(prs) == 0 {
		return nil
	}
	store := cache.Shared()
	keys := make([]cache.SeenKey, len(prs))
	for i, pr := range prs {
		keys[i] = cache.SeenKey{Repo: pr.Repo, Number: pr.Number}
	}
	seenMap, err := store.LoadSeenMap(keys)
	if err != nil {
		logger.Warnf("load seen map: %v", err)
		seenMap = nil
	}
	out := make(map[string]bool, len(prs))
	for _, pr := range prs {
		seenAt, ok := seenMap[cache.SeenKey{Repo: pr.Repo, Number: pr.Number}]
		if !ok || pr.UpdatedAt.After(seenAt) {
			out[prKey(pr)] = true
		}
	}
	return out
}

// UnreadCount returns the number of PRs in the current server state that
// are unread. Used by the menubar title.
func (s *Server) UnreadCount() int {
	s.mu.RLock()
	prs := s.prs
	s.mu.RUnlock()
	return len(s.UnreadMap(prs))
}

// MarkSeen delegates to the shared cache store and triggers a snapshot
// push so the UI and menubar update without waiting for the next poll.
func (s *Server) MarkSeen(repo string, number int) error {
	if err := cache.Shared().MarkSeen(repo, number); err != nil {
		return err
	}
	s.notify()
	return nil
}

// MarkAllSeen marks every PR currently in the server state as read.
func (s *Server) MarkAllSeen() error {
	s.mu.RLock()
	prs := s.prs
	s.mu.RUnlock()
	if len(prs) == 0 {
		return nil
	}
	keys := make([]cache.SeenKey, len(prs))
	for i, pr := range prs {
		keys[i] = cache.SeenKey{Repo: pr.Repo, Number: pr.Number}
	}
	if err := cache.Shared().MarkManySeen(keys); err != nil {
		return err
	}
	s.notify()
	return nil
}

// withSyncStatus attaches per-PR sync statuses to a snapshot.
func (s *Server) withSyncStatus(snap snapshot) snapshot {
	if s.detailCache == nil {
		return snap
	}
	snap.SyncStatus = s.detailCache.AllStatuses()
	return snap
}

func (s *Server) handleJSON(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	data := s.snapshotLocked()
	s.mu.RUnlock()
	data = s.withUnread(data)
	data = s.withSyncStatus(data)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	s.mu.RLock()
	initial := s.snapshotLocked()
	s.mu.RUnlock()
	initial = s.withUnread(initial)
	initial = s.withSyncStatus(initial)
	if b, err := json.Marshal(initial); err == nil {
		fmt.Fprintf(w, "data: %s\n\n", b)
		flusher.Flush()
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-s.updated:
		case <-ticker.C:
		}

		s.mu.RLock()
		data := s.snapshotLocked()
		s.mu.RUnlock()
		data = s.withUnread(data)
		data = s.withSyncStatus(data)

		b, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", b)
		flusher.Flush()
	}
}

func (s *Server) handleSeen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Repo   string `json:"repo"`
		Number int    `json:"number"`
		All    bool   `json:"all"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.All {
		if err := s.MarkAllSeen(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if body.Repo == "" || body.Number == 0 {
			http.Error(w, "repo and number are required", http.StatusBadRequest)
			return
		}
		if err := s.MarkSeen(body.Repo, body.Number); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"ok"}`)
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	select {
	case s.refreshCh <- struct{}{}:
	default:
	}
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, `{"status":"refresh requested"}`)
}

func (s *Server) handlePause(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.TogglePause()
	s.mu.RLock()
	paused := s.paused
	s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"paused":%v}`, paused)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		cfg := s.config
		s.mu.RUnlock()
		json.NewEncoder(w).Encode(cfg) //nolint:errcheck
	case http.MethodPost:
		var cfg SearchConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
			return
		}
		s.SetConfig(cfg)
		go SaveSettings(UISettings{
			Repos:       cfg.Repos,
			Author:      cfg.Author,
			Any:         cfg.Any,
			Bots:        cfg.Bots,
			IgnoredOrgs: cfg.IgnoredOrgs,
		})
		select {
		case s.refreshCh <- struct{}{}:
		default:
		}
		json.NewEncoder(w).Encode(cfg) //nolint:errcheck
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRepos(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	cache := s.repoCache
	cacheAge := time.Since(s.repoCacheAt)
	s.mu.RUnlock()

	if len(cache) == 0 || cacheAge > 5*time.Minute {
		go s.refreshRepoCache()
		if len(cache) == 0 {
			cache = s.reposFromCurrentPRs()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cache) //nolint:errcheck
}

// handleOrgs returns the authenticated user's GitHub org memberships for
// the header's org chooser. Cached for orgsCacheTTL to keep the dropdown
// snappy — org membership changes rarely enough that a 5-minute staleness
// ceiling is fine.
//
// By default the response is filtered through SearchConfig.IgnoredOrgs so
// the dropdown only shows orgs the user cares about. Pass
// ?include-ignored=1 to get the full list — used by the chooser's "manage
// hidden orgs" expansion so the user can unhide.
//
// On fetch failure we return whatever we had cached even if expired; an
// empty slice rather than an error keeps the chooser functional (it can
// still offer "All" / "@me" modes).
func (s *Server) handleOrgs(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	cached := s.orgs
	fresh := time.Since(s.orgsCachedAt) < orgsCacheTTL
	ignored := s.config.IgnoredOrgs
	s.mu.RUnlock()

	if !fresh {
		orgs, err := github.FetchUserOrgs(s.ghOpts)
		if err != nil {
			logger.Warnf("fetch user orgs: %v", err)
		} else {
			s.mu.Lock()
			s.orgs = orgs
			s.orgsCachedAt = time.Now()
			s.mu.Unlock()
			cached = orgs
		}
	}

	if cached == nil {
		cached = []github.Org{}
	}
	if r.URL.Query().Get("include-ignored") != "1" {
		cached = filterIgnoredOrgs(cached, ignored)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cached) //nolint:errcheck
}

// filterIgnoredOrgs returns a copy of orgs with logins in ignored dropped.
// A nil / empty ignored slice is a fast path that just returns orgs.
func filterIgnoredOrgs(orgs []github.Org, ignored []string) []github.Org {
	if len(ignored) == 0 {
		return orgs
	}
	skip := make(map[string]bool, len(ignored))
	for _, o := range ignored {
		skip[o] = true
	}
	out := make([]github.Org, 0, len(orgs))
	for _, o := range orgs {
		if skip[o.Login] {
			continue
		}
		out = append(out, o)
	}
	return out
}

func (s *Server) reposFromCurrentPRs() []repoInfo {
	s.mu.RLock()
	prs := s.prs
	selectedRepos := s.config.Repos
	s.mu.RUnlock()

	selected := make(map[string]bool, len(selectedRepos))
	for _, r := range selectedRepos {
		selected[r] = true
	}

	counts := make(map[string]int)
	for _, pr := range prs {
		counts[pr.Repo]++
	}

	repos := make([]repoInfo, 0, len(counts))
	for repo, count := range counts {
		repos = append(repos, repoInfo{Repo: repo, PRCount: count, Selected: selected[repo]})
	}
	return repos
}

func (s *Server) refreshRepoCache() {
	if s.RepoSearchFn == nil {
		s.mu.Lock()
		s.repoCache = s.reposFromCurrentPRs()
		s.repoCacheAt = time.Now()
		s.mu.Unlock()
		return
	}

	results, err := s.RepoSearchFn()
	if err != nil {
		logger.Warnf("repo search failed: %v", err)
		return
	}

	s.mu.RLock()
	selectedRepos := s.config.Repos
	s.mu.RUnlock()

	selected := make(map[string]bool, len(selectedRepos))
	for _, r := range selectedRepos {
		selected[r] = true
	}

	counts := make(map[string]int)
	for _, pr := range results {
		counts[pr.Repo]++
	}

	repos := make([]repoInfo, 0, len(counts))
	for repo, count := range counts {
		repos = append(repos, repoInfo{Repo: repo, PRCount: count, Selected: selected[repo]})
	}

	s.mu.Lock()
	s.repoCache = repos
	s.repoCacheAt = time.Now()
	s.mu.Unlock()
}

// handleRepoFavicon serves a cached favicon for a repo homepage. Only homepages
// that currently appear on one of the tracked PRs are accepted — this prevents
// the endpoint from being used as an open favicon proxy for arbitrary sites.
func (s *Server) handleRepoFavicon(w http.ResponseWriter, r *http.Request) {
	homepage := r.URL.Query().Get("homepage")
	if homepage == "" {
		http.Error(w, "homepage param required", http.StatusBadRequest)
		return
	}

	if !s.knownHomepage(homepage) {
		http.Error(w, "unknown homepage", http.StatusNotFound)
		return
	}

	store := cache.Shared()
	data, mime, hit, err := store.GetFavicon(homepage)
	if err != nil {
		logger.Warnf("favicon cache read %s: %v", homepage, err)
	}
	if !hit {
		data, mime, err = store.FetchFavicon(r.Context(), homepage)
		if err != nil {
			logger.Debugf("favicon fetch %s: %v", homepage, err)
			http.Error(w, "favicon unavailable", http.StatusNotFound)
			return
		}
	}
	if len(data) == 0 {
		// Negative cache hit — site has no usable favicon.
		http.Error(w, "no favicon", http.StatusNotFound)
		return
	}
	if mime != "" {
		w.Header().Set("Content-Type", mime)
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(data)
}

// knownHomepage reports whether any tracked PR declares this homepage URL. It
// guards the favicon endpoint from being abused as an open proxy.
func (s *Server) knownHomepage(homepage string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, pr := range s.prs {
		if pr.RepoHomepageURL == homepage {
			return true
		}
	}
	return false
}

type prDetail struct {
	PR           *github.PRInfo                `json:"pr,omitempty"`
	Runs         map[int64]*github.WorkflowRun `json:"runs,omitempty"`
	Comments     []github.PRComment            `json:"comments,omitempty"`
	GavelResults *GavelResultsSummary          `json:"gavelResults,omitempty"`
	Error        string                        `json:"error,omitempty"`
}

func (s *Server) handleDetail(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get("repo")
	numStr := r.URL.Query().Get("number")
	if repo == "" || numStr == "" {
		http.Error(w, `{"error":"repo and number params required"}`, http.StatusBadRequest)
		return
	}
	prNumber, err := strconv.Atoi(numStr)
	if err != nil {
		http.Error(w, `{"error":"invalid number"}`, http.StatusBadRequest)
		return
	}

	// Stream progressive detail via SSE so the UI can render sections
	// as they arrive rather than waiting for all GitHub round-trips.
	flusher, ok := w.(http.Flusher)
	if !ok {
		// Fallback: non-streaming response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.fetchPRDetail(repo, prNumber)) //nolint:errcheck
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	emit := func(event string, data any) {
		b, _ := json.Marshal(data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b)
		flusher.Flush()
	}

	// Serve from detail cache if available
	cacheKey := SyncStatusKey(repo, prNumber)
	if entry, ok := s.detailCache.Get(cacheKey); ok {
		d := entry.Detail
		if d.PR != nil {
			emit("pr", map[string]any{"pr": d.PR, "comments": d.Comments})
		}
		if len(d.Runs) > 0 {
			emit("runs", map[string]any{"runs": d.Runs})
		}
		if d.GavelResults != nil {
			emit("gavel", map[string]any{"gavelResults": d.GavelResults})
		}
		emit("done", nil)

		// Trigger background re-sync if stale
		if s.detailCache.IsStale(cacheKey, s.prUpdatedAt(repo, prNumber)) {
			if s.detailSyncer != nil {
				s.detailSyncer.Notify()
			}
		}
		return
	}

	opts := s.ghOpts
	opts.Repo = repo

	// Phase 1: PR metadata + comments (single GraphQL call)
	pr, err := github.FetchPR(opts, prNumber)
	if err != nil {
		logger.Warnf("failed to fetch PR %s#%d: %v", repo, prNumber, err)
		emit("error", map[string]string{"error": err.Error()})
		emit("done", nil)
		return
	}
	comments := prwatch.MergeAndFilter(pr.Comments, pr.ReviewThreads)
	emit("pr", map[string]any{"pr": pr, "comments": comments})

	// Phase 2: Workflow runs + gavel results in parallel
	type runResult struct {
		id  int64
		run *github.WorkflowRun
	}

	// Collect unique run IDs
	var runIDs []int64
	seen := make(map[int64]bool)
	for _, check := range pr.StatusCheckRollup {
		runID, err := github.ExtractRunID(check.DetailsURL)
		if err != nil || seen[runID] {
			continue
		}
		seen[runID] = true
		runIDs = append(runIDs, runID)
	}

	// Start gavel results fetch in parallel
	type gavelResult struct {
		summary *GavelResultsSummary
	}
	gavelCh := make(chan gavelResult, 1)
	allComments := append(pr.Comments, pr.ReviewThreads...)
	artifactID, artifactURL, hasArtifact := github.FindGavelArtifact(allComments)
	if hasArtifact {
		go func() {
			jsonBytes, err := github.DownloadArtifact(opts, artifactID)
			if err != nil {
				logger.Warnf("artifact %d download failed: %v", artifactID, err)
				gavelCh <- gavelResult{&GavelResultsSummary{
					ArtifactID:  artifactID,
					ArtifactURL: artifactURL,
					Error:       err.Error(),
				}}
			} else {
				gavelCh <- gavelResult{computeGavelSummary(jsonBytes, artifactID, artifactURL)}
			}
		}()
	}

	// Fetch workflow runs in parallel
	runCh := make(chan runResult, len(runIDs))
	for _, id := range runIDs {
		go func(runID int64) {
			run, err := github.FetchRunJobs(opts, runID)
			if err != nil {
				logger.Warnf("failed to fetch run %d: %v", runID, err)
				runCh <- runResult{runID, nil}
				return
			}
			runCh <- runResult{runID, run}
		}(id)
	}

	runs := make(map[int64]*github.WorkflowRun, len(runIDs))
	for range runIDs {
		rr := <-runCh
		if rr.run != nil {
			runs[rr.id] = rr.run
		}
	}
	if len(runs) > 0 {
		emit("runs", map[string]any{"runs": runs})
	}

	// Wait for gavel results
	var gavelSummary *GavelResultsSummary
	if hasArtifact {
		gr := <-gavelCh
		gavelSummary = gr.summary
		emit("gavel", map[string]any{"gavelResults": gavelSummary})
	}

	emit("done", nil)

	// Cache the result for future requests
	detail := prDetail{PR: pr, Runs: runs, Comments: comments, GavelResults: gavelSummary}
	prUpdated := s.prUpdatedAt(repo, prNumber)
	s.detailCache.Put(cacheKey, detail, prUpdated)
	s.detailCache.SetStatus(cacheKey, PRSyncStatus{State: SyncUpToDate, LastSynced: time.Now()})
	s.notify()
}

// fetchPRDetail loads full PR metadata, workflow runs, and comments from
// GitHub. Returns a prDetail with Error set on failure rather than an error
// value so callers (handleDetail, export) can embed it directly.
func (s *Server) fetchPRDetail(repo string, number int) prDetail {
	opts := s.ghOpts
	opts.Repo = repo

	result := prDetail{}

	pr, err := github.FetchPR(opts, number)
	if err != nil {
		logger.Warnf("failed to fetch PR %s#%d: %v", repo, number, err)
		result.Error = err.Error()
		return result
	}
	result.PR = pr

	runs := make(map[int64]*github.WorkflowRun)
	seen := make(map[int64]bool)
	for _, check := range pr.StatusCheckRollup {
		runID, err := github.ExtractRunID(check.DetailsURL)
		if err != nil || seen[runID] {
			continue
		}
		seen[runID] = true
		run, err := github.FetchRunJobs(opts, runID)
		if err != nil {
			logger.Warnf("failed to fetch run %d: %v", runID, err)
			continue
		}
		runs[runID] = run
	}
	result.Runs = runs
	result.Comments = prwatch.MergeAndFilter(pr.Comments, pr.ReviewThreads)

	// Scan all comments (including general issue comments, not just review
	// threads) for a gavel sticky comment with an artifact link.
	allComments := append(pr.Comments, pr.ReviewThreads...)
	if artifactID, artifactURL, found := github.FindGavelArtifact(allComments); found {
		jsonBytes, err := github.DownloadArtifact(opts, artifactID)
		if err != nil {
			logger.Warnf("artifact %d download failed: %v", artifactID, err)
			result.GavelResults = &GavelResultsSummary{
				ArtifactID:  artifactID,
				ArtifactURL: artifactURL,
				Error:       err.Error(),
			}
		} else {
			summary := computeGavelSummary(jsonBytes, artifactID, artifactURL)
			result.GavelResults = summary
		}
	}

	return result
}

// populatePRDetail fills a PRViewNode's detail fields using the same fetch
// logic that powers /api/prs/detail. Returns a non-nil error only when the
// PR metadata fetch itself fails (in which case the node is left unchanged).
func (s *Server) populatePRDetail(node *PRViewNode) error {
	detail := s.fetchPRDetail(node.Repo, node.Number)
	if detail.Error != "" {
		return fmt.Errorf("%s", detail.Error)
	}
	node.PR = detail.PR
	node.Runs = detail.Runs
	node.Comments = detail.Comments
	return nil
}

// prUpdatedAt returns the UpdatedAt for a PR in the current server state.
func (s *Server) prUpdatedAt(repo string, number int) time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, pr := range s.prs {
		if pr.Repo == repo && pr.Number == number {
			return pr.UpdatedAt
		}
	}
	return time.Time{}
}

type jobLogsResponse struct {
	JobID int64         `json:"jobId"`
	Logs  string        `json:"logs,omitempty"`
	Steps []github.Step `json:"steps,omitempty"`
	Error string        `json:"error,omitempty"`
}

func (s *Server) handleJobLogs(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get("repo")
	jobIDStr := r.URL.Query().Get("jobId")
	if repo == "" || jobIDStr == "" {
		http.Error(w, `{"error":"repo and jobId params required"}`, http.StatusBadRequest)
		return
	}
	jobID, err := strconv.ParseInt(jobIDStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid jobId"}`, http.StatusBadRequest)
		return
	}
	tail := 100
	if t := r.URL.Query().Get("tail"); t != "" {
		if n, err := strconv.Atoi(t); err == nil && n > 0 {
			tail = n
		}
	}

	opts := s.ghOpts
	opts.Repo = repo

	// We don't know the steps ahead of time — callers that already have the run can pass
	// steps via the runId route. Simpler: fetch the run to learn the job's steps, then
	// call FetchJobLogs. But that's an extra GitHub round-trip. Since the frontend already
	// has the job + step metadata in memory (from /api/prs/detail), it can reconstruct
	// the Job shell and we only need jobID for the log fetch. attachLogsToSteps needs
	// step names to split — so we also need runId here.
	runIDStr := r.URL.Query().Get("runId")
	if runIDStr == "" {
		http.Error(w, `{"error":"runId param required"}`, http.StatusBadRequest)
		return
	}
	runID, err := strconv.ParseInt(runIDStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid runId"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := jobLogsResponse{JobID: jobID}

	run, err := github.FetchRunJobs(opts, runID)
	if err != nil {
		logger.Warnf("failed to fetch run %d: %v", runID, err)
		resp.Error = err.Error()
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
		return
	}

	var job *github.Job
	for i := range run.Jobs {
		if run.Jobs[i].DatabaseID == jobID {
			job = &run.Jobs[i]
			break
		}
	}
	if job == nil {
		http.Error(w, `{"error":"job not found in run"}`, http.StatusNotFound)
		return
	}

	if err := github.FetchJobLogs(opts, job, tail); err != nil {
		logger.Warnf("failed to fetch logs for job %d: %v", jobID, err)
		resp.Error = err.Error()
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
		return
	}

	resp.Logs = job.Logs
	resp.Steps = job.Steps
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}
