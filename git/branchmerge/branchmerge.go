// Package branchmerge lands a branch's commits onto the branch checked out in
// a target checkout, either one cherry-pick per commit (incremental) or as a
// single squash commit, and afterwards removes the branch's worktree and
// deletes the branch once its work is verified to be in the landed commit.
// A failed merge is always rolled back, leaving the checkout where it started.
package branchmerge

import (
	"errors"
	"fmt"
	"strings"

	gavelgit "github.com/flanksource/gavel/git"
	prcreate "github.com/flanksource/gavel/pr/create"
)

type Mode string

const (
	// Squash lands the branch as one commit carrying Options.Message.
	Squash Mode = "squash"
	// Incremental cherry-picks Options.Commits one by one.
	Incremental Mode = "incremental"
)

var (
	ErrInvalidOptions   = errors.New("invalid branch merge options")
	ErrCheckoutNotReady = errors.New("checkout is not ready to merge into")
	ErrNothingToMerge   = errors.New("nothing to merge")
)

type Options struct {
	// Repo is the target checkout; its checked-out branch receives the work.
	Repo string
	// Branch is the source branch. It must not be checked out in Repo.
	Branch string
	// Base, when set, is the branch Repo must have checked out.
	Base string
	// Commits are cherry-picked in order (Incremental only).
	Commits []string
	Mode    Mode
	// Message is the squash commit message (Squash only).
	Message string
}

// validated checks o and returns a copy built only from the checked values, so
// every branch name and commit that reaches a git command line has passed
// validation.
func (o Options) validated() (Options, error) {
	if strings.TrimSpace(o.Repo) == "" || strings.TrimSpace(o.Branch) == "" {
		return Options{}, fmt.Errorf("%w: repo and branch are required", ErrInvalidOptions)
	}
	if err := gavelgit.ValidateBranchName(o.Branch); err != nil {
		return Options{}, fmt.Errorf("%w: %w", ErrInvalidOptions, err)
	}
	out := Options{Repo: o.Repo, Branch: o.Branch, Mode: o.Mode, Message: o.Message}
	if o.Base != "" {
		if err := gavelgit.ValidateBranchName(o.Base); err != nil {
			return Options{}, fmt.Errorf("%w: base: %w", ErrInvalidOptions, err)
		}
		out.Base = o.Base
	}
	switch o.Mode {
	case Incremental:
		if len(o.Commits) == 0 {
			return Options{}, fmt.Errorf("%w: an incremental merge needs the commits to cherry-pick", ErrInvalidOptions)
		}
		for _, commit := range o.Commits {
			if err := gavelgit.ValidateRevision(commit); err != nil {
				return Options{}, fmt.Errorf("%w: %w", ErrInvalidOptions, err)
			}
			out.Commits = append(out.Commits, commit)
		}
	case Squash:
		if strings.TrimSpace(o.Message) == "" {
			return Options{}, fmt.Errorf("%w: a squash merge needs a commit message", ErrInvalidOptions)
		}
		if len(o.Commits) > 0 {
			return Options{}, fmt.Errorf("%w: a squash merge lands the whole branch; commits only apply to %s", ErrInvalidOptions, Incremental)
		}
	default:
		return Options{}, fmt.Errorf("%w: mode %q must be %q or %q", ErrInvalidOptions, o.Mode, Squash, Incremental)
	}
	return out, nil
}

type Result struct {
	TargetBranch string
	LandedSHA    string
	// Commits is how many source commits were landed.
	Commits int
}

// ConflictError reports a merge that conflicted. The merge was aborted, so the
// target checkout is back where it started.
type ConflictError struct {
	// Branch is the target branch the merge was onto.
	Branch string
	Source string
	Mode   Mode
	Paths  []string
	Err    error
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s merge of %s onto %s conflicted in %s; the merge was aborted",
		e.Mode, e.Source, e.Branch, strings.Join(e.Paths, ", "))
}

func (e *ConflictError) Unwrap() error { return e.Err }

// Merge lands opts.Branch onto the branch checked out in opts.Repo.
func Merge(opts Options) (Result, error) {
	opts, err := opts.validated()
	if err != nil {
		return Result{}, err
	}
	target, err := targetBranch(opts)
	if err != nil {
		return Result{}, err
	}
	before, err := captureGit(opts.Repo, "rev-parse", "HEAD")
	if err != nil {
		return Result{}, err
	}
	commits := len(opts.Commits)
	if opts.Mode == Squash {
		commits, err = squash(opts, target, before)
	} else {
		err = cherryPick(opts, target, before)
	}
	if err != nil {
		return Result{}, err
	}
	head, err := captureGit(opts.Repo, "rev-parse", "HEAD")
	if err != nil {
		return Result{}, err
	}
	return Result{TargetBranch: target, LandedSHA: head, Commits: commits}, nil
}

// targetBranch refuses a checkout that cannot take the merge and returns its
// checked-out branch.
func targetBranch(opts Options) (string, error) {
	if err := prcreate.RefuseInProgressOps(opts.Repo); err != nil {
		return "", fmt.Errorf("%w: %w", ErrCheckoutNotReady, err)
	}
	staged, err := captureGit(opts.Repo, "diff", "--cached", "--name-only")
	if err != nil {
		return "", err
	}
	if staged != "" {
		return "", fmt.Errorf("%w: %s has staged changes (%s); commit or unstage them first",
			ErrCheckoutNotReady, opts.Repo, strings.Join(strings.Fields(staged), ", "))
	}
	branch, err := captureGit(opts.Repo, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("%w: %s has a detached HEAD; check out the branch to merge into", ErrCheckoutNotReady, opts.Repo)
	}
	if branch == opts.Branch {
		return "", fmt.Errorf("%w: %s has the source branch %s checked out", ErrCheckoutNotReady, opts.Repo, branch)
	}
	if opts.Base != "" && branch != opts.Base {
		return "", fmt.Errorf("%w: %s has %q checked out, not the base branch %q", ErrCheckoutNotReady, opts.Repo, branch, opts.Base)
	}
	return branch, nil
}
