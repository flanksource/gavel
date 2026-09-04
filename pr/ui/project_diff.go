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
	Path      string `json:"path"`
	Diff      string `json:"diff"`
	Truncated bool   `json:"truncated"`
	Binary    bool   `json:"binary"`
}

func (s *Server) handleProjectDiff(w http.ResponseWriter, r *http.Request) {
	stopFile := rpchttp.Track(r.Context(), "file")
	project, err := GetProject(r.PathValue("name"))
	stopFile()
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	result, err := gatherProjectStatus(project.ResolvedDir(), status.Options{NoRepomap: true, NoResults: true, Context: r.Context()})
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("gather project status: %v", err))
		return
	}
	target, files, err := selectProjectDiffFiles(r.URL.Query().Get("path"), result.Files)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	diff, binary, err := projectWorkingTreeDiff(r.Context(), project.ResolvedDir(), files)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	diff, truncated := gavelgit.TruncateDiff(diff)
	respondJSON(w, http.StatusOK, projectDiffResponse{
		Path: target, Diff: diff, Truncated: truncated, Binary: binary,
	})
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

func projectWorkingTreeDiff(ctx context.Context, workDir string, files []status.FileStatus) (string, bool, error) {
	var patches []string
	covered := 0
	allBinary := true
	for _, file := range files {
		filePatches, fileCovered, err := projectFilePatches(ctx, workDir, file)
		if err != nil {
			return "", false, err
		}
		covered += fileCovered
		for _, patch := range filePatches {
			if strings.TrimSpace(patch) == "" {
				continue
			}
			patches = append(patches, patch)
			if !isBinaryDiff(patch) {
				allBinary = false
			}
		}
	}
	return strings.Join(patches, "\n"), covered == 1 && len(patches) > 0 && allBinary, nil
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

func isBinaryDiff(diff string) bool {
	return strings.Contains(diff, "Binary files ") || strings.Contains(diff, "GIT binary patch")
}
