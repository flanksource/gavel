package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/github/prcreate"
	"github.com/flanksource/gavel/status"
	"golang.org/x/sync/errgroup"
)

// projectGitSummaryConcurrency bounds how many projects git-summary inspects
// at once; each inspection runs several git processes per worktree.
const projectGitSummaryConcurrency = 4

var errUnknownWorktree = errors.New("not a worktree of the project")

type projectGitChanges struct {
	Staged    int `json:"staged"`
	Unstaged  int `json:"unstaged"`
	Both      int `json:"both"`
	Untracked int `json:"untracked"`
	Conflict  int `json:"conflict"`
	Adds      int `json:"adds"`
	Dels      int `json:"dels"`
}

type projectGitWorktree struct {
	gavelgit.Worktree
	Changes projectGitChanges `json:"changes"`
	// Ahead counts the worktree branch's commits not in base; 0 for the base
	// branch itself and for a detached worktree.
	Ahead int `json:"ahead"`
	// LastCommitAt is the committer date of Head, zero for a prunable worktree.
	LastCommitAt time.Time `json:"lastCommitAt"`
	// TouchedAt is the newest modification time among the worktree's
	// uncommitted files; nil when it has none.
	TouchedAt *time.Time `json:"touchedAt,omitempty"`
}

// projectGitResponse is the git state of a project's repository relative to
// its base branch. CurrentBranch and BaseCheckedOut describe the primary
// checkout, which is where branches merge into.
type projectGitResponse struct {
	Base           string                `json:"base"`
	CurrentBranch  string                `json:"currentBranch"`
	BaseCheckedOut bool                  `json:"baseCheckedOut"`
	Worktrees      []projectGitWorktree  `json:"worktrees"`
	Branches       []gavelgit.BranchInfo `json:"branches"`
}

// projectGitSummary is one project's row of /api/projects/git-summary: the
// unmerged and uncommitted line counts plus the linked worktree count (the
// primary checkout excluded) and unmerged branch count. Error
// is set, with the numbers left at 0, when the project could not be inspected.
type projectGitSummary struct {
	Name      string `json:"name"`
	Base      string `json:"base"`
	Adds      int    `json:"adds"`
	Dels      int    `json:"dels"`
	Worktrees int    `json:"worktrees"`
	Branches  int    `json:"branches"`
	Error     string `json:"error,omitempty"`
}

// projectGitBase is the local branch project work merges into: the PR base
// (.gavel.yaml pr.base, else origin/main) without its remote prefix.
func projectGitBase(dir string) (string, error) {
	ref, err := prcreate.ResolveBase(dir, "")
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(strings.TrimPrefix(ref, "refs/heads/"), "origin/"), nil
}

func projectGitState(ctx context.Context, dir string) (projectGitResponse, error) {
	base, err := projectGitBase(dir)
	if err != nil {
		return projectGitResponse{}, err
	}
	worktrees, err := gavelgit.ListWorktrees(dir)
	if err != nil {
		return projectGitResponse{}, err
	}
	branches, err := gavelgit.UnmergedBranches(dir, base)
	if err != nil {
		return projectGitResponse{}, err
	}
	ahead := make(map[string]int, len(branches))
	for _, branch := range branches {
		ahead[branch.Name] = branch.Ahead
	}
	response := projectGitResponse{
		Base: base, CurrentBranch: worktrees[0].Branch, BaseCheckedOut: worktrees[0].Branch == base,
		Worktrees: make([]projectGitWorktree, 0, len(worktrees)), Branches: append([]gavelgit.BranchInfo{}, branches...),
	}
	for _, wt := range worktrees {
		view := projectGitWorktree{Worktree: wt, Ahead: ahead[wt.Branch]}
		if !wt.Prunable {
			if view.Changes, view.TouchedAt, err = projectWorktreeChanges(ctx, wt.Path); err != nil {
				return projectGitResponse{}, err
			}
			if view.LastCommitAt, err = gavelgit.CommitTime(wt.Path, wt.Head); err != nil {
				return projectGitResponse{}, err
			}
		}
		response.Worktrees = append(response.Worktrees, view)
	}
	return response, nil
}

