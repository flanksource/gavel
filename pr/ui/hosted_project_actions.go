// Hosted project actions: test and lint run as sandbox jobs of the organization's repository on
// the branch a member picks, and the repository's branches are listed to pick from.
package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
)

var (
	errProjectActionsNeedSandbox = errors.New("project actions need a sandbox: start gavel serve with --sandbox")
	errHostedActionScope         = errors.New("hosted test and lint run on the whole branch; files and options aren't supported")
)

// ProjectRepo is the organization's repository the project name names.
func (w HostedWorkspaces) ProjectRepo(ctx context.Context, org uuid.UUID, name string) (sandbox.Repo, error) {
	var repos []sandbox.Repo
	if err := w.repoQuery(ctx, org).Where("name = ?", name).Scan(&repos).Error; err != nil {
		return sandbox.Repo{}, fmt.Errorf("read repository %s of organization %s: %w", name, org, err)
	}
	if len(repos) == 0 {
		return sandbox.Repo{}, fmt.Errorf("project %q %w", name, errHostedWorkspace)
	}
	return repos[0], nil
}

// hostedProjectRepo reads the repository the request's project names, answering the request
// itself when it can't.
func (s *Server) hostedProjectRepo(w http.ResponseWriter, r *http.Request) (sandbox.Repo, bool) {
	if s.workspaces == nil {
		writeJSONError(w, http.StatusNotFound, errNoHostedWorkspaces)
		return sandbox.Repo{}, false
	}
	repo, err := s.workspaces.ProjectRepo(r.Context(), s.orgID(), r.PathValue("name"))
	if errors.Is(err, errHostedWorkspace) {
		writeJSONError(w, http.StatusNotFound, err)
		return sandbox.Repo{}, false
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return sandbox.Repo{}, false
	}
	return repo, true
}

// handleProjectBranches lists a hosted project's branches for a member to run an action on.
func (s *Server) handleProjectBranches(w http.ResponseWriter, r *http.Request) {
	if !s.hosted {
		writeJSONError(w, http.StatusNotFound, errors.New("branches are listed for a hosted organization's repositories"))
		return
	}
	repo, ok := s.hostedProjectRepo(w, r)
	if !ok {
		return
	}
	branches, err := s.workspaces.Sources.Branches(r.Context(), repo)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, branches)
}

// startHostedProjectAction runs a hosted project's test or lint as a sandbox job on the branch
// the request picks, as the request's user. Commits only come from agent runs.
func (s *Server) startHostedProjectAction(w http.ResponseWriter, r *http.Request) {
	var request projectActionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	switch request.Action {
	case projectActionCommit:
		writeJSONError(w, commitsFromAgentRuns.status, errors.New(commitsFromAgentRuns.message))
		return
	case projectActionTest, projectActionLint:
	default:
		respondError(w, http.StatusBadRequest, fmt.Sprintf("unsupported project action %q", request.Action))
		return
	}
	if s.agentRuns == nil {
		writeJSONError(w, http.StatusNotImplemented, errProjectActionsNeedSandbox)
		return
	}
	if len(request.Files) > 0 || request.Options != nil {
		writeJSONError(w, http.StatusBadRequest, errHostedActionScope)
		return
	}
	if request.Branch == "" {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("pick the branch to run %s on", request.Action))
		return
	}
	repo, ok := s.hostedProjectRepo(w, r)
	if !ok {
		return
	}
	sha, err := s.workspaces.Sources.BranchCommit(r.Context(), repo, request.Branch)
	if errors.Is(err, sandbox.ErrUnknownBranch) {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	// The job outlives the request that starts it.
	run, err := s.agentRuns.Dispatcher.Start(context.WithoutCancel(r.Context()), tenant.FromContext(r.Context()),
		sandbox.RunSpec{Kind: string(request.Action), RepoID: repo.ID, Ref: "refs/heads/" + request.Branch, SHA: sha}, nil)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	id := run.ID.String()
	respondJSON(w, http.StatusAccepted, projectActionStatus{
		Action: request.Action, RunID: id, Href: "/runs/" + id, Running: true, StartedAt: time.Now().UTC(),
	})
}
