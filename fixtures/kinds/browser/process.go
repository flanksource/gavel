package browser

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"
	"time"

	cexec "github.com/flanksource/clicky/exec"
)

type processOutput struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (o *processOutput) Write(p []byte) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.b.Write(p)
}

func (o *processOutput) String() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.b.String()
}

type managedProcess struct {
	cmd     *exec.Cmd
	done    chan struct{}
	stdout  processOutput
	stderr  processOutput
	waitErr error
}

func startManagedProcess(ctx context.Context, binary string, args []string, cwd string, env map[string]string, output io.Writer) (*managedProcess, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	process := cexec.NewExec(binary, args...).WithExactEnv(env).WithProcessGroup()
	if cwd != "" {
		process.WithCwd(cwd)
	}
	cmd, err := process.PrepareCommand()
	if err != nil {
		return nil, err
	}
	h := &managedProcess{cmd: cmd, done: make(chan struct{})}
	cmd.Stdout = &h.stdout
	if output != nil {
		cmd.Stdout = io.MultiWriter(&h.stdout, output)
	}
	cmd.Stderr = &h.stderr
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %s: %w", binary, err)
	}
	go func() {
		h.waitErr = cmd.Wait()
		close(h.done)
	}()
	return h, nil
}

func (h *managedProcess) pid() int { return h.cmd.Process.Pid }

func (h *managedProcess) exited() bool {
	select {
	case <-h.done:
		return true
	default:
		return false
	}
}

func (h *managedProcess) exitCode() int {
	if !h.exited() {
		return -1
	}
	return h.cmd.ProcessState.ExitCode()
}

func (h *managedProcess) killTree() error {
	if err := syscall.Kill(-h.pid(), syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
		return fmt.Errorf("kill process group %d: %w", h.pid(), err)
	}
	return nil
}

func (h *managedProcess) waitWithin(timeout time.Duration) error {
	select {
	case <-h.done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("process %d was not reaped within %s", h.pid(), timeout)
	}
}
