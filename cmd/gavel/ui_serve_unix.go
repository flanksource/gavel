//go:build unix

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"golang.org/x/sys/unix"
)

// lockfileEnv is the env var name the fork parent uses to hand the child a
// path to a JSON lockfile for handoff coordination. Unset outside the fork
// path; standalone `gavel ui serve` ignores it.
const lockfileEnv = "GAVEL_UI_LOCKFILE"

// openListener either adopts an inherited TCP socket from fd 3 (when spawned
// by `gavel test --ui --auto-stop`) or binds a fresh listener on --port.
//
// Fail LOUD if either path fails: the parent has already printed this URL to
// the user, so silently picking a different port would result in a dead URL.
func openListener(opts UIServeOptions) (net.Listener, error) {
	if opts.ListenerFD > 0 {
		f := os.NewFile(uintptr(opts.ListenerFD), fmt.Sprintf("gavel-ui-listener-fd-%d", opts.ListenerFD))
		l, err := net.FileListener(f)
		if err != nil {
			return nil, fmt.Errorf("adopt inherited listener fd=%d: %w", opts.ListenerFD, err)
		}
		// net.FileListener dup's the fd; we can close our copy now.
		_ = f.Close()
		return l, nil
	}
	host := opts.Addr
	if host == "" {
		host = "localhost"
	}
	addr := net.JoinHostPort(host, strconv.Itoa(opts.Port))
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("bind %s: %w", addr, err)
	}
	return l, nil
}

// notifyHandoff writes the child's PID into the lockfile pointed at by
// $GAVEL_UI_LOCKFILE and releases an exclusive flock on it, signaling the
// fork parent that the child is ready. The parent is polling under a shared
// flock — once our exclusive lock drops, the parent's poll wins, reads our
// pid, and exits its handoff wait.
//
// No-op when the env var is unset (standalone replay mode).
func notifyHandoff(_ UIServeOptions, port int, url string) error {
	path := os.Getenv(lockfileEnv)
	if path == "" {
		return nil
	}

	f, err := os.OpenFile(path, os.O_RDWR, 0o600)
	if err != nil {
		return fmt.Errorf("open lockfile %s: %w", path, err)
	}
	defer f.Close() //nolint:errcheck

	// Exclusive lock. The parent wrote the initial payload under LOCK_EX,
	// released it before cmd.Start, and is now polling under LOCK_SH.
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		return fmt.Errorf("flock LOCK_EX %s: %w", path, err)
	}

	payload := uiLockfilePayload{
		Port:     port,
		URL:      url,
		PID:      os.Getpid(),
		Deadline: time.Now().Add(5 * time.Second).UTC().Format(time.RFC3339Nano),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.WriteAt(data, 0); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}

	// Releasing the exclusive lock is how we signal ready. The parent's poll
	// of LOCK_SH will succeed on the next tick and observe PID != 0.
	return unix.Flock(int(f.Fd()), unix.LOCK_UN)
}

// uiLockfilePayload is the on-disk shape of the handoff lockfile written by
// both parent (initial, PID=0) and child (final, PID=os.Getpid()).
type uiLockfilePayload struct {
	Port     int    `json:"port"`
	URL      string `json:"url"`
	PID      int    `json:"pid"`
	Deadline string `json:"deadline,omitempty"`
}
