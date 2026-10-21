package services

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/flanksource/gavel/fixtures"
)

// Readiness defaults, applied field by field under an entry's `wait:`.
const (
	defaultInterval = 200 * time.Millisecond
	defaultFactor   = 1.5
	defaultMax      = 2 * time.Second
	defaultTimeout  = 60 * time.Second
	probeTimeout    = 5 * time.Second
)

// readinessPolicy layers an entry's `wait:` over the readiness defaults, field
// by field — backoff.factor and backoff.max included. A readiness probe is not
// a CEL expression, so until stays empty.
func readinessPolicy(wait *fixtures.WaitPolicy) fixtures.WaitPolicy {
	interval, ceiling, timeout := fixtures.Duration(defaultInterval), fixtures.Duration(defaultMax), fixtures.Duration(defaultTimeout)
	policy := fixtures.WaitPolicy{Interval: &interval, Timeout: &timeout,
		Backoff: &fixtures.Backoff{Factor: defaultFactor, Max: &ceiling}}
	if wait == nil {
		return policy
	}
	policy.Jitter = wait.Jitter
	if wait.Interval != nil {
		policy.Interval = wait.Interval
	}
	if wait.Timeout != nil {
		policy.Timeout = wait.Timeout
	}
	if b := wait.Backoff; b != nil {
		if b.Factor != 0 {
			policy.Backoff.Factor = b.Factor
		}
		if b.Max != nil {
			policy.Backoff.Max = b.Max
		}
	}
	return policy
}

// probeResult is one readiness observation.
type probeResult struct {
	ready  bool
	status int
	err    error
}

// failure is why the probe does not hold: its error, else the http status it
// got, else nil once ready.
func (r probeResult) failure() error {
	switch {
	case r.err != nil:
		return r.err
	case r.ready:
		return nil
	case r.status != 0:
		return fmt.Errorf("http status %d", r.status)
	default:
		return fmt.Errorf("not ready")
	}
}

// probe runs the entry's ready check once; env and dir are the service's own.
func (r Ready) probe(ctx context.Context, env []string, dir string) probeResult {
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	switch {
	case r.HTTP != "":
		return httpProbe(ctx, r.HTTP, r.Status)
	case r.TCP != "":
		conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", r.TCP)
		if err != nil {
			return probeResult{err: err}
		}
		return probeResult{ready: true, err: conn.Close()}
	default:
		cmd := exec.CommandContext(ctx, "sh", "-c", r.Exec)
		cmd.Env, cmd.Dir = env, dir
		if out, err := cmd.CombinedOutput(); err != nil {
			return probeResult{err: fmt.Errorf("ready exec %q: %w: %s", r.Exec, err, strings.TrimSpace(string(out)))}
		}
		return probeResult{ready: true}
	}
}

func httpProbe(ctx context.Context, url string, want int) probeResult {
	if want == 0 {
		want = http.StatusOK
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return probeResult{err: err}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return probeResult{err: err}
	}
	_ = resp.Body.Close()
	return probeResult{ready: resp.StatusCode == want, status: resp.StatusCode}
}

// outputs is the observable readiness state: `ready`, plus `http.status` for
// an http probe.
func (r probeResult) outputs() map[string]any {
	observed := map[string]any{"ready": r.ready}
	if r.status != 0 {
		observed["http"] = map[string]any{"status": r.status}
	}
	return observed
}
