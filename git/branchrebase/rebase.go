package branchrebase

import (
	"context"
	"errors"
	"fmt"

	gavelgit "github.com/flanksource/gavel/git"
	"github.com/google/uuid"
)

type Options struct {
	Action       string `json:"-"`
	Worktree     string `json:"worktree"`
	Branch       string `json:"branch"`
	Target       string `json:"target"`
	ExpectedHead string `json:"expectedHead"`
	Operation    string `json:"operation"`
	Resolved     bool   `json:"resolved"`
	// Strategy brings the branch up to date by rebasing onto the target
	// (StrategyRebase, the default) or by merging the target in.
	Strategy Strategy `json:"strategy"`
	// RebaseMerges keeps the branch's merge commits when rebasing.
	RebaseMerges bool `json:"rebaseMerges"`
	// NoFastForward records a merge commit even when the branch could fast-forward.
	NoFastForward bool `json:"noFastForward"`
}

type Strategy string

const (
	StrategyRebase Strategy = "rebase"
	StrategyMerge  Strategy = "merge"
)

func validateStrategy(strategy Strategy) error {
	if strategy != StrategyRebase && strategy != StrategyMerge {
		return fmt.Errorf("strategy must be %q or %q, got %q", StrategyRebase, StrategyMerge, strategy)
	}
	return nil
}

type Operation struct {
	ID            string   `json:"id"`
	Worktree      string   `json:"worktree"`
	Branch        string   `json:"branch"`
	Strategy      Strategy `json:"strategy"`
	RebaseMerges  bool     `json:"rebaseMerges,omitempty"`
	NoFastForward bool     `json:"noFastForward,omitempty"`
	Target        string   `json:"target"`
	TargetSHA     string   `json:"targetSha"`
	OriginalHead  string   `json:"originalHead"`
	StashSHA      string   `json:"stashSha"`
	Phase         string   `json:"phase"`
	Error         string   `json:"error"`
	Conflicts     []string `json:"conflicts"`
	Aborted       bool     `json:"aborted"`
}

type Target struct {
	Ref   string `json:"ref"`
	Label string `json:"label"`
}

type Context struct {
	Branch     string     `json:"branch"`
	Head       string     `json:"head"`
	DirtyCount int        `json:"dirtyCount"`
	Targets    []Target   `json:"targets"`
	Operation  *Operation `json:"operation"`
}

func Run(ctx context.Context, opts Options) (*Operation, error) {
	root, err := git(ctx, opts.Worktree, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}
	opts.Worktree = root
	unlock, err := lock(ctx, opts.Worktree)
	if err != nil {
		return nil, err
	}
	defer unlock()
	op, err := load(opts.Worktree)
	if err != nil {
		return nil, err
	}
	if op != nil && (op.Phase == "done" || op.Phase == "aborted") && opts.Action == "start" && opts.Operation == "" {
		op = nil
	}
	if op != nil {
		if opts.Operation != op.ID {
			return op, errors.New("a rebase operation already exists; use its operation ID to resume, finish or abort")
		}
	} else if opts.Action != "start" || opts.Operation != "" {
		return nil, errors.New("no matching rebase operation exists")
	}
	if op == nil {
		op, err = prepare(ctx, opts)
		if err != nil {
			return nil, err
		}
	}
	if op.Phase == "done" || op.Phase == "aborted" {
		if opts.Action == "check" {
			return op, check(ctx, op)
		}
		if opts.Action == "finish" || (opts.Action == "abort" && op.Aborted) {
			return op, nil
		}
		return op, errors.New("operation already completed")
	}
	if opts.Resolved {
		if op.Phase != "restore-conflict" && op.Phase != "restore-applying" {
			return op, errors.New("resolved applies only to interrupted saved-edit restoration")
		}
		if err := check(ctx, op); err != nil {
			return op, err
		}
		op.Phase = "restored"
		if err := save(op); err != nil {
			return op, err
		}
	}
	switch opts.Action {
	case "start":
		err = start(ctx, op)
	case "finish":
		err = finish(ctx, op)
	case "check":
		err = check(ctx, op)
	case "abort":
		err = abort(ctx, op)
	default:
		err = fmt.Errorf("unknown rebase action %q", opts.Action)
	}
	if err != nil {
		op.Error = err.Error()
		paths, pathErr := conflicts(ctx, op.Worktree)
		op.Conflicts = paths
		err = errors.Join(err, pathErr, save(op))
	}
	return op, err
}

