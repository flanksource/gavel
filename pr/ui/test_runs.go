package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	cache "github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/snapshots"
)

// testRunView is the per-run row the Projects tree renders. It mirrors
// snapshots.RunInfo minus the on-disk Path (the detail endpoint resolves the
// file server-side from project+runId, so the path never reaches the client).
type testRunView struct {
	RunID      string    `json:"runId"`
	Kind       string    `json:"kind"`
	Started    time.Time `json:"started,omitempty"`
	Ended      time.Time `json:"ended,omitempty"`
	Repo       string    `json:"repo,omitempty"`
	SHA        string    `json:"sha,omitempty"`
	Frameworks []string  `json:"frameworks,omitempty"`

	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
	Warned  int `json:"warned"`
	Total   int `json:"total"`

	LintViolations int `json:"lintViolations"`
	LintLinters    int `json:"lintLinters"`
}

type projectRuns struct {
	Name string        `json:"name"`
	Dir  string        `json:"dir"`
	Runs []testRunView `json:"runs"`
}

type testRunsResponse struct {
	Projects []projectRuns `json:"projects"`
}

// handleTestRuns serves the .gavel/last.json snapshot grouped by registered
// project.
func (s *Server) handleTestRuns(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	projects, err := collectTestRuns(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := testRunsResponse{Projects: projects}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// handleTestRunsStream pushes the grouped run list on every server update (the
// syncer calls s.notify() after a scan) with a slow ticker as a fallback.
func (s *Server) handleTestRunsStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	writeSnap := func() bool {
		projects, err := collectTestRuns(r.Context())
		if err != nil {
			payload, _ := json.Marshal(map[string]string{"error": err.Error()})
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", payload)
			flusher.Flush()
			return false
		}
		payload, err := json.Marshal(testRunsResponse{Projects: projects})
		if err != nil {
			return false
		}
		fmt.Fprintf(w, "data: %s\n\n", payload)
		flusher.Flush()
		return true
	}

	if !writeSnap() {
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-s.updated:
		case <-ticker.C:
		}
		if !writeSnap() {
			return
		}
	}
}

