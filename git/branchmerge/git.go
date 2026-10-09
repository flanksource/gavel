package branchmerge

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	gavelgit "github.com/flanksource/gavel/git"
)

// RangeCommits lists the commits of from..to in repo, oldest first, as full
// shas.
func RangeCommits(repo, from, to string) ([]string, error) {
	if err := gavelgit.ValidateRevision(from); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidOptions, err)
	}
	if err := gavelgit.ValidateRevision(to); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidOptions, err)
	}
	out, err := captureGit(repo, "rev-list", "--reverse", "--end-of-options", from+".."+to)
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

// commitStaged commits the index with message read from stdin, so free-form
// text never becomes a command-line argument. Hooks are skipped, as a
// cherry-pick skips them.
func commitStaged(dir, message string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "commit", "--no-verify", "-q", "--file=-")
	cmd.Stdin = strings.NewReader(message)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func short(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
