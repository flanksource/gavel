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
	Repos  []string `json:"repos"`
	All    bool     `json:"all,omitempty"`
	Org    string   `json:"org,omitempty"`
	Author string   `json:"author,omitempty"`
	Any    bool     `json:"any,omitempty"`
	Bots   bool     `json:"bots,omitempty"`
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

	RepoSearchFn func() (github.PRSearchResults, error)
	repoCache    []repoInfo
	repoCacheAt  time.Time
}

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
	Unread map[string]bool `json:"unread,omitempty"`
}

func NewServer(interval time.Duration, ghOpts github.Options, config SearchConfig) *Server {
	return &Server{
		interval:  interval,
		ghOpts:    ghOpts,
		config:    config,
		updated:   make(chan struct{}, 1),
		refreshCh: make(chan struct{}, 1),
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
	mux.HandleFunc("/", s.handlePage)
	mux.HandleFunc("/api/prs", s.handleJSON)
	mux.HandleFunc("/api/prs/stream", s.handleSSE)
	mux.HandleFunc("/api/prs/refresh", s.handleRefresh)
	mux.HandleFunc("/api/prs/pause", s.handlePause)
	mux.HandleFunc("/api/prs/detail", s.handleDetail)
	mux.HandleFunc("/api/prs/job-logs", s.handleJobLogs)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/repos", s.handleRepos)
	mux.HandleFunc("/api/activity", s.handleActivity)
	mux.HandleFunc("/api/activity/stream", s.handleActivityStream)
	mux.HandleFunc("/api/activity/reset", s.handleActivityReset)
	mux.HandleFunc("/api/activity/cache", s.handleActivityCache)
	mux.HandleFunc("/favicon.svg", handleFavicon)
	mux.HandleFunc("/brand/gavel-logo.svg", handleLogo)
	mux.HandleFunc("/brand/menubar.png", handleMenubarIcon)
	mux.HandleFunc("/brand/menubar-unread.png", handleMenubarUnreadIcon)
	mux.HandleFunc("/api/prs/seen", s.handleSeen)
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

func (s *Server) handlePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, pageHTML())
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

func (s *Server) handleJSON(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	data := s.snapshotLocked()
	s.mu.RUnlock()
	data = s.withUnread(data)

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
		go SaveSettings(UISettings{Repos: cfg.Repos, Author: cfg.Author, Any: cfg.Any, Bots: cfg.Bots})
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

type prDetail struct {
	PR       *github.PRInfo                `json:"pr,omitempty"`
	Runs     map[int64]*github.WorkflowRun `json:"runs,omitempty"`
	Comments []github.PRComment            `json:"comments,omitempty"`
	Error    string                        `json:"error,omitempty"`
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

	opts := s.ghOpts
	opts.Repo = repo

	w.Header().Set("Content-Type", "application/json")

	flusher, canFlush := w.(http.Flusher)
	_ = canFlush

	result := prDetail{}

	pr, err := github.FetchPR(opts, prNumber)
	if err != nil {
		logger.Warnf("failed to fetch PR %s#%d: %v", repo, prNumber, err)
		result.Error = err.Error()
		json.NewEncoder(w).Encode(result) //nolint:errcheck
		return
	}
	result.PR = pr

	// Fetch workflow runs
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

	// Comments and review threads are populated by the merged FetchPR query above.
	result.Comments = prwatch.MergeAndFilter(pr.Comments, pr.ReviewThreads)

	json.NewEncoder(w).Encode(result) //nolint:errcheck
	if canFlush {
		flusher.Flush()
	}
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
