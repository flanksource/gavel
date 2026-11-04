package main

import (
	"fmt"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/spf13/cobra"
)

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
	lintCommand, _, err := generated.Find([]string{"comment-lint"})
	if err != nil {
		panic(err)
	}
	generated.RemoveCommand(lintCommand)
	rootCmd.AddCommand(lintCommand)
	command.AddCommand(&cobra.Command{Use: "provision", Short: "Provision empty comments tables in the configured database", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		db, err := database.Shared(cmd.Context())
		if err != nil {
			return err
		}
		if db.Disabled() {
			return fmt.Errorf("comments provisioning requires a configured PostgreSQL database")
		}
		return postgres.Provision(cmd.Context(), db.DSN())
	}})
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