func projectWorktreeChanges(ctx context.Context, path string) (projectGitChanges, *time.Time, error) {
	result, err := gatherProjectStatus(path, status.Options{NoRepomap: true, NoResults: true, Context: ctx})
	if err != nil {
		return projectGitChanges{}, nil, fmt.Errorf("gather status of worktree %s: %w", path, err)
	}
	touched, err := lastTouched(path, result.Files)
	if err != nil {
		return projectGitChanges{}, nil, err
	}
	counts := result.Counts()
	return projectGitChanges{
		Staged: counts.Staged, Unstaged: counts.Unstaged, Both: counts.Both, Untracked: counts.Untracked,
		Conflict: counts.Conflict, Adds: counts.Adds, Dels: counts.Dels,
	}, touched, nil
}

// lastTouched is the newest mtime among the changed files of a worktree, in
// UTC; nil when no changed file exists on disk (none changed, or all deleted).
func lastTouched(dir string, files []status.FileStatus) (*time.Time, error) {
	var newest *time.Time
	for _, file := range files {
		info, err := os.Stat(filepath.Join(dir, file.Path))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("stat changed file %s in worktree %s: %w", file.Path, dir, err)
		}
		if at := info.ModTime().UTC(); newest == nil || at.After(*newest) {
			newest = &at
		}
	}
	return newest, nil
}

func (state projectGitResponse) summary(name string) projectGitSummary {
	summary := projectGitSummary{Name: name, Base: state.Base, Branches: len(state.Branches)}
	for _, wt := range state.Worktrees {
		if !wt.Primary {
			summary.Worktrees++
		}
	}
	for _, branch := range state.Branches {
		summary.Adds += branch.Diff.Adds
		summary.Dels += branch.Diff.Dels
	}
	for _, wt := range state.Worktrees {
		summary.Adds += wt.Changes.Adds
		summary.Dels += wt.Changes.Dels
	}
	return summary
}

func (s *Server) handleProjectsGitSummary(w http.ResponseWriter, r *http.Request) {
	projects, err := LoadProjects()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	summaries := make([]projectGitSummary, len(projects))
	var group errgroup.Group
	group.SetLimit(projectGitSummaryConcurrency)
	for i, project := range projects {
		group.Go(func() error {
			// A project's failure is reported on its own row, never the batch.
			state, err := projectGitState(r.Context(), project.ResolvedDir())
			if err != nil {
				summaries[i] = projectGitSummary{Name: project.Name, Error: err.Error()}
				return nil
			}
			summaries[i] = state.summary(project.Name)
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, summaries)
}

func (s *Server) handleProjectGit(w http.ResponseWriter, r *http.Request) {
	project, err := GetProject(r.PathValue("name"))
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	state, err := projectGitState(r.Context(), project.ResolvedDir())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, state)
}

// projectWorkDir is the directory a project request operates in: the
// project's own directory, or the linked worktree the request names. A path
// that is not one of the repository's live worktrees is errUnknownWorktree.
func projectWorkDir(project Project, worktree string) (string, error) {
	dir := project.ResolvedDir()
	if strings.TrimSpace(worktree) == "" {
		return dir, nil
	}
	requested, err := filepath.EvalSymlinks(filepath.Clean(worktree))
	if err != nil {
		return "", fmt.Errorf("%w: worktree %q: %v", errUnknownWorktree, worktree, err)
	}
	worktrees, err := gavelgit.ListWorktrees(dir)
	if err != nil {
		return "", err
	}
	for _, wt := range worktrees {
		if wt.Prunable {
			continue
		}
		resolved, err := filepath.EvalSymlinks(wt.Path)
		if err != nil {
			return "", fmt.Errorf("resolve worktree %s of project %s: %w", wt.Path, project.Name, err)
		}
		if resolved != requested {
			continue
		}
		if wt.Primary {
			return dir, nil
		}
		return wt.Path, nil
	}
	return "", fmt.Errorf("%w: %q is not a worktree of project %s", errUnknownWorktree, worktree, project.Name)
}

// requestWorkDir resolves the ?worktree= of a project request, answering 400
// for a path outside the project's worktrees.
func requestWorkDir(w http.ResponseWriter, r *http.Request, project Project) (string, bool) {
	workDir, err := projectWorkDir(project, r.URL.Query().Get("worktree"))
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, errUnknownWorktree) {
			code = http.StatusBadRequest
		}
		respondError(w, code, err.Error())
		return "", false
	}
	return workDir, true
}
