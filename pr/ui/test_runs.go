package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
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
	projects, err := collectTestRuns()
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
		projects, err := collectTestRuns()
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
func collectTestRuns() ([]projectRuns, error) {
	projects, err := LoadProjects()
	if err != nil {
		return nil, err
	}
	byDir := map[string][]testRunView{}
	for _, project := range projects {
		dir := project.ResolvedDir()
		run, err := snapshots.LastRun(dir)
		if err != nil {
			logger.Warnf("load last test run %s: %v", dir, err)
			continue
		}
		if run != nil {
			byDir[dir] = []testRunView{viewFromInfo(*run)}
		}
	}
	return groupRuns(projects, byDir), nil
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
