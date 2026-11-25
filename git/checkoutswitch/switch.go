package checkoutswitch

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Sync is how Switch brings the branch level with its upstream.
type Sync string

const (
	// SyncNone switches without moving the branch.
	SyncNone Sync = "none"
	// SyncFastForward moves the branch to its upstream when it has no commits
	// the upstream lacks.
	SyncFastForward Sync = "fast-forward"
	// SyncReset moves the branch to its upstream when every commit it would
	// drop is held by another ref or already landed upstream.
	SyncReset Sync = "reset"
)

// Options selects the branch to switch the primary checkout of Repo to.
// ExpectedHead is the checkout's HEAD the caller planned against.
type Options struct {
	Repo         string `json:"repo"`
	Branch       string `json:"branch"`
	ExpectedHead string `json:"expectedHead"`
	Sync         Sync   `json:"sync"`
}

// Result is the checkout after the switch, with what it replaced so the move
// can be undone.
type Result struct {
	Branch             string `json:"branch"`
	Head               string `json:"head"`
	Sync               Sync   `json:"sync"`
	PreviousBranch     string `json:"previousBranch"`
	PreviousHead       string `json:"previousHead"`
	PreviousBranchHead string `json:"previousBranchHead"`
}

// Switch fetches the branch's upstream (unless Sync is none), moves the branch
// as Sync asks, then checks it out in the primary checkout.
func Switch(ctx context.Context, opts Options) (Result, error) {
	if opts.Sync != SyncNone && opts.Sync != SyncFastForward && opts.Sync != SyncReset {
		return Result{}, fmt.Errorf("%w: sync %q must be %q, %q or %q", ErrInvalidOptions, opts.Sync, SyncNone, SyncFastForward, SyncReset)
	}
	if opts.ExpectedHead == "" {
		return Result{}, fmt.Errorf("%w: expectedHead is required", ErrInvalidOptions)
	}
	plan, err := Inspect(ctx, opts.Repo, opts.Branch)
	if err != nil {
		return Result{}, err
	}
	if err := refuse(ctx, plan, opts.ExpectedHead); err != nil {
		return Result{}, err
	}
	if opts.Sync != SyncNone {
		if plan, err = fetchUpstream(ctx, plan); err != nil {
			return Result{}, err
		}
	}
	target, err := syncTarget(plan, opts.Sync)
	if err != nil {
		return Result{}, err
	}
	if err := checkout(ctx, plan, target); err != nil {
		return Result{}, err
	}
	head, err := revParse(ctx, plan.Repo, "HEAD")
	if err != nil {
		return Result{}, err
	}
	return Result{
		Branch: plan.Branch, Head: head, Sync: opts.Sync,
		PreviousBranch: plan.Current, PreviousHead: plan.Head, PreviousBranchHead: plan.BranchHead,
	}, nil
}

func refuse(ctx context.Context, plan Plan, expectedHead string) error {
	switch {
	case plan.Head != expectedHead:
		return fmt.Errorf("%w: HEAD of %s moved from %s to %s; refresh and try again", ErrRefused, plan.Repo, expectedHead, plan.Head)
	case plan.DirtyCount > 0:
		return fmt.Errorf("%w: %d tracked files in %s have uncommitted changes; commit or discard them first", ErrRefused, plan.DirtyCount, plan.Repo)
	case plan.CheckedOutIn != "":
		return fmt.Errorf("%w: %s is checked out in the worktree %s", ErrRefused, plan.Branch, plan.CheckedOutIn)
	}
	return idle(ctx, plan.Repo)
}

// fetchUpstream refreshes the branch's remote-tracking upstream and re-plans
// against it.
func fetchUpstream(ctx context.Context, plan Plan) (Plan, error) {
	if plan.Upstream == "" {
		return plan, fmt.Errorf("%w: %s tracks no upstream to sync with", ErrRefused, plan.Branch)
	}
	remote, err := run(ctx, plan.Repo, "config", "--get", "branch."+plan.Branch+".remote")
	if err != nil {
		return plan, err
	}
	if remote != "." {
		merge, err := run(ctx, plan.Repo, "config", "--get", "branch."+plan.Branch+".merge")
		if err != nil {
			return plan, err
		}
		if _, err := run(ctx, plan.Repo, "fetch", "--no-tags", remote, "+"+merge+":"+plan.upstreamRef); err != nil {
			return plan, err
		}
	}
	refreshed, err := Inspect(ctx, plan.Repo, plan.Branch)
	if err != nil {
		return plan, err
	}
	if refreshed.UpstreamHead == "" {
		return refreshed, fmt.Errorf("%w: the upstream %s of %s no longer exists", ErrRefused, refreshed.Upstream, refreshed.Branch)
	}
	return refreshed, nil
}

// syncTarget is the commit the branch should point at after the switch.
func syncTarget(plan Plan, sync Sync) (string, error) {
	switch sync {
	case SyncFastForward:
		if plan.Ahead > 0 {
			return "", fmt.Errorf("%w: %s has diverged from %s (%d ahead, %d behind) and cannot fast-forward; reset or rebase it instead",
				ErrRefused, plan.Branch, plan.Upstream, plan.Ahead, plan.Behind)
		}
		return plan.UpstreamHead, nil
	case SyncReset:
		if len(plan.Unpreserved) > 0 {
			subjects := make([]string, 0, len(plan.Unpreserved))
			for _, c := range plan.Unpreserved {
				subjects = append(subjects, c.SHA[:8]+" "+c.Subject)
			}
			return "", fmt.Errorf("%w: resetting %s to %s would orphan %d commits no other branch holds:\n%s",
				ErrRefused, plan.Branch, plan.Upstream, len(plan.Unpreserved), strings.Join(subjects, "\n"))
		}
		return plan.UpstreamHead, nil
	}
	return plan.BranchHead, nil
}

// checkout points the branch at target and checks it out. A branch that is not
// checked out moves by compare-and-swap before `git switch`, which still
// refuses to overwrite untracked files; the move is undone if it does.
func checkout(ctx context.Context, plan Plan, target string) error {
	if plan.Current == plan.Branch {
		if target == plan.BranchHead {
			return nil
		}
		_, err := run(ctx, plan.Repo, "reset", "--keep", target)
		return err
	}
	ref := "refs/heads/" + plan.Branch
	if target != plan.BranchHead {
		if _, err := run(ctx, plan.Repo, "update-ref", "-m", "gavel: sync "+plan.Branch+" with "+plan.Upstream, ref, target, plan.BranchHead); err != nil {
			return err
		}
	}
	if _, err := run(ctx, plan.Repo, "switch", "--no-guess", plan.Branch); err != nil {
		if target == plan.BranchHead {
			return err
		}
		_, restoreErr := run(ctx, plan.Repo, "update-ref", "-m", "gavel: restore "+plan.Branch, ref, plan.BranchHead, target)
		return errors.Join(err, restoreErr)
	}
	return nil
}
