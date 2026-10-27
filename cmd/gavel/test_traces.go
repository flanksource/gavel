package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/flanksource/commons-db/cmd/query/traceserve"
	"github.com/flanksource/commons-db/tracing/stacks"
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
	// sigquit is the stack transport of the Go test binaries the run
	// supervises, registered on the host for the process capture's stacks.
	sigquit *testrunner.SigquitDumper
}

// wantsRunTrace reports whether a run opens a trace host: one the Processes
// tab or a diagnostics snapshot reads, or one whose fixtures can arm traces.
func wantsRunTrace(opts testrunner.RunOptions) bool {
	return opts.UI || opts.Diagnostics || opts.Fixtures || opts.FixtureRunner != nil
}

// startRunTrace opens the run's store and, when the run is watched, starts the
// process capture of gavel's own tree, its stacks dumped over gavel's sigquit
// besides the host's transports. A run watched in a UI (--ui, or
// --process-facets as pr/ui's project runs pass) reads every facet and
// environment; others read none.
func startRunTrace(opts testrunner.RunOptions, started time.Time) (*runTrace, error) {
	if opts.ProcessFacets && !opts.UI && !opts.Diagnostics {
		return nil, errors.New("--process-facets needs a process trace: pass --ui or --diagnostics")
	}
	host, err := traces.Open(opts.Dirs.StateRoot, traces.RunID(started))
	if err != nil {
		return nil, err
	}
	trace := &runTrace{host: host}
	if !opts.UI && !opts.Diagnostics {
		return trace, nil
	}
	trace.sigquit = testrunner.NewSigquitDumper()
	if err := host.RegisterDumper(trace.sigquit); err != nil {
		return nil, errors.Join(err, host.Close(context.Background()))
	}
	options := traces.ProcessTraceOptions{RootPID: os.Getpid(), Facets: opts.UI || opts.ProcessFacets}
	if trace.process, err = host.StartProcessTrace(options); err != nil {
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
	var sessions *traces.ProcessSessions
	if t.process != nil {
		running := t.process.Sessions()
		sessions = &running
	}
	srv.SetTraceRun(t.host.RunID(), t.processRefs())
	srv.ServeTraces(t.host.RunID(), handler, sessions)
	return nil
}

// close takes a last reading of the process capture — gavel's final
// goroutines into its stacks, a re-observed tree — stops the run's captures
// and closes its store, which stays on disk for the saved snapshot; the
// oldest runs' stores beyond traces.KeepRuns go.
func (t *runTrace) close(stateRoot string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var processErr error
	if t.process != nil {
		processErr = errors.Join(
			t.process.CheckpointStacks(ctx, stacks.CheckpointParams{PID: os.Getpid(), Transport: stacks.TransportInProcess, Reason: "final"}),
			t.process.CheckpointTree(ctx),
			t.process.Stop(ctx),
		)
	}
	return errors.Join(processErr, t.host.Close(ctx), traces.Prune(stateRoot, traces.KeepRuns))
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
