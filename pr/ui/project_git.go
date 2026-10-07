package ui

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/git/gitstate"
	"golang.org/x/sync/errgroup"
)

// projectGitSummaryConcurrency bounds how many projects git-summary inspects
// at once; each inspection runs several git processes per worktree.
const projectGitSummaryConcurrency = 4

var errUnknownWorktree = errors.New("not a worktree of the project")

func (s *Server) handleProjectsGitSummary(w http.ResponseWriter, r *http.Request) {
	projects, err := LoadProjects()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	ctx := s.requestContext(r)
	summaries := make([]gitstate.Summary, len(projects))
	var group errgroup.Group
	group.SetLimit(projectGitSummaryConcurrency)
	for i, project := range projects {
		group.Go(func() error {
			// A project's failure is reported on its own row, never the batch.
			state, err := ctx.GitState().Get(ctx, project.ResolvedDir())
			if err != nil {
				summaries[i] = gitstate.Summary{Name: project.Name, Error: err.Error()}
				return nil
			}
			summaries[i] = state.Summary(project.Name)
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
	ctx := s.requestContext(r)
	state, err := ctx.GitState().Get(ctx, project.ResolvedDir())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, state)
}

// projectWorkDir is the directory a project request operates in: the
// project's own directory, or the linked worktree the request names. A path
// that is not one of the repository's live worktrees is errUnknownWorktree.
func projectWorkDir(ctx gavelctx.Context, project Project, worktree string) (string, error) {
	dir := project.ResolvedDir()
	if strings.TrimSpace(worktree) == "" {
		return dir, nil
	}
	requested, err := filepath.EvalSymlinks(filepath.Clean(worktree))
	if err != nil {
		return "", fmt.Errorf("%w: worktree %q: %v", errUnknownWorktree, worktree, err)
	}
	wt, found, err := ctx.GitState().WorktreeOf(ctx, dir, requested)
	if err != nil {
		return "", fmt.Errorf("worktrees of project %s: %w", project.Name, err)
	}
	if !found {
		return "", fmt.Errorf("%w: %q is not a worktree of project %s", errUnknownWorktree, worktree, project.Name)
	}
	if wt.Primary {
		return dir, nil
	}
	return wt.Path, nil
}

// requestWorkDir resolves the ?worktree= of a project request, answering 400
// for a path outside the project's worktrees.
func (s *Server) requestWorkDir(w http.ResponseWriter, r *http.Request, project Project) (string, bool) {
	workDir, err := projectWorkDir(s.requestContext(r), project, r.URL.Query().Get("worktree"))
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
