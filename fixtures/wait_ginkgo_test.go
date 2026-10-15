package fixtures

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type fakeClock struct {
	now   time.Time
	slept []time.Duration
}

func (c *fakeClock) Now() time.Time { return c.now }

func (c *fakeClock) Sleep(ctx context.Context, d time.Duration) error {
	c.slept = append(c.slept, d)
	c.now = c.now.Add(d)
	return ctx.Err()
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}
}

// sleepUntilCancelled never finishes a delay on its own: only a cancelled
// context ends its Sleep.
type sleepUntilCancelled struct{ *fakeClock }

func (sleepUntilCancelled) Sleep(ctx context.Context, _ time.Duration) error {
	<-ctx.Done()
	return ctx.Err()
}

func fixedRand(value float64) func() float64 { return func() float64 { return value } }

func dur(d time.Duration) *Duration {
	value := Duration(d)
	return &value
}

var _ = Describe("backoffDelay", func() {
	backoff := &Backoff{Factor: 2, Max: dur(5 * time.Second)}

	It("grows interval·factor^n and caps it at max", func() {
		var delays []time.Duration
		for n := 0; n < 5; n++ {
			delays = append(delays, backoffDelay(time.Second, backoff, jitter{}, n, fixedRand(0.5)))
		}
		Expect(delays).To(Equal([]time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 5 * time.Second, 5 * time.Second}))
	})

	It("keeps a percentage jitter within ±fraction of the capped delay", func() {
		spread, err := parseJitter("20%")
		Expect(err).ToNot(HaveOccurred())
		low := backoffDelay(2*time.Second, backoff, spread, 1, fixedRand(0))
		high := backoffDelay(2*time.Second, backoff, spread, 1, fixedRand(0.999999))
		Expect(low).To(Equal(3200 * time.Millisecond))
		Expect(high).To(BeNumerically("~", 4800*time.Millisecond, time.Millisecond))
		for _, r := range []float64{0, 0.1, 0.5, 0.9, 0.99} {
			Expect(backoffDelay(2*time.Second, backoff, spread, 1, fixedRand(r))).To(
				And(BeNumerically(">=", 3200*time.Millisecond), BeNumerically("<=", 4800*time.Millisecond)))
		}
	})

	It("applies an absolute jitter and never goes negative", func() {
		spread, err := parseJitter("500ms")
		Expect(err).ToNot(HaveOccurred())
		Expect(backoffDelay(300*time.Millisecond, nil, spread, 0, fixedRand(0))).To(Equal(time.Duration(0)))
		Expect(backoffDelay(time.Second, nil, spread, 3, fixedRand(0))).To(Equal(500 * time.Millisecond))
	})

	It("rejects a jitter that is neither a percentage nor a duration", func() {
		_, err := parseJitter("lots")
		Expect(err).To(MatchError(ContainSubstring(`jitter "lots"`)))
	})
})

