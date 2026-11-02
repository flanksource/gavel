package branchrebase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gofrs/flock"
)

func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_EDITOR=true")
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

func statePath(dir string) (string, error) {
	out, err := git(context.Background(), dir, "rev-parse", "--path-format=absolute", "--git-path", "gavel-rebase.json")
	if err != nil {
		return "", err
	}
	return out, nil
}

func load(dir string) (*Operation, error) {
	path, err := statePath(dir)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var op Operation
	if err := json.Unmarshal(raw, &op); err != nil {
		return nil, fmt.Errorf("read rebase operation: %w", err)
	}
	if op.ID == "" || op.Worktree != dir {
		return nil, errors.New("invalid rebase recovery metadata")
	}
	return &op, nil
}

func save(op *Operation) error {
	path, err := statePath(op.Worktree)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(op)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path+".new", raw, 0600); err != nil {
		return err
	}
	return os.Rename(path+".new", path)
}

func lock(ctx context.Context, dir string) (func(), error) {
	path, err := statePath(dir)
	if err != nil {
		return nil, err
	}
	l := flock.New(filepath.Clean(path) + ".lock")
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ok, err := l.TryLock()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("another rebase command is running for this worktree")
	}
	return func() { mustUnlock(l) }, nil
}

func mustUnlock(lock *flock.Flock) {
	if err := lock.Unlock(); err != nil {
		panic(fmt.Errorf("release rebase filesystem lock: %w", err))
	}
}

func idle(ctx context.Context, dir string) error {
	for _, name := range []string{"rebase-merge", "rebase-apply", "MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD", "sequencer"} {
		path, err := git(ctx, dir, "rev-parse", "--path-format=absolute", "--git-path", name)
		if err != nil {
			return err
		}
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s has %s in progress; resolve it before continuing", dir, name)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func conflicts(ctx context.Context, dir string) ([]string, error) {
	out, err := gitOutput(ctx, dir, "diff", "--name-only", "--diff-filter=U", "-z")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return []string{}, nil
	}
	return strings.Split(strings.TrimSuffix(out, "\x00"), "\x00"), nil
}
