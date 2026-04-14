package testui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner/bench"
	"github.com/flanksource/gavel/testrunner/parsers"
)

type Server struct {
	mu      sync.RWMutex
	tests   []parsers.Test
	lint    []*linters.LinterResult
	lintRun bool
	benchCmp *bench.BenchComparison
	done    bool
	updated chan struct{}
	gitRoot string

	rerunMu sync.Mutex
	rerunFn RerunFunc
}

func NewServer() *Server {
	return &Server{
		updated: make(chan struct{}, 1),
	}
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

// MarkDone flips the snapshot to done without requiring a test stream.
// Used by `gavel lint --ui` where there is no test channel to drain.
func (s *Server) MarkDone() {
	s.mu.Lock()
	s.done = true
	s.mu.Unlock()
	s.notify()
}

func (s *Server) StreamFrom(ch <-chan []parsers.Test) {
	go func() {
		for tests := range ch {
			s.mu.Lock()
			s.tests = tests
			s.mu.Unlock()
			s.notify()
		}
		s.mu.Lock()
		s.done = true
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
	mux.HandleFunc("/", s.handlePage)
	mux.HandleFunc("/api/tests", s.handleJSON)
	mux.HandleFunc("/api/tests/stream", s.handleSSE)
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
	Tests   []parsers.Test           `json:"tests"`
	Lint    []*linters.LinterResult  `json:"lint,omitempty"`
	LintRun bool                     `json:"lint_run,omitempty"`
	Bench   *bench.BenchComparison   `json:"bench,omitempty"`
	Done    bool                     `json:"done"`
}

func (s *Server) snapshot() snapshot {
	return snapshot{Tests: s.tests, Lint: s.lint, LintRun: s.lintRun, Bench: s.benchCmp, Done: s.done}
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
		data := s.snapshot()
		s.mu.RUnlock()

		b, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", b)
		flusher.Flush()

		if data.Done {
			fmt.Fprintf(w, "event: done\ndata: {}\n\n")
			flusher.Flush()
			return
		}
	}
}
