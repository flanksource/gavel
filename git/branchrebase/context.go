package branchrebase

import (
	"context"
	"strings"

	gavelgit "github.com/flanksource/gavel/git"
)

func Inspect(ctx context.Context, dir string) (Context, error) {
	root, err := git(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return Context{}, err
	}
	op, err := load(root)
	if err != nil {
		return Context{}, err
	}
	branch, err := git(ctx, root, "branch", "--show-current")
	if err != nil {
		return Context{}, err
	}
	head, err := git(ctx, root, "rev-parse", "HEAD")
	if err != nil {
		return Context{}, err
	}
	dirty, err := git(ctx, root, "status", "--porcelain", "--untracked-files=all")
	if err != nil {
		return Context{}, err
	}
	branches, err := gavelgit.ListBranches(root)
	if err != nil {
		return Context{}, err
	}
	result := Context{Branch: branch, Head: head, Operation: op, Targets: []Target{{Ref: "refs/remotes/origin/main", Label: "origin/main"}}}
	if dirty != "" {
		result.DirtyCount = len(strings.Split(dirty, "\n"))
	}
	for _, b := range branches {
		if b.Name != branch {
			result.Targets = append(result.Targets, Target{Ref: "refs/heads/" + b.Name, Label: b.Name})
		}
	}
	if op != nil {
		op.Conflicts, err = conflicts(ctx, root)
		if err != nil {
			return Context{}, err
		}
		if op.Phase == "done" || op.Phase == "aborted" {
			result.Operation = nil
		}
	}
	return result, nil
}
