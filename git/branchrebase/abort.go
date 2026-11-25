package branchrebase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func abort(ctx context.Context, op *Operation) error {
	if op.Phase == "restore" || op.Phase == "restore-conflict" || op.Phase == "restore-applying" || op.Phase == "restored" {
		return errors.New("rebase already completed; finish saved-edit recovery instead of aborting")
	}
	if err := abortRebase(ctx, op); err != nil {
		return err
	}
	head, err := git(ctx, op.Worktree, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if op.Phase == "stash" {
		if err := recoverStash(ctx, op); err != nil {
			return err
		}
	}
	if head != op.OriginalHead {
		if op.StashSHA != "" {
			return fmt.Errorf("HEAD %s no longer matches the original commit %s; saved edits stay in stash %s: check out %s and abort again, or finish the rebase on %s", head, op.OriginalHead, op.StashSHA, op.OriginalHead, op.Branch)
		}
		op.Aborted, op.Phase, op.Error, op.Conflicts = true, "aborted", "", []string{}
		return save(op)
	}
	op.Aborted, op.Phase = true, "aborting"
	if err := save(op); err != nil {
		return err
	}
	return restore(ctx, op)
}

func abortRebase(ctx context.Context, op *Operation) error {
	if op.Strategy == StrategyMerge {
		return abortMerge(ctx, op)
	}
	for _, name := range []string{"rebase-merge", "rebase-apply"} {
		dir, err := git(ctx, op.Worktree, "rev-parse", "--path-format=absolute", "--git-path", name)
		if err != nil {
			return err
		}
		if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return err
		}
		for file, expected := range map[string]string{"orig-head": op.OriginalHead, "onto": op.TargetSHA, "head-name": "refs/heads/" + op.Branch} {
			actual, err := os.ReadFile(filepath.Join(dir, file))
			if err != nil {
				return err
			}
			if strings.TrimSpace(string(actual)) != expected {
				return fmt.Errorf("refusing to abort an unrelated rebase: %s does not match", file)
			}
		}
		_, err = git(ctx, op.Worktree, "rebase", "--abort")
		return err
	}
	return idle(ctx, op.Worktree)
}

// abortMerge undoes the operation's own paused merge: MERGE_HEAD must be the
// pinned target and ORIG_HEAD the branch's original commit.
func abortMerge(ctx context.Context, op *Operation) error {
	mergeHead, err := git(ctx, op.Worktree, "rev-parse", "--path-format=absolute", "--git-path", "MERGE_HEAD")
	if err != nil {
		return err
	}
	if _, err := os.Stat(mergeHead); errors.Is(err, os.ErrNotExist) {
		return idle(ctx, op.Worktree)
	} else if err != nil {
		return err
	}
	for file, expected := range map[string]string{"MERGE_HEAD": op.TargetSHA, "ORIG_HEAD": op.OriginalHead} {
		path, err := git(ctx, op.Worktree, "rev-parse", "--path-format=absolute", "--git-path", file)
		if err != nil {
			return err
		}
		actual, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.TrimSpace(string(actual)) != expected {
			return fmt.Errorf("refusing to abort an unrelated merge: %s does not match", file)
		}
	}
	_, err = git(ctx, op.Worktree, "merge", "--abort")
	return err
}

func recoverStash(ctx context.Context, op *Operation) error {
	list, err := git(ctx, op.Worktree, "stash", "list", "--format=%H%x00%gs")
	if err != nil {
		return err
	}
	for _, line := range strings.Split(list, "\n") {
		sha, msg, ok := strings.Cut(line, "\x00")
		if ok && strings.HasSuffix(msg, ": gavel-rebase:"+op.ID) {
			op.StashSHA = sha
			return nil
		}
	}
	return nil
}
