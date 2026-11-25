package main

import (
	"encoding/json"
	"errors"

	"github.com/flanksource/gavel/git/branchrebase"
	"github.com/spf13/cobra"
)

func newGitRebaseCommand() *cobra.Command {
	root := &cobra.Command{Use: "rebase", Short: "Rebase a worktree while preserving local edits"}
	for _, action := range []string{"start", "check", "finish", "abort"} {
		var options branchrebase.Options
		cmd := &cobra.Command{Use: action, Short: action + " a recoverable worktree rebase", RunE: func(cmd *cobra.Command, _ []string) error {
			options.Action = action
			result, err := branchrebase.Run(cmd.Context(), options)
			return errors.Join(err, json.NewEncoder(cmd.OutOrStdout()).Encode(result))
		}}
		cmd.Flags().StringVar(&options.Worktree, "worktree", "", "Worktree to operate in (required)")
		cmd.Flags().StringVar(&options.Branch, "branch", "", "Expected source branch (start only)")
		cmd.Flags().StringVar(&options.Target, "target", "", "Full local ref or refs/remotes/origin/main (start only)")
		cmd.Flags().StringVar(&options.ExpectedHead, "expected-head", "", "Expected source commit (start only)")
		cmd.Flags().StringVar((*string)(&options.Strategy), "strategy", string(branchrebase.StrategyRebase), "rebase onto the target, or merge it in (start only)")
		cmd.Flags().BoolVar(&options.RebaseMerges, "rebase-merges", false, "Keep merge commits when rebasing (start only)")
		cmd.Flags().BoolVar(&options.NoFastForward, "no-ff", false, "Record a merge commit even when the branch could fast-forward (start only)")
		cmd.Flags().StringVar(&options.Operation, "operation", "", "Operation ID returned by a failed start")
		cmd.Flags().BoolVar(&options.Resolved, "resolved", false, "Confirm saved edits were manually restored and all restoration conflicts resolved")
		root.AddCommand(cmd)
	}
	return root
}
