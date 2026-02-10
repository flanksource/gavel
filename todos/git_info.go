package todos

import (
	"os/exec"
	"strings"
)

// GetGitInfo retrieves the current git state for a working directory.
// Returns branch name, short commit SHA, and whether there are uncommitted changes.
func GetGitInfo(workDir string) (branch, commit string, dirty bool, err error) {
	branch, err = runGitCommand(workDir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", "", false, err
	}

	commit, err = runGitCommand(workDir, "rev-parse", "--short", "HEAD")
	if err != nil {
		return branch, "", false, err
	}

	status, err := runGitCommand(workDir, "status", "--porcelain")
	if err != nil {
		return branch, commit, false, err
	}
	dirty = len(strings.TrimSpace(status)) > 0

	return branch, commit, dirty, nil
}

func runGitCommand(workDir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
