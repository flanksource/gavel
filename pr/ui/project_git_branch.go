package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/flanksource/captain/pkg/aiflags"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/git/branchmerge"
	"github.com/flanksource/gavel/git/gitstate"
	prcreate "github.com/flanksource/gavel/pr/create"
)

var (
	errUnknownBranch       = errors.New("unknown branch")
	errBranchWorktreeDirty = errors.New("branch worktree has uncommitted changes")
)

// projectBranchRef is a branch of a project's repository resolved against the
// project's base branch.
type projectBranchRef struct {
	project Project
	dir     string
	base    string
	branch  string
	head    string
}

// resolveProjectBranch reads {name} and the branch, answering 404 for an
// unknown project or branch.
func resolveProjectBranch(w http.ResponseWriter, r *http.Request, branch string) (projectBranchRef, bool) {
	project, err := GetProject(r.PathValue("name"))
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return projectBranchRef{}, false
	}
	ref := projectBranchRef{project: project, dir: project.ResolvedDir(), branch: strings.TrimSpace(branch)}
	if ref.branch == "" {
		respondError(w, http.StatusBadRequest, "branch is required")
		return projectBranchRef{}, false
	}
	if ref.base, err = gitstate.Base(ref.dir); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return projectBranchRef{}, false
	}
	if ref.head = localBranchHead(ref.dir, ref.branch); ref.head == "" {
		respondError(w, http.StatusNotFound, fmt.Errorf("%w %q in project %s", errUnknownBranch, ref.branch, project.Name).Error())
		return projectBranchRef{}, false
	}
	return ref, true
}

// localBranchHead is the tip sha of a local branch, or "" when branch is not a
// valid branch name or no such local branch exists.
func localBranchHead(dir, branch string) string {
	if exec.Command("git", "-C", dir, "check-ref-format", "--branch", branch).Run() != nil {
		return ""
	}
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch+"^{commit}").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// commitRange is merge-base(base, branch)..branch as full shas.
func (ref projectBranchRef) commitRange(file string) (gavelgit.CommitDiffOptions, error) {
	mergeBase, err := gavelgit.MergeBase(ref.dir, "refs/heads/"+ref.base, ref.head)
	if err != nil {
		return gavelgit.CommitDiffOptions{}, err
	}
	return gavelgit.CommitDiffOptions{Base: mergeBase, Head: ref.head, File: strings.TrimSpace(file)}, nil
}

func (s *Server) projectBranchRange(w http.ResponseWriter, r *http.Request) (projectBranchRef, gavelgit.CommitDiffOptions, bool) {
	query := r.URL.Query()
	ref, ok := resolveProjectBranch(w, r, query.Get("branch"))
	if !ok {
		return ref, gavelgit.CommitDiffOptions{}, false
	}
	opts, err := ref.commitRange(query.Get("file"))
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return ref, opts, false
	}
	if err := opts.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return ref, opts, false
	}
	return ref, opts, true
}

// handleProjectBranchFiles lists the files a branch changed since it forked
// from base, in the /api/todos/commits/files shape. The branch head and base
// tip are the last ref scan's, and the file list is the range's cached one,
// so only the first read of a range runs git.
func (s *Server) handleProjectBranchFiles(w http.ResponseWriter, r *http.Request) {
	project, err := GetProject(r.PathValue("name"))
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	branch := strings.TrimSpace(r.URL.Query().Get("branch"))
	if branch == "" {
		respondError(w, http.StatusBadRequest, "branch is required")
		return
	}
	tracker, ok := s.requestGitTracker(w)
	if !ok {
		return
	}
	ctx, dir := s.requestContext(r), project.ResolvedDir()
	state, err := tracker.State(ctx, dir)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	repoID, err := tracker.Track(ctx, dir)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	head, found, err := tracker.Store().BranchHead(ctx, repoID, branch)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !found {
		respondError(w, http.StatusNotFound, fmt.Errorf("%w %q in project %s", errUnknownBranch, branch, project.Name).Error())
		return
	}
	files, mergeBase, err := storedRangeFiles(ctx, tracker, dir, gitstate.RangeKey{Base: state.BaseSHA, Head: head})
	if err != nil {
		respondError(w, todoCommitDiffErrorStatus(err), err.Error())
		return
	}
	respondJSON(w, http.StatusOK, todoCommitFilesResponse{Hash: head, Base: mergeBase, Files: files})
}