var _ = Describe("waiter", func() {
	var clock *fakeClock
	var statuses []string
	var observed int

	observe := func(context.Context) (map[string]any, error) {
		status := statuses[min(observed, len(statuses)-1)]
		observed++
		return map[string]any{"status": status}, nil
	}
	newWaiter := func(policy WaitPolicy) waiter {
		return waiter{policy: policy, observe: observe, vars: func(o map[string]any) map[string]any { return o },
			check: evalStepCEL, clock: clock, rand: fixedRand(0.5)}
	}

	BeforeEach(func() {
		clock, observed = newFakeClock(), 0
	})

	It("persists only terminal outputs and elapsed time including probe execution", func() {
		w := newWaiter(WaitPolicy{Until: "status == 'done'", Interval: dur(time.Second), Timeout: dur(time.Minute)})
		w.observe = func(context.Context) (map[string]any, error) {
			observed++
			clock.now = clock.now.Add(250 * time.Millisecond)
			if observed == 1 {
				return map[string]any{"status": "intermediate-secret"}, nil
			}
			return map[string]any{"status": "done"}, nil
		}
		report, _ := w.run(context.Background())
		raw, err := json.Marshal(report)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(raw)).To(And(ContainSubstring(`"durationMs":1500`), ContainSubstring(`"outputs":{"status":"done"}`), Not(ContainSubstring("observations")), Not(ContainSubstring("intermediate-secret"))))
	})

	It("succeeds on the first observation the until expression holds", func() {
		statuses = []string{"pending", "running", "done"}
		report, outputs := newWaiter(WaitPolicy{Until: "status == 'done'", Interval: dur(time.Second),
			Backoff: &Backoff{Factor: 2}, Timeout: dur(time.Minute)}).run(context.Background())

		Expect(report.Outcome).To(Equal(WaitSatisfied))
		Expect(report.DurationMs).To(Equal(int64(3000)))
		Expect(report.Outputs).To(Equal(map[string]any{"status": "done"}))
		Expect(clock.slept).To(Equal([]time.Duration{time.Second, 2 * time.Second}))
		Expect(outputs).To(HaveKeyWithValue("status", "done"))
	})

	It("stops early when failWhen holds, with the CEL trace of the verdict", func() {
		statuses = []string{"running", "failed", "done"}
		report, _ := newWaiter(WaitPolicy{Until: "status == 'done'", FailWhen: "status == 'failed'",
			Interval: dur(time.Second), Timeout: dur(time.Minute)}).run(context.Background())

		Expect(report.Outcome).To(Equal(WaitFailed))
		Expect(report.DurationMs).To(Equal(int64(1000)))
		Expect(report.Outputs).To(Equal(map[string]any{"status": "failed"}))
		Expect(report.Error).To(ContainSubstring("status == 'failed'"))
		Expect(report.Trace).To(ContainSubstring("failed"))
	})

	It("times out at the deadline carrying the last observed outputs", func() {
		statuses = []string{"queued", "running"}
		report, outputs := newWaiter(WaitPolicy{Until: "status == 'done'", Interval: dur(time.Second),
			Timeout: dur(3 * time.Second)}).run(context.Background())

		Expect(report.Outcome).To(Equal(WaitTimeout))
		Expect(observed).To(Equal(4))
		Expect(report.DurationMs).To(Equal(int64(3000)))
		Expect(report.Outputs).To(HaveKeyWithValue("status", "running"))
		Expect(outputs).To(HaveKeyWithValue("status", "running"))
		Expect(report.Error).To(ContainSubstring("status == 'done'"))
		Expect(report.Trace).ToNot(BeEmpty())
	})

	It("clamps the last delay to the deadline instead of overshooting it", func() {
		statuses = []string{"running"}
		report, _ := newWaiter(WaitPolicy{Until: "status == 'done'", Interval: dur(2 * time.Second),
			Timeout: dur(3 * time.Second)}).run(context.Background())

		Expect(report.Outcome).To(Equal(WaitTimeout))
		Expect(clock.slept).To(Equal([]time.Duration{2 * time.Second, time.Second}))
	})

	It("records an observation error and keeps polling", func() {
		calls := 0
		w := newWaiter(WaitPolicy{Until: "status == 'up'", Interval: dur(time.Second), Timeout: dur(time.Minute)})
		w.observe = func(context.Context) (map[string]any, error) {
			calls++
			if calls == 1 {
				return nil, errors.New("connection refused")
			}
			return map[string]any{"status": "up"}, nil
		}
		report, _ := w.run(context.Background())

		Expect(report.Outcome).To(Equal(WaitSatisfied))
		Expect(calls).To(Equal(2))
		Expect(report.Error).To(BeEmpty())
		Expect(report.Outputs).To(Equal(map[string]any{"status": "up"}))
	})

	It("fails loudly when the until expression cannot be evaluated", func() {
		statuses = []string{"running"}
		report, _ := newWaiter(WaitPolicy{Until: "missing.field == 'x'", Interval: dur(time.Second),
			Timeout: dur(time.Minute)}).run(context.Background())

		Expect(report.Outcome).To(Equal(WaitFailed))
		Expect(report.Error).To(ContainSubstring("missing.field"))
	})

	It("retains the terminal trace when the CEL evaluator supplies one with an error", func() {
		statuses = []string{"running"}
		w := newWaiter(WaitPolicy{Until: "status == 'done'", Interval: dur(time.Second), Timeout: dur(time.Minute)})
		w.check = func(string, map[string]any) (bool, string, error) {
			return false, "status == 'done' -> evaluation failed", errors.New("evaluation failed")
		}
		report, _ := w.run(context.Background())
		Expect(report.Trace).To(Equal("status == 'done' -> evaluation failed"))
		Expect(report.Error).To(ContainSubstring("evaluation failed"))
		Expect(report.Outcome).To(Equal(WaitFailed))
	})

	It("keeps an empty terminal observation rather than earlier outputs after a probe error", func() {
		w := newWaiter(WaitPolicy{Until: "status == 'done'", Interval: dur(time.Second), Timeout: dur(time.Second)})
		w.observe = func(context.Context) (map[string]any, error) {
			observed++
			if observed == 1 {
				return map[string]any{"status": "earlier-value"}, nil
			}
			return nil, errors.New("connection lost")
		}
		report, outputs := w.run(context.Background())
		Expect(outputs).To(BeEmpty())
		Expect(report.Outputs).To(BeEmpty())
		Expect(report.Error).To(ContainSubstring("connection lost"))
		Expect(report.DurationMs).To(Equal(int64(1000)))
	})

	It("records elapsed time and final output when cancelled during a delay", func() {
		ctx, cancel := context.WithCancel(context.Background())
		w := newWaiter(WaitPolicy{Until: "status == 'done'", Interval: dur(time.Second), Timeout: dur(time.Minute)})
		w.observe = func(context.Context) (map[string]any, error) {
			clock.now = clock.now.Add(250 * time.Millisecond)
			cancel()
			return map[string]any{"status": "running"}, nil
		}
		report, _ := w.run(ctx)
		Expect(report.Outcome).To(Equal(WaitTimeout))
		Expect(report.Error).To(ContainSubstring("context canceled"))
		Expect(report.DurationMs).To(Equal(int64(1250)))
		Expect(report.Outputs).To(Equal(map[string]any{"status": "running"}))
	})
})

