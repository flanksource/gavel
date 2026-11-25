package main

import (
	"slices"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// finishedRun is a run saved on its own, outside a successful run's sha-keyed
// snapshot: one the runner returned an error for, with what completed, a
// terminal fixtures run, or a UI rerun.
type finishedRun struct {
	Opts        testrunner.RunOptions
	Tests       []parsers.Test
	Lint        []*linters.LinterResult
	Started     time.Time
	Trace       *runTrace
	Performance *fixtures.BenchmarkPerformance
	Err         error
}

// saveFinishedRun writes run's per-run snapshot, carrying its error and exit
// code 1 when it has one, and returns it.
func saveFinishedRun(run finishedRun) testui.Snapshot {
	snapshot := buildTestSnapshot(run.Opts, run.Tests, run.Lint, run.Started, time.Now().UTC(), run.Trace)
	snapshot.Performance = run.Performance
	if run.Opts.RunKind != "" {
		snapshot.Metadata.Kind = run.Opts.RunKind
	}
	if run.Err != nil {
		snapshot.Error = run.Err.Error()
		failed := 1
		snapshot.Metadata.ExitCode = &failed
	}
	persistPerRunSnapshot(run.Opts.Dirs.StateRoot, &snapshot, run.Started)
	return snapshot
}

// saveFixtureRerun saves a UI rerun as the run on screen, before with the
// rerun's tests over it, so any fixture step of it can be resumed. It returns
// the saved run's id, empty for a run without fixtures.
func saveFixtureRerun(before []parsers.Test, run finishedRun) string {
	run.Tests = testui.MergeRerunTests(before, run.Tests)
	if !slices.ContainsFunc(run.Tests, holdsFixture) {
		return ""
	}
	return saveFinishedRun(run).Metadata.SavedRun
}

func holdsFixture(test parsers.Test) bool {
	return test.Framework == parsers.Fixture || slices.ContainsFunc(test.Children, holdsFixture)
}

// persistRunSnapshot writes a completed run as both the per-run timestamped
// snapshot and the sha-keyed snapshot (refreshing last.json, which a bare
// --failed resolves).
func persistRunSnapshot(stateRoot string, snapshot *testui.Snapshot, started time.Time) {
	persistPerRunSnapshot(stateRoot, snapshot, started)
	if path, err := snapshots.Save(stateRoot, snapshot); err != nil {
		logger.Warnf("persist snapshot: %v", err)
	} else {
		logger.V(1).Infof("wrote snapshot to %s", path)
	}
}

func persistPerRunSnapshot(stateRoot string, snapshot *testui.Snapshot, started time.Time) {
	if path, err := snapshots.SavePerRun(stateRoot, snapshot, started, ""); err != nil {
		logger.Warnf("persist per-run snapshot: %v", err)
	} else {
		logger.V(1).Infof("wrote per-run snapshot to %s", path)
	}
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
			Rerun:        opts.FixtureCommand,
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
