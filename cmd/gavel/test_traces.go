package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/flanksource/commons-db/cmd/query/traceserve"
	"github.com/flanksource/gavel/testrunner"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/traces"
)

// runTrace is a run's trace host: its store at .gavel/traces/<runID>, shared by
// the fixtures it runs, and, under --ui or --diagnostics, the process capture
// rooted at gavel itself that the Processes tab reads.
type runTrace struct {
	host    *traces.Host
	process *traces.ProcessTrace
}

// wantsRunTrace reports whether a run opens a trace host: one the Processes
// tab or a diagnostics snapshot reads, or one whose fixtures can arm traces.
func wantsRunTrace(opts testrunner.RunOptions) bool {
	return opts.UI || opts.Diagnostics || opts.Fixtures || opts.FixtureRunner != nil
}

// startRunTrace opens the run's store and, when the run is watched, starts the
// process capture of gavel's own tree.
func startRunTrace(opts testrunner.RunOptions, started time.Time) (*runTrace, error) {
	host, err := traces.Open(opts.WorkDir, traces.RunID(started))
	if err != nil {
		return nil, err
	}
	trace := &runTrace{host: host}
	if !opts.UI && !opts.Diagnostics {
		return trace, nil
	}
	if trace.process, err = host.StartProcessTrace(os.Getpid()); err != nil {
		return nil, errors.Join(fmt.Errorf("process trace: %w", err), host.Close(context.Background()))
	}
	return trace, nil
}

// serveOn mounts the run's store on the test UI at testui.QueryMount, with the
// process capture the Processes tab reads and checkpoints.
func (t *runTrace) serveOn(srv *testui.Server) error {
	handler, err := traceserve.Handler(traceserve.Options{Runtime: t.host.Runtime(), Authorize: traceserve.ReadOnly})
	if err != nil {
		return fmt.Errorf("serve run traces: %w", err)
	}
	session := ""
	if t.process != nil {
		session = t.process.Session()
	}
	srv.SetTraceRun(t.host.RunID(), t.processRefs())
	srv.ServeTraces(t.host.RunID(), handler, session)
	return nil
}

// close takes a last checkpoint of the process capture, so the store keeps
// the run's final goroutines, stops the run's captures and closes its store,
// which stays on disk for the saved snapshot; the oldest runs' stores beyond
// traces.KeepRuns go.
func (t *runTrace) close(workDir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var processErr error
	if t.process != nil {
		processErr = errors.Join(t.process.Checkpoint(ctx), t.process.Stop(ctx))
	}
	return errors.Join(processErr, t.host.Close(ctx), traces.Prune(workDir, traces.KeepRuns))
}

// runID names the run's store, empty for a run that opened none.
func (t *runTrace) runID() string {
	if t == nil {
		return ""
	}
	return t.host.RunID()
}

// processRefs name the run's process capture, nil when the run is not watched.
func (t *runTrace) processRefs() *traces.ProcessRefs {
	if t == nil || t.process == nil {
		return nil
	}
	refs := t.process.Refs()
	return &refs
}
