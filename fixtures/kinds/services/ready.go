package services

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"net"
	"net/http"
	"os/exec"
	"strconv"
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

// schedule is the probe cadence an entry's `wait:` describes: interval grown
// by factor up to max, ± jitter, within timeout.
//
// It reads fixtures.WaitPolicy's fields directly. Once the generic CEL waiter
// lands in package fixtures, the delay computation should move onto it.
type schedule struct {
	interval, max, timeout time.Duration
	factor                 float64
	jitterFraction         float64
	jitterAbsolute         time.Duration
}

func scheduleFor(policy *fixtures.WaitPolicy) (schedule, error) {
	s := schedule{interval: defaultInterval, max: defaultMax, timeout: defaultTimeout, factor: defaultFactor}
	if policy == nil {
		return s, nil
	}
	if policy.Interval != nil {
		s.interval = policy.Interval.Std()
	}
	if policy.Timeout != nil {
		s.timeout = policy.Timeout.Std()
	}
	if b := policy.Backoff; b != nil {
		if b.Factor != 0 {
			s.factor = b.Factor
		}
		if b.Max != nil {
			s.max = b.Max.Std()
		}
	}
	if s.interval <= 0 || s.timeout <= 0 || s.factor < 1 {
		return s, fmt.Errorf("wait: interval and timeout must be positive and backoff.factor >= 1 (got interval=%s timeout=%s factor=%v)", s.interval, s.timeout, s.factor)
	}
	return s, s.parseJitter(policy.Jitter)
}

func (s *schedule) parseJitter(jitter string) error {
	if jitter == "" {
		return nil
	}
	if pct, ok := strings.CutSuffix(jitter, "%"); ok {
		value, err := strconv.ParseFloat(pct, 64)
		if err != nil || value < 0 || value > 100 {
			return fmt.Errorf("wait.jitter %q: a percentage must be between 0%% and 100%%", jitter)
		}
		s.jitterFraction = value / 100
		return nil
	}
	value, err := time.ParseDuration(jitter)
	if err != nil || value < 0 {
		return fmt.Errorf("wait.jitter %q must be a percentage (\"20%%\") or a duration (\"500ms\")", jitter)
	}
	s.jitterAbsolute = value
	return nil
}

// delay is the wait before observation attempt+1; random returns [0,1).
func (s schedule) delay(attempt int, random func() float64) time.Duration {
	d := float64(s.interval) * math.Pow(s.factor, float64(attempt))
	if s.max > 0 && d > float64(s.max) {
		d = float64(s.max)
	}
	spread := float64(s.jitterAbsolute)
	if s.jitterFraction > 0 {
		spread = d * s.jitterFraction
	}
	return time.Duration(max(0, d+(random()*2-1)*spread))
}

// probeResult is one readiness observation.
type probeResult struct {
	ready  bool
	status int
	err    error
}

func (r probeResult) String() string {
	switch {
	case r.err != nil:
		return r.err.Error()
	case r.status != 0:
		return fmt.Sprintf("http status %d", r.status)
	default:
		return "not ready"
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

// awaitReady probes on the schedule until the probe passes, the process
// exits (done closes), ctx ends or the timeout expires. Every failure names
// the last observation.
func awaitReady(ctx context.Context, sched schedule, probe func(context.Context) probeResult, done <-chan struct{}) (probeResult, error) {
	deadline := time.Now().Add(sched.timeout)
	var last probeResult
	for attempt := 0; ; attempt++ {
		select {
		case <-done:
			return last, errExited
		default:
		}
		if last = probe(ctx); last.ready {
			return last, nil
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return last, fmt.Errorf("not ready after %s: last probe: %s", sched.timeout, last)
		}
		timer := time.NewTimer(min(sched.delay(attempt, rand.Float64), remaining))
		select {
		case <-ctx.Done():
			timer.Stop()
			return last, ctx.Err()
		case <-done:
			timer.Stop()
			return last, errExited
		case <-timer.C:
		}
	}
}
