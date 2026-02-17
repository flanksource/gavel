package main

import (
	"fmt"
	"os"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/types"
	"github.com/spf13/cobra"
)

var fixturesCmd = &cobra.Command{
	Use:          "fixtures [fixture-files...]",
	Short:        "Run fixture-based tests from markdown tables and command blocks",
	Args:         cobra.MinimumNArgs(1),
	RunE:         runFixtures,
	SilenceUsage: true,
}

func runFixtures(cmd *cobra.Command, args []string) error {
	wd, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	runner, err := fixtures.NewRunner(fixtures.RunnerOptions{
		Paths:          args,
		Format:         clicky.Flags.ResolveFormat(),
		NoColor:        clicky.Flags.NoColor,
		WorkDir:        wd,
		MaxWorkers:     clicky.Flags.MaxConcurrent,
		Logger:         logger.StandardLogger(),
		ExecutablePath: executablePath,
	})
	if err != nil {
		return fmt.Errorf("failed to create fixture runner: %w", err)
	}

	return runner.Run()
}

func init() {
	rootCmd.AddCommand(fixturesCmd)
}
