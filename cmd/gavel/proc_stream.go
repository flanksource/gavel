package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/procfile"
)

// streamTailLines is how much trailing context each process's log shows before
// the live stream takes over during a restart/stop.
const streamTailLines = 10

// streamUntilReady tails the tracked processes' logs while polling them to a
// ready/failed outcome. With follow it keeps streaming after they settle, until
// the user interrupts. names selects the subset to track (empty tracks every
// non-stopped process in report). It returns a non-nil error naming any process
// that failed to start (a crash, or one that never reached running) so the
// caller can exit non-zero.
func streamUntilReady(workDir, pf string, report *procfile.StatusReport, names []string, follow bool) error {
	tracked := trackedProcNames(report.Processes, names)
	if len(tracked) == 0 {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream := startStreaming(ctx, workDir, pf, tracked)

	failErr := awaitTrackedReady(ctx, workDir, pf, tracked)
	if follow && failErr == nil {
		waitForInterruptOrDone(ctx)
	}
	cancel()
	<-stream
	return failErr
}

// stopAndStream tails the running processes' logs while the stop is carried out,
// so the user sees shutdown output live, then returns the post-stop status. With
// follow it keeps streaming until interrupted. Streaming starts before the stop
// is issued (the whole-daemon stop blocks until the supervisor exits) so the
// drain is captured as it happens.
func stopAndStream(workDir, pf string, names []string, follow bool) (any, error) {
	before, err := procfile.Status(workDir, pf)
	if err != nil {
		return nil, err
	}
	tracked := runningProcNames(before.Processes, names)
	if len(tracked) == 0 {
		return procfile.Stop(workDir, pf, names)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream := startStreaming(ctx, workDir, pf, tracked)

	report, stopErr := procfile.Stop(workDir, pf, names)
	if stopErr != nil {
		cancel()
		<-stream
		return nil, stopErr
	}
	if follow {
		waitForInterruptOrDone(ctx)
	}
	cancel()
	<-stream
	return report, nil
}

// startStreaming runs procfile.Stream in the background, returning a channel that
// closes when it has flushed and exited (after ctx is cancelled).
func startStreaming(ctx context.Context, workDir, pf string, names []string) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := procfile.Stream(ctx, workDir, pf, names, streamTailLines, os.Stdout); err != nil && ctx.Err() == nil {
			logger.Warnf("stream process logs: %v", err)
		}
	}()
	return done
}

// awaitTrackedReady polls the tracked processes until each settles, returning a
// non-nil error listing the ones that failed to start. ctx cancellation (e.g. a
// failed start elsewhere being torn down) ends the wait with no error.
func awaitTrackedReady(ctx context.Context, workDir, pf string, names []string) error {
	deadline := time.Now().Add(procReadyTimeout)
	trackers := make(map[string]*procTracker, len(names))
	for _, n := range names {
		trackers[n] = &procTracker{deadline: deadline}
	}
	var failures []string
	for len(trackers) > 0 {
		rep, err := procfile.Status(workDir, pf)
		if err != nil {
			return err
		}
		now := time.Now()
		for name, tr := range trackers {
			ps, ok := findProcState(rep.Processes, name)
			if !ok {
				delete(trackers, name)
				failures = append(failures, fmt.Sprintf("%s disappeared from status", name))
				continue
			}
			switch outcome, ferr := tr.observe(ps, now); outcome {
			case outcomeReady, outcomeWarn:
				delete(trackers, name)
			case outcomeFailed:
				delete(trackers, name)
				failures = append(failures, ferr.Error())
			}
		}
		if len(trackers) == 0 {
			break
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(procReadyPoll):
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("%d process(es) failed to start: %s", len(failures), strings.Join(failures, "; "))
	}
	return nil
}

// runningProcNames returns the names of the processes that are currently up
// (running/starting/restarting), restricted to names when non-empty. Stopped
// processes are skipped so a stop only tails what was actually running.
func runningProcNames(procs []procfile.ProcState, names []string) []string {
	want := make(map[string]bool, len(names))
	for _, n := range names {
		want[n] = true
	}
	out := make([]string, 0, len(procs))
	for _, p := range procs {
		if len(names) > 0 && !want[p.Name] {
			continue
		}
		switch p.Status {
		case procfile.StatusRunning, procfile.StatusStarting, procfile.StatusRestarting:
			out = append(out, p.Name)
		}
	}
	return out
}

// waitForInterruptOrDone blocks until SIGINT/SIGTERM or ctx cancellation, so
// --follow streams until the user presses Ctrl-C.
func waitForInterruptOrDone(ctx context.Context) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)
	select {
	case <-sig:
	case <-ctx.Done():
	}
}
