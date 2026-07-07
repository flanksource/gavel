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
	RunE:         runTodosSync,
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
	todosSyncCmd.Flags().StringVar(&todosDir, "dir", "", "TODOs directory (default: .todos; only with --provider=todos)")
	todosSyncCmd.Flags().StringSliceVar(&todosSyncMarkers, "markers", []string{"TODO", "FIXME"}, "Source comment markers to sync")
	todosSyncCmd.Flags().StringArrayVar(&todosSyncIgnore, "ignore", nil, "Additional path glob to ignore during source scan")
	todosSyncCmd.Flags().BoolVar(&todosSyncDryRun, "dry-run", false, "Report planned sync changes without creating or updating TODOs")
}
