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
	"net"
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
	// ServiceName is the canonical label for the gavel background service.
	// Used as the launchd label stem (com.flanksource.<ServiceName>) and
	// the systemd user unit stem (<ServiceName>.service).
	ServiceName = "gavel"

	// legacyServiceName is the previous ServiceName value. Install /
	// Uninstall clean it up so upgraders don't end up running two
	// competing daemons side by side.
	legacyServiceName = "gavel-pr-ui"

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

// TailLog returns the last n lines of LogFile. A missing logfile is treated
// as "no log yet" and returns an empty slice (no error) so `system status`
// can be run before `system start`. Lines are returned without trailing
// newlines, in original order (oldest first).
//
// For large logs we read the tail in 8KiB chunks from the end rather than
// loading the whole file — the daemon log can grow to many MB over time.
func TailLog(n int) ([]string, error) {
	if n <= 0 {
		return nil, nil
	}
	path, err := LogFile()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}
	size := info.Size()
	if size == 0 {
		return nil, nil
	}

	const chunk = 8 * 1024
	var (
		buf  []byte
		off  = size
		seen int
	)
	for off > 0 && seen <= n {
		read := min(int64(chunk), off)
		off -= read
		tmp := make([]byte, read)
		if _, err := f.ReadAt(tmp, off); err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		buf = append(tmp, buf...)
		seen = strings.Count(string(buf), "\n")
	}

	lines := strings.Split(strings.TrimRight(string(buf), "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines, nil
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

// ReadStatus reports whether the pr UI daemon is running. When a launchd /
// systemd user service is installed, status is sourced from the service
// manager (launchctl print / systemctl show). Otherwise we fall back to the
// pidfile-based check — a stale pidfile is reported Running=false, Stale=true.
func ReadStatus() (Status, error) {
	installed, err := IsInstalled()
	if err != nil {
		return Status{}, err
	}
	if installed {
		return serviceStatus()
	}
	return readPidStatus()
}

// readPidStatus is the pidfile-based status check. Extracted so ReadStatus can
// dispatch cleanly between service-manager and pidfile sources.
func readPidStatus() (Status, error) {
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

// Start stops any existing instance then starts a fresh pr UI.
//
// When a launchd / systemd user service is installed, Start asks the service
// manager to (re)start the service — the service file fixes the binary and
// flags, so opts.Extra / opts.BinaryPath are ignored and a warning is logged
// if the caller supplied them.
//
// Otherwise Start spawns a detached `gavel pr list --all --ui --menu-bar`
// directly, writes its PID to the pidfile, and returns that PID. The child
// is fully detached: Setsid so it survives the parent's shell, and Release
// so we don't keep a zombie-holding handle on it.
func Start(opts StartOptions) (int, error) {
	installed, err := IsInstalled()
	if err != nil {
		return 0, err
	}
	if installed {
		if len(opts.Extra) > 0 || opts.BinaryPath != "" {
			logger.Warnf("service is installed; ignoring StartOptions (the service file fixes binary and flags — uninstall and reinstall to change them)")
		}
		// Truncate before serviceStart so the resulting log contains only
		// this run — launchd/systemd reopen the file descriptor on process
		// launch, so truncating here is safe.
		if err := truncateLogFile(); err != nil {
			logger.Warnf("truncate log before service start: %v", err)
		}
		if err := serviceStart(); err != nil {
			return 0, err
		}
		st, err := serviceStatus()
		if err != nil {
			return 0, err
		}
		return st.PID, nil
	}
	return startDetached(opts)
}

// DefaultUIPort is the port the pr UI binds to when the user hasn't
// overridden it. `pr list --port=0` auto-scans from here upward.
const DefaultUIPort = 9092

// UIPort is the historical name of DefaultUIPort kept as an alias so old
// callers (WaitForReady, tests) keep compiling.
const UIPort = DefaultUIPort

// portFileName is where the daemon persists the actually-bound UI port so
// consumers (system status, WaitForReady, menubar deep-links) can find it
// when --port=0 picked something other than DefaultUIPort.
const portFileName = "pr-ui.port"

// PortFile returns the absolute path of the persisted port file.
func PortFile() (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, portFileName), nil
}

// WriteUIPort persists the port the daemon is actually listening on. Called
// once at startup after the listener binds. Failures to write the port file
// shouldn't kill the daemon — consumers fall back to DefaultUIPort — so the
// caller logs but doesn't propagate the error.
func WriteUIPort(port int) error {
	path, err := PortFile()
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(port)+"\n"), 0o644)
}

