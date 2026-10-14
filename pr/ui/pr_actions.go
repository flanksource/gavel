package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/todos/types"
)

// prActionRequest is the body for approval requests. nodeId is
// the GraphQL node ID the UI already holds from the loaded PR detail; repo and
// number identify which cached detail to invalidate once the action lands.
type prActionRequest struct {
	Repo   string `json:"repo"`
	Number int    `json:"number"`
	NodeID string `json:"nodeId"`
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

type prMergePayload struct {
	URL string `json:"url"`
	model.MergeInput
}

// handlePRMerge completes a linked hosted PR using its native merge strategy.
func (s *Server) handlePRMerge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}
	var body prMergePayload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	link, err := types.PRFromURL(body.URL)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	repo, err := linkedPRRepository(link.URL)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	if _, err := github.MergeMethodFor(body.Method); err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	client, number, err := provider.ResolveURL(provider.Options{WorkDir: s.ghOpts.WorkDir, Token: s.ghOpts.Token}, link.URL)
	if err == nil {
		err = client.MergePR(r.Context(), number, body.MergeInput)
	}
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err)
		return
	}
	s.afterPRAction(repo, number)
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

// handlePRUpdateBranch updates the PR's head branch by merging the base branch
// into it — the equivalent of GitHub's "Update branch" button. This is used
// when the PR is behind its target branch.
func (s *Server) handlePRUpdateBranch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Repo   string `json:"repo"`
		Number int    `json:"number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Repo) == "" || body.Number == 0 {
		http.Error(w, `{"error":"repo and number are required"}`, http.StatusBadRequest)
		return
	}

	opts := s.ghOpts
	opts.Repo = body.Repo

	if err := github.UpdatePullRequestBranch(opts, body.Number); err != nil {
		logger.Warnf("update branch PR %s#%d: %v", body.Repo, body.Number, err)
		writeJSONError(w, http.StatusBadGateway, err)
		return
	}

	s.afterPRAction(body.Repo, body.Number)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"ok"}`)
}
