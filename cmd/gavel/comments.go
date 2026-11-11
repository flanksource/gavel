package main

import (
	"fmt"
	"os"
	"time"

	"github.com/flanksource/clicky"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/spf13/cobra"
)

// actAsAssertedCommentActor runs the root's setup, then acts as the actor
// GAVEL_COMMENT_ACTOR asserts (an agent replying to a thread it was sent).
// Cobra runs only the nearest PersistentPreRunE, so the root's is chained here.
func actAsAssertedCommentActor(cmd *cobra.Command, args []string) error {
	if err := rootCmd.PersistentPreRunE(cmd, args); err != nil {
		return err
	}
	value := os.Getenv(commentapi.ActorEnv)
	if value == "" {
		return nil
	}
	actor, err := commentapi.DecodeActor(value)
	if err != nil {
		return fmt.Errorf("%s: %w", commentapi.ActorEnv, err)
	}
	cmd.SetContext(commentapi.WithActor(cmd.Context(), actor))
	return nil
}

func init() {
	if err := ui.RegisterComments(); err != nil {
		panic(err)
	}
	generated := &cobra.Command{Use: "gavel"}
	clicky.GenerateCLI(generated)
	command, _, err := generated.Find([]string{"comments"})
	if err != nil {
		panic(err)
	}
	generated.RemoveCommand(command)
	rootCmd.AddCommand(command)
	command.PersistentPreRunE = actAsAssertedCommentActor
	lintCommand, _, err := generated.Find([]string{"comment-lint"})
	if err != nil {
		panic(err)
	}
	generated.RemoveCommand(lintCommand)
	rootCmd.AddCommand(lintCommand)
	command.AddCommand(commentsDeliverCommand(), commentsRoutesCommand())
	command.AddCommand(&cobra.Command{Use: "dispatch", Short: "Process one explicitly admitted native publication for the local tenant", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		service, err := ui.OpenComments(cmd.Context())
		if err != nil {
			return err
		}
		op, err := service.ProcessProvider(cmd.Context(), time.Minute)
		if err != nil {
			return err
		}
		if op.State != "done" {
			return fmt.Errorf("publication %s is %s: %s", op.ID, op.State, op.Error)
		}
		return nil
	}})
}
