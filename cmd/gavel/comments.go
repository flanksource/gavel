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

// listCommentsAtPath makes `comments list [path]` list the threads of that
// file or directory, defaulting to the current directory's project when no
// scope or path is given. Clicky's generated list passes only flags, so the
// positional path becomes --path.
func listCommentsAtPath(cmd *cobra.Command, args []string) error {
	path := cmd.Flags().Lookup("path")
	if path == nil {
		return fmt.Errorf("%s has no --path flag", cmd.CommandPath())
	}
	switch {
	case len(args) == 1 && path.Changed:
		return fmt.Errorf("pass the path as an argument or --path, not both")
	case len(args) == 1:
		return cmd.Flags().Set("path", args[0])
	case !path.Changed && !cmd.Flags().Changed("scope"):
		return cmd.Flags().Set("path", ".")
	}
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
	command.PreRunE = listCommentsAtPath
	list, _, err := command.Find([]string{"list"})
	if err != nil {
		panic(err)
	}
	list.Use, list.Args, list.PreRunE = "list [path]", cobra.MaximumNArgs(1), listCommentsAtPath
	lintCommand, _, err := generated.Find([]string{"comment-lint"})
	if err != nil {
		panic(err)
	}
	generated.RemoveCommand(lintCommand)
	rootCmd.AddCommand(lintCommand)
	command.AddCommand(commentsDeliverCommand(), commentsAgentsCommand())
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
