package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitIgnoredPaths returns the supplied paths that Git considers ignored.
// Native Git is the source of truth so repository, info, global, and XDG
// excludes all use Git's own matching semantics.
func GitIgnoredPaths(paths []string, workDir string) (map[string]struct{}, error) {
	ignored := make(map[string]struct{})
	if len(paths) == 0 {
		return ignored, nil
	}

	gitRoot, err := GitTopLevel(workDir)
	if err != nil || gitRoot == "" {
		return ignored, err
	}
	absoluteWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve work directory %q: %w", workDir, err)
	}
	relToOriginal := make(map[string][]string, len(paths))
	relPaths := make([]string, 0, len(paths))
	for _, path := range paths {
		absolute := path
		if !filepath.IsAbs(absolute) {
			absolute = filepath.Join(absoluteWorkDir, absolute)
		}
		rel, err := filepath.Rel(absoluteWorkDir, absolute)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		rel = filepath.ToSlash(rel)
		if _, exists := relToOriginal[rel]; !exists {
			relPaths = append(relPaths, rel)
		}
		relToOriginal[rel] = append(relToOriginal[rel], path)
	}
	if len(relPaths) == 0 {
		return ignored, nil
	}

	cmd := exec.Command("git", "check-ignore", "--no-index", "-z", "--stdin")
	cmd.Dir = absoluteWorkDir
	cmd.Stdin = strings.NewReader(strings.Join(relPaths, "\x00"))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 1 {
			return ignored, nil
		}
		return nil, fmt.Errorf("git check-ignore in %q: %w: %s", absoluteWorkDir, err, strings.TrimSpace(stderr.String()))
	}

	for _, rel := range strings.Split(stdout.String(), "\x00") {
		for _, original := range relToOriginal[rel] {
			ignored[original] = struct{}{}
		}
	}
	return ignored, nil
}

// GitTopLevel resolves the native Git worktree root. Directories outside a
// worktree return an empty root because linting non-Git source trees is valid.
func GitTopLevel(workDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = workDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 128 && strings.Contains(stderr.String(), "not a git repository") {
			return "", nil
		}
		return "", fmt.Errorf("resolve git root for %q: %w: %s", workDir, err, strings.TrimSpace(stderr.String()))
	}
	root := strings.TrimSpace(string(output))
	if root == "" {
		return "", fmt.Errorf("resolve git root for %q: empty output", workDir)
	}
	return root, nil
}
