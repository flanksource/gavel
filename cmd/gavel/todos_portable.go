package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/todos/portable"
	"github.com/spf13/cobra"
)

var (
	todosImportDirectory string
	todosExportDirectory string
	todosExportForce     bool
)

var todosImportCmd = &cobra.Command{
	Use:          "import [todo-files...]",
	Short:        "Import .todos Markdown into native PostgreSQL TODOs",
	SilenceUsage: true,
	Args:         cobra.ArbitraryArgs,
	Long: `Explicitly import portable .todos Markdown into the current native PostgreSQL
workspace. With no file arguments, every valid Markdown TODO under --dir is
imported. This command does not select or enable a runtime file provider.`,
	Example: `  gavel todos import
  gavel todos import --dir ./archive/todos
  gavel todos import .todos/fix-parser.md .todos/add-retry.md`,
	RunE: runTodosImport,
}

var todosExportCmd = &cobra.Command{
	Use:          "export [todo-refs...]",
	Short:        "Export native PostgreSQL TODOs as .todos Markdown",
	SilenceUsage: true,
	Args:         cobra.ArbitraryArgs,
	Long: `Explicitly export portable fields from the current native PostgreSQL workspace
as .todos Markdown. With no references, every issue in the workspace is
exported. This command never changes runtime TODO storage.`,
	Example: `  gavel todos export
  gavel todos export --dir ./backup/todos
  gavel todos export 3f2a1b "Fix parser"`,
	RunE: runTodosExport,
}

func init() {
	todosCmd.AddCommand(todosImportCmd, todosExportCmd)
	todosImportCmd.Flags().StringVar(&todosImportDirectory, "dir", portable.DefaultDirectory, "Directory to read when no files are supplied")
	todosExportCmd.Flags().StringVar(&todosExportDirectory, "dir", portable.DefaultDirectory, "Directory to write exported Markdown files")
	todosExportCmd.Flags().BoolVar(&todosExportForce, "force", false, "Replace an unrelated file at an export path")
}

func runTodosImport(command *cobra.Command, files []string) error {
	ctx := command.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("resolve portable TODO import workspace: %w", err)
	}
	project, err := ui.ProjectForDir(workDir)
	if err != nil {
		return err
	}
	db, err := database.Require(ctx, "gavel todos import")
	if err != nil {
		return err
	}
	result, err := portable.Import(ctx, db, project.WorkspaceOptions(), todosImportDirectory, files)
	if err != nil {
		return err
	}
	fmt.Fprintf(command.OutOrStdout(), "Imported %d created, %d updated, %d unchanged TODOs from %s\n",
		result.Created, result.Updated, result.Unchanged, result.Directory)
	return nil
}

func runTodosExport(command *cobra.Command, refs []string) error {
	ctx := command.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("resolve portable TODO export workspace: %w", err)
	}
	project, err := ui.ProjectForDir(workDir)
	if err != nil {
		return err
	}
	db, err := database.Require(ctx, "gavel todos export")
	if err != nil {
		return err
	}
	result, err := portable.Export(ctx, db, project.WorkspaceOptions(), todosExportDirectory, refs, todosExportForce)
	if err != nil {
		return err
	}
	noun := "TODOs"
	if result.Exported == 1 {
		noun = "TODO"
	}
	fmt.Fprintf(command.OutOrStdout(), "Exported %d %s to %s\n", result.Exported, noun, strings.TrimSpace(result.Directory))
	return nil
}
