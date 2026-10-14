package fixtures

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"
)

// Wait outcomes recorded under Metadata["wait"].outcome.
const (
	WaitSatisfied = "satisfied"
	WaitFailed    = "failed"
	WaitTimeout   = "timeout"
)

// stepClock is the time source of waits and retries; tests inject a fake so
// the delay schedule is asserted without sleeping.
type stepClock interface {
	Now() time.Time
	Sleep(ctx context.Context, d time.Duration) error
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

func (realClock) Sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// jitter is a ± spread applied to every delay: a fraction of it ("20%") or an
// absolute duration ("500ms").
type jitter struct {
	fraction float64
	absolute time.Duration
}

func parseJitter(s string) (jitter, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return jitter{}, nil
	}
	if pct, ok := strings.CutSuffix(s, "%"); ok {
		value, err := strconv.ParseFloat(strings.TrimSpace(pct), 64)
		if err != nil || value < 0 || value > 100 {
			return jitter{}, fmt.Errorf("jitter %q: want a percentage between 0%% and 100%%", s)
		}
		return jitter{fraction: value / 100}, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil || d < 0 {
		return jitter{}, fmt.Errorf("jitter %q: want a percentage (\"20%%\") or a non-negative duration (\"500ms\")", s)
	}
	return jitter{absolute: d}, nil
}

// backoffDelay is the n-th (0-based) delay of a schedule: base·factor^n capped
// at backoff.max, then spread by ±jitter using rnd ∈ [0,1). Never negative.
func backoffDelay(base time.Duration, backoff *Backoff, spread jitter, n int, rnd func() float64) time.Duration {
	delay := float64(base)
	if backoff != nil && backoff.Factor > 0 {
		delay *= math.Pow(backoff.Factor, float64(n))
		if backoff.Max != nil && delay > float64(backoff.Max.Std()) {
			delay = float64(backoff.Max.Std())
		}
	}
	amount := spread.fraction * delay
	if spread.absolute > 0 {
		amount = float64(spread.absolute)
	}
	delay += (2*rnd() - 1) * amount
	if delay < 0 {
		return 0
	}
	return time.Duration(delay)
}

// validateWait rejects a wait policy that could not run to a verdict.
func validateWait(policy WaitPolicy) error {
	switch {
	case strings.TrimSpace(policy.Until) == "":
		return fmt.Errorf("wait: until is required")
	case policy.Interval == nil || policy.Interval.Std() <= 0:
		return fmt.Errorf("wait: interval must be a positive duration")
	case policy.Timeout == nil || policy.Timeout.Std() <= 0:
		return fmt.Errorf("wait: timeout must be a positive duration")
	case policy.Backoff != nil && policy.Backoff.Factor < 1:
		return fmt.Errorf("wait: backoff factor must be >= 1, got %v", policy.Backoff.Factor)
	}
	_, err := parseJitter(policy.Jitter)
	return err
}

// Built-in wait timing, applied under every other layer once a wait is active
// (the step declared `wait:` or its kind has a DefaultWait): the first re-poll
// after 1s, growing ×1.5 up to 10s, with a 2m deadline — or the step's own
// `timeout:` when it declares one. `until` has no built-in: a wait without
// one is rejected.
const (
	BuiltinWaitInterval      = time.Second
	BuiltinWaitBackoffFactor = 1.5
	BuiltinWaitBackoffMax    = 10 * time.Second
	BuiltinWaitTimeout       = 2 * time.Minute
)

// builtinWait is the lowest wait layer; stepTimeout, when set, replaces the
// built-in deadline.
func builtinWait(stepTimeout *Duration) *WaitPolicy {
	interval, max, timeout := Duration(BuiltinWaitInterval), Duration(BuiltinWaitBackoffMax), Duration(BuiltinWaitTimeout)
	if stepTimeout != nil {
		timeout = *stepTimeout
	}
	return &WaitPolicy{Interval: &interval, Timeout: &timeout,
		Backoff: &Backoff{Factor: BuiltinWaitBackoffFactor, Max: &max}}
}

// mergeWait layers the step's wait over the file defaults over the kind's
// default, field by field. File defaults only fill fields: a wait runs when the
// step or its kind asks for one, never because the file set timing defaults.
func mergeWait(step, file, kind *WaitPolicy) *WaitPolicy {
	if step == nil && kind == nil {
		return nil
	}
	merged := WaitPolicy{}
	for _, layer := range []*WaitPolicy{kind, file, step} {
		if layer == nil {
			continue
		}
		merged.Until = firstNonEmpty(layer.Until, merged.Until)
		merged.FailWhen = firstNonEmpty(layer.FailWhen, merged.FailWhen)
		merged.Jitter = firstNonEmpty(layer.Jitter, merged.Jitter)
		merged.Interval = firstDuration(layer.Interval, merged.Interval)
		merged.Timeout = firstDuration(layer.Timeout, merged.Timeout)
		if layer.Backoff != nil {
			merged.Backoff = layer.Backoff
		}
	}
	return &merged
}

// mergeRetry layers the step's retry over the file defaults, field by field.
func mergeRetry(step, file *RetryPolicy) *RetryPolicy {
	if step == nil && file == nil {
		return nil
	}
	merged := RetryPolicy{}
	for _, layer := range []*RetryPolicy{file, step} {
		if layer == nil {
			continue
		}
		if layer.Attempts > 0 {
			merged.Attempts = layer.Attempts
		}
		merged.When = firstNonEmpty(layer.When, merged.When)
		merged.Jitter = firstNonEmpty(layer.Jitter, merged.Jitter)
		merged.Delay = firstDuration(layer.Delay, merged.Delay)
		if layer.Backoff != nil {
			merged.Backoff = layer.Backoff
		}
	}
	return &merged
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func firstDuration(values ...*Duration) *Duration {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}

// observationTimeLayout is RFC3339 with milliseconds, the precision a UI parses.
const observationTimeLayout = "2006-01-02T15:04:05.000Z07:00"

// WaitObservation is one poll of a step's observable state.
type WaitObservation struct {
	At       string         `json:"at"` // RFC3339, millisecond precision
	DelayMs  int64          `json:"delayMs"`
	Until    bool           `json:"until"`
	FailWhen bool           `json:"failWhen"`
	Outputs  map[string]any `json:"outputs"`
	Error    string         `json:"error,omitempty"`
}

// WaitReport is recorded under Metadata["wait"]: the effective policy, how
// the wait ended, and every observation so a timeout shows its trajectory.
type WaitReport struct {
	Policy       WaitPolicy        `json:"policy"`
	Outcome      string            `json:"outcome"`
	Observations []WaitObservation `json:"observations"`
	Error        string            `json:"error,omitempty"`
	Trace        string            `json:"trace,omitempty"`
}

// celCheck evaluates a boolean CEL expression, returning the trace of a false
// or failed evaluation.
type celCheck func(expression string, vars map[string]any) (bool, string, error)

// waiter polls observe on the policy's backoff schedule until until holds,
// failWhen holds, or the timeout passes. The first observation is immediate.
type waiter struct {
	policy  WaitPolicy
	observe func(ctx context.Context) (map[string]any, error)
	vars    func(outputs map[string]any) map[string]any
	check   celCheck
	clock   stepClock
	rand    func() float64
}

func (w waiter) run(ctx context.Context) (WaitReport, map[string]any) {
	report := WaitReport{Policy: w.policy, Observations: []WaitObservation{}}
	spread, err := parseJitter(w.policy.Jitter)
	if err != nil {
		report.Outcome, report.Error = WaitFailed, err.Error()
		return report, nil
	}
	deadline := w.clock.Now().Add(w.policy.Timeout.Std())
	var last map[string]any
	var delay time.Duration
	for n := 0; ; n++ {
		obs, done := w.observeOnce(ctx, delay, &report)
		if len(obs.Outputs) > 0 {
			last = obs.Outputs
		}
		if done {
			return report, last
		}
		now := w.clock.Now()
		if !now.Before(deadline) {
			return w.timedOut(report, last), last
		}
		delay = backoffDelay(w.policy.Interval.Std(), w.policy.Backoff, spread, n, w.rand)
		if remaining := deadline.Sub(now); delay > remaining {
			delay = remaining
		}
		if err := w.clock.Sleep(ctx, delay); err != nil {
			report.Outcome, report.Error = WaitTimeout, fmt.Sprintf("wait interrupted: %v", err)
			return report, last
		}
	}
}

// observeOnce takes one observation and evaluates the policy against it; done
// reports a terminal verdict (satisfied or failed) already written to report.
func (w waiter) observeOnce(ctx context.Context, delay time.Duration, report *WaitReport) (WaitObservation, bool) {
	outputs, err := w.observe(ctx)
	obs := WaitObservation{At: w.clock.Now().UTC().Format(observationTimeLayout), DelayMs: delay.Milliseconds(), Outputs: outputs}
	if obs.Outputs == nil {
		obs.Outputs = map[string]any{}
	}
	if err != nil {
		obs.Error = err.Error()
		report.Observations = append(report.Observations, obs)
		return obs, false
	}
	vars := w.vars(outputs)
	if w.policy.FailWhen != "" {
		failed, _, err := w.check(w.policy.FailWhen, vars)
		if err != nil {
			return w.evalFailed(report, obs, "failWhen", w.policy.FailWhen, err)
		}
		obs.FailWhen = failed
		if failed {
			_, trace, _ := w.check("!("+w.policy.FailWhen+")", vars)
			report.Observations = append(report.Observations, obs)
			report.Outcome, report.Trace = WaitFailed, trace
			report.Error = fmt.Sprintf("wait failWhen %s is true", w.policy.FailWhen)
			return obs, true
		}
	}
	satisfied, _, err := w.check(w.policy.Until, vars)
	if err != nil {
		return w.evalFailed(report, obs, "until", w.policy.Until, err)
	}
	obs.Until = satisfied
	report.Observations = append(report.Observations, obs)
	if satisfied {
		report.Outcome = WaitSatisfied
	}
	return obs, satisfied
}

func (w waiter) evalFailed(report *WaitReport, obs WaitObservation, field, expression string, err error) (WaitObservation, bool) {
	obs.Error = err.Error()
	report.Observations = append(report.Observations, obs)
	report.Outcome = WaitFailed
	report.Error = fmt.Sprintf("wait %s %s: %v", field, expression, err)
	return obs, true
}

func (w waiter) timedOut(report WaitReport, last map[string]any) WaitReport {
	report.Outcome = WaitTimeout
	report.Error = fmt.Sprintf("wait until %s did not hold within %s", w.policy.Until, w.policy.Timeout.Std())
	if last != nil {
		_, report.Trace, _ = w.check(w.policy.Until, w.vars(last))
	}
	if n := len(report.Observations); n > 0 && report.Observations[n-1].Error != "" {
		report.Error += "; last observation failed: " + report.Observations[n-1].Error
	}
	return report
}

// defaultRand is the jitter source of real runs.
func defaultRand() float64 { return rand.Float64() }
