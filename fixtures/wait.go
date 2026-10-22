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
	if strings.TrimSpace(policy.Until) == "" {
		return fmt.Errorf("wait: until is required")
	}
	return policy.ValidateTiming()
}

// ValidateTiming rejects a schedule (interval, timeout, backoff, jitter) a
// wait could not run on.
func (policy WaitPolicy) ValidateTiming() error {
	switch {
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

type waitObservation struct {
	Until, FailWhen bool
	Outputs         map[string]any
	Error           string
}

// WaitReport records the terminal observation and actual time spent waiting.
type WaitReport struct {
	Policy     WaitPolicy     `json:"policy"`
	Outcome    string         `json:"outcome"`
	DurationMs int64          `json:"durationMs"`
	Outputs    map[string]any `json:"outputs"`
	Error      string         `json:"error,omitempty"`
	// Expression is the CEL expression whose evaluation decided the outcome:
	// failWhen when it held or could not be evaluated, otherwise until. Trace is
	// always its trace.
	Expression string `json:"expression,omitempty"`
	Trace      string `json:"trace,omitempty"`
	// Cause is the step kind's domain reason (FailureCauser) a wait that did not
	// succeed ended as it did, read from its last outputs.
	Cause string `json:"cause,omitempty"`
	// celVars are the variables a held failWhen was evaluated against.
	celVars map[string]any
}

// celCheck evaluates a boolean CEL expression, returning the trace of a false
// or failed evaluation.
type celCheck func(expression string, vars map[string]any) (bool, string, error)

// WaitProbe is one observation judged in Go instead of by CEL, such as a
// service readiness probe.
type WaitProbe struct {
	Outputs map[string]any
	// Until satisfies the wait; FailWhen ends it, failed, at once.
	Until, FailWhen bool
	// Error is why the probe does not hold. It is recorded on the observation
	// and names the last probe when the wait fails or times out.
	Error error
}

// WaitOptions drives Wait: Probe polled on Policy's timing.
type WaitOptions struct {
	// Policy supplies the timing only; until/failWhen are CEL and rejected.
	Policy WaitPolicy
	Probe  func(ctx context.Context) WaitProbe
	// Wake, when it closes, cuts the current delay short so a terminal change
	// (a process exiting) is probed at once.
	Wake <-chan struct{}
}

// Wait polls opts.Probe on the shared backoff schedule until it holds, fails
// or the timeout passes. An unrunnable policy is an error; every other ending
// is the report's outcome.
func Wait(ctx context.Context, opts WaitOptions) (WaitReport, error) {
	switch {
	case opts.Probe == nil:
		return WaitReport{}, fmt.Errorf("wait: a probe is required")
	case opts.Policy.Until != "" || opts.Policy.FailWhen != "":
		return WaitReport{}, fmt.Errorf("wait: until/failWhen are CEL; a probed wait is judged by its probe")
	}
	if err := opts.Policy.ValidateTiming(); err != nil {
		return WaitReport{}, err
	}
	report, _ := waiter{policy: opts.Policy, probe: opts.Probe, wake: opts.Wake, clock: realClock{}, rand: defaultRand}.run(ctx)
	return report, nil
}

// waiter polls on the policy's backoff schedule until the wait is satisfied,
// fails, or the timeout passes. The first observation is immediate. A probe
// judges each observation in Go; otherwise observe's outputs are judged by
// the policy's until/failWhen CEL.
type waiter struct {
	policy  WaitPolicy
	observe func(ctx context.Context) (map[string]any, error)
	vars    func(outputs map[string]any) map[string]any
	check   celCheck
	probe   func(ctx context.Context) WaitProbe
	wake    <-chan struct{}
	clock   stepClock
	rand    func() float64
}

func (w waiter) run(ctx context.Context) (report WaitReport, outputs map[string]any) {
	started := w.clock.Now()
	report = WaitReport{Policy: w.policy, Outputs: map[string]any{}}
	defer func() { report.DurationMs = w.clock.Now().Sub(started).Milliseconds() }()
	spread, err := parseJitter(w.policy.Jitter)
	if err != nil {
		report.Outcome, report.Error = WaitFailed, err.Error()
		return report, nil
	}
	deadline := started.Add(w.policy.Timeout.Std())
	var delay time.Duration
	for n := 0; ; n++ {
		obs, done := w.observeOnce(ctx, &report)
		report.Outputs = obs.Outputs
		if done {
			return report, report.Outputs
		}
		now := w.clock.Now()
		if !now.Before(deadline) {
			return w.timedOut(report, obs), report.Outputs
		}
		delay = backoffDelay(w.policy.Interval.Std(), w.policy.Backoff, spread, n, w.rand)
		if remaining := deadline.Sub(now); delay > remaining {
			delay = remaining
		}
		if err := w.sleep(ctx, delay); err != nil {
			report.Outcome, report.Error = WaitTimeout, fmt.Sprintf("wait interrupted: %v", err)
			return report, report.Outputs
		}
	}
}

// sleep waits out delay, returning early (without error) once wake closes.
func (w waiter) sleep(ctx context.Context, delay time.Duration) error {
	if w.wake == nil {
		return w.clock.Sleep(ctx, delay)
	}
	sleepCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		select {
		case <-w.wake:
			cancel()
		case <-sleepCtx.Done():
		}
	}()
	if err := w.clock.Sleep(sleepCtx, delay); err != nil && ctx.Err() != nil {
		return err
	}
	return nil
}

// probeOnce records one Go-judged probe; done reports a terminal verdict
// already written to report.
func (w waiter) probeOnce(ctx context.Context, report *WaitReport) (waitObservation, bool) {
	probe := w.probe(ctx)
	obs := waitObservation{Until: probe.Until, FailWhen: probe.FailWhen, Outputs: probe.Outputs}
	if obs.Outputs == nil {
		obs.Outputs = map[string]any{}
	}
	if probe.Error != nil {
		obs.Error = probe.Error.Error()
	}

	switch {
	case probe.FailWhen && probe.Error == nil:
		report.Outcome, report.Error = WaitFailed, "wait probe set FailWhen without an Error; a failing probe must say why"
	case probe.FailWhen:
		report.Outcome, report.Error = WaitFailed, obs.Error
	case probe.Until:
		report.Outcome = WaitSatisfied
	default:
		return obs, false
	}
	return obs, true
}

// observeOnce takes one observation and evaluates the policy against it; done
// reports a terminal verdict (satisfied or failed) already written to report.
func (w waiter) observeOnce(ctx context.Context, report *WaitReport) (waitObservation, bool) {
	if w.probe != nil {
		return w.probeOnce(ctx, report)
	}
	outputs, err := w.observe(ctx)
	obs := waitObservation{Outputs: outputs}
	if obs.Outputs == nil {
		obs.Outputs = map[string]any{}
	}
	if err != nil {
		obs.Error = err.Error()
		return obs, false
	}
	vars := w.vars(outputs)
	if w.policy.FailWhen != "" {
		failed, trace, err := w.check(w.policy.FailWhen, vars)

		if err != nil {
			return w.evalFailed(report, obs, "failWhen", w.policy.FailWhen, trace, err)

		}
		obs.FailWhen = failed
		if failed {
			_, trace, _ := w.check("!("+w.policy.FailWhen+")", vars)
			report.Outcome, report.Expression, report.Trace, report.celVars = WaitFailed, w.policy.FailWhen, trace, vars
			report.Error = fmt.Sprintf("wait failWhen %s is true", w.policy.FailWhen)
			return obs, true
		}
	}
	satisfied, trace, err := w.check(w.policy.Until, vars)
	if err != nil {
		return w.evalFailed(report, obs, "until", w.policy.Until, trace, err)

	}
	obs.Until = satisfied

	if satisfied {
		report.Outcome, report.Expression = WaitSatisfied, w.policy.Until
	}
	return obs, satisfied
}

func (w waiter) evalFailed(report *WaitReport, obs waitObservation, field, expression, trace string, err error) (waitObservation, bool) {
	obs.Error = err.Error()

	report.Outcome, report.Expression = WaitFailed, expression
	report.Trace = trace
	report.Error = fmt.Sprintf("wait %s %s: %v", field, expression, err)
	return obs, true
}

func (w waiter) timedOut(report WaitReport, last waitObservation) WaitReport {
	report.Outcome = WaitTimeout
	if w.probe != nil {
		lastProbe := firstNonEmpty(last.Error, "not ready")
		report.Error = fmt.Sprintf("not ready after %s: last probe: %s", w.policy.Timeout.Std(), lastProbe)
		return report
	}
	report.Expression = w.policy.Until
	report.Error = fmt.Sprintf("wait until %s did not hold within %s", w.policy.Until, w.policy.Timeout.Std())
	if last.Outputs != nil && last.Error == "" {
		_, report.Trace, _ = w.check(w.policy.Until, w.vars(last.Outputs))
	}
	if last.Error != "" {
		report.Error += "; last observation failed: " + last.Error
	}
	return report
}

// defaultRand is the jitter source of real runs.
func defaultRand() float64 { return rand.Float64() }
