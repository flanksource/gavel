// Runs sandbox jobs as docker or podman containers on one host: read-only, unprivileged, on a
// bridge network whose host firewall blocks private ranges.
package sandbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// containerRunArgs are the `run` arguments for job on network. envFile holds SecretEnv; empty
// when the job has none.
func containerRunArgs(job Job, network, envFile string) ([]string, error) {
	if job.Image == "" {
		return nil, fmt.Errorf("sandbox job has no image")
	}
	if len(job.Command) == 0 {
		return nil, fmt.Errorf("sandbox job has no command")
	}
	limits := job.Limits.withDefaults()
	cpus, err := cpuCores(limits.CPU)
	if err != nil {
		return nil, err
	}
	memory, err := memoryBytes(limits.Memory)
	if err != nil {
		return nil, err
	}
	env, err := sortedEnv(job.Env)
	if err != nil {
		return nil, err
	}
	args := []string{
		"run", "-d", "--name", jobName(job.RunID),
		"--network", network,
		"--user", "1000:1000", "--read-only",
		// The job runs the repository's code: go test builds into /tmp and installed tools run
		// from the home directory, so every tmpfs allows exec. A tmpfs given options is root's,
		// mode 0755, so the job's own are made the job user's.
		"--tmpfs", "/tmp:exec", "--tmpfs", "/workspace:exec,size=10g,uid=1000,gid=1000", "--tmpfs", "/home/gavel:exec,uid=1000,gid=1000",
		"--cap-drop", "ALL", "--security-opt", "no-new-privileges", "--pids-limit", "1024",
		"--cpus", cpus, "--memory", strconv.FormatInt(memory, 10),
	}
	for _, pair := range env {
		args = append(args, "-e", pair)
	}
	if envFile != "" {
		args = append(args, "--env-file", envFile)
	}
	args = append(args,
		"--label", "gavel.dev/run="+job.RunID.String(),
		"--label", "gavel.dev/org="+job.Scope.OrgID.String(),
		// The command is the whole argv, replacing the image's entrypoint, as a Kubernetes
		// container's command does.
		"--entrypoint", job.Command[0],
		job.Image)
	return append(args, job.Command[1:]...), nil
}

// ContainerOptions configures the container executor.
type ContainerOptions struct {
	// Network is the bridge jobs join, e.g. gavel-sandbox; see CheckHost.
	Network string
}

type containerExecutor struct {
	binary  string
	options ContainerOptions
}

// NewContainer returns an executor that runs jobs with binary, docker or podman.
func NewContainer(binary string, options ContainerOptions) Executor {
	return &containerExecutor{binary: binary, options: options}
}

func (e *containerExecutor) Name() string { return e.binary }

func (e *containerExecutor) Submit(ctx context.Context, job Job) (Handle, error) {
	envFile, err := writeEnvFile(job.SecretEnv)
	if err != nil {
		return nil, err
	}
	if envFile != "" {
		// The container reads it on start; it isn't needed after.
		defer os.Remove(envFile)
	}
	args, err := containerRunArgs(job, e.options.Network, envFile)
	if err != nil {
		return nil, err
	}
	name := jobName(job.RunID)
	if _, err := e.run(ctx, args...); err != nil {
		return nil, fmt.Errorf("start sandbox job %s: %w", name, err)
	}
	return &containerHandle{executor: e, name: name, timeout: job.Limits.withDefaults().Timeout, started: time.Now()}, nil
}

// run runs the container binary and returns its stdout; a failure carries its stderr.
func (e *containerExecutor) run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, e.binary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if message := strings.TrimSpace(stderr.String()); message != "" {
			return "", fmt.Errorf("%s %s: %w: %s", e.binary, args[0], err, message)
		}
		return "", fmt.Errorf("%s %s: %w", e.binary, args[0], err)
	}
	return stdout.String(), nil
}

// writeEnvFile writes env to a temporary file only its owner can read, for --env-file. It
// returns "" when env is empty.
func writeEnvFile(env map[string]string) (string, error) {
	if len(env) == 0 {
		return "", nil
	}
	pairs, err := sortedEnv(env)
	if err != nil {
		return "", err
	}
	file, err := os.CreateTemp("", "gavel-run-*.env")
	if err != nil {
		return "", fmt.Errorf("create the sandbox job's secret env file: %w", err)
	}
	if _, err := file.WriteString(strings.Join(pairs, "\n") + "\n"); err != nil {
		_ = file.Close()
		_ = os.Remove(file.Name())
		return "", fmt.Errorf("write the sandbox job's secret env file: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(file.Name())
		return "", fmt.Errorf("write the sandbox job's secret env file: %w", err)
	}
	return file.Name(), nil
}

type containerHandle struct {
	executor *containerExecutor
	name     string
	timeout  time.Duration
	started  time.Time
}

func (h *containerHandle) Ref() string { return h.name }

func (h *containerHandle) Logs(ctx context.Context) (io.ReadCloser, error) {
	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, h.executor.binary, "logs", "-f", h.name)
	reader, writer := io.Pipe()
	cmd.Stdout, cmd.Stderr = writer, writer
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("follow the logs of sandbox job %s: %w", h.name, err)
	}
	go func() {
		err := cmd.Wait()
		if err != nil && ctx.Err() == nil {
			err = fmt.Errorf("follow the logs of sandbox job %s: %w", h.name, err)
		} else {
			err = nil
		}
		_ = writer.CloseWithError(err)
	}()
	return &followedLogs{Reader: reader, stop: cancel}, nil
}

type followedLogs struct {
	io.Reader
	stop context.CancelFunc
}

func (l *followedLogs) Close() error {
	l.stop()
	return nil
}

func (h *containerHandle) Wait(ctx context.Context) (Result, error) {
	deadline, cancel := context.WithDeadline(ctx, h.started.Add(h.timeout))
	defer cancel()
	out, err := h.executor.run(deadline, "wait", h.name)
	if deadline.Err() == context.DeadlineExceeded && ctx.Err() == nil {
		cancelErr := h.Cancel(context.WithoutCancel(ctx))
		return Result{}, errors.Join(fmt.Errorf("sandbox job %s timed out after %s", h.name, h.timeout), cancelErr)
	}
	if err != nil {
		return Result{}, fmt.Errorf("wait for sandbox job %s: %w", h.name, err)
	}
	code, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return Result{}, fmt.Errorf("wait for sandbox job %s: %s printed %q, not an exit code", h.name, h.executor.binary, strings.TrimSpace(out))
	}
	if _, err := h.executor.run(ctx, "rm", h.name); err != nil {
		return Result{ExitCode: code}, fmt.Errorf("remove sandbox job %s: %w", h.name, err)
	}
	return Result{ExitCode: code}, nil
}

func (h *containerHandle) Cancel(ctx context.Context) error {
	if _, err := h.executor.run(ctx, "rm", "-f", h.name); err != nil {
		return fmt.Errorf("cancel sandbox job %s: %w", h.name, err)
	}
	return nil
}
