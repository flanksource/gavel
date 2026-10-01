package land

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/github/prcreate"
	"github.com/flanksource/gavel/todos/native"
)

// merge cherry-picks commits onto repo's checked-out branch. A failed pick is
// aborted, leaving the checkout where it started.
func merge(repo, runBranch string, commits []string) (native.RunLanding, error) {
	if err := prcreate.RefuseInProgressOps(repo); err != nil {
		return native.RunLanding{}, fmt.Errorf("%w: %w", ErrCheckoutNotReady, err)
	}
	staged, err := captureGit(repo, "diff", "--cached", "--name-only")
	if err != nil {
		return native.RunLanding{}, err
	}
	if staged != "" {
		return native.RunLanding{}, fmt.Errorf("%w: %s has staged changes (%s); commit or unstage them first",
			ErrCheckoutNotReady, repo, strings.Join(strings.Fields(staged), ", "))
	}
	branch, err := captureGit(repo, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return native.RunLanding{}, fmt.Errorf("%w: %s has a detached HEAD; check out the branch to land onto", ErrCheckoutNotReady, repo)
	}
	if branch == runBranch {
		return native.RunLanding{}, fmt.Errorf("%w: %s has the run's own branch %s checked out", ErrCheckoutNotReady, repo, branch)
	}
	before, err := captureGit(repo, "rev-parse", "HEAD")
	if err != nil {
		return native.RunLanding{}, err
	}
	if out, err := combinedGit(repo, append([]string{"cherry-pick"}, commits...)...); err != nil {
		return native.RunLanding{}, abortCherryPick(repo, branch, before, fmt.Errorf("git cherry-pick: %s: %w", out, err))
	}
	head, err := captureGit(repo, "rev-parse", "HEAD")
	if err != nil {
		return native.RunLanding{}, err
	}
	return native.RunLanding{Via: native.LandingMerge, TargetBranch: branch, LandedSHA: head}, nil
}

// abortCherryPick undoes a failed cherry-pick and reports why it failed: a
// *ConflictError naming the unmerged paths, or the pick's own error.
func abortCherryPick(repo, branch, before string, cause error) error {
	unmerged, err := captureGit(repo, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return errors.Join(cause, err)
	}
	inProgress, err := cherryPickInProgress(repo)
	if err != nil {
		return errors.Join(cause, err)
	}
	if inProgress {
		if out, err := combinedGit(repo, "cherry-pick", "--abort"); err != nil {
			return errors.Join(cause, fmt.Errorf("git cherry-pick --abort in %s: %s: %w", repo, out, err))
		}
	}
	after, err := captureGit(repo, "rev-parse", "HEAD")
	if err != nil {
		return errors.Join(cause, err)
	}
	if after != before {
		return errors.Join(cause, fmt.Errorf("aborting the cherry-pick left %s at %s, not %s", branch, short(after), short(before)))
	}
	if paths := strings.Fields(unmerged); len(paths) > 0 {
		return &ConflictError{Branch: branch, Paths: paths, Err: cause}
	}
	return fmt.Errorf("cherry-pick the run's commits onto %s: %w", branch, cause)
}

func cherryPickInProgress(repo string) (bool, error) {
	for _, name := range []string{"CHERRY_PICK_HEAD", "sequencer"} {
		path, err := captureGit(repo, "rev-parse", "--path-format=absolute", "--git-path", name)
		if err != nil {
			return false, err
		}
		if _, err := os.Stat(path); err == nil {
			return true, nil
		} else if !os.IsNotExist(err) {
			return false, err
		}
	}
	return false, nil
}

// cleanup removes the run's worktree, if it is still there, and deletes its
// branch once every commit on it is confirmed to be in landing.LandedSHA.
func cleanup(wt *api.WorktreeState, landing *native.RunLanding) error {
	if wt.Path != "" {
		if _, err := os.Stat(wt.Path); err == nil {
			if out, err := combinedGit(wt.Repo, "worktree", "remove", filepath.Clean(wt.Path)); err != nil {
				return fmt.Errorf("git worktree remove %s: %s: %w", wt.Path, out, err)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("inspect worktree %s: %w", wt.Path, err)
		}
	}
	if !branchExists(wt.Repo, wt.Branch) {
		return nil
	}
	cherry, err := captureGit(wt.Repo, "cherry", landing.LandedSHA, "refs/heads/"+wt.Branch, wt.Setup)
	if err != nil {
		return fmt.Errorf("check that %s landed: %w", wt.Branch, err)
	}
	var unlanded []string
	for _, line := range strings.Split(cherry, "\n") {
		if sha, ok := strings.CutPrefix(line, "+ "); ok {
			unlanded = append(unlanded, short(sha))
		}
	}
	if len(unlanded) > 0 {
		return fmt.Errorf("kept branch %s: %s not in %s", wt.Branch, strings.Join(unlanded, ", "), short(landing.LandedSHA))
	}
	if out, err := combinedGit(wt.Repo, "branch", "-D", wt.Branch); err != nil {
		return fmt.Errorf("git branch -D %s: %s: %w", wt.Branch, out, err)
	}
	deletedAt := time.Now().UTC()
	landing.BranchDeletedAt = &deletedAt
	return nil
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
