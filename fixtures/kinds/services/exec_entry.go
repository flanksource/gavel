package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	cexec "github.com/flanksource/clicky/exec"

	"github.com/flanksource/gavel/fixtures"
)

const (
	stopGrace    = 5 * time.Second
	logTailLines = 30
)

var errExited = errors.New("exited before it was ready")

// execService is a supervised `sh -c` process in its own process group, with
// stdout/stderr in a log file under the step's TmpDir.
type execService struct {
	name    string
	ready   Ready
	cmd     *exec.Cmd
	logPath string
	done    chan struct{}
	waitErr error
}

// startExec starts the entry and blocks until its ready probe passes. Its
// outputs record the readiness probes under `wait`, also when it never became
// ready. On any failure the process is stopped and the error carries the log
// tail.
func startExec(ctx context.Context, step *fixtures.StepContext, entry Entry, port int) (*execService, map[string]any, error) {
	svc, err := launch(step, entry)
	if err != nil {
		return nil, nil, fmt.Errorf("entry %q: %w", entry.Name, err)
	}
	if err := fixtures.ReportProcess(ctx, svc.cmd.Process.Pid); err != nil {
		return nil, nil, errors.Join(fmt.Errorf("entry %q: %w", entry.Name, err), svc.stop())
	}
	report, err := fixtures.Wait(ctx, fixtures.WaitOptions{Policy: readinessPolicy(entry.Wait), Probe: svc.readiness, Wake: svc.done})
	if err != nil {
		return nil, nil, errors.Join(fmt.Errorf("entry %q: %w", entry.Name, err), svc.stop())
	}
	readiness, err := fixtures.PlainValue(report)
	if err != nil {
		return nil, nil, errors.Join(fmt.Errorf("entry %q readiness report: %w", entry.Name, err), svc.stop())
	}
	if report.Outcome != fixtures.WaitSatisfied {
		stopErr := svc.stop()
		if report.Outcome == fixtures.WaitFailed {
			stopErr = nil // the exit is already the reported failure
		}
		return nil, map[string]any{"type": "exec", "ready": false, "wait": readiness}, errors.Join(fmt.Errorf("service %q %s\n--- %s (last %d lines) ---\n%s",
			entry.Name, report.Error, svc.logPath, logTailLines, svc.logTail()), stopErr)
	}

	outputs := map[string]any{
		"type": "exec", "pid": svc.cmd.Process.Pid, "log": svc.logPath, "ready": true, "env": stringMap(entry.Env), "wait": readiness,
	}
	if port != 0 {
		outputs["port"] = port
		outputs["host"] = "127.0.0.1"
		outputs["url"] = "http://127.0.0.1:" + strconv.Itoa(port)
	}
	return svc, outputs, nil
}

func launch(step *fixtures.StepContext, entry Entry) (*execService, error) {
	logDir := filepath.Join(step.TmpDir, "services")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}
	logPath := filepath.Join(logDir, entry.Name+".log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = logFile.Close() }()

	dir := entry.Cwd
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(step.WorkDir, dir)
	}
	process := cexec.NewExec("sh", "-c", entry.Exec).WithCwd(dir).WithProcessGroup().WithExactEnv(step.ChildEnv())
	if err := step.Opts.ProcessRuntime.Apply(process); err != nil {
		return nil, err
	}
	cmd, err := process.WithEnv(entry.Env).PrepareCommand()
	if err != nil {
		return nil, err
	}
	cmd.Stdout, cmd.Stderr = logFile, logFile
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %q: %w", entry.Exec, err)
	}
	svc := &execService{name: entry.Name, ready: *entry.Ready, cmd: cmd, logPath: logPath, done: make(chan struct{})}
	go func() {
		svc.waitErr = cmd.Wait()
		close(svc.done)
	}()
	return svc, nil
}

// readiness probes once for the readiness wait; a process that exited fails
// the wait at once.
func (s *execService) readiness(ctx context.Context) fixtures.WaitProbe {
	select {
	case <-s.done:
		return fixtures.WaitProbe{Outputs: map[string]any{"ready": false}, FailWhen: true, Error: fmt.Errorf("%w (%v)", errExited, s.waitErr)}
	default:
	}
	result := s.ready.probe(ctx, s.cmd.Env, s.cmd.Dir)
	return fixtures.WaitProbe{Outputs: result.outputs(), Until: result.ready, Error: result.failure()}
}

