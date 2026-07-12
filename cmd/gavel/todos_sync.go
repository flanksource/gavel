package main

import (
	"context"
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/todosync"
	"github.com/spf13/cobra"
)

var (
	todosSyncMarkers []string
	todosSyncIgnore  []string
	todosSyncDryRun  bool
)

var todosSyncCmd = &cobra.Command{
	Use:          "sync [paths...]",
	SilenceUsage: true,
	Short:        "Sync source TODO/FIXME comments into TODO issues",
	Long: `Scan the source tree for TODO/FIXME comments and create or update a TODO issue
for each. Restrict to specific paths with positional args; change which markers
are picked up with --markers. Use --dry-run to preview the changes.`,
	Example: `  gavel todos sync
  gavel todos sync ./pkg/parser
  gavel todos sync --markers TODO,FIXME,HACK --dry-run`,
	RunE: runTodosSync,
}

func runTodosSync(cmd *cobra.Command, args []string) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	provider, err := newTodosProvider(workDir, todosDir)
	if err != nil {
		return err
	}
	result, err := todosync.SyncSourceComments(context.Background(), provider, todosync.SourceCommentSyncOptions{
		WorkDir: workDir,
		Paths:   args,
		Markers: todosSyncMarkers,
		Ignore:  todosSyncIgnore,
		DryRun:  todosSyncDryRun,
	})
	if err != nil {
		return err
	}
	fmt.Println(clicky.MustFormat(result))
	return nil
}

func init() {
	todosCmd.AddCommand(todosSyncCmd)
	todosSyncCmd.Flags().StringVar(&todosDir, "dir", "", "Deprecated; runtime TODOs are stored in PostgreSQL (must be omitted)")
	todosSyncCmd.Flags().StringSliceVar(&todosSyncMarkers, "markers", []string{"TODO", "FIXME"}, "Source comment markers to sync")
	todosSyncCmd.Flags().StringArrayVar(&todosSyncIgnore, "ignore", nil, "Additional path glob to ignore during source scan")
	todosSyncCmd.Flags().BoolVar(&todosSyncDryRun, "dry-run", false, "Report planned sync changes without creating or updating TODOs")
}
