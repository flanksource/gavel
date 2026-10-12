package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/todos"
	"github.com/spf13/cobra"
)

type TodosTransferOptions struct {
	TodoTargetOptions
	To string `flag:"to" required:"true" help:"Target project name (from gavel projects)"`
}

var todosTransferCmd *cobra.Command

func runTodosTransfer(opts TodosTransferOptions) error {
	ref, err := opts.One()
	if err != nil {
		return err
	}
	if opts.To == "" {
		return fmt.Errorf("--to <project> is required")
	}
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	target, err := ui.GetProject(opts.To)
	if err != nil {
		return err
	}
	targetDir, err := filepath.Abs(target.ResolvedDir())
	if err != nil {
		return fmt.Errorf("resolve target project dir: %w", err)
	}
	if targetDir == filepath.Clean(workDir) {
		return fmt.Errorf("project %q points at the current workspace; nothing to transfer", opts.To)
	}
	source, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}
	targetProvider, err := openRuntimeTodosProvider(context.Background(), targetDir)
	if err != nil {
		return fmt.Errorf("open target native TODO workspace %q: %w", target.Name, err)
	}

	created, err := todos.Transfer(context.Background(), source, targetProvider, ref)
	if err != nil {
		return err
	}

	fmt.Printf("Moved %q to project %q (%s)\n\n", created.Title, target.Name, targetDir)
	fmt.Println(created.PrettyDetailed().ANSI())
	return nil
}

func init() {
	todosTransferCmd = clicky.AddNamedCommand("transfer", todosCmd, TodosTransferOptions{}, func(opts TodosTransferOptions) (any, error) {
		return nil, runTodosTransfer(opts)
	})
	todosTransferCmd.Short = "Move a TODO from the current workspace to another project"
}
