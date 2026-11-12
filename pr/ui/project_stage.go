package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	rpchttp "github.com/flanksource/clicky/rpc/http"
	"github.com/flanksource/gavel/git/patchselect"
	"github.com/flanksource/gavel/status"
)

type projectStageRequest struct {
	Paths   []string `json:"paths"`
	Unstage bool     `json:"unstage,omitempty"`
}

type projectStageLinesRequest struct {
	Path    string             `json:"path"`
	Unstage bool               `json:"unstage,omitempty"`
	Lines   []patchselect.Line `json:"lines"`
}

type projectDiscardRequest struct {
	Paths []string `json:"paths"`
}

type projectOKResponse struct {
	OK bool `json:"ok"`
}

// projectIndexTarget is a validated index mutation: the repository root git
// runs in and the status entries the request names, read live.
type projectIndexTarget struct {
	workDir string
	root    string
	entries []status.FileStatus
}

// handleProjectStage stages (git add) or unstages (git restore --staged) whole
// status entries.
func (s *Server) handleProjectStage(w http.ResponseWriter, r *http.Request) {
	var request projectStageRequest
	target, ok := s.decodeIndexRequest(w, r, &request, func() []string { return request.Paths })
	if !ok {
		return
	}
	err := stageEntries(r.Context(), target, request.Unstage)
	s.respondIndexMutation(w, target, err)
}

// handleProjectStageLines stages or unstages individual changed lines of one
// file by applying a line-filtered patch to the index.
func (s *Server) handleProjectStageLines(w http.ResponseWriter, r *http.Request) {
	var request projectStageLinesRequest
	target, ok := s.decodeIndexRequest(w, r, &request, func() []string { return []string{request.Path} })
	if !ok {
		return
	}
	err := stageLines(r.Context(), target, request)
	s.respondIndexMutation(w, target, err)
}

// handleProjectDiscard drops the changes of status entries: tracked paths are
// restored to HEAD in the index and worktree, untracked ones are removed.
func (s *Server) handleProjectDiscard(w http.ResponseWriter, r *http.Request) {
	var request projectDiscardRequest
	target, ok := s.decodeIndexRequest(w, r, &request, func() []string { return request.Paths })
	if !ok {
		return
	}
	err := discardEntries(r.Context(), target)
	s.respondIndexMutation(w, target, err)
}

func (s *Server) decodeIndexRequest(w http.ResponseWriter, r *http.Request, request any, paths func() []string) (projectIndexTarget, bool) {
	project, err := getProject(r.Context(), r.PathValue("name"))
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return projectIndexTarget{}, false
	}
	workDir, ok := s.requestWorkDir(w, r, project)
	if !ok {
		return projectIndexTarget{}, false
	}
	if err := decodeStrict(r, request); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return projectIndexTarget{}, false
	}
	target, err := resolveIndexTarget(r.Context(), workDir, paths())
	if err != nil {
		respondError(w, indexErrorStatus(err), err.Error())
		return projectIndexTarget{}, false
	}
	return target, true
}

// errIndexRequest marks a request the caller can fix: a bad path or selection.
var errIndexRequest = errors.New("invalid index request")

func indexRequestErrorf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", errIndexRequest, fmt.Sprintf(format, args...))
}