// ReadUIPort returns the port the daemon last wrote via WriteUIPort. When
// the file is missing or unparseable (fresh machine, pre-rename install)
// it returns DefaultUIPort — the historical hardcoded value — so consumers
// keep working without special-casing.
func ReadUIPort() int {
	path, err := PortFile()
	if err != nil {
		return DefaultUIPort
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return DefaultUIPort
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || n <= 0 {
		return DefaultUIPort
	}
	return n
}

// ScanFreePort binds the first available TCP port starting at `start`,
// trying up to `tries` consecutive numbers. Returns the bound listener so
// the caller doesn't race losing the port between scan and bind — whoever
// needs the port simply uses the returned listener.
//
// Used by `pr list --port=0`: start=DefaultUIPort, tries=50 gives the UI
// a 50-port search window (9092-9141) which is plenty for the "one daemon
// per user + a few leftover sockets in TIME_WAIT" case.
func ScanFreePort(start, tries int) (int, net.Listener, error) {
	if tries < 1 {
		tries = 1
	}
	var lastErr error
	for i := 0; i < tries; i++ {
		port := start + i
		l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		if err == nil {
			return port, l, nil
		}
		lastErr = err
	}
	return 0, nil, fmt.Errorf("no free port in range %d-%d: %w", start, start+tries-1, lastErr)
}

// Readiness is the outcome of WaitForReady: the daemon is Ready, it Crashed
// (pidfile points at a dead pid — or disappeared), or we TimedOut waiting
// for the UI port to bind.
type Readiness int

const (
	// ReadinessReady means the daemon is running AND the UI port accepts TCP
	// connections.
	ReadinessReady Readiness = iota
	// ReadinessCrashed means the process exited during the wait window.
	ReadinessCrashed
	// ReadinessTimedOut means the process is still running but the UI port
	// never came up within the timeout.
	ReadinessTimedOut
)

func (r Readiness) String() string {
	switch r {
	case ReadinessReady:
		return "ready"
	case ReadinessCrashed:
		return "crashed"
	case ReadinessTimedOut:
		return "timed out"
	default:
		return "unknown"
	}
}

// WaitForReady polls every 500ms until the daemon is fully up — process
// running AND localhost:<UI port> accepting TCP connections — or until it
// crashes or the timeout elapses. The UI port is read from the persisted
// port file (falls back to DefaultUIPort when missing), so --port=0 and
// --port=N installs both work. Callers typically dump the log file on any
// outcome other than ReadinessReady.
func WaitForReady(timeout time.Duration) (Readiness, error) {
	return waitForReady(timeout, ReadUIPort())
}

// waitForReady is the testable core — the port is injected so unit tests
// can point it at a guaranteed-closed port instead of the real :9092 (which
// might be held by a prior daemon on the dev box).
//
// Quirk: immediately after `launchctl kickstart -k` launchd can briefly
// report `state = not running` while it tears down the old process and
// spawns the new one. We tolerate that by only flipping to Crashed once
// we've seen the process running at least once — OR once the grace window
// has passed without ever seeing it alive (a genuine "never started").
func waitForReady(timeout time.Duration, port int) (Readiness, error) {
	deadline := time.Now().Add(timeout)
	// launchctl grace window: tolerate a brief "not running" immediately
	// after kickstart. 3s is plenty for launchd to fork the new process.
	graceDeadline := time.Now().Add(3 * time.Second)
	everSeenRunning := false
	for {
		st, err := ReadStatus()
		if err != nil {
			return ReadinessCrashed, err
		}
		if st.Running {
			everSeenRunning = true
			if tcpPortOpen(port) {
				return ReadinessReady, nil
			}
		} else if everSeenRunning || time.Now().After(graceDeadline) {
			// Process was running and is now gone → crashed.
			// Or we never saw it start within the grace window → crashed.
			return ReadinessCrashed, nil
		}
		if time.Now().After(deadline) {
			return ReadinessTimedOut, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// tcpPortOpen returns true if we can open a TCP connection to localhost:port.
// A short dial timeout keeps the poll cheap.
func tcpPortOpen(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 200*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// truncateLogFile zeros LogFile() so each restart begins with a fresh log.
// A missing file is fine (nothing to truncate); the first StoreHTTP creates
// it. Errors are returned so callers can log a warning — truncation failure
// shouldn't block a restart.
func truncateLogFile() error {
	path, err := LogFile()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("truncate %s: %w", path, err)
	}
	return f.Close()
}

// startDetached is the pidfile-based direct-spawn path. Used when no service
// is installed (dev/CI scenarios where a persistent LaunchAgent/unit isn't
// wanted).
func startDetached(opts StartOptions) (int, error) {
	if err := stopDetached(5 * time.Second); err != nil {
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
		args = []string{"pr", "list", "--all", "--ui", "--menu-bar", "--port=0", "--persist-port"}
	}

	logPath, err := LogFile()
	if err != nil {
		return 0, err
	}
	// O_TRUNC so each restart begins with a clean log — operators inspecting
	// `gavel system status` / pr-ui.log don't have to hunt past stale output
	// from prior runs.
	logFD, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return 0, fmt.Errorf("open log file %s: %w", logPath, err)
	}
	defer logFD.Close()

	fmt.Fprintf(logFD, "--- pr-ui start %s ---\n", time.Now().Format(time.RFC3339))

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

// Stop terminates the pr UI daemon. When a launchd / systemd user service is
// installed, Stop asks the service manager to stop the unit. Otherwise it
// signals SIGTERM to the pid recorded in the pidfile and waits up to timeout
// for it to exit, escalating to SIGKILL on timeout. Returns nil when no
// process is running (nothing to do).
func Stop(timeout time.Duration) error {
	installed, err := IsInstalled()
	if err != nil {
		return err
	}
	if installed {
		return serviceStop()
	}
	return stopDetached(timeout)
}

// stopDetached is the pidfile-based stop path.
func stopDetached(timeout time.Duration) error {
	st, err := readPidStatus()
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
