package ui

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/commons/logger"
	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/git/gitstate"
	"github.com/flanksource/gavel/status"
)

var errUnknownWorktree = errors.New("not a worktree of the project")

// handleProjectsGitSummary answers every project's row from the stored git
// state; no git runs once each repository has been tracked.
func (s *Server) handleProjectsGitSummary(w http.ResponseWriter, r *http.Request) {
	projects, err := LoadProjects()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	tracker, ok := s.requestGitTracker(w)
	if !ok {
		return
	}
	dirs := make([]string, 0, len(projects))
	for _, project := range projects {
		dirs = append(dirs, project.ResolvedDir())
	}
	states, errs := tracker.States(s.requestContext(r), dirs)
	summaries := make([]gitstate.Summary, 0, len(projects))
	for _, project := range projects {
		// A project's failure is reported on its own row, never the batch.
		if err := errs[project.ResolvedDir()]; err != nil {
			summaries = append(summaries, gitstate.Summary{Name: project.Name, Error: err.Error()})
			continue
		}
		summaries = append(summaries, states[project.ResolvedDir()].Summary(project.Name))
	}
	respondJSON(w, http.StatusOK, summaries)
}

func (s *Server) handleProjectGit(w http.ResponseWriter, r *http.Request) {
	project, err := GetProject(r.PathValue("name"))
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	tracker, ok := s.requestGitTracker(w)
	if !ok {
		return
	}
	state, err := tracker.State(s.requestContext(r), project.ResolvedDir())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, state)
}

// gitChangeCounts returns, by project directory, the number of uncommitted
// files (staged, unstaged, untracked and conflicted) in the primary checkout,
// read from the stored git state in one States call so the proc-status cadence
// never runs git itself. A directory whose git state cannot be read (not a git
// work tree, or no tracker in this process) has no entry: the proc-status wire
// field is omitted, which reads as "no git info" rather than zero changes.
func gitChangeCounts(ctx gavelctx.Context, projects []Project) map[string]int {
	tracker, err := ctx.GitTracker()
	if err != nil {
		logger.Debugf("proc status git changes: %v", err)
		return map[string]int{}
	}
	dirs := make([]string, 0, len(projects))
	for _, p := range projects {
		if dir := p.ResolvedDir(); dir != "" {
			dirs = append(dirs, dir)
		}
	}
	states, errs := tracker.States(ctx, dirs)
	for dir, err := range errs {
		logger.Debugf("git state %s: %v", dir, err)
	}
	counts := make(map[string]int, len(states))
	for dir, state := range states {
		primary, ok := state.Primary()
		if !ok {
			logger.Debugf("git state of %s lists no primary checkout", dir)
			continue
		}
		counts[dir] = primary.Changes.Files()
	}
	return counts
}

// projectWorkDir is the directory a project request operates in: the
// project's own directory, or the linked worktree the request names, looked
// up in the stored git state. A path that is not one of the repository's live
// worktrees is errUnknownWorktree.
func projectWorkDir(ctx gavelctx.Context, project Project, worktree string) (string, error) {
	dir := project.ResolvedDir()
	if strings.TrimSpace(worktree) == "" {
		return dir, nil
	}
	tracker, err := ctx.GitTracker()
	if err != nil {
		return "", err
	}
	requested, err := filepath.EvalSymlinks(filepath.Clean(worktree))
	if err != nil {
		return "", fmt.Errorf("%w: worktree %q: %v", errUnknownWorktree, worktree, err)
	}
	wt, found, err := tracker.WorktreeOf(ctx, dir, requested)
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

// workDirErrorStatus maps a projectWorkDir failure: 400 for a path outside the
// project's worktrees, 503 without a git state tracker, else 500.
func workDirErrorStatus(err error) int {
	if errors.Is(err, errUnknownWorktree) {
		return http.StatusBadRequest
	}
	return gitErrorStatus(err)
}

// requestWorkDir resolves the ?worktree= of a project request, answering 400
// for a path outside the project's worktrees.
func (s *Server) requestWorkDir(w http.ResponseWriter, r *http.Request, project Project) (string, bool) {
	workDir, err := projectWorkDir(s.requestContext(r), project, r.URL.Query().Get("worktree"))
	if err != nil {
		respondError(w, workDirErrorStatus(err), err.Error())
		return "", false
	}
	return workDir, true
}

// storedWorktreeFiles is the worktree holding workDir, one of project's, and
// the uncommitted files its last status scan recorded.
func storedWorktreeFiles(ctx gavelctx.Context, project Project, workDir string) (gitstate.Worktree, []status.FileStatus, error) {
	tracker, err := ctx.GitTracker()
	if err != nil {
		return gitstate.Worktree{}, nil, err
	}
	state, err := tracker.State(ctx, project.ResolvedDir())
	if err != nil {
		return gitstate.Worktree{}, nil, err
	}
	wt, err := containingWorktree(state, workDir)
	if err != nil {
		return gitstate.Worktree{}, nil, err
	}
	files, err := tracker.Files(ctx, project.ResolvedDir(), wt.Path)
	if err != nil {
		return gitstate.Worktree{}, nil, err
	}
	return wt, files, nil
}

// containingWorktree is the live worktree of state holding dir, the deepest
// when worktrees nest: the worktree whose stored status covers a project
// directory that may be a subdirectory of its checkout. It reads only the
// filesystem, never git.
func containingWorktree(state gitstate.State, dir string) (gitstate.Worktree, error) {
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return gitstate.Worktree{}, fmt.Errorf("resolve %s: %w", dir, err)
	}
	var (
		best  gitstate.Worktree
		depth = -1
	)
	for _, wt := range state.Worktrees {
		if wt.Prunable {
			continue
		}
		root, err := filepath.EvalSymlinks(wt.Path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return gitstate.Worktree{}, fmt.Errorf("resolve worktree %s: %w", wt.Path, err)
		}
		within := resolved == root || strings.HasPrefix(resolved, root+string(filepath.Separator))
		if within && len(root) > depth {
			best, depth = wt, len(root)
		}
	}
	if depth < 0 {
		return gitstate.Worktree{}, fmt.Errorf("%s is in none of the recorded worktrees of its repository", dir)
	}
	return best, nil
}
