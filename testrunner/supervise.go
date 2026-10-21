package testrunner

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/flanksource/clicky/exec"
	commonsCtx "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
)

// PackageTimeoutReason is the diagnostics reason a package that ran past its
// deadline is captured with, in the log and on the run's process trace.
func PackageTimeoutReason(label string) string { return "package-timeout:" + label }

// packageSupervision is what the supervisor of one package subprocess saw:
// whether it had to kill it, and the Go test binaries it asked for a
// goroutine dump first.
type packageSupervision struct {
	timedOut atomic.Bool
	// settled closes once the supervisor is done with the process: a binary
	// that exits on its SIGQUIT can end go test before the dump wait does.
	settled chan struct{}
	dumped  []goTestBinary
}

func (s *packageSupervision) TimedOut() bool { return s.timedOut.Load() }

// Dumped are the binaries the supervisor of a timed-out package SIGQUIT,
// once it has finished with them; call it only when TimedOut.
func (s *packageSupervision) Dumped() []goTestBinary {
	<-s.settled
	return s.dumped
}

// supervisePackage arms the package's deadline (the per-package timeout and
// the global --timeout) and returns the context to run the process under.
// When a deadline fires while the process runs, the supervisor captures
// diagnostics, asks the Go test binaries under the process for a goroutine
// dump (SIGQUIT) and waits for it, then kills the tree. The process is run
// under a context the supervisor cancels only after that, and with no timer of
// its own, so nothing kills it before its dump is taken.
func (o *TestOrchestrator) supervisePackage(parent commonsCtx.Context, process *exec.Process, label string) (commonsCtx.Context, context.CancelFunc, *packageSupervision) {
	supervision := &packageSupervision{settled: make(chan struct{})}
	if o.TestTimeout > 0 {
		process.WithTimeout(o.TestTimeout)
	}
	timeout := process.Timeout
	process.WithTimeout(0)

	base := context.Context(parent)
	if o.Context != nil {
		merged, cancelMerge := mergeParentContext(base, o.Context)
		base = merged
		go func() {
			<-merged.Done()
			cancelMerge()
		}()
	}
	var (
		pkgCtx    context.Context
		cancelPkg context.CancelFunc
	)
	if timeout > 0 {
		pkgCtx, cancelPkg = context.WithTimeout(base, timeout)
	} else {
		pkgCtx, cancelPkg = context.WithCancel(base)
	}
	runCtx, cancelRun := context.WithCancel(context.Background())

	go func() {
		<-pkgCtx.Done()
		defer close(supervision.settled)
		defer cancelRun()
		if !process.IsRunning() {
			return
		}
		supervision.timedOut.Store(true)
		logger.Warnf("test package %s exceeded deadline — capturing diagnostics before killing tree", label)
		captureGlobalDiagnostics(PackageTimeoutReason(label))
		if o.deadlineExceeded(pkgCtx) {
			supervision.dumped = quitGoTestBinaries(process.Pid(), goroutineDumpWait)
		}
		if err := process.KillTree(); err != nil {
			logger.Warnf("kill-tree failed for %s: %v", label, err)
		}
	}()

	return commonsCtx.NewContext(runCtx), func() { cancelPkg(); cancelRun() }, supervision
}

// deadlineExceeded tells a deadline (the package's or the run's --timeout)
// from a stop, which kills without waiting for goroutine dumps.
func (o *TestOrchestrator) deadlineExceeded(pkgCtx context.Context) bool {
	return errors.Is(pkgCtx.Err(), context.DeadlineExceeded) ||
		o.Context != nil && errors.Is(o.Context.Err(), context.DeadlineExceeded)
}

// mergeParentContext returns a child context that is cancelled when either
// primary or parent is cancelled. Cancelling child never cancels the parents.
func mergeParentContext(primary, parent context.Context) (context.Context, context.CancelFunc) {
	child, cancel := context.WithCancel(primary)
	if parent == nil {
		return child, cancel
	}
	go func() {
		select {
		case <-parent.Done():
			cancel()
		case <-child.Done():
		}
	}()
	return child, cancel
}

// captureGlobalDiagnostics is set by the CLI layer. It is called with the
// reason (PackageTimeoutReason) before each subprocess kill so timeout output
// captures live goroutine, task, and process state.
var captureGlobalDiagnostics = func(string) {}

// SetCaptureGlobalDiagnostics installs the diagnostics hook invoked by the
// per-package supervisor before it kills a timed-out subprocess. The caller
// is responsible for ensuring the hook itself is safe to run concurrently.
func SetCaptureGlobalDiagnostics(fn func(reason string)) {
	if fn == nil {
		captureGlobalDiagnostics = func(string) {}
		return
	}
	captureGlobalDiagnostics = fn
}
