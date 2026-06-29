package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
)

// prActionRequest is the shared body for the PR mutation endpoints. nodeId is
// the GraphQL node ID the UI already holds from the loaded PR detail; repo and
// number identify which cached detail to invalidate once the action lands.
type prActionRequest struct {
	Repo   string `json:"repo"`
	Number int    `json:"number"`
	NodeID string `json:"nodeId"`
	// Merge-only fields.
	Method string `json:"method,omitempty"` // rebase|squash|merge
	Auto   bool   `json:"auto,omitempty"`   // enable auto-merge instead of merging now
	// Approve-only field.
	Body string `json:"body,omitempty"`
}

// decodePRAction decodes and validates the fields shared by every PR action
// endpoint. It writes the error response itself and returns ok=false when the
// request is malformed, so callers can `if !ok { return }`.
func decodePRAction(w http.ResponseWriter, r *http.Request) (prActionRequest, bool) {
	var body prActionRequest
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return body, false
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return body, false
	}
	if strings.TrimSpace(body.NodeID) == "" {
		http.Error(w, `{"error":"nodeId is required"}`, http.StatusBadRequest)
		return body, false
	}
	return body, true
}

// afterPRAction drops the stale cached detail for the PR and requests an
// immediate list refresh so the new state (merged / approved / auto-merge set)
// surfaces without waiting for the next poll tick.
func (s *Server) afterPRAction(repo string, number int) {
	if s.detailCache != nil && repo != "" && number != 0 {
		s.detailCache.Invalidate(SyncStatusKey(repo, number))
	}
	select {
	case s.refreshCh <- struct{}{}:
	default:
	}
}

// handlePRMerge merges the PR now, or enables GitHub auto-merge when auto=true.
// Both routes through the user's configured token; GitHub rejects the action
// (not mergeable, no permission, auto-merge disallowed) with an error that is
// surfaced verbatim rather than swallowed.
func (s *Server) handlePRMerge(w http.ResponseWriter, r *http.Request) {
	body, ok := decodePRAction(w, r)
	if !ok {
		return
	}
	opts := s.ghOpts
	opts.Repo = body.Repo

	var err error
	if body.Auto {
		err = github.EnableAutoMerge(opts, body.NodeID, body.Method)
	} else {
		err = github.MergePullRequest(opts, body.NodeID, body.Method)
	}
	if err != nil {
		logger.Warnf("merge PR %s#%d (auto=%v, method=%s): %v", body.Repo, body.Number, body.Auto, body.Method, err)
		writeJSONError(w, http.StatusBadGateway, err)
		return
	}

	s.afterPRAction(body.Repo, body.Number)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"ok"}`)
}

// handlePRApprove submits an APPROVE review on the PR with an optional comment.
func (s *Server) handlePRApprove(w http.ResponseWriter, r *http.Request) {
	body, ok := decodePRAction(w, r)
	if !ok {
		return
	}
	opts := s.ghOpts
	opts.Repo = body.Repo

	if err := github.ApprovePullRequest(opts, body.NodeID, body.Body); err != nil {
		logger.Warnf("approve PR %s#%d: %v", body.Repo, body.Number, err)
		writeJSONError(w, http.StatusBadGateway, err)
		return
	}

	s.afterPRAction(body.Repo, body.Number)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"ok"}`)
}
