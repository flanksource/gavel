package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/git/branchmerge"
	"github.com/spf13/cobra"
)

// newGitMergeCommand is the landing step of the branch-merge prompt.
func newGitMergeCommand() *cobra.Command {
	options := branchmerge.Options{Repo: "."}
	var mode, worktree string
	cmd := &cobra.Command{
		Use:    "merge <branch>",
		Short:  "Land a branch into the current checkout",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			opts := options
			opts.Branch, opts.Mode = args[0], branchmerge.Mode(mode)
			worktrees, err := git.ListWorktrees(opts.Repo)
			if err != nil {
				return err
			}
			if !slices.ContainsFunc(worktrees, func(wt git.Worktree) bool {
				return wt.Branch == opts.Branch && !wt.Primary && wt.Path == filepath.Clean(worktree)
			}) {
				return fmt.Errorf("%s is not the linked worktree for %s", worktree, opts.Branch)
			}
			switch opts.Mode {
			case branchmerge.Squash:
				if strings.TrimSpace(opts.Message) == "" {
					opts.Message = fmt.Sprintf("Squash merge branch %s", opts.Branch)
				}
			case branchmerge.Incremental:
				opts.FastForwardOnly = true
				opts.Commits, err = branchmerge.RangeCommits(opts.Repo, "HEAD", "refs/heads/"+opts.Branch)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("mode must be squash or incremental, got %q", mode)
			}
			merged, err := branchmerge.Merge(opts)
			if err != nil {
				return err
			}
			_, err = branchmerge.Cleanup(branchmerge.CleanupOptions{
				Repo: opts.Repo, Worktree: worktree, Branch: opts.Branch, LandedSHA: merged.LandedSHA, Mode: opts.Mode,
			})
			return err
		},
	}
	cmd.Flags().StringVar(&options.Base, "base", "", "Branch checked out in the target checkout")
	cmd.Flags().StringVar(&mode, "mode", "squash", "How to land the branch")
	cmd.Flags().StringVar(&options.Message, "message", "", "Squash commit message")
	cmd.Flags().StringVar(&worktree, "worktree", "", "Source worktree to remove after landing")
	cmd.Flags().BoolVar(&options.Autostash, "autostash", true, "Save and restore local changes in the target checkout")
	for _, name := range []string{"base", "worktree"} {
		if err := cmd.MarkFlagRequired(name); err != nil {
			panic(err)
		}
	}
	return cmd
}
