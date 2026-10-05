package branchmerge

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// RangeCommits lists the commits of from..to in repo, oldest first, as full
// shas.
func RangeCommits(repo, from, to string) ([]string, error) {
	out, err := captureGit(repo, "rev-list", "--reverse", from+".."+to)
	if err != nil {
		return nil, err
	}
	return strings.Fields(out), nil
}

func branchExists(repo, branch string) bool {
	return exec.Command("git", "-C", repo, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch).Run() == nil
}

// captureGit returns git's trimmed stdout, or an error carrying its stderr.
func captureGit(dir string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("git %s in %s: %s", strings.Join(args, " "), dir, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("git %s in %s: %w", strings.Join(args, " "), dir, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// combinedGit runs git and returns its trimmed combined output, for commands
// whose failure explanation goes to either stream.
func combinedGit(dir string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func short(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