// observe re-probes readiness for Observer.Observe.
func (s *execService) observe(ctx context.Context) map[string]any {
	select {
	case <-s.done:
		return map[string]any{"ready": false, "error": fmt.Sprintf("exited: %v", s.waitErr)}
	default:
	}
	result := s.ready.probe(ctx, s.cmd.Env, s.cmd.Dir)
	observed := result.outputs()
	if result.err != nil {
		observed["error"] = result.err.Error()
	}
	return observed
}

// stop sends SIGTERM to the process group, escalating to SIGKILL after
// stopGrace. Exits caused by those signals are expected; any other exit
// status (a crash during the fixture) is reported.
func (s *execService) stop() error {
	pgid := -s.cmd.Process.Pid
	signalErr := signalGroup(pgid, syscall.SIGTERM)
	var forced error
	deadline := time.Now().Add(stopGrace)
	for groupAlive(pgid) && time.Now().Before(deadline) {
		time.Sleep(25 * time.Millisecond)
	}
	if groupAlive(pgid) {
		forced = errors.Join(fmt.Errorf("service %q required SIGKILL after %s", s.name, stopGrace), signalGroup(pgid, syscall.SIGKILL))
	}
	select {
	case <-s.done:
	case <-time.After(stopGrace):
		return errors.Join(signalErr, forced, fmt.Errorf("service %q (pid %d) did not exit after teardown", s.name, s.cmd.Process.Pid))
	}
	// The leader is the `sh -c` wrapper: it can exit while the service it
	// started is still handling SIGTERM, so teardown waits for the whole group.
	forced = errors.Join(forced, awaitGroupExit(s.name, pgid))
	// The leader is reaped now. On darwin, signalling a group whose members
	// are all zombies fails with EPERM — that is a group that already exited,
	// not a live process we could not stop.
	if errors.Is(signalErr, syscall.EPERM) {
		signalErr = nil
	}
	waitErr := s.waitErr
	var exit *exec.ExitError
	if errors.As(waitErr, &exit) {
		if status, ok := exit.Sys().(syscall.WaitStatus); ok && status.Signaled() &&
			(status.Signal() == syscall.SIGTERM || status.Signal() == syscall.SIGKILL) {
			waitErr = nil
		}
	}
	if waitErr != nil {
		waitErr = fmt.Errorf("service %q: %w\n%s", s.name, waitErr, s.logTail())
	}
	return errors.Join(signalErr, forced, waitErr)
}

func groupAlive(pgid int) bool {
	err := syscall.Kill(pgid, 0)
	return err == nil
}

// signalGroup signals a process group; a group that no longer exists is not
// an error.
func signalGroup(pgid int, signal syscall.Signal) error {
	if err := syscall.Kill(pgid, signal); err != nil && !errors.Is(err, syscall.ESRCH) {
		return fmt.Errorf("signal %s to process group %d: %w", signal, -pgid, err)
	}
	return nil
}

// awaitGroupExit waits up to stopGrace for every member of the process group
// to exit, then SIGKILLs what is left. EPERM means only zombies remain (darwin).
func awaitGroupExit(name string, pgid int) error {
	deadline := time.Now().Add(stopGrace)
	for syscall.Kill(pgid, 0) == nil {
		if time.Now().After(deadline) {
			return errors.Join(fmt.Errorf("service %q left processes in group %d after %s; sent SIGKILL", name, -pgid, stopGrace),
				signalGroup(pgid, syscall.SIGKILL))
		}
		time.Sleep(20 * time.Millisecond)
	}
	return nil
}

func (s *execService) logTail() string {
	data, err := os.ReadFile(s.logPath)
	if err != nil {
		return fmt.Sprintf("<read log: %v>", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	return strings.Join(lines[max(0, len(lines)-logTailLines):], "\n")
}

func stringMap(in map[string]string) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
