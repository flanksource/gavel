package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/flanksource/clicky"
	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/commons/properties"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/record"
	"github.com/flanksource/gavel/verify"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func runFixtures(cmd *cobra.Command, args []string) error {
	if fixturesSchema {
		return writeFixturesSchema(os.Stdout)
	}

	wd, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	cfg, err := verify.LoadGavelConfig(wd)
	if err != nil {
		return fmt.Errorf("load .gavel.yaml: %w", err)
	}

	recordSpec, err := record.Parse(fixturesRecord)
	if err != nil {
		return fmt.Errorf("--record: %w", err)
	}
	limits := fixtures.BenchmarkLimits{MaxDurationMS: float64(fixturesMaxTime) / 1e6,
		MaxSQLDurationMS: float64(fixturesMaxSQLTime) / 1e6, MaxSQLQueryMS: float64(fixturesMaxSQLQuery) / 1e6}
	for _, item := range []struct {
		name  string
		value float64
	}{{"max-time", limits.MaxDurationMS}, {"max-sql-time", limits.MaxSQLDurationMS}, {"max-sql-query", limits.MaxSQLQueryMS}} {
		if cmd.Flags().Changed(item.name) && item.value <= 0 {
			return fmt.Errorf("--%s requires a positive duration", item.name)
		}
	}
	if cmd.Flags().Changed("max-deviation-pct") {
		limits.MaxDeviationPct = &fixturesMaxDeviationPct
	}
	if cmd.Flags().Changed("max-sql-queries") {
		limits.MaxSQLQueries = &fixturesMaxSQLQueries
	}
	if cmd.Flags().Changed("max-slow-sql") {
		limits.MaxSlowSQL = &fixturesMaxSlowSQL
	}
	for _, item := range []struct {
		flag  string
		value string
		dest  *uint64
	}{{"max-memory", fixturesMaxMemory, &limits.MaxRSSBytes}, {"max-io-read", fixturesMaxIORead, &limits.MaxDiskReadBytes}, {"max-io-write", fixturesMaxIOWrite, &limits.MaxDiskWriteBytes}} {
		if item.value == "" {
			continue
		}
		bytes, err := properties.ParseBytes(item.value)
		if err != nil || bytes <= 0 {
			return fmt.Errorf("--%s requires a positive byte size: %q", item.flag, item.value)
		}
		*item.dest = uint64(bytes)
	}

	runnerOpts := fixtures.RunnerOptions{
		Paths:          args,
		ResolveSpec:    fixtureSpecResolver(wd, cfg),
		Format:         clicky.Flags.ResolveFormat(),
		NoColor:        clicky.Flags.NoColor,
		WorkDir:        wd,
		MaxWorkers:     fixtureMaxWorkers(cmd),
		Logger:         logger.StandardLogger(),
		ExecutablePath: executablePath,
		UpdateGolden:   fixturesUpdateGolden,
		Record:         recordSpec,
		Benchmark:      fixturesBenchmark,
		Baseline:       fixturesBaseline,
		Profile:        fixturesProfile,
		Limits:         limits,
		Display: lo.ToPtr(fixtures.DisplayOptionsForVerbosity(clicky.Flags.LevelCount, fixtures.DisplayOptions{
			ShowPassed: fixturesShowPassed,
			ShowStdout: fixtures.ParseOutputMode(fixturesShowStdout),
			ShowStderr: fixtures.ParseOutputMode(fixturesShowStderr),
		}, cmd.Flags().Changed("show-stdout"), cmd.Flags().Changed("show-stderr"))),
	}
	if fixturesUI.UI {
		opts, detach := fixtureUIRunOptions(fixtureUIRunRequest{Runner: &runnerOpts, UI: fixturesUI})
		_, err := runTests(opts, detach)
		return err
	}

	runner, err := fixtures.NewRunner(runnerOpts)
	if err != nil {
		return fmt.Errorf("failed to create fixture runner: %w", err)
	}

	clickytask.SetLiveRenderer(fixtureLiveRenderer{})
	defer clickytask.SetLiveRenderer(nil)
	tree, runErr := runner.Run()
	if tree != nil {
		if runner.BenchmarkReport() != nil && strings.EqualFold(runnerOpts.Format, "json") {
			data, err := os.ReadFile(runner.BenchmarkReport().Path)
			if err != nil {
				return fmt.Errorf("read fixture benchmark snapshot: %w", err)
			}
			if _, err := os.Stdout.Write(append(data, '\n')); err != nil {
				return fmt.Errorf("write fixture benchmark snapshot: %w", err)
			}
		} else if len(tree.Children) == 1 {
			fmt.Println(clicky.MustFormat(*tree.Children[0]))
		} else {
			fmt.Println(clicky.MustFormat(*tree))
		}
		if runner.BenchmarkReport() != nil && !strings.EqualFold(runnerOpts.Format, "json") {
			fmt.Println(clicky.MustFormat(*runner.BenchmarkReport()))
		}
	}
	return runErr
}

// fixtureMaxWorkers resolves fixture parallelism from --max-concurrent, but only
// when it was actually typed. That flag is clicky's process-wide task default
// (4); inheriting it silently would make `gavel fixtures run` parallel while
// `gavel test` runs the same files serially, which is how a suite pinned to one
// worker ends up racing depending on which entry point invoked it.
func fixtureMaxWorkers(cmd *cobra.Command) int {
	if cmd.Flags().Changed("max-concurrent") {
		return clicky.Flags.MaxConcurrent
	}
	return 0
}
