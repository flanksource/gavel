// Package sandbox runs gavel jobs in isolated Kubernetes Jobs or containers for hosted mode.
// Jobs never receive database or GitHub credentials; they report back through the run API.
package sandbox

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Job is one sandboxed run.
type Job struct {
	RunID   uuid.UUID
	Scope   tenant.Scope
	Image   string
	Command []string // e.g. ["gavel","runner","exec","--run","<id>"]
	// Env holds non-secret settings: GAVEL_SERVER, GAVEL_RUN_ID, GAVEL_DB=off, CI=1.
	Env map[string]string
	// SecretEnv holds credentials, GAVEL_RUN_TOKEN and model API keys, which never appear on a
	// command line or in a pod spec.
	SecretEnv map[string]string
	Limits    Limits
}

// Limits bounds a job. CPU and Memory are Kubernetes quantities ("2", "500m", "4Gi").
type Limits struct {
	CPU     string
	Memory  string
	Timeout time.Duration
}

// DefaultLimits apply to whatever a job leaves unset.
var DefaultLimits = Limits{CPU: "2", Memory: "4Gi", Timeout: 60 * time.Minute}

// withDefaults fills unset limits from DefaultLimits.
func (l Limits) withDefaults() Limits {
	if l.CPU == "" {
		l.CPU = DefaultLimits.CPU
	}
	if l.Memory == "" {
		l.Memory = DefaultLimits.Memory
	}
	if l.Timeout == 0 {
		l.Timeout = DefaultLimits.Timeout
	}
	return l
}

// Result is how a job ended.
type Result struct {
	ExitCode int
}

// Handle follows a submitted job.
type Handle interface {
	// Logs follows the job's output until it exits.
	Logs(ctx context.Context) (io.ReadCloser, error)
	Wait(ctx context.Context) (Result, error)
	Cancel(ctx context.Context) error
	// Ref names the job in its backend: a Kubernetes Job name or a container name.
	Ref() string
}

// Executor submits jobs to a backend.
type Executor interface {
	Submit(ctx context.Context, job Job) (Handle, error)
	Name() string
}

// jobName is the backend name of run's job.
func jobName(run uuid.UUID) string {
	return "gavel-run-" + run.String()
}

// sortedEnv lists env as KEY=VALUE, sorted, rejecting a name that would split differently.
func sortedEnv(env map[string]string) ([]string, error) {
	pairs := make([]string, 0, len(env))
	for key, value := range env {
		if key == "" || strings.ContainsAny(key, "=\n\x00") {
			return nil, fmt.Errorf("sandbox job has an invalid environment variable name %q", key)
		}
		if strings.ContainsAny(value, "\n\x00") {
			return nil, fmt.Errorf("sandbox job's environment variable %s has a newline or NUL in its value", key)
		}
		pairs = append(pairs, key+"="+value)
	}
	sort.Strings(pairs)
	return pairs, nil
}

// sortedKeys lists env's names in order.
func sortedKeys(env map[string]string) []string {
	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// cpuCores is a Kubernetes CPU quantity as a number of cores.
func cpuCores(quantity string) (string, error) {
	parsed, err := resource.ParseQuantity(quantity)
	if err != nil || parsed.Sign() <= 0 {
		return "", fmt.Errorf("sandbox job's CPU limit %q is not a positive quantity such as 2 or 500m", quantity)
	}
	return fmt.Sprintf("%g", parsed.AsApproximateFloat64()), nil
}

// memoryBytes is a Kubernetes memory quantity in bytes.
func memoryBytes(quantity string) (int64, error) {
	parsed, err := resource.ParseQuantity(quantity)
	if err != nil || parsed.Sign() <= 0 {
		return 0, fmt.Errorf("sandbox job's memory limit %q is not a positive quantity such as 4Gi", quantity)
	}
	return parsed.Value(), nil
}
