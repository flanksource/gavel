package commit

import (
	"fmt"
	"os/exec"
	"strings"
)

func stagedDiff(workDir string) (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	cmd.Dir = workDir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff --cached: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}

func stagedFiles(workDir string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	cmd.Dir = workDir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --cached --name-only: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func addFiles(workDir string, files []string) error {
	if len(files) == 0 {
		return nil
	}
	args := append([]string{"add", "--"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func commitWithMessage(workDir, msg string) (string, error) {
	cmd := exec.Command("git", "commit", "-m", msg)
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git commit: %w: %s", err, strings.TrimSpace(string(out)))
	}
	hashCmd := exec.Command("git", "rev-parse", "HEAD")
	hashCmd.Dir = workDir
	out, err := hashCmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
