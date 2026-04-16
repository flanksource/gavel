package testui

import (
	"encoding/json"
	"net/http"
)

// RerunRequest is the payload accepted by POST /api/rerun.
type RerunRequest struct {
	PackagePaths []string `json:"package_paths,omitempty"`
	WorkDir      string   `json:"work_dir,omitempty"`
	TestName     string   `json:"test_name,omitempty"`
	Suite        []string `json:"suite,omitempty"`
	Framework    string   `json:"framework,omitempty"`
	Lint         bool     `json:"lint,omitempty"`
	LintFiles    []string `json:"lint_files,omitempty"`
	LintLinters  []string `json:"lint_linters,omitempty"`
}

// RerunFunc is invoked to rerun the tests described by req.
// It runs synchronously; the handler returns once it completes.
type RerunFunc func(req RerunRequest) error

// SetRerunFunc installs the rerun callback. Call this before serving traffic.
func (s *Server) SetRerunFunc(fn RerunFunc) {
	s.rerunFn = fn
}

func (s *Server) handleRerun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.rerunFn == nil {
		http.Error(w, "rerun not supported", http.StatusNotImplemented)
		return
	}

	var req RerunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	if !s.rerunMu.TryLock() {
		http.Error(w, "rerun already in progress", http.StatusConflict)
		return
	}
	defer s.rerunMu.Unlock()

	if err := s.rerunFn(req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
