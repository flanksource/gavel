package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/git/branchmerge"
	"github.com/flanksource/gavel/git/gitstate"
	prcreate "github.com/flanksource/gavel/pr/create"
)

type projectBranchMergeRequest struct {
	Branch  string           `json:"branch"`
	Mode    branchmerge.Mode `json:"mode"`
	Message string           `json:"message,omitempty"`
}

type projectBranchMergeResponse struct {
	TargetBranch    string           `json:"targetBranch"`
	LandedSHA       string           `json:"landedSha"`
	Mode            branchmerge.Mode `json:"mode"`
	Commits         int              `json:"commits"`
	WorktreeRemoved bool             `json:"worktreeRemoved"`
	BranchDeleted   bool             `json:"branchDeleted"`
}

type projectBranchPRRequest struct {
	Branch string `json:"branch"`
	Draft  bool   `json:"draft"`
}

type projectBranchPRResponse struct {
	Number      int    `json:"number"`
	URL         string `json:"url"`
	TopicBranch string `json:"topicBranch"`
}

func decodeStrict(r *http.Request, into any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(into); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	return nil
}

// handleProjectBranchMerge lands a branch onto the base branch checked out in
// the project's primary checkout, squashed or commit by commit, then removes
// the branch's worktree and deletes the branch. Main is not pushed.
func (s *Server) handleProjectBranchMerge(w http.ResponseWriter, r *http.Request) {
	var request projectBranchMergeRequest
	if err := decodeStrict(r, &request); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if request.Mode != branchmerge.Squash && request.Mode != branchmerge.Incremental {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("mode %q must be %q or %q", request.Mode, branchmerge.Squash, branchmerge.Incremental))
		return
	}
	ref, ok := resolveProjectBranch(w, r, request.Branch)
	if !ok {
		return
	}
	response, err := s.mergeProjectBranch(r.Context(), ref, request)
	// A failed merge can still have moved refs or removed the worktree.
	s.touchGit(ref.dir)
	if err != nil {
		respondBranchError(w, err)
		return
	}
	s.notify()
	respondJSON(w, http.StatusOK, response)
}

func (s *Server) mergeProjectBranch(ctx context.Context, ref projectBranchRef, request projectBranchMergeRequest) (projectBranchMergeResponse, error) {
	wt, err := ref.branchWorktree()
	if err != nil {
		return projectBranchMergeResponse{}, err
	}
	if err := refuseDirtyWorktree(ctx, wt); err != nil {
		return projectBranchMergeResponse{}, err
	}
	commits, err := branchmerge.RangeCommits(ref.dir, "refs/heads/"+ref.base, ref.head)
	if err != nil {
		return projectBranchMergeResponse{}, err
	}
	if len(commits) == 0 {
		return projectBranchMergeResponse{}, fmt.Errorf("%w: %s has no commits that are not on %s", branchmerge.ErrNothingToMerge, ref.branch, ref.base)
	}
	opts := branchmerge.Options{Repo: ref.dir, Branch: ref.branch, Base: ref.base, Mode: request.Mode, Message: strings.TrimSpace(request.Message)}
	if request.Mode == branchmerge.Incremental {
		opts.Commits = commits
	} else if opts.Message == "" {
		if opts.Message, err = s.squashMessage(ctx, ref.dir, commits); err != nil {
			return projectBranchMergeResponse{}, err
		}
	}
	merged, err := branchmerge.Merge(opts)
	if err != nil {
		return projectBranchMergeResponse{}, err
	}
	cleanup := branchmerge.CleanupOptions{Repo: ref.dir, Branch: ref.branch, LandedSHA: merged.LandedSHA, Mode: request.Mode}
	if !wt.Primary {
		cleanup.Worktree = wt.Path
	}
	cleaned, err := branchmerge.Cleanup(cleanup)
	if err != nil {
		return projectBranchMergeResponse{}, fmt.Errorf("merged %s onto %s at %s, but its cleanup failed: %w", ref.branch, merged.TargetBranch, merged.LandedSHA, err)
	}
	return projectBranchMergeResponse{
		TargetBranch: merged.TargetBranch, LandedSHA: merged.LandedSHA, Mode: request.Mode, Commits: merged.Commits,
		WorktreeRemoved: cleaned.WorktreeRemoved, BranchDeleted: cleaned.BranchDeleted,
	}, nil
}

// refuseDirtyWorktree refuses a merge whose branch is checked out in a linked
// worktree holding uncommitted changes, which the cleanup would destroy.
func refuseDirtyWorktree(ctx context.Context, wt gavelgit.Worktree) error {
	if wt.Path == "" || wt.Primary {
		return nil
	}
	// Read live, never memoized: a stale "clean" must not let a merge destroy work.
	changes, _, _, err := gitstate.WorktreeChanges(ctx, wt.Path)
	if err != nil {
		return err
	}
	if changes != (gitstate.Changes{}) {
		return fmt.Errorf("%w: %s on %s; commit or discard them first", errBranchWorktreeDirty, wt.Path, wt.Branch)
	}
	return nil
}

// handleProjectBranchPR opens a pull request carrying a branch's commits not in
// the PR base, through the same prcreate flow a todo run's PR landing uses.
func (s *Server) handleProjectBranchPR(w http.ResponseWriter, r *http.Request) {
	var request projectBranchPRRequest
	if err := decodeStrict(r, &request); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	ref, ok := resolveProjectBranch(w, r, request.Branch)
	if !ok {
		return
	}
	baseRef, err := prcreate.ResolveBase(ref.dir, "")
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	commits, err := branchmerge.RangeCommits(ref.dir, baseRef, ref.head)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(commits) == 0 {
		respondBranchError(w, fmt.Errorf("%w: %s has no commits that are not on %s", branchmerge.ErrNothingToMerge, ref.branch, baseRef))
		return
	}
	result, err := prcreate.Create(r.Context(), ref.dir, prcreate.Input{
		SHAs: commits, Base: baseRef, Draft: request.Draft, Deps: s.projectPRDeps(),
	})
	s.touchGit(ref.dir)
	if err != nil {
		respondBranchError(w, err)
		return
	}
	s.notify()
	respondJSON(w, http.StatusOK, projectBranchPRResponse{Number: result.PR.Number, URL: result.PR.URL, TopicBranch: result.TopicBranch})
}