func indexErrorStatus(err error) int {
	if errors.Is(err, errIndexRequest) || errors.Is(err, patchselect.ErrBinary) || errors.Is(err, patchselect.ErrNoSelection) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func (s *Server) respondIndexMutation(w http.ResponseWriter, target projectIndexTarget, err error) {
	s.touchGit(target.workDir)
	s.notify()
	if err != nil {
		respondError(w, indexErrorStatus(err), err.Error())
		return
	}
	respondJSON(w, http.StatusOK, projectOKResponse{OK: true})
}

func resolveIndexTarget(ctx context.Context, workDir string, requested []string) (projectIndexTarget, error) {
	if len(requested) == 0 {
		return projectIndexTarget{}, indexRequestErrorf("at least one path is required")
	}
	// Read live, never memoized: a mutation must validate against the files as they are now.
	result, err := gatherProjectStatus(workDir, status.Options{Context: ctx, NoRepomap: true, NoResults: true})
	if err != nil {
		return projectIndexTarget{}, fmt.Errorf("gather project status: %w", err)
	}
	entries, err := resolveProjectChanges(requested, result.Files)
	if err != nil {
		return projectIndexTarget{}, err
	}
	root, err := runProjectGit(ctx, projectGitCommand{dir: workDir, args: []string{"rev-parse", "--show-toplevel"}})
	if err != nil {
		return projectIndexTarget{}, err
	}
	return projectIndexTarget{workDir: workDir, root: strings.TrimSpace(root), entries: entries}, nil
}

// resolveProjectChanges maps requested paths to the live status entries they
// name: an entry itself, every entry under a requested folder, or a file
// inside a wholly untracked directory entry.
func resolveProjectChanges(requested []string, files []status.FileStatus) ([]status.FileStatus, error) {
	var entries []status.FileStatus
	seen := map[string]bool{}
	add := func(file status.FileStatus) {
		if !seen[file.Path] {
			seen[file.Path] = true
			entries = append(entries, file)
		}
	}
	for _, raw := range requested {
		cleaned, err := normalizeProjectPath(strings.TrimSuffix(raw, "/"))
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errIndexRequest, err)
		}
		matches := matchProjectChanges(cleaned, files)
		if len(matches) == 0 {
			return nil, indexRequestErrorf("%q is not a current project change", raw)
		}
		for _, match := range matches {
			if match.State == status.StateConflict {
				return nil, indexRequestErrorf("%q has unresolved conflicts", match.Path)
			}
			add(match)
		}
	}
	return entries, nil
}

func matchProjectChanges(requested string, files []status.FileStatus) []status.FileStatus {
	var exact, nested []status.FileStatus
	for _, file := range files {
		switch {
		case file.Path == requested:
			exact = append(exact, file)
		case strings.HasPrefix(file.Path, requested+"/"):
			nested = append(nested, file)
		}
	}
	if len(exact) > 0 {
		return exact
	}
	return nested
}

// pathspecs are the git paths of entries, with each rename's previous path
// when includePrevious is set.
func pathspecs(entries []status.FileStatus, includePrevious bool) []string {
	var paths []string
	for _, entry := range entries {
		paths = append(paths, entry.Path)
		if includePrevious && entry.PreviousPath != "" {
			paths = append(paths, entry.PreviousPath)
		}
	}
	return paths
}

func stageEntries(ctx context.Context, target projectIndexTarget, unstage bool) error {
	if unstage {
		for _, entry := range target.entries {
			if entry.State == status.StateUntracked {
				return indexRequestErrorf("%q is untracked, not staged", entry.Path)
			}
		}
		return target.git(ctx, append([]string{"restore", "--staged", "--"}, pathspecs(target.entries, true)...)...)
	}
	if err := target.git(ctx, append([]string{"add", "--"}, pathspecs(target.entries, false)...)...); err != nil {
		return err
	}
	var renamedFrom []string
	for _, entry := range target.entries {
		if entry.PreviousPath != "" && (entry.StagedKind == status.KindRenamed || entry.WorkKind == status.KindRenamed) {
			renamedFrom = append(renamedFrom, entry.PreviousPath)
		}
	}
	if len(renamedFrom) == 0 {
		return nil
	}
	// The previous path of a rename is gone from the worktree, so `git add`
	// cannot name it; removing it from the index stages that half of the rename.
	return target.git(ctx, append([]string{"rm", "--cached", "--quiet", "--ignore-unmatch", "--"}, renamedFrom...)...)
}

