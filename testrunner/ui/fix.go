package testui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

// LintFixRequest asks an AI to repair the open findings named by fingerprint,
// optionally with another model than the reviewer's and a cap on the
// fix→re-lint cycles. Only these findings are sent and checked; the rest of
// the run, ignored and dismissed findings included, is left alone.
type LintFixRequest struct {
	Fingerprints  []string `json:"fingerprints"`
	Model         string   `json:"model,omitempty"`
	MaxIterations int      `json:"max_iterations,omitempty"`
}

// LintFixRecord is the outcome of a scoped AI fix, saved with the run that
// re-linted after it so history shows review → fix and what each fix cost.
type LintFixRecord struct {
	Fingerprints []string `json:"fingerprints"`
	Model        string   `json:"model,omitempty"`
	Iterations   int      `json:"iterations"`
	StopReason   string   `json:"stop_reason"`
	CostUSD      float64  `json:"cost_usd"`
	// Remaining are the requested fingerprints the final re-lint still reports.
	Remaining []string `json:"remaining,omitempty"`
}

// LintFixFunc runs a scoped fix synchronously, streaming the agent's output to
// output, and leaves the re-linted results on the server.
type LintFixFunc func(req LintFixRequest, output *RerunOutputBuffer) error

// SetLintFixFunc installs the AI fix callback. Call this before serving traffic.
func (s *Server) SetLintFixFunc(fn LintFixFunc) {
	s.lintFixFn = fn
}

func (s *Server) handleLintFix(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.lintFixFn == nil {
		http.Error(w, "AI fix not supported by this run", http.StatusNotImplemented)
		return
	}
	var req LintFixRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(req.Fingerprints) == 0 {
		http.Error(w, "fingerprints are required", http.StatusBadRequest)
		return
	}
	if req.MaxIterations < 0 {
		http.Error(w, fmt.Sprintf("max_iterations must not be negative, got %d", req.MaxIterations), http.StatusBadRequest)
		return
	}
	if missing := s.notOpen(req.Fingerprints); len(missing) > 0 {
		http.Error(w, "not open findings in this run (ignored, dismissed or already fixed): "+strings.Join(missing, ", "), http.StatusConflict)
		return
	}
	if !s.rerunMu.TryLock() {
		http.Error(w, "a rerun or fix is already in progress", http.StatusConflict)
		return
	}
	defer s.rerunMu.Unlock()

	buf := NewRerunOutputBuffer(fmt.Sprintf("ai-fix %d finding(s)", len(req.Fingerprints)))
	s.mu.Lock()
	s.rerunOutput = buf
	s.mu.Unlock()

	if err := s.lintFixFn(req, buf); err != nil {
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
	s.mu.RLock()
	record := s.lintFix
	s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(record)
}

// SetLintFix records the outcome of the last scoped fix for the snapshot.
func (s *Server) SetLintFix(record *LintFixRecord) {
	s.mu.Lock()
	s.lintFix = record
	s.mu.Unlock()
	s.notify()
}

// notOpen returns the fingerprints that are not an open violation of the
// current lint results.
func (s *Server) notOpen(fingerprints []string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var open []string
	for _, result := range s.lint {
		if result == nil {
			continue
		}
		for _, v := range result.Violations {
			open = append(open, v.Fingerprint)
		}
	}
	var missing []string
	for _, fp := range fingerprints {
		if !slices.Contains(open, fp) {
			missing = append(missing, fp)
		}
	}
	return missing
}