// handleTestRun streams a single run's snapshot JSON, read from disk on demand.
// The file is resolved from (project|dir, runId) — never from a client-supplied
// path — and validated to live inside the workspace's .gavel directory. The
// Tests tab addresses workspaces by project name; every todo surface is
// dir-keyed, so both spellings resolve here.
func (s *Server) handleTestRun(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	project := strings.TrimSpace(query.Get("project"))
	runID := query.Get("runId")

	// Presence, not emptiness, picks the addressing mode: a todo surface on the
	// server's default workspace sends a bare `dir=`.
	if query.Has("project") == query.Has("dir") {
		http.Error(w, `{"error":"exactly one of project or dir is required"}`, http.StatusBadRequest)
		return
	}
	workDir := s.resolveTodoDir(strings.TrimSpace(query.Get("dir")))
	if query.Has("project") {
		p, err := GetProject(project)
		if err != nil {
			respondError(w, statusForProjectErr(err), err.Error())
			return
		}
		workDir = p.ResolvedDir()
	}

	path, err := resolveRunPath(workDir, runID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusBadRequest)
		return
	}
	if path == "" {
		http.Error(w, `{"error":"run not found"}`, http.StatusNotFound)
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Warnf("read test run %s: %v", path, err)
		http.Error(w, `{"error":"failed to read run"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

// collectTestRuns groups each project's .gavel/last.json snapshot under the
// live project list so the tree shows only the current run.
//
// The `last` pointer is always read from disk, so the view never lags a run that
// just finished. Only the run's counts come from the cache, and only when the
// syncer has already scanned that exact run — run-*.json files are immutable, so
// a row for a given run id can never be a stale view of it. Decoding those
// snapshots was 61% of this server's CPU while the Tests tab was open: the
// pointer file is a few hundred bytes, but the run it names carries the whole
// test tree and lint violation set, and there is one per project per request.
func collectTestRuns(ctx context.Context) ([]projectRuns, error) {
	projects, err := LoadProjects()
	if err != nil {
		return nil, err
	}
	store := cache.Shared()
	byDir := map[string][]testRunView{}
	for _, project := range projects {
		dir := project.ResolvedDir()
		view, err := lastRunView(ctx, store, dir)
		if err != nil {
			logger.Warnf("load last test run %s: %v", dir, err)
			continue
		}
		if view != nil {
			byDir[dir] = []testRunView{*view}
		}
	}
	return groupRuns(projects, byDir), nil
}

// lastRunView resolves a workspace's current run, preferring the cached summary
// of the run the `last` pointer names and decoding the snapshot only when the
// scanner has not reached it yet (or no cache is configured).
func lastRunView(ctx context.Context, store *cache.Store, dir string) (*testRunView, error) {
	if dir == "" {
		return nil, nil
	}
	pointer, err := snapshots.LoadPointer(dir, snapshots.PointerLast)
	if err != nil || pointer == nil {
		return nil, err
	}
	if row := cachedRunRow(ctx, store, dir, pointer); row != nil {
		// The pointer's own name, not the row's run id: this view is always "the
		// run `last` names", and the detail endpoint resolves PointerLast through
		// the pointer. Reporting the underlying run id only on a cache hit would
		// make the same request answer differently before and after a sweep.
		return viewFromCache(*row, snapshots.PointerLast), nil
	}
	run, err := snapshots.LastRun(dir)
	if err != nil || run == nil {
		return nil, err
	}
	view := viewFromInfo(*run)
	return &view, nil
}

// pointerRunID is the cache key for the snapshot a pointer names: the file's
// stem. The reader and the syncer must derive it identically or every lookup
// misses, so both call this.
//
// Note this is NOT a run-*.json stem. Save writes sha-<id>.json and points
// `last` at it; SavePerRun writes run-*.json. The Tests tab renders the former
// and the sweep enumerates the latter, which is why caching the sweep alone left
// this view uncached.
func pointerRunID(pointer *snapshots.Pointer) string {
	if pointer == nil || strings.TrimSpace(pointer.Path) == "" {
		return ""
	}
	stem := strings.TrimSuffix(filepath.Base(pointer.Path), ".json")
	if stem == "" || stem == snapshots.PointerLast {
		return ""
	}
	return stem
}

// cachedRunRow looks up the run a pointer names. A cache miss, a disabled cache
// and a read error are all "decode the snapshot instead", so the Tests tab keeps
// working before the first sweep and with no database configured.
func cachedRunRow(ctx context.Context, store *cache.Store, dir string, pointer *snapshots.Pointer) *cache.TestRunCache {
	runID := pointerRunID(pointer)
	if runID == "" {
		return nil
	}
	row, err := store.GetTestRun(ctx, dir, runID)
	if err != nil {
		logger.Debugf("read cached test run %s/%s: %v", dir, runID, err)
		return nil
	}
	return row
}

func viewFromCache(row cache.TestRunCache, runID string) *testRunView {
	var frameworks []string
	if len(row.Frameworks) > 0 {
		if err := json.Unmarshal(row.Frameworks, &frameworks); err != nil {
			logger.Debugf("decode cached frameworks for %s: %v", row.RunID, err)
		}
	}
	return &testRunView{
		RunID:          runID,
		Kind:           row.Kind,
		Started:        timeFromNanos(row.StartedTS),
		Ended:          timeFromNanos(row.EndedTS),
		Repo:           row.Repo,
		SHA:            row.SHA,
		Frameworks:     frameworks,
		Passed:         row.Passed,
		Failed:         row.Failed,
		Skipped:        row.Skipped,
		Warned:         row.Warned,
		Total:          row.Total,
		LintViolations: row.LintViolations,
		LintLinters:    row.LintLinters,
	}
}

// timeFromNanos mirrors test_run_syncer.nanos: 0 means the field was unset, not
// the unix epoch.
func timeFromNanos(ts int64) time.Time {
	if ts == 0 {
		return time.Time{}
	}
	return time.Unix(0, ts).UTC()
}

// groupRuns projects the current runs onto the live project list. Every project
// gets a non-nil Runs slice: a nil slice marshals to JSON null, which breaks the
// client's ProjectRuns.runs (typed TestRunView[], non-null) on `.runs.length`.
func groupRuns(projects []Project, byDir map[string][]testRunView) []projectRuns {
	out := make([]projectRuns, 0, len(projects))
	for _, p := range projects {
		dir := p.ResolvedDir()
		runs := byDir[dir]
		if runs == nil {
			runs = []testRunView{}
		}
		out = append(out, projectRuns{Name: p.Name, Dir: dir, Runs: runs})
	}
	return out
}

// resolveRunPath maps (workspace dir, runId) to .gavel/last.json or a
// run-*.json snapshot. Timestamped run IDs must be bare file stems so the
// result always lands inside <dir>/.gavel.
func resolveRunPath(dir, runID string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("workspace has no directory")
	}
	if runID == "" {
		return "", fmt.Errorf("runId required")
	}
	if runID == snapshots.PointerLast {
		run, err := snapshots.LastRun(dir)
		if err != nil || run == nil {
			return "", err
		}
		return run.Path, nil
	}
	if strings.ContainsAny(runID, `/\`) || strings.Contains(runID, "..") || !strings.HasPrefix(runID, snapshots.PerRunPrefix) {
		return "", fmt.Errorf("invalid runId")
	}
	gavelDir := filepath.Clean(filepath.Join(dir, snapshots.Dir))
	path := filepath.Join(gavelDir, runID+".json")
	if !strings.HasPrefix(path, gavelDir+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid runId")
	}
	if _, err := os.Stat(path); err != nil {
		return "", nil
	}
	return path, nil
}

func viewFromInfo(run snapshots.RunInfo) testRunView {
	return testRunView{
		RunID:          run.RunID,
		Kind:           run.Kind,
		Started:        run.Started,
		Ended:          run.Ended,
		Repo:           run.Repo,
		SHA:            run.SHA,
		Frameworks:     run.Frameworks,
		Passed:         run.Passed,
		Failed:         run.Failed,
		Skipped:        run.Skipped,
		Warned:         run.Warned,
		Total:          run.Total,
		LintViolations: run.LintViolations,
		LintLinters:    run.LintLinters,
	}
}
