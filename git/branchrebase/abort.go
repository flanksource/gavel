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
	if head != op.OriginalHead {
		return errors.New("HEAD no longer matches the original commit; refusing to reset it")
	}
	if op.Phase == "stash" {
		if err := recoverStash(ctx, op); err != nil {
			return err
		}
	}
	op.Aborted, op.Phase = true, "aborting"
	if err := save(op); err != nil {
		return err
	}
	return restore(ctx, op)
}

func abortRebase(ctx context.Context, op *Operation) error {
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
