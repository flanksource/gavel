package branchrebase

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

func stash(ctx context.Context, op *Operation) error {
	unlock, err := lockStash(ctx, op.Worktree)
	if err != nil {
		return err
	}
	defer unlock()
	if err := recoverStash(ctx, op); err != nil {
		return err
	}
	if op.StashSHA != "" {
		return save(op)
	}
	dirty, err := git(ctx, op.Worktree, "status", "--porcelain", "--untracked-files=all")
	if err != nil || dirty == "" {
		return err
	}
	if _, err := git(ctx, op.Worktree, "stash", "push", "--include-untracked", "--message", "gavel-rebase:"+op.ID); err != nil {
		return err
	}
	if err := recoverStash(ctx, op); err != nil {
		return err
	}
	if op.StashSHA == "" {
		return errors.New("stash was created but its operation identity cannot be found")
	}
	return save(op)
}

func check(ctx context.Context, op *Operation) error {
	if err := idle(ctx, op.Worktree); err != nil {
		return err
	}
	branch, err := git(ctx, op.Worktree, "branch", "--show-current")
	if err != nil {
		return err
	}
	if branch != op.Branch {
		return fmt.Errorf("expected original branch %s, found %s", op.Branch, branch)
	}
	paths, err := conflicts(ctx, op.Worktree)
	if err != nil {
		return err
	}
	if len(paths) > 0 {
		return fmt.Errorf("unresolved conflicts: %s", strings.Join(paths, ", "))
	}
	if !op.Aborted {
		if op.TargetSHA == "" {
			return errors.New("rebase target has not been resolved")
		}
		if _, err := git(ctx, op.Worktree, "merge-base", "--is-ancestor", op.TargetSHA, "HEAD"); err != nil {
			return err
		}
		if _, err := git(ctx, op.Worktree, "diff", "--check", op.TargetSHA+"...HEAD"); err != nil {
			return err
		}
	}
	return nil
}

func finish(ctx context.Context, op *Operation) error {
	if op.Phase == "restore-conflict" || op.Phase == "restore-applying" {
		return errors.New("saved edits need recovery; resolve them and finish with --resolved, retaining the edits uncommitted")
	}
	if err := check(ctx, op); err != nil {
		return err
	}
	if op.Phase != "restore" && op.Phase != "restore-conflict" && op.Phase != "restored" && op.Phase != "aborting" {
		dirty, err := git(ctx, op.Worktree, "status", "--porcelain", "--untracked-files=all")
		if err != nil {
			return err
		}
		if dirty != "" {
			return errors.New("unexpected edits before stash restoration; preserve them before finalizing")
		}
		op.Phase = "restore"
		if err := save(op); err != nil {
			return err
		}
	}
	return restore(ctx, op)
}

func restore(ctx context.Context, op *Operation) error {
	if op.StashSHA != "" && op.Phase != "restore-conflict" && op.Phase != "restored" {
		op.Phase = "restore-applying"
		if err := save(op); err != nil {
			return err
		}
		if _, err := git(ctx, op.Worktree, "stash", "apply", "--index", op.StashSHA); err != nil {
			op.Phase = "restore-conflict"
			return errors.Join(fmt.Errorf("restore saved edits (stash %s retained): %w", op.StashSHA, err), save(op))
		}
	}
	paths, err := conflicts(ctx, op.Worktree)
	if err != nil {
		return err
	}
	if len(paths) > 0 {
		return errors.New("resolve saved-edit conflicts before finalizing; leave the edits uncommitted")
	}
	op.Phase = "restored"
	if err := save(op); err != nil {
		return err
	}
	if err := dropStash(ctx, op); err != nil {
		return err
	}
	op.Phase, op.Error, op.Conflicts = "done", "", []string{}
	if op.Aborted {
		op.Phase = "aborted"
	}
	return save(op)
}

func dropStash(ctx context.Context, op *Operation) error {
	if op.StashSHA == "" {
		return nil
	}
	unlock, err := lockStash(ctx, op.Worktree)
	if err != nil {
		return err
	}
	defer unlock()
	list, err := git(ctx, op.Worktree, "stash", "list", "--format=%H%x00%gd")
	if err != nil {
		return err
	}
	for _, line := range strings.Split(list, "\n") {
		sha, ref, ok := strings.Cut(line, "\x00")
		if ok && sha == op.StashSHA {
			_, err := git(ctx, op.Worktree, "stash", "drop", ref)
			return err
		}
	}
	return nil
}
