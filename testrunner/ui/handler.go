package testui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner/bench"
	"github.com/flanksource/gavel/testrunner/parsers"
)

type Server struct {
	mu       sync.RWMutex
	tests    []parsers.Test
	lint     []*linters.LinterResult
	lintRun  bool
	benchCmp *bench.BenchComparison
	done     bool
	run      *RunMetadata
	updated  chan struct{}
	gitRoot  string
	diag     *DiagnosticsManager

	rerunMu sync.Mutex
	rerunFn RerunFunc
}

func NewServer() *Server {
	return &Server{
		updated: make(chan struct{}, 1),
	}
}

type RunMetadata struct {
	Sequence   int       `json:"sequence,omitempty"`
	Kind       string    `json:"kind,omitempty"`
	StartedAt  time.Time `json:"started_at,omitempty"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
}

func cloneRunMetadata(run *RunMetadata) *RunMetadata {
	if run == nil {
		return nil
	}
	cloned := *run
	return &cloned
}

func (s *Server) BeginRun(kind string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := 1
	if s.run != nil && s.run.Sequence > 0 {
		next = s.run.Sequence + 1
	}
	s.run = &RunMetadata{
		Sequence:  next,
		Kind:      kind,
		StartedAt: time.Now().UTC(),
	}
	s.done = false
	s.notify()
}

func (s *Server) finishRunLocked() {
	if s.run == nil || !s.run.FinishedAt.IsZero() {
		return
	}
	s.run.FinishedAt = time.Now().UTC()
}

func (s *Server) SetResults(tests []parsers.Test) {
	s.mu.Lock()
	s.tests = tests
	s.done = true
	s.mu.Unlock()
	s.notify()
}

// SetLintResults stores lint results so they appear in the next snapshot.
func (s *Server) SetLintResults(results []*linters.LinterResult) {
	s.mu.Lock()
	s.lint = results
	s.lintRun = true
	s.mu.Unlock()
	s.notify()
}

// SetBenchComparison stores a benchmark comparison so it appears in the next snapshot.
func (s *Server) SetBenchComparison(cmp *bench.BenchComparison) {
	s.mu.Lock()
	s.benchCmp = cmp
	s.done = true
	s.finishRunLocked()
	s.mu.Unlock()
	s.notify()
}

// SetGitRoot records the git root used for resolving relative paths and
// locating the .gavel.yaml written by the ignore endpoint.
func (s *Server) SetGitRoot(root string) {
	s.mu.Lock()
	s.gitRoot = root
	s.mu.Unlock()
}

func (s *Server) GitRoot() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.gitRoot
}

// MarkDone flips the snapshot to done without requiring a test stream.
// Used by `gavel lint --ui` where there is no test channel to drain.
func (s *Server) MarkDone() {
	s.mu.Lock()
	s.done = true
	s.finishRunLocked()
	s.mu.Unlock()
	s.notify()
}

func (s *Server) StreamFrom(ch <-chan []parsers.Test) {
	go func() {
		for tests := range ch {
			s.mu.Lock()
			s.tests = tests
			s.done = false
			s.mu.Unlock()
			s.notify()
		}
		s.mu.Lock()
		s.done = true
		s.finishRunLocked()
		s.mu.Unlock()
		s.notify()
	}()
}

func (s *Server) notify() {
	select {
	case s.updated <- struct{}{}:
	default:
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoute)
	mux.HandleFunc("/api/tests", s.handleJSON)
	mux.HandleFunc("/api/tests/stream", s.handleSSE)
	mux.HandleFunc("/api/diagnostics", s.handleDiagnosticsJSON)
	mux.HandleFunc("/api/diagnostics/collect", s.handleDiagnosticsCollect)
	mux.HandleFunc("/api/rerun", s.handleRerun)
	mux.HandleFunc("/api/lint/ignore", s.handleLintIgnore)
	mux.HandleFunc("/api/benchmarks", s.handleBenchJSON)
	return mux
}

func (s *Server) handleBenchJSON(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	cmp := s.benchCmp
	s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	if cmp == nil {
		w.Write([]byte("null")) //nolint:errcheck
		return
	}
	json.NewEncoder(w).Encode(cmp) //nolint:errcheck
}

func (s *Server) handleRoute(w http.ResponseWriter, r *http.Request) {
	req, ok := parseRouteRequest(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if req.IsExport {
		if req.Tab == viewTabDiagnostics {
			http.NotFound(w, r)
			return
		}
		s.handleExport(w, r, req)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, pageHTML())
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request, req routeRequest) {
	s.mu.RLock()
	report, err := s.buildExportReport(req)
	s.mu.RUnlock()
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
    <title>Test Results</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://code.iconify.design/iconify-icon/2.0.0/iconify-icon.min.js"></script>
</head>
<body>
    <div id="root"></div>
    <script>` + bundleJS + `</script>
</body>
</html>`
}

func (s *Server) handleJSON(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	data := s.snapshot()
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}

type snapshot struct {
	Tests                []parsers.Test          `json:"tests"`
	Lint                 []*linters.LinterResult `json:"lint,omitempty"`
	LintRun              bool                    `json:"lint_run,omitempty"`
	Bench                *bench.BenchComparison  `json:"bench,omitempty"`
	DiagnosticsAvailable bool                    `json:"diagnostics_available,omitempty"`
	Run                  *RunMetadata            `json:"run,omitempty"`
	Done                 bool                    `json:"done"`
}

