//go:build unix

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/flanksource/commons/logger"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"golang.org/x/sys/unix"
)

// handoffDetachedUI forks a detached `gavel ui serve` child that inherits the
// UI listener FD so the URL the user already saw keeps working after the
// parent `gavel test` exits.
//
// Protocol:
//
//  1. Write results JSON to .tmp/gavel-ui/results-<pid>.json.
//  2. Write an initial lockfile at .tmp/gavel-ui/port-<port>.lock under
//     LOCK_EX (JSON: {port, url, pid:0, deadline}). Release the lock.
//  3. Dup the listener FD, spawn a Setsid child via exec.Cmd with the
//     listener attached as fd 3 via ExtraFiles, stdio redirected to
//     .tmp/gavel-ui/serve-<port>.log, env GAVEL_UI_LOCKFILE=<path>.
//  4. Close the parent's listener (the child's dup keeps the socket alive).
//  5. Poll the lockfile under LOCK_SH every 50ms for up to 5s waiting for the
//     child to flip pid -> non-zero. Log a warning if the window expires.
//  6. Print `UI (detached): <url>` and return.
//
// Any error at steps 1–3 aborts the handoff; the caller logs the warning and
// returns normally so the test exit code is still the real signal.
func handoffDetachedUI(
	listener net.Listener,
	snapshot testui.Snapshot,
	autoStop time.Duration,
	idleTimeout time.Duration,
) error {
	if listener == nil {
		return fmt.Errorf("no UI listener to hand off")
	}
	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		return fmt.Errorf("UI listener is not a TCPListener: %T", listener)
	}
	port := tcpListener.Addr().(*net.TCPAddr).Port

	baseDir := filepath.Join(".tmp", "gavel-ui")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", baseDir, err)
	}

	pid := os.Getpid()
	resultsPath := filepath.Join(baseDir, fmt.Sprintf("results-%d.json", pid))
	lockPath := filepath.Join(baseDir, fmt.Sprintf("port-%d.lock", port))
	logPath := filepath.Join(baseDir, fmt.Sprintf("serve-%d.log", port))

	if err := writeSnapshotJSON(resultsPath, snapshot); err != nil {
		return fmt.Errorf("write snapshot: %w", err)
	}

	url := fmt.Sprintf("http://localhost:%d", port)
	if err := writeInitialLockfile(lockPath, port, url); err != nil {
		return fmt.Errorf("write lockfile: %w", err)
	}

	// Dup the listener FD. net.TCPListener.File() returns a dup'd *os.File,
	// which is what we want: cmd.ExtraFiles needs a file we can pass to the
	// child, and the original listener's FD stays under our control.
	listenerFile, err := tcpListener.File()
	if err != nil {
		return fmt.Errorf("dup listener fd: %w", err)
	}
	// The child will receive this as fd 3 (0, 1, 2 are stdin/stdout/stderr).
	const childListenerFD = 3

	self, err := os.Executable()
	if err != nil {
		listenerFile.Close() //nolint:errcheck
		return fmt.Errorf("resolve self: %w", err)
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		listenerFile.Close() //nolint:errcheck
		return fmt.Errorf("open serve log: %w", err)
	}

	cmd := exec.Command(self, "ui", "serve",
		fmt.Sprintf("--listener-fd=%d", childListenerFD),
		fmt.Sprintf("--auto-stop=%s", durationOrDefault(autoStop, 30*time.Minute)),
		fmt.Sprintf("--idle-timeout=%s", durationOrDefault(idleTimeout, 5*time.Minute)),
		resultsPath,
	)
	cmd.ExtraFiles = []*os.File{listenerFile}
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil
	cmd.Env = append(os.Environ(), "GAVEL_UI_LOCKFILE="+lockPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		listenerFile.Close() //nolint:errcheck
		logFile.Close()      //nolint:errcheck
		return fmt.Errorf("spawn detached UI: %w", err)
	}
	// Parent no longer needs the dup'd listener or the log file handle.
	_ = listenerFile.Close()
	_ = logFile.Close()
	// Close the parent's original listener so the child is the sole owner.
	// The kernel keeps the socket alive via the child's inherited fd.
	_ = tcpListener.Close()

	// Release the child process from our wait queue — we don't want to
	// reap it on exit, and Setsid already gave it its own session.
	if err := cmd.Process.Release(); err != nil {
		logger.V(1).Infof("process.Release: %v", err)
	}

	// Wait for the child to flip pid -> non-zero in the lockfile.
	childPID, err := waitForHandoff(lockPath, 5*time.Second)
	if err != nil {
		logger.Warnf("Detached UI child did not report ready within 5s: %v", err)
	}

	logger.V(1).Infof("Detached UI child pid=%d serving %s", childPID, url)
	fmt.Fprintf(os.Stderr, "UI (detached): %s\n", url)
	return nil
}

func durationOrDefault(d, fallback time.Duration) time.Duration {
	if d <= 0 {
		return fallback
	}
	return d
}

func writeSnapshotJSON(path string, snapshot testui.Snapshot) error {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// writeInitialLockfile creates the lockfile with PID=0 and a 5s handoff
// deadline, writes it under LOCK_EX, then releases the lock so the child
// can take LOCK_EX when it's ready.
func writeInitialLockfile(path string, port int, url string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck

	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		return fmt.Errorf("flock LOCK_EX: %w", err)
	}

	payload := uiLockfilePayload{
		Port:     port,
		URL:      url,
		PID:      0,
		Deadline: time.Now().Add(5 * time.Second).UTC().Format(time.RFC3339Nano),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := f.WriteAt(data, 0); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	return unix.Flock(int(f.Fd()), unix.LOCK_UN)
}

// waitForHandoff polls the lockfile under LOCK_SH every 50ms until pid flips
// to non-zero or the deadline passes. Returns the child's pid on success.
func waitForHandoff(path string, timeout time.Duration) (int, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		pid, ok, err := readLockfilePID(path)
		if err == nil && ok {
			return pid, nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return 0, fmt.Errorf("handoff deadline exceeded")
}

func readLockfilePID(path string) (int, bool, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return 0, false, err
	}
	defer f.Close() //nolint:errcheck
	if err := unix.Flock(int(f.Fd()), unix.LOCK_SH); err != nil {
		return 0, false, err
	}
	defer unix.Flock(int(f.Fd()), unix.LOCK_UN) //nolint:errcheck

	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false, err
	}
	var payload uiLockfilePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return 0, false, err
	}
	if payload.PID == 0 {
		return 0, false, nil
	}
	return payload.PID, true, nil
}
