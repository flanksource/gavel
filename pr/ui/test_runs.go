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
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/snapshots"
)

// testRunView is the per-run row the Tests tab renders. It mirrors
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

// handleTestRuns serves the run list grouped by registered project. It reads
// from the DB cache (populated by TestRunSyncer) and nudges the syncer so a
// stale-ish first paint refreshes behind the user.
func (s *Server) handleTestRuns(w http.ResponseWriter, r *http.Request) {
	s.notifyTestRunSyncer()
	w.Header().Set("Content-Type", "application/json")
	resp := testRunsResponse{Projects: collectTestRuns(r.Context())}
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

	s.notifyTestRunSyncer()
	writeSnap := func() {
		resp := testRunsResponse{Projects: collectTestRuns(r.Context())}
		if b, err := json.Marshal(resp); err == nil {
			fmt.Fprintf(w, "data: %s\n\n", b)
			flusher.Flush()
		}
	}

	writeSnap()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-s.updated:
		case <-ticker.C:
		}
		writeSnap()
	}
}

// handleTestRun streams a single run's snapshot JSON, read from disk on demand.
// The file is resolved from (project, runId) — never from a client-supplied
// path — and validated to live inside the workspace's .gavel directory.
func (s *Server) handleTestRun(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	runID := r.URL.Query().Get("runId")
	p, ok := GetProject(project)
	if !ok {
		http.Error(w, `{"error":"unknown project"}`, http.StatusNotFound)
		return
	}
	path, err := resolveRunPath(p.ResolvedDir(), runID)
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

// collectTestRuns groups cached runs under the live project list (so renames /
// removals reflect immediately). Falls back to a direct .gavel scan when no DB
// is configured, mirroring CachedGriteProvider's cache-optional design.
func collectTestRuns(ctx context.Context) []projectRuns {
	projects := LoadProjects()
	store := cache.Shared()

	byDir := map[string][]testRunView{}
	if store.Disabled() {
		for _, p := range projects {
			dir := p.ResolvedDir()
			infos, err := snapshots.ListRuns(dir, time.Time{})
			if err != nil {
				logger.Warnf("scan test runs %s: %v", dir, err)
			}
			byDir[dir] = viewsFromInfos(infos)
		}
	} else {
		rows, err := store.ListTestRuns(ctx)
		if err != nil {
			logger.Warnf("list test runs: %v", err)
		}
		for _, row := range rows {
			byDir[row.WorkspaceDir] = append(byDir[row.WorkspaceDir], viewFromRow(row))
		}
	}
	return groupRuns(projects, byDir)
}

// groupRuns projects the cached runs onto the live project list. Every project
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

// resolveRunPath maps (workspace dir, runId) to the run-*.json on disk. runId
// must be a bare per-run file stem (no separators, no traversal) so the result
// always lands inside <dir>/.gavel. Returns "" when the file is absent, an
// error when runId is malformed.
func resolveRunPath(dir, runID string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("workspace has no directory")
	}
	if runID == "" {
		return "", fmt.Errorf("runId required")
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

func viewFromRow(row cache.TestRunCache) testRunView {
	var frameworks []string
	if len(row.Frameworks) > 0 {
		_ = json.Unmarshal(row.Frameworks, &frameworks)
	}
	v := testRunView{
		RunID:          row.RunID,
		Kind:           row.Kind,
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
	if row.StartedTS > 0 {
		v.Started = time.Unix(0, row.StartedTS).UTC()
	}
	if row.EndedTS > 0 {
		v.Ended = time.Unix(0, row.EndedTS).UTC()
	}
	return v
}

func viewsFromInfos(infos []snapshots.RunInfo) []testRunView {
	out := make([]testRunView, 0, len(infos))
	for _, r := range infos {
		out = append(out, testRunView{
			RunID:          r.RunID,
			Kind:           r.Kind,
			Started:        r.Started,
			Ended:          r.Ended,
			Repo:           r.Repo,
			SHA:            r.SHA,
			Frameworks:     r.Frameworks,
			Passed:         r.Passed,
			Failed:         r.Failed,
			Skipped:        r.Skipped,
			Warned:         r.Warned,
			Total:          r.Total,
			LintViolations: r.LintViolations,
			LintLinters:    r.LintLinters,
		})
	}
	return out
}
