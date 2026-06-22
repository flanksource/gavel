package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/todos"
	"github.com/spf13/cobra"
)

var transferToProject string

var todosTransferCmd = &cobra.Command{
	Use:          "transfer <ref> --to <project>",
	SilenceUsage: true,
	Short:        "Move a TODO from the current workspace to another project",
	Args:         cobra.ExactArgs(1),
	RunE:         runTodosTransfer,
}

func runTodosTransfer(_ *cobra.Command, args []string) error {
	if transferToProject == "" {
		return fmt.Errorf("--to <project> is required")
	}
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	target, ok := ui.GetProject(transferToProject)
	if !ok {
		return fmt.Errorf("%w: %q", ui.ErrProjectNotFound, transferToProject)
	}
	targetDir, err := filepath.Abs(target.ResolvedDir())
	if err != nil {
		return fmt.Errorf("resolve target project dir: %w", err)
	}
	if targetDir == filepath.Clean(workDir) {
		return fmt.Errorf("project %q points at the current workspace; nothing to transfer", transferToProject)
	}

	source, err := newTodosProvider(workDir, todosDir)
	if err != nil {
		return err
	}

	created, err := todos.Transfer(context.Background(), source, ui.ProviderForProject(target), args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Moved %q to project %q (%s)\n\n", created.Title, target.Name, targetDir)
	fmt.Println(created.PrettyDetailed().ANSI())
	return nil
}

func init() {
	todosCmd.AddCommand(todosTransferCmd)
	todosTransferCmd.Flags().StringVar(&transferToProject, "to", "", "Target project name (from `gavel projects`)")
	todosTransferCmd.Flags().StringVar(&todosDir, "dir", "", "Source TODOs directory (default: .todos)")
}
