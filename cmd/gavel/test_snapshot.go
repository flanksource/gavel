package main

import (
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// failedRun is a run the runner returned an error for, with what completed.
type failedRun struct {
	Opts        testrunner.RunOptions
	Tests       []parsers.Test
	Lint        []*linters.LinterResult
	Started     time.Time
	Trace       *runTrace
	Performance *fixtures.BenchmarkPerformance
	Err         error
}

// saveFailedRun writes run's per-run snapshot, carrying its error and exit
// code 1, and returns it.
func saveFailedRun(run failedRun) testui.Snapshot {
	snapshot := buildTestSnapshot(run.Opts, run.Tests, run.Lint, run.Started, time.Now().UTC(), run.Trace)
	snapshot.Performance = run.Performance
	snapshot.Error = run.Err.Error()
	failed := 1
	snapshot.Metadata.ExitCode = &failed
	if path, err := snapshots.SavePerRun(run.Opts.Dirs.StateRoot, &snapshot, run.Started, ""); err != nil {
		logger.Warnf("persist per-run snapshot: %v", err)
	} else {
		logger.V(1).Infof("wrote per-run snapshot to %s", path)
	}
	return snapshot
}

func buildTestSnapshot(
	opts testrunner.RunOptions,
	tests []parsers.Test,
	lint []*linters.LinterResult,
	started time.Time,
	ended time.Time,
	trace *runTrace,
) testui.Snapshot {
	if started.IsZero() {
		started = time.Now().UTC()
	}
	if ended.IsZero() {
		ended = time.Now().UTC()
	}
	return testui.Snapshot{
		Metadata: &testui.SnapshotMetadata{
			Version:      version,
			Started:      started.UTC(),
			Ended:        ended.UTC(),
			Kind:         "initial",
			Sequence:     1,
			Args:         snapshotArgs(opts),
			Changes:      opts.Selection.Snapshot(),
			RunID:        trace.runID(),
			StateRoot:    opts.Dirs.StateRoot,
			ProcessTrace: trace.processRefs(),
		},
		Git: snapshots.GitInfo(opts.WorkDir),
		Status: testui.SnapshotStatus{
			Running: false,
			LintRun: opts.Lint,
		},
		Tests: tests,
		Lint:  lint,
	}
}

func installTimeoutDiagnosticsHook(reporter *runDiagnosticsReporter) {
	testrunner.SetCaptureGlobalDiagnostics(reporter.Capture)
}

func snapshotArgs(opts testrunner.RunOptions) map[string]any {
	args := map[string]any{
		"starting_paths": append([]string(nil), opts.StartingPaths...),
		"extra_args":     append([]string(nil), opts.ExtraArgs...),
		"tags":           append([]string(nil), opts.Tags...),
		"runner_args":    append([]string(nil), opts.PassThroughArgs...),
		"show_passed":    opts.ShowPassed,
		"ignore":         append([]string(nil), opts.Ignore...),
		"show_stdout":    string(opts.ShowStdout),
		"show_stderr":    string(opts.ShowStderr),
		"work_dir":       opts.WorkDir,
		"dry_run":        opts.DryRun,
		"recursive":      opts.Recursive,
		"nodes":          opts.Nodes,
		"concurrency":    opts.Concurrency,
		"ui":             opts.UI,
		"addr":           opts.Addr,
		"diagnostics":    opts.Diagnostics,
		"process_facets": opts.ProcessFacets,
		"skip_hooks":     opts.SkipHooks,
		"auto_stop":      durationString(opts.AutoStop),
		"idle_timeout":   durationString(opts.IdleTimeout),
		"timeout":        durationString(opts.Timeout),
		"lint_timeout":   durationString(opts.LintTimeout),
		"test_timeout":   durationString(opts.TestTimeout),
		"lint":           opts.Lint,
		"cache":          opts.Cache,
		"changed":        opts.Changed,
		"since":          opts.Since,
		"changed_files":  append([]string(nil), opts.ChangedFiles...),
		"bench":          opts.Bench,
		"fixtures":       opts.Fixtures,
		"fixture_files":  append([]string(nil), opts.FixtureFiles...),
	}
	if opts.FixtureRunner != nil {
		args["fixture_benchmark"] = opts.FixtureRunner.Benchmark
		args["fixture_baseline"] = opts.FixtureRunner.Baseline
		args["fixture_profile"] = opts.FixtureRunner.Profile
	}
	return args
}

func fixtureBenchmarkMode(opts testrunner.RunOptions) string {
	if opts.FixtureRunner == nil {
		return ""
	}
	if opts.FixtureRunner.Benchmark {
		return "benchmark"
	}
	if opts.FixtureRunner.Profile {
		return "profile"
	}
	return ""
}

func durationString(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	return d.String()
}
