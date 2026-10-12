package branchmerge

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// cherryPick picks opts.Commits onto the target. A failed pick is aborted.
func cherryPick(opts Options, target, before string) error {
	out, err := combinedGit(opts.Repo, append([]string{"cherry-pick", "--end-of-options"}, opts.Commits...)...)
	if err == nil {
		return nil
	}
	return abortMerge(opts, target, before, fmt.Errorf("git cherry-pick: %s: %w", out, err))
}

// squash stages the whole branch onto the target with `git merge --squash` and
// commits it as one commit, returning how many source commits it carried.
// Hooks are skipped, as a cherry-pick skips them.
func squash(opts Options, target, before string) (int, error) {
	source := "refs/heads/" + opts.Branch
	count, err := captureGit(opts.Repo, "rev-list", "--count", "HEAD.."+source)
	if err != nil {
		return 0, err
	}
	commits, err := strconv.Atoi(count)
	if err != nil {
		return 0, fmt.Errorf("git rev-list --count HEAD..%s in %s: unexpected output %q", source, opts.Repo, count)
	}
	if commits == 0 {
		return 0, fmt.Errorf("%w: %s has no commits that are not already on %s", ErrNothingToMerge, opts.Branch, target)
	}
	if out, err := combinedGit(opts.Repo, "merge", "--squash", "--no-commit", source); err != nil {
		return 0, abortMerge(opts, target, before, fmt.Errorf("git merge --squash: %s: %w", out, err))
	}
	if _, err := captureGit(opts.Repo, "diff", "--cached", "--quiet"); err == nil {
		return 0, abortMerge(opts, target, before,
			fmt.Errorf("%w: squashing %s onto %s staged no changes", ErrNothingToMerge, opts.Branch, target))
	}
	if out, err := commitStaged(opts.Repo, opts.Message); err != nil {
		return 0, abortMerge(opts, target, before, fmt.Errorf("git commit: %s: %w", out, err))
	}
	return commits, nil
}

// abortMerge rolls a failed cherry-pick or squash back and reports why it
// failed: a *ConflictError naming the unmerged paths, or the cause itself.
func abortMerge(opts Options, target, before string, cause error) error {
	unmerged, err := captureGit(opts.Repo, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return errors.Join(cause, err)
	}
	if err := rollback(opts, before); err != nil {
		return errors.Join(cause, err)
	}
	after, err := captureGit(opts.Repo, "rev-parse", "HEAD")
	if err != nil {
		return errors.Join(cause, err)
	}
	if after != before {
		return errors.Join(cause, fmt.Errorf("aborting the merge left %s at %s, not %s", target, short(after), short(before)))
	}
	if paths := strings.Fields(unmerged); len(paths) > 0 {
		return &ConflictError{Branch: target, Source: opts.Branch, Mode: opts.Mode, Paths: paths, Err: cause}
	}
	return fmt.Errorf("%s merge of %s onto %s: %w", opts.Mode, opts.Branch, target, cause)
}

func rollback(opts Options, before string) error {
	if opts.Mode == Incremental {
		inProgress, err := cherryPickInProgress(opts.Repo)
		if err != nil || !inProgress {
			return err
		}
		if out, err := combinedGit(opts.Repo, "cherry-pick", "--abort"); err != nil {
			return fmt.Errorf("git cherry-pick --abort in %s: %s: %w", opts.Repo, out, err)
		}
		return nil
	}
	if out, err := combinedGit(opts.Repo, "reset", "--merge", before); err != nil {
		return fmt.Errorf("git reset --merge in %s: %s: %w", opts.Repo, out, err)
	}
	msg, err := captureGit(opts.Repo, "rev-parse", "--path-format=absolute", "--git-path", "SQUASH_MSG")
	if err != nil {
		return err
	}
	if err := os.Remove(msg); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove %s: %w", msg, err)
	}
	return nil
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
