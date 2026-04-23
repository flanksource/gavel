package testui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
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
// The output buffer receives live process stdout/stderr for streaming to the UI.
// It runs synchronously; the handler returns once it completes.
type RerunFunc func(req RerunRequest, output *RerunOutputBuffer) error

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

	command := rerunCommandLabel(req)
	buf := NewRerunOutputBuffer(command)
	s.mu.Lock()
	s.rerunOutput = buf
	s.mu.Unlock()

	err := s.rerunFn(req, buf)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			buf.Cancel()
			w.WriteHeader(http.StatusAccepted)
			return
		}
		buf.Finish(false)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.Finish(true)
	w.WriteHeader(http.StatusAccepted)
}

func rerunCommandLabel(req RerunRequest) string {
	if req.Lint {
		parts := []string{"lint"}
		if len(req.LintLinters) > 0 {
			parts = append(parts, strings.Join(req.LintLinters, ","))
		}
		if len(req.LintFiles) > 0 {
			parts = append(parts, fmt.Sprintf("(%d files)", len(req.LintFiles)))
		}
		return strings.Join(parts, " ")
	}
	parts := []string{"test"}
	if req.Framework != "" {
		parts = append(parts, req.Framework)
	}
	if req.TestName != "" {
		parts = append(parts, req.TestName)
	} else if len(req.PackagePaths) > 0 {
		parts = append(parts, strings.Join(req.PackagePaths, " "))
	}
	return strings.Join(parts, " ")
}
