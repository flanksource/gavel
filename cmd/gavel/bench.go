package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/formatters"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/bench"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/spf13/cobra"
)

var (
	benchRunCount    int
	benchRunTimeout  string
	benchRunBenchtime string
	benchRunOut      string
	benchRunPattern  string
	benchRunMem      bool
	benchRunExtra    []string

	benchCompareBase      string
	benchCompareHead      string
	benchCompareBaseLabel string
	benchCompareHeadLabel string
	benchCompareThreshold float64
	benchCompareUI        bool
	benchCompareOpts      clicky.FormatOptions
)

var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "Run Go benchmarks and compare base vs head results",
}

var benchRunCmd = &cobra.Command{
	Use:          "run [packages...]",
	Short:        "Run benchmarks for the given packages and write structured JSON",
	SilenceUsage: true,
	RunE:         runBenchRun,
}

var benchCompareCmd = &cobra.Command{
	Use:          "compare",
	Short:        "Compare two bench JSON files and report deltas (exits 1 on regression)",
	SilenceUsage: true,
	RunE:         runBenchCompare,
}

func runBenchRun(cmd *cobra.Command, args []string) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return err
	}

	extra := []string{
		fmt.Sprintf("-count=%d", benchRunCount),
		fmt.Sprintf("-timeout=%s", benchRunTimeout),
		fmt.Sprintf("-benchtime=%s", benchRunBenchtime),
	}
	if benchRunMem {
		extra = append(extra, "-benchmem")
	}
	extra = append(extra, benchRunExtra...)

	pattern := benchRunPattern
	if pattern == "" {
		pattern = "."
	}

	opts := testrunner.RunOptions{
		WorkDir:       workDir,
		StartingPaths: args,
		ExtraArgs:     extra,
		Bench:         pattern,
		ShowPassed:    true, // benchmark entries are Passed=true; required for them to flow through
		Recursive:     true,
		ShowStdout:    testrunner.OutputNever,
		ShowStderr:    testrunner.OutputOnFailure,
	}

	result, err := testrunner.Run(opts)
	if err != nil {
		return err
	}
	tests, _ := result.([]parsers.Test)
	runs := testsToBenchRuns(tests)
	if len(runs) == 0 {
		return fmt.Errorf("no benchmark results collected; did the packages contain Benchmark* funcs?")
	}

	data, err := json.MarshalIndent(runs, "", "  ")
	if err != nil {
		return err
	}
	if benchRunOut == "" || benchRunOut == "-" {
		fmt.Println(string(data))
		return nil
	}
	return os.WriteFile(benchRunOut, data, 0644)
}

// testsToBenchRuns flattens the hierarchical test tree, picking out benchmark leaves.
func testsToBenchRuns(tests []parsers.Test) []bench.BenchRun {
	var out []bench.BenchRun
	var walk func(t parsers.Test)
	walk = func(t parsers.Test) {
		if t.Benchmark != nil && len(t.Benchmark.Samples) > 0 {
			out = append(out, bench.BenchRun{
				Name:        t.Name,
				Package:     t.Package,
				Samples:     t.Benchmark.Samples,
				Iterations:  t.Benchmark.Iterations,
				BytesPerOp:  t.Benchmark.BytesPerOp,
				AllocsPerOp: t.Benchmark.AllocsPerOp,
				MBPerSec:    t.Benchmark.MBPerSec,
			})
		}
		for _, c := range t.Children {
			walk(c)
		}
	}
	for _, t := range tests {
		walk(t)
	}
	return out
}

func runBenchCompare(cmd *cobra.Command, args []string) error {
	if benchCompareBase == "" || benchCompareHead == "" {
		return fmt.Errorf("--base and --head are required")
	}
	base, err := loadBenchRuns(benchCompareBase)
	if err != nil {
		return fmt.Errorf("loading base: %w", err)
	}
	head, err := loadBenchRuns(benchCompareHead)
	if err != nil {
		return fmt.Errorf("loading head: %w", err)
	}

	cmp := bench.Compare(base, head, benchCompareThreshold)
	cmp.BaseLabel = benchCompareBaseLabel
	cmp.HeadLabel = benchCompareHeadLabel

	clicky.MustPrint(cmp, benchCompareOpts)
	if cmp.HasRegression {
		exitCode = 1
	}

	if benchCompareUI {
		srv := startTestUI()
		if srv != nil {
			srv.SetBenchComparison(&cmp)
			waitForInterrupt()
		}
	}
	return nil
}

func waitForInterrupt() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}

func loadBenchRuns(path string) ([]bench.BenchRun, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var runs []bench.BenchRun
	if err := json.Unmarshal(data, &runs); err != nil {
		return nil, err
	}
	return runs, nil
}

func init() {
	benchRunCmd.Flags().IntVar(&benchRunCount, "count", 6, "Number of times to run each benchmark (go test -count)")
	benchRunCmd.Flags().StringVar(&benchRunTimeout, "timeout", "20m", "Test execution timeout (go test -timeout)")
	benchRunCmd.Flags().StringVar(&benchRunBenchtime, "benchtime", "1s", "Duration per benchmark (go test -benchtime)")
	benchRunCmd.Flags().StringVar(&benchRunOut, "out", "", "Write JSON results to this file (default: stdout)")
	benchRunCmd.Flags().StringVar(&benchRunPattern, "pattern", ".", "Benchmark name regex (go test -bench)")
	benchRunCmd.Flags().BoolVar(&benchRunMem, "benchmem", true, "Include memory allocation stats (-benchmem)")
	benchRunCmd.Flags().StringSliceVar(&benchRunExtra, "extra", nil, "Extra flags passed through to go test")

	benchCompareCmd.Flags().StringVar(&benchCompareBase, "base", "", "Path to base bench JSON file (required)")
	benchCompareCmd.Flags().StringVar(&benchCompareHead, "head", "", "Path to head bench JSON file (required)")
	benchCompareCmd.Flags().StringVar(&benchCompareBaseLabel, "base-label", "base", "Display label for the base run")
	benchCompareCmd.Flags().StringVar(&benchCompareHeadLabel, "head-label", "head", "Display label for the head run")
	benchCompareCmd.Flags().Float64Var(&benchCompareThreshold, "threshold", bench.DefaultThreshold, "Regression threshold in percent")
	benchCompareCmd.Flags().BoolVar(&benchCompareUI, "ui", false, "Launch browser UI with the comparison preloaded")
	formatters.BindPFlags(benchCompareCmd.Flags(), &benchCompareOpts)

	benchCmd.AddCommand(benchRunCmd, benchCompareCmd)
	rootCmd.AddCommand(benchCmd)
}
