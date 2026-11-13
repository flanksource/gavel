package testui

import (
	"encoding/json"
	"net/http"

	"github.com/flanksource/gavel/linters"
)

// DismissRequest sets violations aside, or with Restore brings them back, by
// fingerprint. A dismissal lasts while this server runs and never writes
// .gavel.yaml; it keeps a finding out of verdicts and AI fixes for this review.
type DismissRequest struct {
	Fingerprints []string `json:"fingerprints"`
	Restore      bool     `json:"restore,omitempty"`
}

type DismissResponse struct {
	Moved int `json:"moved"`
}

func (s *Server) handleLintDismiss(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req DismissRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(req.Fingerprints) == 0 {
		http.Error(w, "fingerprints are required", http.StatusBadRequest)
		return
	}
	set := make(map[string]bool, len(req.Fingerprints))
	for _, fingerprint := range req.Fingerprints {
		set[fingerprint] = true
	}

	s.mu.Lock()
	if s.dismissed == nil {
		s.dismissed = map[string]bool{}
	}
	var moved int
	if req.Restore {
		for fingerprint := range set {
			delete(s.dismissed, fingerprint)
		}
		moved = linters.RestoreViolations(s.lint, set)
	} else {
		for fingerprint := range set {
			s.dismissed[fingerprint] = true
		}
		moved = linters.DismissViolations(s.lint, set)
	}
	s.mu.Unlock()
	s.notify()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(DismissResponse{Moved: moved})
}
