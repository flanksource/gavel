package main

import (
	"time"

	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// attachRecorderTee inserts the recorder between the runner and any
// downstream consumer of opts.Updates. Each batch of test results from the
// runner is funneled into a fresh Snapshot (Status.Running=true) and fed to
// the recorder. The batch is also forwarded unchanged to downstream so the
// UI server (when --ui) keeps receiving its updates.
//
// Returns the new send channel that should replace opts.Updates. When the
// runner closes the new channel (single-root via Done(); multi-root via the
// deferred close in runner.go:413), the goroutine drains and closes the
// downstream channel too, preserving existing close semantics.
func attachRecorderTee(
	rec *snapshots.Recorder,
	opts testrunner.RunOptions,
	runStarted time.Time,
	downstream chan<- []parsers.Test,
) chan<- []parsers.Test {
	tee := make(chan []parsers.Test, 16)
	go func() {
		defer func() {
			if downstream != nil {
				close(downstream)
			}
		}()
		for batch := range tee {
			rec.Update(buildRunningSnapshot(opts, batch, runStarted))
			if downstream != nil {
				downstream <- batch
			}
		}
	}()
	return tee
}

// buildRunningSnapshot constructs an in-flight snapshot from a streaming
// test batch. Marked Status.Running=true so consumers (the future PR
// watcher, IDE plugins, etc.) can distinguish a live snapshot from
// last.json's settled state.
func buildRunningSnapshot(
	opts testrunner.RunOptions,
	tests []parsers.Test,
	runStarted time.Time,
) *testui.Snapshot {
	return &testui.Snapshot{
		Metadata: &testui.SnapshotMetadata{
			Version:  version,
			Started:  runStarted.UTC(),
			Kind:     "initial",
			Sequence: 1,
			Args:     snapshotArgs(opts),
		},
		Git: snapshotGitInfo(opts.WorkDir),
		Status: testui.SnapshotStatus{
			Running: true,
			LintRun: false,
		},
		Tests: tests,
	}
}
