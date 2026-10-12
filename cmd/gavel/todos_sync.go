package main

import (
	"context"
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/todosync"
	"github.com/spf13/cobra"
)

type TodosSyncOptions struct {
	Paths   []string `args:"true"`
	Markers []string `flag:"markers" default:"TODO,FIXME" help:"Source comment markers to sync"`
	Ignore  []string `flag:"ignore" help:"Additional path glob to ignore during source scan"`
	DryRun  bool     `flag:"dry-run" help:"Report planned sync changes without updating TODOs"`
}

var todosSyncCmd *cobra.Command

func runTodosSync(opts TodosSyncOptions) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}
	result, err := todosync.SyncSourceComments(context.Background(), provider, todosync.SourceCommentSyncOptions{
		WorkDir: workDir,
		Paths:   opts.Paths,
		Markers: opts.Markers,
		Ignore:  opts.Ignore,
		DryRun:  opts.DryRun,
	})
	if err != nil {
		return err
	}
	fmt.Println(clicky.MustFormat(result))
	return nil
}

func init() {
	todosSyncCmd = clicky.AddNamedCommand("sync", todosCmd, TodosSyncOptions{}, func(opts TodosSyncOptions) (any, error) {
		return nil, runTodosSync(opts)
	})
	todosSyncCmd.Short = "Sync source TODO/FIXME comments into TODO issues"
}