var _ = Describe("waiter with a Go probe", func() {
	var clock *fakeClock
	policy := WaitPolicy{Interval: dur(time.Second), Backoff: &Backoff{Factor: 2, Max: dur(3 * time.Second)}, Timeout: dur(10 * time.Second)}

	BeforeEach(func() { clock = newFakeClock() })

	probeWaiter := func(probes ...WaitProbe) (waiter, *int) {
		calls := 0
		return waiter{policy: policy, clock: clock, rand: fixedRand(0.5), probe: func(context.Context) WaitProbe {
			calls++
			return probes[min(calls, len(probes))-1]
		}}, &calls
	}
	refused := WaitProbe{Outputs: map[string]any{"ready": false}, Error: errors.New("connection refused")}

	It("records only the final probe and elapsed backoff time", func() {
		w, _ := probeWaiter(refused, WaitProbe{Outputs: map[string]any{"ready": false}, Error: errors.New("http status 503")},
			WaitProbe{Outputs: map[string]any{"ready": true}, Until: true})
		report, outputs := w.run(context.Background())

		Expect(report).To(Equal(WaitReport{Policy: policy, Outcome: WaitSatisfied, DurationMs: 3000, Outputs: map[string]any{"ready": true}}))
		Expect(outputs).To(Equal(map[string]any{"ready": true}))
	})

	It("fails fast on a failWhen probe, naming its error", func() {
		w, calls := probeWaiter(refused, WaitProbe{Outputs: map[string]any{"ready": false}, FailWhen: true, Error: errors.New("exited before it was ready")})
		report, _ := w.run(context.Background())

		Expect(*calls).To(Equal(2))
		Expect(report.Outcome).To(Equal(WaitFailed))
		Expect(report.Error).To(Equal("exited before it was ready"))
		Expect(report.Outputs).To(Equal(map[string]any{"ready": false}))
	})

	It("names the broken probe contract when a failWhen probe gives no error", func() {
		w, _ := probeWaiter(WaitProbe{Outputs: map[string]any{"ready": false}, FailWhen: true})
		report, _ := w.run(context.Background())

		Expect([]string{report.Outcome, report.Error}).To(Equal([]string{WaitFailed, "wait probe set FailWhen without an Error; a failing probe must say why"}))
	})

	It("times out at the deadline naming the last probe", func() {
		w, _ := probeWaiter(refused)
		report, _ := w.run(context.Background())

		Expect(report.Outcome).To(Equal(WaitTimeout))
		Expect(report.Error).To(Equal("not ready after 10s: last probe: connection refused"))
		Expect(clock.slept).To(Equal([]time.Duration{time.Second, 2 * time.Second, 3 * time.Second, 3 * time.Second, time.Second}))
	})

	It("cuts the current delay short when wake closes, so the next probe runs at once", func() {
		wake := make(chan struct{})
		close(wake)
		w, calls := probeWaiter(refused, WaitProbe{Outputs: map[string]any{"ready": false}, FailWhen: true, Error: errors.New("exited")})
		w.wake, w.clock = wake, sleepUntilCancelled{clock}
		report, _ := w.run(context.Background())

		Expect(*calls).To(Equal(2))
		Expect(report.Outcome).To(Equal(WaitFailed))
	})
})