// handleProjectBranchDiff returns a branch's diff since it forked from base,
// optionally narrowed to one file, in the /api/todos/commits/diff shape.
func (s *Server) handleProjectBranchDiff(w http.ResponseWriter, r *http.Request) {
	ref, opts, ok := s.projectBranchRange(w, r)
	if !ok {
		return
	}
	result, err := gavelgit.CommitDiff(ref.dir, opts)
	if err != nil {
		respondError(w, todoCommitDiffErrorStatus(err), err.Error())
		return
	}
	respondJSON(w, http.StatusOK, todoCommitDiffResponse{
		Diff: result.Diff, Truncated: result.Truncated, Binary: result.Binary, Path: opts.File, Commit: opts.Head,
	})
}

// branchWorktree is the linked worktree that has the branch checked out, or
// the zero Worktree when none does.
func (ref projectBranchRef) branchWorktree() (gavelgit.Worktree, error) {
	worktrees, err := gavelgit.ListWorktrees(ref.dir)
	if err != nil {
		return gavelgit.Worktree{}, err
	}
	for _, wt := range worktrees {
		if wt.Branch == ref.branch {
			return wt, nil
		}
	}
	return gavelgit.Worktree{}, nil
}

// projectPRDeps is the GitHub and AI content seam for branch PRs and squash
// messages; specs set Server.prDeps to stub both.
func (s *Server) projectPRDeps() prcreate.Deps {
	if s.prDeps != nil {
		return *s.prDeps
	}
	return prcreate.DefaultDeps()
}

// squashMessage is the AI-generated PR title and body of the commits.
func (s *Server) squashMessage(ctx context.Context, dir string, commits []string) (string, error) {
	deps := s.projectPRDeps()
	if deps.GenerateContent == nil {
		return "", errors.New("PR content generation is not configured")
	}
	input, err := prcreate.ContentInput(dir, commits, aiflags.ModelFlags{}, nil)
	if err != nil {
		return "", err
	}
	content, err := deps.GenerateContent(ctx, input)
	if err != nil {
		return "", fmt.Errorf("generate the squash commit message: %w", err)
	}
	message := strings.TrimSpace(content.Title)
	if message == "" {
		return "", errors.New("generate the squash commit message: the generated title is empty")
	}
	if body := strings.TrimSpace(content.Body); body != "" {
		message += "\n\n" + body
	}
	return message, nil
}

// respondBranchError maps a branch merge or PR failure onto its status: 409
// for refusals and conflicts (listing the conflicting paths), 400 for invalid
// options and 500 for anything else.
func respondBranchError(w http.ResponseWriter, err error) {
	var conflict *branchmerge.ConflictError
	var prConflict *prcreate.ConflictError
	switch {
	case errors.As(err, &conflict):
		respondJSON(w, http.StatusConflict, map[string]any{"error": err.Error(), "conflicts": conflict.Paths})
	case errors.As(err, &prConflict):
		paths, pathsErr := unmergedPaths(prConflict.Worktree)
		if pathsErr != nil {
			err = errors.Join(err, pathsErr)
		}
		respondJSON(w, http.StatusConflict, map[string]any{"error": err.Error(), "conflicts": paths})
	case errors.Is(err, branchmerge.ErrInvalidOptions):
		respondError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, branchmerge.ErrCheckoutNotReady), errors.Is(err, branchmerge.ErrNothingToMerge),
		errors.Is(err, errBranchWorktreeDirty):
		respondError(w, http.StatusConflict, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, err.Error())
	}
}

func unmergedPaths(dir string) ([]string, error) {
	out, err := exec.Command("git", "-C", dir, "diff", "--name-only", "--diff-filter=U").Output()
	if err != nil {
		return []string{}, fmt.Errorf("list the conflicting paths in %s: %w", dir, err)
	}
	return append([]string{}, strings.Fields(string(out))...), nil
}
