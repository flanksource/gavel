package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/todos/portable"
	"github.com/spf13/cobra"
)

type TodosImportOptions struct {
	Files []string `args:"true"`
	Dir   string   `flag:"dir" default:".todos" help:"Directory to read when no files are supplied"`
}

type TodosExportOptions struct {
	Refs  []string `args:"true"`
	Dir   string   `flag:"dir" default:".todos" help:"Directory to write exported Markdown files"`
	Force bool     `flag:"force" help:"Replace an unrelated file at an export path"`
}

var todosImportCmd *cobra.Command
var todosExportCmd *cobra.Command

func init() {
	todosImportCmd = clicky.AddNamedCommandWithContext("import", todosCmd, TodosImportOptions{}, func(ctx context.Context, opts TodosImportOptions) (any, error) {
		return nil, runTodosImport(ctx, opts)
	})
	todosImportCmd.Short = "Import .todos Markdown into native PostgreSQL TODOs"
	todosExportCmd = clicky.AddNamedCommandWithContext("export", todosCmd, TodosExportOptions{}, func(ctx context.Context, opts TodosExportOptions) (any, error) {
		return nil, runTodosExport(ctx, opts)
	})
	todosExportCmd.Short = "Export native PostgreSQL TODOs as .todos Markdown"
}

func runTodosImport(ctx context.Context, opts TodosImportOptions) error {
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
	result, err := portable.Import(ctx, db, project.WorkspaceOptions(), opts.Dir, opts.Files)
	if err != nil {
		return err
	}
	fmt.Printf("Imported %d created, %d updated, %d unchanged TODOs from %s\n",
		result.Created, result.Updated, result.Unchanged, result.Directory)
	return nil
}

func runTodosExport(ctx context.Context, opts TodosExportOptions) error {
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
	result, err := portable.Export(ctx, db, project.WorkspaceOptions(), opts.Dir, opts.Refs, opts.Force)
	if err != nil {
		return err
	}
	noun := "TODOs"
	if result.Exported == 1 {
		noun = "TODO"
	}
	fmt.Printf("Exported %d %s to %s\n", result.Exported, noun, strings.TrimSpace(result.Directory))
	return nil
}
