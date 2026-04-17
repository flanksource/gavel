// Package service manages the detached `gavel pr list --all --ui` background
// process and the launchd (macOS) / systemd --user (Linux) service files that
// keep it running across logins.
//
// The daemon is "owned" by a pidfile at StateDir()/pr-ui.pid. A single user
// runs at most one instance; `system start` tears down any existing process
// before spawning a fresh one so flags can change between invocations.
package service

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/flanksource/commons/logger"
)

const (
	// ServiceName is the canonical label for the pr UI background service.
	// Used as the launchd label and systemd unit stem.
	ServiceName = "gavel-pr-ui"

	pidFileName = "pr-ui.pid"
	logFileName = "pr-ui.log"
)

// StateDir returns the per-user directory where the pidfile and log live.
// Mirrors pr/ui/settings.go which uses ~/.config/gavel.
func StateDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	dir := filepath.Join(home, ".config", "gavel")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create state dir %s: %w", dir, err)
	}
	return dir, nil
}

// PidFile returns the absolute path of the pidfile.
func PidFile() (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, pidFileName), nil
}

// LogFile returns the absolute path of the stdout/stderr log.
func LogFile() (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, logFileName), nil
}

// BinaryPath returns the absolute path to the current gavel binary. Service
// files and detached Start() both need an absolute path so they don't depend
// on $PATH at service-activation time.
func BinaryPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve current executable: %w", err)
	}
	abs, err := filepath.Abs(exe)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path of %s: %w", exe, err)
	}
	return abs, nil
}

// Status describes whether the pr UI daemon is currently running.
type Status struct {
	Running bool
	PID     int
	PidFile string
	LogFile string
	// Stale is true when the pidfile points at a PID that is no longer alive.
	Stale bool
}

// ReadStatus inspects the pidfile and reports whether the recorded process is
// still alive. A stale pidfile is reported Running=false, Stale=true.
func ReadStatus() (Status, error) {
	pf, err := PidFile()
	if err != nil {
		return Status{}, err
	}
	lf, err := LogFile()
	if err != nil {
		return Status{}, err
	}
	st := Status{PidFile: pf, LogFile: lf}
	pid, err := readPid(pf)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return st, nil
		}
		return st, err
	}
	st.PID = pid
	if processAlive(pid) {
		st.Running = true
		return st, nil
	}
	st.Stale = true
	return st, nil
}

func readPid(pf string) (int, error) {
	b, err := os.ReadFile(pf)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(b))
	if s == "" {
		return 0, fmt.Errorf("pidfile %s is empty", pf)
	}
	pid, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("pidfile %s has invalid pid %q: %w", pf, s, err)
	}
	return pid, nil
}

func writePid(pf string, pid int) error {
	return os.WriteFile(pf, []byte(strconv.Itoa(pid)+"\n"), 0o644)
}

// processAlive returns true if signal 0 can be delivered to pid (i.e. the
// process exists and belongs to the current user). EPERM — "exists but not
// ours" — is rare for a user-owned pidfile and we treat it as "not ours".
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}

// StartOptions configures the detached Start().
type StartOptions struct {
	// Extra flags forwarded after `pr list`. When empty, Start uses the
	// canonical `--all --ui` flag set. Callers overriding this should include
	// `--ui` themselves.
	Extra []string
	// BinaryPath overrides os.Executable() — used by tests.
	BinaryPath string
}

// Start stops any existing detached instance then spawns a fresh
// `gavel pr list --all --ui` in the background, writing its PID to the
// pidfile and redirecting stdout/stderr to LogFile. Returns the new PID.
//
// The child is fully detached: Setsid so it survives the parent's shell, and
// Release so we don't keep a zombie-holding handle on it.
func Start(opts StartOptions) (int, error) {
	if err := Stop(5 * time.Second); err != nil {
		// Stop is best-effort — a failure to signal a previous process
		// shouldn't block starting a new one, but we surface the warning.
		logger.Warnf("pr-ui stop before start: %v", err)
	}

	bin := opts.BinaryPath
	if bin == "" {
		b, err := BinaryPath()
		if err != nil {
			return 0, err
		}
		bin = b
	}

	args := opts.Extra
	if len(args) == 0 {
		args = []string{"pr", "list", "--all", "--ui"}
	}

	logPath, err := LogFile()
	if err != nil {
		return 0, err
	}
	logFD, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return 0, fmt.Errorf("open log file %s: %w", logPath, err)
	}
	defer logFD.Close()

	fmt.Fprintf(logFD, "\n--- pr-ui start %s ---\n", time.Now().Format(time.RFC3339))

	cmd := exec.Command(bin, args...)
	cmd.Stdout = logFD
	cmd.Stderr = logFD
	cmd.Stdin = nil
	cmd.SysProcAttr = detachedSysProcAttr()

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("start %s %s: %w", bin, strings.Join(args, " "), err)
	}

	pid := cmd.Process.Pid
	pf, err := PidFile()
	if err != nil {
		_ = cmd.Process.Kill()
		return 0, err
	}
	if err := writePid(pf, pid); err != nil {
		_ = cmd.Process.Kill()
		return 0, fmt.Errorf("write pidfile: %w", err)
	}

	if err := cmd.Process.Release(); err != nil {
		logger.Warnf("release child process: %v", err)
	}
	return pid, nil
}

// Stop signals SIGTERM to the pid recorded in the pidfile and waits up to
// timeout for it to exit, escalating to SIGKILL on timeout. Returns nil when
// no process is running (nothing to do).
func Stop(timeout time.Duration) error {
	st, err := ReadStatus()
	if err != nil {
		return err
	}
	if !st.Running {
		if st.Stale {
			logger.V(1).Infof("removing stale pidfile %s", st.PidFile)
			_ = os.Remove(st.PidFile)
		}
		return nil
	}
	p, err := os.FindProcess(st.PID)
	if err != nil {
		return fmt.Errorf("find pid %d: %w", st.PID, err)
	}
	if err := p.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("signal SIGTERM to pid %d: %w", st.PID, err)
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !processAlive(st.PID) {
			_ = os.Remove(st.PidFile)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	logger.Warnf("pid %d did not exit within %s, sending SIGKILL", st.PID, timeout)
	if err := p.Signal(syscall.SIGKILL); err != nil {
		return fmt.Errorf("signal SIGKILL to pid %d: %w", st.PID, err)
	}
	_ = os.Remove(st.PidFile)
	return nil
}
