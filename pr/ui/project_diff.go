package ui

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

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
	project, err := GetProject(r.PathValue("name"))
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	result, err := gatherProjectStatus(project.ResolvedDir(), status.Options{NoRepomap: true})
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("gather project status: %v", err))
		return
	}
	target, files, err := selectProjectDiffFiles(r.URL.Query().Get("path"), result.Files)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	diff, binary, err := projectWorkingTreeDiff(project.ResolvedDir(), files)
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

func projectWorkingTreeDiff(workDir string, files []status.FileStatus) (string, bool, error) {
	var patches []string
	allBinary := true
	for _, file := range files {
		filePatches, err := projectFilePatches(workDir, file)
		if err != nil {
			return "", false, err
		}
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
	return strings.Join(patches, "\n"), len(files) == 1 && len(patches) > 0 && allBinary, nil
}

func projectFilePatches(workDir string, file status.FileStatus) ([]string, error) {
	if file.State == status.StateUntracked {
		info, err := os.Lstat(filepath.Join(workDir, filepath.FromSlash(file.Path)))
		if err != nil {
			return nil, fmt.Errorf("inspect untracked file %q: %w", file.Path, err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("untracked diff target %q is not a regular file", file.Path)
		}
		patch, err := runProjectGitDiff(workDir, true, "--no-index", "--no-color", "--", "/dev/null", file.Path)
		if err != nil {
			return nil, fmt.Errorf("diff untracked file %q: %w", file.Path, err)
		}
		return []string{patch}, nil
	}

	paths := []string{"--"}
	if file.PreviousPath != "" {
		paths = append(paths, file.PreviousPath)
	}
	paths = append(paths, file.Path)
	stagedArgs := append([]string{"--cached", "--find-renames", "--no-color"}, paths...)
	staged, err := runProjectGitDiff(workDir, false, stagedArgs...)
	if err != nil {
		return nil, fmt.Errorf("diff staged file %q: %w", file.Path, err)
	}
	workArgs := append([]string{"--find-renames", "--no-color"}, paths...)
	unstaged, err := runProjectGitDiff(workDir, false, workArgs...)
	if err != nil {
		return nil, fmt.Errorf("diff unstaged file %q: %w", file.Path, err)
	}
	return []string{staged, unstaged}, nil
}

func runProjectGitDiff(workDir string, allowChanges bool, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"diff"}, args...)...)
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
