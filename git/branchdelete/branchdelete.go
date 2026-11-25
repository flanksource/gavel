// Package branchdelete deletes a local branch together with its linked
// worktree, after showing what would be lost: the worktree's uncommitted files
// and the branch's commits that are not in the base branch.
package branchdelete

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gavelgit "github.com/flanksource/gavel/git"
)

var (
	// ErrNeedsConfirmation refuses a delete that would lose uncommitted files or
	// unmerged commits the caller did not acknowledge.
	ErrNeedsConfirmation = errors.New("delete needs confirmation")
	// ErrStale refuses a delete whose plan no longer matches the branch.
	ErrStale = errors.New("branch changed since the delete was planned")
	// ErrRefused refuses branches that cannot be deleted at all.
	ErrRefused = errors.New("branch cannot be deleted")
)

type Commit struct {
	SHA     string `json:"sha"`
	Subject string `json:"subject"`
}

// Plan is what deleting Branch would remove.
type Plan struct {
	Branch string `json:"branch"`
	Head   string `json:"head"`
	Base   string `json:"base"`
	// Worktree is the branch's linked worktree, removed with it; "" when none.
	Worktree string `json:"worktree,omitempty"`
	// Uncommitted are the worktree's uncommitted paths, untracked included.
	Uncommitted []string `json:"uncommitted"`
	// Unmerged are the branch's commits whose change is not in Base.
	Unmerged []Commit `json:"unmerged"`
}

type Options struct {
	// Repo is any checkout of the repository.
	Repo   string
	Branch string
	Base   string
	// Worktree is the checkout holding Branch; zero when it has none.
	Worktree gavelgit.Worktree
	// ExpectedHead, DiscardUncommitted and DeleteUnmerged apply to Delete only.
	ExpectedHead       string
	DiscardUncommitted bool
	DeleteUnmerged     bool
}

type Result struct {
	WorktreeRemoved bool `json:"worktreeRemoved"`
	BranchDeleted   bool `json:"branchDeleted"`
}

// Inspect plans deleting opts.Branch without changing anything.
func Inspect(ctx context.Context, opts Options) (Plan, error) {
	if err := gavelgit.ValidateBranchName(opts.Branch); err != nil {
		return Plan{}, fmt.Errorf("%w: %w", ErrRefused, err)
	}
	if opts.Branch == opts.Base {
		return Plan{}, fmt.Errorf("%w: %s is the base branch", ErrRefused, opts.Branch)
	}
	if opts.Worktree.Primary {
		return Plan{}, fmt.Errorf("%w: %s is checked out in the main checkout %s; switch it to another branch first", ErrRefused, opts.Branch, opts.Worktree.Path)
	}
	head, err := git(ctx, opts.Repo, "rev-parse", "--verify", "--end-of-options", "refs/heads/"+opts.Branch)
	if err != nil {
		return Plan{}, err
	}
	plan := Plan{Branch: opts.Branch, Head: head, Base: opts.Base, Worktree: opts.Worktree.Path, Uncommitted: []string{}, Unmerged: []Commit{}}
	if plan.Worktree != "" && !opts.Worktree.Prunable {
		if plan.Uncommitted, err = uncommitted(ctx, plan.Worktree); err != nil {
			return Plan{}, err
		}
	}
	plan.Unmerged, err = unmerged(ctx, opts.Repo, opts.Base, opts.Branch)
	return plan, err
}

// Delete removes the branch's worktree and then the branch. It re-plans first
// and refuses when the branch moved or when uncommitted files or unmerged
// commits would be lost without the matching acknowledgement.
func Delete(ctx context.Context, opts Options) (Result, error) {
	var result Result
	plan, err := Inspect(ctx, opts)
	if err != nil {
		return result, err
	}
	if opts.ExpectedHead == "" || plan.Head != opts.ExpectedHead {
		return result, fmt.Errorf("%w: %s is at %s, expected %q", ErrStale, plan.Branch, plan.Head, opts.ExpectedHead)
	}
	if len(plan.Uncommitted) > 0 && !opts.DiscardUncommitted {
		return result, fmt.Errorf("%w: %s has %d uncommitted files", ErrNeedsConfirmation, plan.Worktree, len(plan.Uncommitted))
	}
	if len(plan.Unmerged) > 0 && !opts.DeleteUnmerged {
		return result, fmt.Errorf("%w: %s has %d commits not in %s", ErrNeedsConfirmation, plan.Branch, len(plan.Unmerged), plan.Base)
	}
	if plan.Worktree != "" {
		args := []string{"worktree", "remove"}
		if len(plan.Uncommitted) > 0 {
			args = append(args, "--force")
		}
		if _, err := git(ctx, opts.Repo, append(args, "--", filepath.Clean(plan.Worktree))...); err != nil {
			return result, err
		}
		result.WorktreeRemoved = true
	}
	if _, err := git(ctx, opts.Repo, "branch", "-D", "--", plan.Branch); err != nil {
		return result, err
	}
	result.BranchDeleted = true
	return result, nil
}

func uncommitted(ctx context.Context, worktree string) ([]string, error) {
	if _, err := os.Stat(worktree); err != nil {
		return nil, fmt.Errorf("inspect worktree %s: %w", worktree, err)
	}
	out, err := gitOutput(ctx, worktree, "status", "--porcelain", "--untracked-files=all", "-z")
	if err != nil {
		return nil, err
	}
	paths := []string{}
	entries := strings.Split(strings.TrimSuffix(out, "\x00"), "\x00")
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		if len(entry) < 4 {
			continue
		}
		paths = append(paths, entry[3:])
		// A rename or copy is followed by its source path.
		if entry[0] == 'R' || entry[0] == 'C' {
			i++
		}
	}
	return paths, nil
}

// unmerged lists the branch's commits whose patch is not in base (git cherry).
func unmerged(ctx context.Context, repo, base, branch string) ([]Commit, error) {
	if err := gavelgit.ValidateBranchName(base); err != nil {
		return nil, fmt.Errorf("base: %w", err)
	}
	out, err := git(ctx, repo, "cherry", "-v", "refs/heads/"+base, "refs/heads/"+branch)
	if err != nil {
		return nil, err
	}
	commits := []Commit{}
	for _, line := range strings.Split(out, "\n") {
		rest, ok := strings.CutPrefix(line, "+ ")
		if !ok {
			continue
		}
		sha, subject, _ := strings.Cut(rest, " ")
		commits = append(commits, Commit{SHA: sha, Subject: subject})
	}
	return commits, nil
}

func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s in %s: %w\n%s", strings.Join(args, " "), dir, err, out)
	}
	return string(out), nil
}

func git(ctx context.Context, dir string, args ...string) (string, error) {
	out, err := gitOutput(ctx, dir, args...)
	return strings.TrimSpace(out), err
}
