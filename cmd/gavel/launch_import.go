package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/flanksource/gavel/resultstore"
	"github.com/spf13/cobra"
)

func launchImportCommand() *cobra.Command {
	var root, from, to string
	command := &cobra.Command{Use: "import", Short: "Copy terminal executions to another result adapter in the same namespace", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) (result error) {
		if from == to {
			return errors.New("source and destination adapters must differ")
		}
		dir := filepath.Join(root, ".gavel", "runs")
		source, err := openLaunchStore(cmd.Context(), from, dir)
		if err != nil {
			return err
		}
		defer func() { result = errors.Join(result, source.Close()) }()
		target, err := openLaunchStore(cmd.Context(), to, dir)
		if err != nil {
			return err
		}
		defer func() { result = errors.Join(result, target.Close()) }()
		runs, err := source.List(cmd.Context(), resultstore.ListOptions{})
		if err != nil {
			return err
		}
		for _, run := range runs {
			if !run.Terminal() {
				return fmt.Errorf("execution %s is still active; stop it before importing", run.Metadata.ExecutionID)
			}
		}
		count := 0
		for _, run := range runs {
			created, err := target.Import(cmd.Context(), run)
			if err != nil {
				return err
			}
			if created {
				count++
			}
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Imported %d terminal executions from %s to %s\n", count, from, to)
		return err
	}}
	command.Flags().StringVar(&root, "root", ".", "Launch storage namespace root")
	command.Flags().StringVar(&from, "from", "", "Source adapter: json or sqlite")
	command.Flags().StringVar(&to, "to", "", "Destination adapter: json or sqlite")
	return command
}