var _ = Describe("Wait", func() {
	probe := func(context.Context) WaitProbe { return WaitProbe{Until: true} }
	timing := WaitPolicy{Interval: dur(time.Second), Timeout: dur(time.Second)}

	It("rejects a CEL until or failWhen, which a Go probe cannot honour", func() {
		withUntil := timing
		withUntil.Until = "ready"
		_, err := Wait(context.Background(), WaitOptions{Policy: withUntil, Probe: probe})
		Expect(err).To(MatchError(ContainSubstring("until/failWhen")))
	})

	It("rejects an invalid timing policy before probing", func() {
		invalid := timing
		invalid.Jitter = "150%"
		_, err := Wait(context.Background(), WaitOptions{Policy: invalid, Probe: probe})
		Expect(err).To(MatchError(ContainSubstring(`jitter "150%"`)))
	})
})

var _ = Describe("validateWait", func() {
	It("requires until, a positive interval and a timeout", func() {
		Expect(validateWait(WaitPolicy{Interval: dur(time.Second), Timeout: dur(time.Second)})).To(MatchError(ContainSubstring("until")))
		Expect(validateWait(WaitPolicy{Until: "true", Timeout: dur(time.Second)})).To(MatchError(ContainSubstring("interval")))
		Expect(validateWait(WaitPolicy{Until: "true", Interval: dur(time.Second)})).To(MatchError(ContainSubstring("timeout")))
		Expect(validateWait(WaitPolicy{Until: "true", Interval: dur(time.Second), Timeout: dur(time.Second), Backoff: &Backoff{Factor: 0.5}})).To(MatchError(ContainSubstring("factor")))
		Expect(validateWait(WaitPolicy{Until: "true", Interval: dur(time.Second), Timeout: dur(time.Second), Jitter: "10%"})).To(Succeed())
	})
})

var _ = Describe("mergeWait", func() {
	It("layers step over file defaults over the kind default, field by field", func() {
		kind := &WaitPolicy{Until: "ready", Interval: dur(time.Second), Timeout: dur(time.Minute)}
		file := &WaitPolicy{Interval: dur(2 * time.Second), Jitter: "20%", Timeout: dur(2 * time.Minute)}
		step := &WaitPolicy{Timeout: dur(30 * time.Second)}

		merged := mergeWait(step, file, kind)

		Expect(merged).To(Equal(&WaitPolicy{Until: "ready", Interval: dur(2 * time.Second), Jitter: "20%", Timeout: dur(30 * time.Second)}))
	})

	It("does not activate a wait from file defaults alone", func() {
		Expect(mergeWait(nil, &WaitPolicy{Interval: dur(time.Second)}, nil)).To(BeNil())
	})
})

