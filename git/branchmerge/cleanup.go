package branchmerge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gavelgit "github.com/flanksource/gavel/git"
)

type CleanupOptions struct {
	Repo string
	// Worktree is the branch's worktree path; empty when it has none.
	Worktree  string
	Branch    string
	LandedSHA string
	// Limit, for an Incremental landing, excludes commits at or before it
	// from the landed check (git cherry's <limit>).
	Limit string
	Mode  Mode
}

type CleanupResult struct {
	WorktreeRemoved bool
	BranchDeleted   bool
}

// Cleanup removes the branch's worktree, if it is still there, and deletes the
// branch once its work is confirmed to be in LandedSHA: every commit by patch
// id for an Incremental landing, and the branch merging into LandedSHA as a
// no-op for a Squash landing (patch ids cannot match a squash commit).
func Cleanup(opts CleanupOptions) (CleanupResult, error) {
	var result CleanupResult
	opts, err := opts.validated()
	if err != nil {
		return result, err
	}
	if opts.Worktree != "" {
		if _, err := os.Stat(opts.Worktree); err == nil {
			if out, err := combinedGit(opts.Repo, "worktree", "remove", filepath.Clean(opts.Worktree)); err != nil {
				return result, fmt.Errorf("git worktree remove %s: %s: %w", opts.Worktree, out, err)
			}
			result.WorktreeRemoved = true
		} else if !os.IsNotExist(err) {
			return result, fmt.Errorf("inspect worktree %s: %w", opts.Worktree, err)
		}
	}
	if !branchExists(opts.Repo, opts.Branch) {
		return result, nil
	}
	verify := verifyCherry
	if opts.Mode == Squash {
		verify = verifySquash
	}
	if err := verify(opts); err != nil {
		return result, err
	}
	if out, err := combinedGit(opts.Repo, "branch", "-D", opts.Branch); err != nil {
		return result, fmt.Errorf("git branch -D %s: %s: %w", opts.Branch, out, err)
	}
	result.BranchDeleted = true
	return result, nil
}

// validated checks o and returns a copy built only from the checked values, so
// the branch and commits that reach git have all passed validation.
func (o CleanupOptions) validated() (CleanupOptions, error) {
	if err := gavelgit.ValidateBranchName(o.Branch); err != nil {
		return CleanupOptions{}, fmt.Errorf("%w: %w", ErrInvalidOptions, err)
	}
	if err := gavelgit.ValidateRevision(o.LandedSHA); err != nil {
		return CleanupOptions{}, fmt.Errorf("%w: landed commit: %w", ErrInvalidOptions, err)
	}
	out := CleanupOptions{Repo: o.Repo, Worktree: o.Worktree, Branch: o.Branch, LandedSHA: o.LandedSHA, Mode: o.Mode}
	if o.Limit != "" {
		if err := gavelgit.ValidateRevision(o.Limit); err != nil {
			return CleanupOptions{}, fmt.Errorf("%w: limit: %w", ErrInvalidOptions, err)
		}
		out.Limit = o.Limit
	}
	return out, nil
}

func verifyCherry(opts CleanupOptions) error {
	args := []string{"cherry", opts.LandedSHA, "refs/heads/" + opts.Branch}
	if opts.Limit != "" {
		args = append(args, opts.Limit)
	}
	cherry, err := captureGit(opts.Repo, args...)
	if err != nil {
		return fmt.Errorf("check that %s landed: %w", opts.Branch, err)
	}
	var unlanded []string
	for _, line := range strings.Split(cherry, "\n") {
		if sha, ok := strings.CutPrefix(line, "+ "); ok {
			unlanded = append(unlanded, short(sha))
		}
	}
	if len(unlanded) > 0 {
		return fmt.Errorf("kept branch %s: %s not in %s", opts.Branch, strings.Join(unlanded, ", "), short(opts.LandedSHA))
	}
	return nil
}

// verifySquash checks that merging the branch into LandedSHA would change
// nothing, i.e. everything the branch carries is already in the landed tree.
func verifySquash(opts CleanupOptions) error {
	landedTree, err := captureGit(opts.Repo, "rev-parse", opts.LandedSHA+"^{tree}")
	if err != nil {
		return fmt.Errorf("check that %s landed: %w", opts.Branch, err)
	}
	out, err := combinedGit(opts.Repo, "merge-tree", "--write-tree", "--no-messages", opts.LandedSHA, "refs/heads/"+opts.Branch)
	if err != nil {
		return fmt.Errorf("kept branch %s: merging it into %s conflicts: %s: %w", opts.Branch, short(opts.LandedSHA), out, err)
	}
	if merged := strings.TrimSpace(strings.SplitN(out, "\n", 2)[0]); merged != landedTree {
		return fmt.Errorf("kept branch %s: it holds changes not in %s", opts.Branch, short(opts.LandedSHA))
	}
	return nil
}
