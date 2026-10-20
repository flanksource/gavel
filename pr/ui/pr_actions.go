package ui

import (
	"encoding/json"
	"errors"
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
	// Update-branch only: the head the UI saw, so GitHub refuses a moved branch.
	HeadSHA string `json:"headSha,omitempty"`
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

// prRebaseRequest is what the AI rebase (prompts/library/pr-rebase.prompt)
// needs to rebase a PR the update-branch refused: the local checkout it works
// in, the remote it force-pushes to, and the exact head and base commits —
// fetched into the checkout here, so the prompt's worktree can start from them.
type prRebaseRequest struct {
	Repo    string `json:"repo"`
	Remote  string `json:"remote"`
	Number  int    `json:"number"`
	URL     string `json:"url"`
	Branch  string `json:"branch"`
	Base    string `json:"base"`
	BaseSHA string `json:"baseSha"`
	Head    string `json:"head"`
}

// prUpdateConflict is the 409 body of an update-branch refused over conflicts.
// Rebase is set when the AI rebase can run; otherwise RebaseBlocked says why.
type prUpdateConflict struct {
	Error         string           `json:"error"`
	Conflict      bool             `json:"conflict"`
	Rebase        *prRebaseRequest `json:"rebase,omitempty"`
	RebaseBlocked string           `json:"rebaseBlocked,omitempty"`
}

// handlePRUpdateBranch rebases the PR's head branch onto its base — GitHub's
// "Update with rebase". A refusal over conflicts answers 409 with what the AI
// rebase needs, so the UI can offer it instead of a dead end.
func (s *Server) handlePRUpdateBranch(w http.ResponseWriter, r *http.Request) {
	body, ok := decodePRAction(w, r)
	if !ok {
		return
	}
	opts := s.ghOpts
	opts.Repo = body.Repo

	err := github.UpdatePullRequestBranch(opts, github.UpdateBranchOptions{NodeID: body.NodeID, HeadSHA: body.HeadSHA})
	w.Header().Set("Content-Type", "application/json")
	switch {
	case errors.Is(err, github.ErrUpdateBranchConflict):
		conflict := prUpdateConflict{Error: err.Error(), Conflict: true}
		rebase, blocked := s.prRebaseRequest(body.Repo, body.Number)
		if blocked != nil {
			logger.Warnf("update branch PR %s#%d: AI rebase unavailable: %v", body.Repo, body.Number, blocked)
			conflict.RebaseBlocked = blocked.Error()
		}
		conflict.Rebase = rebase
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(conflict) //nolint:errcheck
	case err != nil:
		logger.Warnf("update branch PR %s#%d: %v", body.Repo, body.Number, err)
		writeJSONError(w, http.StatusBadGateway, err)
	default:
		s.afterPRAction(body.Repo, body.Number)
		fmt.Fprint(w, `{"status":"ok"}`)
	}
}

// prRebaseRequest resolves the local checkout of repo from the project catalog
// and fetches the PR's head and base into it. The rebase force-pushes the head
// branch, so a PR from a fork — whose branch is not on the checkout's remote —
// is refused rather than pushed somewhere else.
func (s *Server) prRebaseRequest(repo string, number int) (*prRebaseRequest, error) {
	projects, err := LoadProjects(LoadProjectsOptions{})
	if err != nil {
		return nil, err
	}
	project, ok := ProjectForRepo(projects, repo)
	if !ok {
		return nil, fmt.Errorf("no project has a local checkout of %s; add one to rebase it with AI", repo)
	}
	opts := s.ghOpts
	opts.Repo, opts.WorkDir = repo, project.Dir
	pr, err := github.FetchPR(opts, number)
	if err != nil {
		return nil, err
	}
	if pr.HeadRepositoryURL != "" && !strings.HasSuffix(strings.ToLower(pr.HeadRepositoryURL), "/"+strings.ToLower(repo)) {
		return nil, fmt.Errorf("the head branch lives in %s, not %s; rebase it from a checkout of the fork", pr.HeadRepositoryURL, repo)
	}
	remote, err := github.RepoRemote(opts)
	if err != nil {
		return nil, err
	}
	if err := github.EnsurePRCommits(opts, pr); err != nil {
		return nil, err
	}
	return &prRebaseRequest{
		Repo: project.Dir, Remote: remote, Number: pr.Number, URL: pr.URL,
		Branch: pr.HeadRefName, Base: pr.BaseRefName, BaseSHA: pr.BaseRefOID, Head: pr.HeadRefOID,
	}, nil
}
