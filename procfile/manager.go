package procfile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
)

// stopTimeout is how long Stop waits for the supervisor to exit after SIGTERM
// before escalating to SIGKILL.
const stopTimeout = 10 * time.Second

// resolveTarget discovers the Procfile from workDir (defaulting to the current
// directory) and anchors the project root at the Procfile's directory, so
// `proc stop`/`status` from any subdirectory address the same state.
func resolveTarget(workDir, pfOverride string) (root, procfile string, err error) {
	if workDir == "" {
		if workDir, err = os.Getwd(); err != nil {
			return "", "", fmt.Errorf("resolve working dir: %w", err)
		}
	}
	workDir, _ = filepath.Abs(workDir)
	procfile = Find(workDir, pfOverride)
	if procfile == "" {
		return "", "", fmt.Errorf("no Procfile found at or above %s (use --procfile to point at one)", workDir)
	}
	return filepath.Dir(procfile), procfile, nil
}

// loadConfig reads the procfile section of the merged .gavel.yaml for root.
func loadConfig(root string) (verify.ProcfileConfig, error) {
	cfg, err := verify.LoadGavelConfig(root)
	if err != nil {
		return verify.ProcfileConfig{}, err
	}
	return cfg.Procfile, nil
}

// Run builds and runs the supervisor inline (the `gavel proc run` form). With
// foreground=true it multiplexes process output to stdout; the detached daemon
// started by Start re-execs this with foreground=false.
func Run(workDir, pfOverride string, names []string, foreground bool) error {
	root, pf, err := resolveTarget(workDir, pfOverride)
	if err != nil {
		return err
	}
	cfg, err := loadConfig(root)
	if err != nil {
		return err
	}
	sup, err := NewSupervisor(Options{Root: root, Procfile: pf, Names: names, Foreground: foreground, Config: cfg})
	if err != nil {
		return err
	}
	return sup.Run()
}

// Start launches the supervisor as a detached background daemon and waits for it
// to come up. It refuses to start a second daemon for the same root.
func Start(workDir, pfOverride string, names []string) (*StatusReport, error) {
	root, pf, err := resolveTarget(workDir, pfOverride)
	if err != nil {
		return nil, err
	}
	return startFor(root, pf, names)
}

func startFor(root, pf string, names []string) (*StatusReport, error) {
	dir, err := StateDir(root)
	if err != nil {
		return nil, err
	}
	if st, err := ReadState(dir); err != nil {
		return nil, err
	} else if st.Running() {
		return nil, fmt.Errorf("gavel proc is already running (supervisor pid %d) — use `gavel proc restart`", st.SupervisorPID)
	}
	if err := spawnDetached(root, pf, names, dir); err != nil {
		return nil, err
	}
	if err := waitRunning(dir, 5*time.Second); err != nil {
		return nil, err
	}
	return statusFor(root, pf)
}

// spawnDetached re-execs `gavel proc run --detached` in a new session with its
// stdout/stderr redirected to supervisor.log. cmd.Dir is the project root so the
// child rediscovers the same Procfile and state directory.
func spawnDetached(root, pf string, names []string, dir string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve gavel binary: %w", err)
	}
	bin, _ := filepath.Abs(exe)

	args := []string{"proc", "run", "--detached", "--procfile", pf}
	args = append(args, names...)

	logPath := SupervisorLogPath(dir)
	logFD, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open supervisor log %s: %w", logPath, err)
	}
	defer logFD.Close()

	cmd := exec.Command(bin, args...)
	cmd.Dir = root
	cmd.Stdout = logFD
	cmd.Stderr = logFD
	cmd.Stdin = nil
	cmd.SysProcAttr = detachedSysProcAttr()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start detached supervisor: %w", err)
	}
	if err := cmd.Process.Release(); err != nil {
		logger.Warnf("release supervisor process: %v", err)
	}
	return nil
}

// waitRunning polls state.json until the supervisor reports running, surfacing
// the tail of the supervisor log on timeout so startup failures are visible.
func waitRunning(dir string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		st, err := ReadState(dir)
		if err == nil && st.Running() {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	tail, _ := utils.TailFile(SupervisorLogPath(dir), 20)
	if len(tail) > 0 {
		return fmt.Errorf("supervisor did not come up within %s; last log lines:\n%s", timeout, strings.Join(tail, "\n"))
	}
	return fmt.Errorf("supervisor did not come up within %s", timeout)
}

// Stop stops the whole daemon (no names) by signalling the supervisor, or the
// named processes via the control socket. It is idempotent when nothing runs.
func Stop(workDir, pfOverride string, names []string) (*StatusReport, error) {
	root, pf, err := resolveTarget(workDir, pfOverride)
	if err != nil {
		return nil, err
	}
	dir, err := StateDir(root)
	if err != nil {
		return nil, err
	}
	st, err := ReadState(dir)
	if err != nil {
		return nil, err
	}
	if !st.Running() {
		return statusFor(root, pf)
	}
	if len(names) == 0 {
		if err := signalAndWait(st.SupervisorPID, stopTimeout); err != nil {
			return nil, err
		}
		return statusFor(root, pf)
	}
	if _, err := sendControl(root, ctrlRequest{Action: actionStop, Names: names}); err != nil {
		return nil, err
	}
	return statusFor(root, pf)
}

// Restart restarts the named processes (or all) on a running supervisor, or
// starts the daemon when none is running.
func Restart(workDir, pfOverride string, names []string) (*StatusReport, error) {
	root, pf, err := resolveTarget(workDir, pfOverride)
	if err != nil {
		return nil, err
	}
	dir, err := StateDir(root)
	if err != nil {
		return nil, err
	}
	st, err := ReadState(dir)
	if err != nil {
		return nil, err
	}
	if !st.Running() {
		return startFor(root, pf, names)
	}
	if _, err := sendControl(root, ctrlRequest{Action: actionRestart, Names: names}); err != nil {
		return nil, err
	}
	return statusFor(root, pf)
}

// Status returns the merged Procfile + supervisor status.
func Status(workDir, pfOverride string) (*StatusReport, error) {
	root, pf, err := resolveTarget(workDir, pfOverride)
	if err != nil {
		return nil, err
	}
	return statusFor(root, pf)
}

func statusFor(root, pf string) (*StatusReport, error) {
	report, err := gather(root, pf)
	if err != nil {
		return nil, err
	}
	return &StatusReport{report}, nil
}

// List returns the configured processes and their commands.
func List(workDir, pfOverride string) (*ListReport, error) {
	root, pf, err := resolveTarget(workDir, pfOverride)
	if err != nil {
		return nil, err
	}
	report, err := gather(root, pf)
	if err != nil {
		return nil, err
	}
	return &ListReport{report}, nil
}

// signalAndWait sends SIGTERM to pid, waits up to timeout for it to exit, then
// escalates to SIGKILL. Mirrors service.stopDetached for the supervisor daemon.
func signalAndWait(pid int, timeout time.Duration) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find supervisor pid %d: %w", pid, err)
	}
	if err := p.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("signal SIGTERM to supervisor pid %d: %w", pid, err)
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !utils.ProcessAlive(pid) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	logger.Warnf("supervisor pid %d did not exit within %s, sending SIGKILL", pid, timeout)
	if err := p.Signal(syscall.SIGKILL); err != nil {
		return fmt.Errorf("signal SIGKILL to supervisor pid %d: %w", pid, err)
	}
	return nil
}
