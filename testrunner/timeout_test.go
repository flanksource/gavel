package testrunner

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/flanksource/clicky/exec"
	commonsCtx "github.com/flanksource/commons/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSupervisePackageKillsOnTimeout(t *testing.T) {
	prev := captureGlobalDiagnostics
	defer SetCaptureGlobalDiagnostics(prev)
	var captured atomic.Int32
	SetCaptureGlobalDiagnostics(func() { captured.Add(1) })

	process := exec.NewExec("sleep", "30")

	orch := &TestOrchestrator{
		RunOptions: RunOptions{
			TestTimeout: 150 * time.Millisecond,
		},
	}
	ctx, cancel, timedOutPtr := orch.supervisePackage(commonsCtx.NewContext(context.Background()), process, "fake/pkg")
	defer cancel()

	result := process.Run().Result()
	// Give the supervisor goroutine a moment to run its kill path.
	time.Sleep(50 * time.Millisecond)

	assert.NotNil(t, result)
	require.NotNil(t, timedOutPtr)
	assert.True(t, *timedOutPtr, "supervisor must mark timedOut=true")
	assert.NotEqual(t, 0, int(captured.Load()), "captureGlobalDiagnostics must be called before kill")
	_ = ctx
}

func TestSupervisePackageLetsFastCommandsFinish(t *testing.T) {
	process := exec.NewExec("true")
	orch := &TestOrchestrator{
		RunOptions: RunOptions{TestTimeout: 5 * time.Second},
	}
	_, cancel, timedOutPtr := orch.supervisePackage(commonsCtx.NewContext(context.Background()), process, "fake/pkg")
	defer cancel()

	_ = process.Run().Result()
	assert.False(t, *timedOutPtr)
}

func TestSupervisePackageRespectsGlobalContext(t *testing.T) {
	prev := captureGlobalDiagnostics
	defer SetCaptureGlobalDiagnostics(prev)
	SetCaptureGlobalDiagnostics(func() {})

	globalCtx, cancelGlobal := context.WithCancel(context.Background())

	process := exec.NewExec("sleep", "30")

	orch := &TestOrchestrator{
		RunOptions: RunOptions{
			Context:     globalCtx,
			TestTimeout: 10 * time.Second, // per-package won't fire
		},
	}
	_, cancel, timedOutPtr := orch.supervisePackage(commonsCtx.NewContext(context.Background()), process, "fake/pkg")
	defer cancel()

	// Cancel globally after the process has started.
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancelGlobal()
	}()

	_ = process.Run().Result()
	time.Sleep(50 * time.Millisecond)
	assert.True(t, *timedOutPtr, "global cancellation must trip per-package supervisor")
}

// TestSupervisePackageForceKillsSignalTrappingProcess drives a subprocess
// that traps SIGINT and SIGTERM. With WithProcessGroup + KillTree the kill
// is atomic (SIGKILL to -pgid) and Run() must unblock via cmd.WaitDelay. The
// test asserts both: pid is reaped AND Run() goroutine returns — the latter
// was the real gap that let testrunner/ui hang for 5m14s.
func TestSupervisePackageForceKillsSignalTrappingProcess(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("signal trapping behaviour is POSIX-specific")
	}

	prev := captureGlobalDiagnostics
	defer SetCaptureGlobalDiagnostics(prev)
	SetCaptureGlobalDiagnostics(func() {})

	process := exec.NewExec("sh", "-c", "trap '' INT TERM; sleep 60").WithProcessGroup()

	orch := &TestOrchestrator{
		RunOptions: RunOptions{TestTimeout: 150 * time.Millisecond},
	}
	_, cancel, timedOutPtr := orch.supervisePackage(commonsCtx.NewContext(context.Background()), process, "stubborn/pkg")
	defer cancel()

	runDone := make(chan struct{})
	go func() {
		_ = process.Run().Result()
		close(runDone)
	}()

	require.NoError(t, waitUntil(func() bool { return process.Pid() > 0 }, 5*time.Second, 25*time.Millisecond),
		"expected a pid to observe before the supervisor escalates")
	pid := process.Pid()

	// Atomic pgid SIGKILL + cmd.WaitDelay (2s) should reap the pid well
	// inside 5s of the 150ms deadline.
	require.NoError(t, waitUntil(func() bool { return !pidAlive(pid) }, 5*time.Second, 25*time.Millisecond),
		"subprocess pid %d should be reaped via pgid SIGKILL", pid)

	assert.True(t, *timedOutPtr, "supervisor must mark timedOut=true after kill")

	select {
	case <-runDone:
	case <-time.After(5 * time.Second):
		t.Fatal("process.Run() did not return after KillTree — the 5m14s regression is back")
	}
}

// TestSupervisePackageReapsGrandchildren mirrors the real-world testrunner/ui
// shape: a parent that backgrounds a grandchild that also traps signals.
// The pgid kill must reap both, and Run() must unblock.
func TestSupervisePackageReapsGrandchildren(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX-only scenario")
	}

	prev := captureGlobalDiagnostics
	defer SetCaptureGlobalDiagnostics(prev)
	SetCaptureGlobalDiagnostics(func() {})

	script := `trap '' INT TERM; (trap '' INT TERM; sleep 60 & echo $! > /tmp/gavel_supervise_child.pid; wait) & wait`
	process := exec.NewExec("sh", "-c", script).WithProcessGroup()

	orch := &TestOrchestrator{
		RunOptions: RunOptions{TestTimeout: 200 * time.Millisecond},
	}
	_, cancel, timedOutPtr := orch.supervisePackage(commonsCtx.NewContext(context.Background()), process, "stubborn/tree")
	defer cancel()

	runDone := make(chan struct{})
	go func() {
		_ = process.Run().Result()
		close(runDone)
	}()

	require.NoError(t, waitUntil(func() bool { return process.Pid() > 0 }, 5*time.Second, 25*time.Millisecond))
	parentPID := process.Pid()

	var childPID int
	_ = waitUntil(func() bool {
		data, err := os.ReadFile("/tmp/gavel_supervise_child.pid")
		if err != nil {
			return false
		}
		for _, b := range data {
			if b >= '0' && b <= '9' {
				childPID = childPID*10 + int(b-'0')
			} else {
				break
			}
		}
		return childPID > 0
	}, 3*time.Second, 25*time.Millisecond)
	defer os.Remove("/tmp/gavel_supervise_child.pid")

	require.NoError(t, waitUntil(func() bool { return !pidAlive(parentPID) }, 5*time.Second, 25*time.Millisecond),
		"parent %d must die via pgid SIGKILL", parentPID)
	if childPID > 0 {
		require.NoError(t, waitUntil(func() bool { return !pidAlive(childPID) }, 5*time.Second, 25*time.Millisecond),
			"grandchild %d must die via pgid SIGKILL", childPID)
	}

	assert.True(t, *timedOutPtr)

	select {
	case <-runDone:
	case <-time.After(5 * time.Second):
		t.Fatal("process.Run() did not return after KillTree against grandchildren")
	}
}

// pidAlive reports whether a Unix pid is still alive using signal 0, which
// performs the permission + existence checks without delivering a signal.
func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}

// waitUntil polls cond every interval until it returns true or the timeout
// elapses.
func waitUntil(cond func() bool, timeout, interval time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return nil
		}
		time.Sleep(interval)
	}
	return fmt.Errorf("condition not met within %s", timeout)
}