var _ = Describe("retrier", func() {
	var clock *fakeClock

	BeforeEach(func() { clock = newFakeClock() })

	It("keeps independent attempt evidence when the kind reuses its output map", func() {
		calls := 0
		shared := map[string]any{"phase": "failed-launch"}
		r := retrier{policy: RetryPolicy{Attempts: 2, When: "true"}, check: evalStepCEL, clock: clock, rand: fixedRand(0.5), vars: retryVarsFor}
		_, attempts := r.run(context.Background(), func(context.Context) (StepOutcome, error) {
			calls++
			if calls == 2 {
				shared["phase"] = "launched"
				return StepOutcome{Result: FixtureResult{Status: task.StatusPASS, Stdout: "second"}, Outputs: shared}, nil
			}
			return StepOutcome{Result: FixtureResult{Status: task.StatusFAIL, Stdout: "first"}, Outputs: shared}, nil
		})
		raw, err := json.Marshal(attempts)
		Expect(err).NotTo(HaveOccurred())
		var stored []map[string]any
		Expect(json.Unmarshal(raw, &stored)).To(Succeed())
		Expect(stored[0]).To(HaveKeyWithValue("result", HaveKeyWithValue("stdout", "first")))
		Expect(string(raw)).To(ContainSubstring("failed-launch"))
		Expect(stored[1]).To(HaveKeyWithValue("result", HaveKeyWithValue("stdout", "second")))
	})

	failing := func(calls *int) func(context.Context) (StepOutcome, error) {
		return func(context.Context) (StepOutcome, error) {
			*calls++
			return StepOutcome{Result: FixtureResult{Status: task.StatusFAIL, ExitCode: 75, Error: "connection refused"}}, nil
		}
	}

	It("does not retry when the when expression is false", func() {
		calls := 0
		r := retrier{policy: RetryPolicy{Attempts: 3, When: "exitCode == 1", Delay: dur(time.Second)},
			check: evalStepCEL, clock: clock, rand: fixedRand(0.5), vars: retryVarsFor}
		_, attempts := r.run(context.Background(), failing(&calls))

		Expect(calls).To(Equal(1))
		Expect(attempts).To(HaveLen(1))
		Expect(attempts[0].Status).To(Equal(AttemptFailed))
		Expect(clock.slept).To(BeEmpty())
	})

	It("stops at the attempts cap with delays growing by the backoff", func() {
		calls := 0
		r := retrier{policy: RetryPolicy{Attempts: 3, When: "exitCode == 75 && error.contains('refused')",
			Delay: dur(time.Second), Backoff: &Backoff{Factor: 2}}, check: evalStepCEL, clock: clock,
			rand: fixedRand(0.5), vars: retryVarsFor}
		outcome, attempts := r.run(context.Background(), failing(&calls))

		Expect(calls).To(Equal(3))
		Expect(attempts).To(HaveLen(3))
		Expect(attempts[2].Attempt).To(Equal(3))
		Expect(clock.slept).To(Equal([]time.Duration{time.Second, 2 * time.Second}))
		Expect(outcome.Result.Status).To(Equal(task.StatusFAIL))
	})

	It("never retries without a when expression", func() {
		calls := 0
		r := retrier{policy: RetryPolicy{Attempts: 5}, check: evalStepCEL, clock: clock, rand: fixedRand(0.5), vars: retryVarsFor}
		_, attempts := r.run(context.Background(), failing(&calls))

		Expect(calls).To(Equal(1))
		Expect(attempts).To(HaveLen(1))
	})

	It("exposes outputs returned alongside an Execute error to when, and keeps them on the outcome", func() {
		calls := 0
		r := retrier{policy: RetryPolicy{Attempts: 2, When: "launch.failed && error.contains('admission')"},
			check: evalStepCEL, clock: clock, rand: fixedRand(0.5), vars: retryVarsFor}
		outcome, attempts := r.run(context.Background(), func(context.Context) (StepOutcome, error) {
			calls++
			return StepOutcome{Outputs: map[string]any{"launch": map[string]any{"failed": true}}}, errors.New("admission rejected")
		})

		Expect(calls).To(Equal(2))
		Expect(attempts).To(HaveLen(2))
		Expect(outcome.Result.Status).To(Equal(task.StatusERR))
		Expect(outcome.Result.Error).To(Equal("admission rejected"))
		Expect(outcome.Outputs).To(HaveKeyWithValue("launch", map[string]any{"failed": true}))
	})

	It("returns after the first passing attempt", func() {
		calls := 0
		r := retrier{policy: RetryPolicy{Attempts: 3, When: "true"}, check: evalStepCEL, clock: clock, rand: fixedRand(0.5), vars: retryVarsFor}
		_, attempts := r.run(context.Background(), func(context.Context) (StepOutcome, error) {
			calls++
			status := task.StatusFAIL
			if calls == 2 {
				status = task.StatusPASS
			}
			return StepOutcome{Result: FixtureResult{Status: status}}, nil
		})

		Expect(calls).To(Equal(2))
		Expect(attempts[1].Status).To(Equal(AttemptPassed))
	})

	It("cleans a failed attempt before starting its retry", func() {
		calls, cleaned := 0, false
		r := retrier{policy: RetryPolicy{Attempts: 2, When: "true"}, check: evalStepCEL, clock: clock, rand: fixedRand(0.5), vars: retryVarsFor}
		_, attempts := r.run(context.Background(), func(context.Context) (StepOutcome, error) {
			calls++
			if calls == 2 {
				Expect(cleaned).To(BeTrue())
				return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}}, nil
			}
			return StepOutcome{Result: FixtureResult{Status: task.StatusFAIL}, Cleanup: func(context.Context) error { cleaned = true; return nil }}, nil
		})
		Expect(attempts).To(HaveLen(2))
	})

	It("stops retrying when the failed attempt cannot be cleaned", func() {
		calls := 0
		r := retrier{policy: RetryPolicy{Attempts: 2, When: "true"}, check: evalStepCEL, clock: clock, rand: fixedRand(0.5), vars: retryVarsFor}
		outcome, _ := r.run(context.Background(), func(context.Context) (StepOutcome, error) {
			calls++
			return StepOutcome{Result: FixtureResult{Status: task.StatusFAIL}, Cleanup: func(context.Context) error { return errors.New("cleanup refused") }}, nil
		})
		Expect(calls).To(Equal(1))
		Expect(outcome.Result.Error).To(ContainSubstring("cleanup refused"))
	})
})