func (s *Server) snapshot() snapshot {
	tests := s.tests
	taskTests := virtualTaskTests()
	if len(taskTests) > 0 {
		merged := make([]parsers.Test, 0, len(taskTests)+len(s.tests))
		merged = append(merged, taskTests...)
		merged = append(merged, s.tests...)
		tests = merged
	}
	return snapshot{
		Tests:                tests,
		Lint:                 s.lint,
		LintRun:              s.lintRun,
		Bench:                s.benchCmp,
		DiagnosticsAvailable: s.diag != nil,
		Run:                  cloneRunMetadata(s.run),
		Done:                 s.done && tasksDone(),
	}
}

func virtualTaskTests() []parsers.Test {
	snapshots := clickytask.SnapshotAll(TestTaskGroupName, LintTaskGroupName)
	if len(snapshots) == 0 {
		return nil
	}

	childrenByGroup := make(map[string][]parsers.Test)
	groupOrder := make([]string, 0, 2)
	groups := make(map[string]clickytask.TaskSnapshot)

	for _, snap := range snapshots {
		if snap.Type == "group" {
			groups[snap.ID] = snap
			groupOrder = append(groupOrder, snap.ID)
			continue
		}
		if snap.Group == "" {
			continue
		}
		childrenByGroup[snap.Group] = append(childrenByGroup[snap.Group], taskSnapshotToTest(snap))
	}

	var out []parsers.Test
	for _, id := range groupOrder {
		group, ok := groups[id]
		if !ok {
			continue
		}
		test := taskSnapshotToTest(group)
		test.Children = childrenByGroup[id]
		out = append(out, test)
	}
	return out
}

func taskSnapshotToTest(snap clickytask.TaskSnapshot) parsers.Test {
	t := parsers.Test{
		Name:      snap.Name,
		Framework: parsers.Framework("task"),
	}
	if snap.Type == "task" {
		t.Command = snap.Name
	}
	if snap.Message != "" {
		t.Message = snap.Message
	} else if snap.Error != "" {
		t.Message = snap.Error
	}
	if len(snap.Logs) > 0 {
		lines := make([]string, 0, len(snap.Logs))
		for _, entry := range snap.Logs {
			lines = append(lines, entry.Message)
		}
		t.Stderr = strings.Join(lines, "\n")
	}
	t.Context = map[string]any{
		"duration": snap.Duration,
		"status":   snap.Status,
		"type":     snap.Type,
	}

	switch snap.Status {
	case "running", "pending":
		t.Pending = true
	case "success", "PASS":
		t.Passed = true
	case "failed", "FAIL", "ERR", "warning":
		t.Failed = true
	default:
		t.Skipped = true
	}

	return t
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

	// Send initial snapshot immediately
	s.mu.RLock()
	initial := s.snapshot()
	s.mu.RUnlock()
	lastPayload := ""
	if b, err := json.Marshal(initial); err == nil {
		fmt.Fprintf(w, "data: %s\n\n", b)
		flusher.Flush()
		lastPayload = string(b)
	}

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-s.updated:
		case <-ticker.C:
		}

		s.mu.RLock()
		data := s.snapshot()
		s.mu.RUnlock()

		b, _ := json.Marshal(data)
		payload := string(b)
		if payload != lastPayload {
			fmt.Fprintf(w, "data: %s\n\n", b)
			flusher.Flush()
			lastPayload = payload
		}

		if data.Done {
			fmt.Fprintf(w, "event: done\ndata: {}\n\n")
			flusher.Flush()
			return
		}
	}
}

func tasksDone() bool {
	snapshots := clickytask.SnapshotAll(TestTaskGroupName, LintTaskGroupName)
	if len(snapshots) == 0 {
		return true
	}

	for _, snap := range snapshots {
		switch snap.Status {
		case "running", "pending":
			return false
		}
	}

	return true
}

func (s *Server) EnableDiagnostics(rootPID int) {
	s.mu.Lock()
	s.diag = NewDiagnosticsManager(rootPID, nil)
	s.mu.Unlock()
	s.notify()
}

func (s *Server) SetDiagnosticsManager(manager *DiagnosticsManager) {
	s.mu.Lock()
	s.diag = manager
	s.mu.Unlock()
	s.notify()
}

func (s *Server) diagnosticsManager() *DiagnosticsManager {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.diag
}

func (s *Server) handleDiagnosticsJSON(w http.ResponseWriter, _ *http.Request) {
	manager := s.diagnosticsManager()
	if manager == nil {
		http.Error(w, "diagnostics unavailable", http.StatusNotFound)
		return
	}
	snapshot, err := manager.Snapshot()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot) //nolint:errcheck
}

func (s *Server) handleDiagnosticsCollect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	manager := s.diagnosticsManager()
	if manager == nil {
		http.NotFound(w, r)
		return
	}

	var req StackCaptureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.PID == 0 {
		http.Error(w, "pid is required", http.StatusBadRequest)
		return
	}

	details, err := manager.CollectStack(req.PID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.notify()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details) //nolint:errcheck
}