func discardEntries(ctx context.Context, target projectIndexTarget) error {
	var tracked, untracked []status.FileStatus
	for _, entry := range target.entries {
		if entry.State == status.StateUntracked {
			untracked = append(untracked, entry)
		} else {
			tracked = append(tracked, entry)
		}
	}
	if len(tracked) > 0 {
		args := append([]string{"restore", "--source=HEAD", "--staged", "--worktree", "--"}, pathspecs(tracked, true)...)
		if err := target.git(ctx, args...); err != nil {
			return err
		}
	}
	if len(untracked) == 0 {
		return nil
	}
	// git clean keeps ignored files and never leaves the repository.
	if err := target.git(ctx, append([]string{"clean", "--force", "-d", "--quiet", "--"}, pathspecs(untracked, false)...)...); err != nil {
		return err
	}
	return removeEmptyParents(target.workDir, pathspecs(untracked, false))
}

// removeEmptyParents deletes the folders a discard emptied, so discarding an
// untracked folder's files removes the folder too. os.Remove refuses a
// non-empty directory, which ends the walk.
func removeEmptyParents(workDir string, paths []string) error {
	for _, path := range paths {
		for dir := filepath.Dir(filepath.FromSlash(path)); dir != "."; dir = filepath.Dir(dir) {
			err := os.Remove(filepath.Join(workDir, dir))
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				if entries, readErr := os.ReadDir(filepath.Join(workDir, dir)); readErr == nil && len(entries) > 0 {
					break
				}
				return fmt.Errorf("remove emptied folder %s: %w", dir, err)
			}
		}
	}
	return nil
}

func stageLines(ctx context.Context, target projectIndexTarget, request projectStageLinesRequest) error {
	if len(target.entries) != 1 || target.entries[0].Path != request.Path {
		return indexRequestErrorf("%q must name a single changed file", request.Path)
	}
	entry := target.entries[0]
	if entry.State == status.StateUntracked {
		if request.Unstage {
			return indexRequestErrorf("%q is untracked, not staged", entry.Path)
		}
		if err := target.git(ctx, "add", "--intent-to-add", "--", entry.Path); err != nil {
			return err
		}
	}
	err := applySelectedLines(ctx, target, entry.Path, request)
	if err != nil && entry.State == status.StateUntracked {
		if resetErr := target.git(ctx, "reset", "--quiet", "--", entry.Path); resetErr != nil {
			return fmt.Errorf("%w; undo intent-to-add of %q: %v", err, entry.Path, resetErr)
		}
	}
	return err
}

func applySelectedLines(ctx context.Context, target projectIndexTarget, path string, request projectStageLinesRequest) error {
	diffArgs := []string{"diff", "--no-color", "--no-ext-diff", "-U3", "--", path}
	applyArgs := []string{"apply", "--cached", "-"}
	if request.Unstage {
		diffArgs = []string{"diff", "--cached", "--no-color", "--no-ext-diff", "-U3", "--", path}
		applyArgs = []string{"apply", "--cached", "--reverse", "-"}
	}
	diff, err := target.output(ctx, "", diffArgs...)
	if err != nil {
		return err
	}
	patch, err := patchselect.Filter(diff, patchselect.Options{Lines: request.Lines, Reverse: request.Unstage})
	if err != nil {
		return fmt.Errorf("%w: %q: %w", errIndexRequest, path, err)
	}
	_, err = target.output(ctx, patch, applyArgs...)
	return err
}

func (t projectIndexTarget) git(ctx context.Context, args ...string) error {
	_, err := t.output(ctx, "", args...)
	return err
}

// output runs git in the repository root with literal pathspecs, so a path
// holding '*' or ':' names only itself.
func (t projectIndexTarget) output(ctx context.Context, stdin string, args ...string) (string, error) {
	return runProjectGit(ctx, projectGitCommand{dir: t.root, args: append([]string{"--literal-pathspecs"}, args...), stdin: stdin})
}

type projectGitCommand struct {
	dir   string
	args  []string
	stdin string
}

func runProjectGit(ctx context.Context, command projectGitCommand) (string, error) {
	stopGit := rpchttp.Track(ctx, "git")
	defer stopGit()
	cmd := exec.CommandContext(ctx, "git", command.args...)
	cmd.Dir = command.dir
	if command.stdin != "" {
		cmd.Stdin = strings.NewReader(command.stdin)
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(command.args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return string(output), nil
}
