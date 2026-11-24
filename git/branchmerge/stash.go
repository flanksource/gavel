package branchmerge

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/commons/logger"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/google/uuid"
)

func lockMerge(repo string) (func(), error) {
	lock, err := gavelgit.StashLock(context.Background(), repo)
	if err != nil {
		return nil, err
	}
	acquired, err := lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("lock %s for merging: %w", repo, err)
	}
	if !acquired {
		return nil, fmt.Errorf("%w: another Gavel operation is using the stash in %s", ErrCheckoutNotReady, repo)
	}
	return func() {
		if err := lock.Unlock(); err != nil {
			logger.Errorf("unlock %s after merging: %v", repo, err)
		}
	}, nil
}

func saveTargetEdits(opts Options) (string, error) {
	if !opts.Autostash {
		return "", nil
	}
	dirty, err := captureGit(opts.Repo, "status", "--porcelain", "--untracked-files=normal", "--ignore-submodules=dirty")
	if err != nil || dirty == "" {
		return "", err
	}
	message := "gavel merge " + opts.Branch + " " + uuid.NewString()
	if out, err := combinedGit(opts.Repo, "stash", "push", "--include-untracked", "--message", message); err != nil {
		return "", fmt.Errorf("save local changes in %q: %s: %w", message, out, err)
	}
	list, err := captureGit(opts.Repo, "stash", "list", "--format=%H%x00%gs")
	if err != nil {
		return "", fmt.Errorf("find saved edits in stash %q: %w", message, err)
	}
	for _, line := range strings.Split(list, "\n") {
		sha, subject, _ := strings.Cut(line, "\x00")
		if strings.HasSuffix(subject, ": "+message) {
			return sha, nil
		}
	}
	return "", fmt.Errorf("could not find the stash %q for the local changes in %s", message, opts.Repo)
}

func restoreTargetEdits(repo, sha string) error {
	if sha == "" {
		return nil
	}
	if out, err := combinedGit(repo, "stash", "apply", "--index", "--", sha); err != nil {
		return fmt.Errorf("could not restore local changes in %s (stash %s retained): %s: %w", repo, sha, out, err)
	}
	list, err := captureGit(repo, "stash", "list", "--format=%H%x00%gd")
	if err != nil {
		return fmt.Errorf("local changes restored but could not locate stash %s for removal: %w", sha, err)
	}
	for _, line := range strings.Split(list, "\n") {
		id, ref, _ := strings.Cut(line, "\x00")
		if id == sha {
			if _, err := captureGit(repo, "stash", "drop", "--", ref); err != nil {
				return fmt.Errorf("local changes restored but could not drop stash %s: %w", sha, err)
			}
			break
		}
	}
	return nil
}
