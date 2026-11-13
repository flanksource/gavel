package main

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func runDashboardCommit(ctx context.Context, args []string, output io.Writer) error {
	if len(args) == 0 || args[0] != "commit" {
		return fmt.Errorf("expected a commit command, received %q", args)
	}
	parent := &cobra.Command{Use: "gavel", SilenceUsage: true, SilenceErrors: true}
	cmd := registerCommitCommand(parent, func(opts CommitOptions) (any, error) {
		result, err := executeCommit(ctx, opts, output)
		if result != nil {
			if _, writeErr := fmt.Fprintln(output, result.Pretty().String()); writeErr != nil {
				return nil, writeErr
			}
		}
		return nil, err
	})
	cmd.SetOut(output)
	cmd.SetErr(output)
	parent.SetArgs(args)
	return parent.ExecuteContext(ctx)
}
