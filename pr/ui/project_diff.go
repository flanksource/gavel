package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	rpchttp "github.com/flanksource/clicky/rpc/http"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/status"
)

type projectDiffResponse struct {
	Path string `json:"path"`
	Diff string `json:"diff"`
	// Staged (HEAD→index) and Unstaged (index→worktree, or the /dev/null diff
	// of an untracked file) are set only when the target is a single file, so
	// the UI can stage individual lines against them.
	Staged    string `json:"staged"`
	Unstaged  string `json:"unstaged"`
	Truncated bool   `json:"truncated"`
	Binary    bool   `json:"binary"`
}

func (s *Server) handleProjectDiff(w http.ResponseWriter, r *http.Request) {
	stopFile := rpchttp.Track(r.Context(), "file")
	project, err := getProject(r.Context(), r.PathValue("name"))
	stopFile()
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	workDir, ok := s.requestWorkDir(w, r, project)
	if !ok {
		return
	}
	// The files come from the stored status scan; the patch below stays live.
	_, changed, err := storedWorktreeFiles(s.requestContext(r), project, workDir)
	if err != nil {
		respondError(w, gitErrorStatus(err), fmt.Sprintf("project changes: %v", err))
		return
	}
	target, files, err := selectProjectDiffFiles(r.URL.Query().Get("path"), changed)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	working, err := projectWorkingTreeDiff(r.Context(), workDir, files)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response := projectDiffResponse{Path: target, Binary: working.binary}
	var truncated [3]bool
	response.Diff, truncated[0] = gavelgit.TruncateDiff(working.diff)
	if len(files) == 1 && files[0].Path == target {
		response.Staged, truncated[1] = gavelgit.TruncateDiff(working.staged)
		response.Unstaged, truncated[2] = gavelgit.TruncateDiff(working.unstaged)
	}
	response.Truncated = truncated[0] || truncated[1] || truncated[2]
	respondJSON(w, http.StatusOK, response)
}

// projectWorkingDiff is the combined patch of the selected entries plus, for a
// single file, its staged and unstaged halves.
type projectWorkingDiff struct {
	diff, staged, unstaged string
	binary                 bool
}

func selectProjectDiffFiles(target string, files []status.FileStatus) (string, []status.FileStatus, error) {
	target = filepath.ToSlash(filepath.Clean(strings.TrimSpace(target)))
	if target == "" || target == "." || filepath.IsAbs(target) || target == ".." || strings.HasPrefix(target, "../") {
		return "", nil, fmt.Errorf("invalid project diff path %q", target)
	}
	selected := make([]status.FileStatus, 0)
	for _, file := range files {
		path := filepath.ToSlash(file.Path)
		if path == target || strings.HasPrefix(path, target+"/") {
			selected = append(selected, file)
		}
	}
	if len(selected) == 0 {
		return "", nil, fmt.Errorf("%q is not a current project change", target)
	}
	sort.Slice(selected, func(left, right int) bool { return selected[left].Path < selected[right].Path })
	return target, selected, nil
}

func projectWorkingTreeDiff(ctx context.Context, workDir string, files []status.FileStatus) (projectWorkingDiff, error) {
	var (
		patches []string
		working projectWorkingDiff
	)
	covered := 0
	allBinary := true
	for _, file := range files {
		filePatches, fileCovered, err := projectFilePatches(ctx, workDir, file)
		if err != nil {
			return projectWorkingDiff{}, err
		}
		covered += fileCovered
		if len(files) == 1 && fileCovered == 1 {
			working.staged, working.unstaged = splitFilePatches(file, filePatches)
		}
		for _, patch := range filePatches {
			if strings.TrimSpace(patch) == "" {
				continue
			}
			patches = append(patches, patch)
			if !gavelgit.IsBinaryDiff(patch) {
				allBinary = false
			}
		}
	}
	working.diff = strings.Join(patches, "\n")
	working.binary = covered == 1 && len(patches) > 0 && allBinary
	return working, nil
}

// splitFilePatches names the patches projectFilePatches rendered for one
// file: a tracked file's staged and unstaged diffs, or an untracked file's
// creation diff, which is unstaged.
func splitFilePatches(file status.FileStatus, patches []string) (staged, unstaged string) {
	if file.State == status.StateUntracked {
		return "", patches[0]
	}
	return patches[0], patches[1]
}

// projectFilePatches renders one status entry as unified-diff patches and
// reports how many working-tree files those patches covered — git status
// collapses a wholly untracked directory into a single entry.
func projectFilePatches(ctx context.Context, workDir string, file status.FileStatus) (patches []string, covered int, err error) {
	if file.State == status.StateUntracked {
		targets, err := untrackedDiffTargets(ctx, workDir, file.Path)
		if err != nil {
			return nil, 0, err
		}
		for _, target := range targets {
			patch, err := runProjectGitDiff(ctx, workDir, true, "--no-index", "--no-color", "--", "/dev/null", target)
			if err != nil {
				return nil, 0, fmt.Errorf("diff untracked file %q: %w", target, err)
			}
			patches = append(patches, patch)
		}
		return patches, len(targets), nil
	}

	paths := []string{"--"}
	if file.PreviousPath != "" {
		paths = append(paths, file.PreviousPath)
	}
	paths = append(paths, file.Path)
	stagedArgs := append([]string{"--cached", "--find-renames", "--no-color"}, paths...)
	staged, err := runProjectGitDiff(ctx, workDir, false, stagedArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("diff staged file %q: %w", file.Path, err)
	}
	workArgs := append([]string{"--find-renames", "--no-color"}, paths...)
	unstaged, err := runProjectGitDiff(ctx, workDir, false, workArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("diff unstaged file %q: %w", file.Path, err)
	}
	return []string{staged, unstaged}, 1, nil
}

// untrackedDiffTargets resolves an untracked status entry to the working-tree
// files it stands for. A directory entry is expanded through git so nested
// .gitignore rules keep applying to what the diff shows.
func untrackedDiffTargets(ctx context.Context, workDir, path string) ([]string, error) {
	stopFile := rpchttp.Track(ctx, "file")
	info, err := os.Lstat(filepath.Join(workDir, filepath.FromSlash(path)))
	stopFile()
	if err != nil {
		return nil, fmt.Errorf("inspect untracked path %q: %w", path, err)
	}
	if info.Mode().IsRegular() {
		return []string{path}, nil
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("untracked diff target %q is not a regular file", path)
	}

	stopGit := rpchttp.Track(ctx, "git")
	cmd := exec.CommandContext(ctx, "git", "ls-files", "--others", "--exclude-standard", "-z", "--", path)
	cmd.Dir = workDir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	stopGit()
	if err != nil {
		return nil, fmt.Errorf("list untracked files under %q: %w\nOutput: %s", path, err, strings.TrimSpace(stderr.String()))
	}
	var targets []string
	for _, target := range strings.Split(string(output), "\x00") {
		if target != "" {
			targets = append(targets, target)
		}
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("untracked directory %q holds no diffable files", path)
	}
	return targets, nil
}

func runProjectGitDiff(ctx context.Context, workDir string, allowChanges bool, args ...string) (string, error) {
	stopGit := rpchttp.Track(ctx, "git")
	defer stopGit()
	cmd := exec.CommandContext(ctx, "git", append([]string{"diff"}, args...)...)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err == nil {
		return string(output), nil
	}
	var exitError *exec.ExitError
	if allowChanges && errors.As(err, &exitError) && exitError.ExitCode() == 1 {
		return string(output), nil
	}
	return "", fmt.Errorf("git diff: %w\nOutput: %s", err, string(output))
}
