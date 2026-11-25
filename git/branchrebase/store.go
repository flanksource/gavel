package branchrebase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	gavelgit "github.com/flanksource/gavel/git"
	"github.com/gofrs/flock"
)

var commitIDPattern = regexp.MustCompile(`^(?:[0-9a-fA-F]{40}|[0-9a-fA-F]{64})$`)

func validateTarget(ref string) error {
	if ref == "refs/remotes/origin/main" {
		return nil
	}
	if !strings.HasPrefix(ref, "refs/heads/") {
		return errors.New("target must be origin/main or a local branch ref")
	}
	return gavelgit.ValidateBranchName(strings.TrimPrefix(ref, "refs/heads/"))
}

func validateCommitID(sha string) error {
	if !commitIDPattern.MatchString(sha) {
		return fmt.Errorf("invalid rebase commit ID %q", sha)
	}
	return nil
}

func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	if strings.TrimSpace(dir) == "" {
		return "", errors.New("git requires a worktree directory")
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return "", fmt.Errorf("open git worktree: %w", err)
	}
	defer root.Close()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = root.Name()
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
	if err := validateTarget(op.Target); err != nil {
		return nil, fmt.Errorf("invalid rebase recovery target: %w", err)
	}
	// Operations recorded before strategies existed were all rebases.
	if op.Strategy == "" {
		op.Strategy = StrategyRebase
	}
	if err := validateStrategy(op.Strategy); err != nil {
		return nil, fmt.Errorf("invalid rebase recovery strategy: %w", err)
	}
	for _, sha := range []string{op.OriginalHead, op.TargetSHA, op.StashSHA} {
		if sha == "" {
			continue
		}
		if err := validateCommitID(sha); err != nil {
			return nil, fmt.Errorf("invalid rebase recovery commit ID: %w", err)
		}
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