func prepare(ctx context.Context, opts Options) (*Operation, error) {
	if opts.Branch == "" || opts.ExpectedHead == "" {
		return nil, errors.New("branch and expectedHead are required")
	}
	if err := idle(ctx, opts.Worktree); err != nil {
		return nil, err
	}
	branch, err := git(ctx, opts.Worktree, "branch", "--show-current")
	if err != nil {
		return nil, err
	}
	if branch != opts.Branch {
		return nil, fmt.Errorf("expected branch %q, found %q", opts.Branch, branch)
	}
	head, err := git(ctx, opts.Worktree, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}
	if head != opts.ExpectedHead {
		return nil, errors.New("HEAD changed; refresh before rebasing")
	}
	if err := validateTarget(opts.Target); err != nil {
		return nil, err
	}
	if opts.Target == "refs/heads/"+branch {
		return nil, errors.New("cannot rebase a branch onto itself")
	}
	if opts.Strategy == "" {
		opts.Strategy = StrategyRebase
	}
	if err := validateStrategy(opts.Strategy); err != nil {
		return nil, err
	}
	if opts.RebaseMerges && opts.Strategy != StrategyRebase {
		return nil, errors.New("rebaseMerges applies only to the rebase strategy")
	}
	if opts.NoFastForward && opts.Strategy != StrategyMerge {
		return nil, errors.New("noFastForward applies only to the merge strategy")
	}
	op := &Operation{
		ID: uuid.NewString(), Worktree: opts.Worktree, Branch: branch, Strategy: opts.Strategy,
		RebaseMerges: opts.RebaseMerges, NoFastForward: opts.NoFastForward,
		Target: opts.Target, OriginalHead: head, Phase: "fetch", Conflicts: []string{},
	}
	return op, save(op)
}

func start(ctx context.Context, op *Operation) error {
	if op.Phase == "fetch" || op.Phase == "stash" {
		branch, err := git(ctx, op.Worktree, "branch", "--show-current")
		if err != nil {
			return err
		}
		head, err := git(ctx, op.Worktree, "rev-parse", "HEAD")
		if err != nil {
			return err
		}
		if branch != op.Branch || head != op.OriginalHead {
			return errors.New("branch or HEAD changed since the rebase started")
		}
		if err := idle(ctx, op.Worktree); err != nil {
			return err
		}
	}
	if op.Phase == "restore" || op.Phase == "restore-conflict" || op.Phase == "restored" {
		return finish(ctx, op)
	}
	if op.Phase == "fetch" {
		if op.Target == "refs/remotes/origin/main" {
			if _, err := git(ctx, op.Worktree, "fetch", "--no-tags", "origin", "+refs/heads/main:refs/remotes/origin/main"); err != nil {
				return err
			}
		}
		if err := gavelgit.ValidateRevision(op.Target); err != nil {
			return err
		}
		sha, err := git(ctx, op.Worktree, "rev-parse", "--verify", "--end-of-options", op.Target+"^{commit}")
		if err != nil {
			return err
		}
		if err := validateCommitID(sha); err != nil {
			return err
		}
		op.TargetSHA, op.Phase = sha, "stash"
		if err := save(op); err != nil {
			return err
		}
	}
	if op.Phase == "stash" {
		if err := stash(ctx, op); err != nil {
			return err
		}
		op.Phase = "rebase"
		if err := save(op); err != nil {
			return err
		}
	}
	if err := idle(ctx, op.Worktree); err != nil {
		return err
	}
	branch, err := git(ctx, op.Worktree, "branch", "--show-current")
	if err != nil {
		return err
	}
	if branch != op.Branch {
		return errors.New("checked-out branch changed since the rebase started")
	}
	head, err := git(ctx, op.Worktree, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if head != op.OriginalHead {
		return finish(ctx, op)
	}
	if err := validateCommitID(op.TargetSHA); err != nil {
		return err
	}
	if _, err := git(ctx, op.Worktree, updateArgs(op)...); err != nil {
		return err
	}
	return finish(ctx, op)
}

// updateArgs is the git command that brings the branch up to the pinned target
// with the operation's strategy.
func updateArgs(op *Operation) []string {
	if op.Strategy == StrategyMerge {
		args := []string{"-c", "core.editor=true", "merge", "--no-edit", "--no-autostash"}
		if op.NoFastForward {
			args = append(args, "--no-ff")
		}
		return append(args, "--", op.TargetSHA)
	}
	args := []string{"-c", "core.editor=true", "rebase", "--no-autostash"}
	if op.RebaseMerges {
		args = append(args, "--rebase-merges")
	}
	return append(args, "--", op.TargetSHA)
}
