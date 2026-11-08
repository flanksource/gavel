package main

import (
	"time"

	"github.com/flanksource/gavel/runstatus"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

type recorderTee struct {
	Recorder *snapshots.Recorder
	// Reporter publishes the run's test counts in its status.
	Reporter   *runstatus.Reporter
	Options    testrunner.RunOptions
	Started    time.Time
	Downstream chan<- []parsers.Test
}

// attachRecorderTee inserts the recorder and status reporter between the
// runner and any downstream consumer of opts.Updates. Each batch of test
// results from the runner is funneled into a fresh Snapshot
// (Status.Running=true) and fed to the recorder, and its counts to the
// reporter. The batch is also forwarded unchanged to downstream so the UI
// server (when --ui) keeps receiving its updates.
//
// Returns the new send channel that should replace opts.Updates. When the
// runner closes the new channel (single-root via Done(); multi-root via the
// deferred close in runner.go:413), the goroutine drains and closes the
// downstream channel too, preserving existing close semantics.
func attachRecorderTee(t recorderTee) chan<- []parsers.Test {
	tee := make(chan []parsers.Test, 16)
	go func() {
		defer func() {
			if t.Downstream != nil {
				close(t.Downstream)
			}
		}()
		for batch := range tee {
			t.Recorder.Update(buildRunningSnapshot(t.Options, batch, t.Started))
			t.Reporter.SetTests(batch)
			if t.Downstream != nil {
				t.Downstream <- batch
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
		Git: snapshots.GitInfo(opts.WorkDir),
		Status: testui.SnapshotStatus{
			Running: true,
			LintRun: false,
		},
		Tests: tests,
	}
}
