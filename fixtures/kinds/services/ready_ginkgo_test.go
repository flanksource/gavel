package services

import (
	"time"

	"github.com/flanksource/gavel/fixtures"

	// Not dot-imported: ginkgo's Entry would collide with this package's Entry.
	g "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
)

func duration(d time.Duration) *fixtures.Duration {
	value := fixtures.Duration(d)
	return &value
}

var _ = g.Describe("readiness schedule", func() {
	const interval = 100 * time.Millisecond
	policy := &fixtures.WaitPolicy{
		Interval: duration(interval),
		Backoff:  &fixtures.Backoff{Factor: 2, Max: duration(500 * time.Millisecond)},
		Timeout:  duration(time.Second),
	}

	g.It("grows the delay geometrically up to backoff.max without jitter", func() {
		sched, err := scheduleFor(policy)
		Expect(err).NotTo(HaveOccurred())
		midpoint := func() float64 { return 0.5 }
		delays := []time.Duration{}
		for attempt := range 5 {
			delays = append(delays, sched.delay(attempt, midpoint))
		}
		Expect(delays).To(Equal([]time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond, 500 * time.Millisecond, 500 * time.Millisecond}))
	})

	g.DescribeTable("keeps jittered delays within the declared spread",
		func(jitter string, low, high time.Duration) {
			withJitter := *policy
			withJitter.Jitter = jitter
			sched, err := scheduleFor(&withJitter)
			Expect(err).NotTo(HaveOccurred())
			Expect(sched.delay(0, func() float64 { return 0 })).To(Equal(low))
			Expect(sched.delay(0, func() float64 { return 0.999999999 })).To(BeNumerically("~", high, time.Microsecond))
		},
		g.Entry("percentage", "20%", 80*time.Millisecond, 120*time.Millisecond),
		g.Entry("absolute", "30ms", 70*time.Millisecond, 130*time.Millisecond),
		g.Entry("absolute larger than the delay clamps at zero", "150ms", time.Duration(0), 250*time.Millisecond),
	)

	g.DescribeTable("rejects an invalid policy",
		func(mutate func(*fixtures.WaitPolicy), message string) {
			invalid := *policy
			mutate(&invalid)
			_, err := scheduleFor(&invalid)
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		g.Entry("jitter over 100%", func(p *fixtures.WaitPolicy) { p.Jitter = "150%" }, "between 0% and 100%"),
		g.Entry("jitter that is neither", func(p *fixtures.WaitPolicy) { p.Jitter = "lots" }, `must be a percentage ("20%") or a duration`),
		g.Entry("shrinking backoff", func(p *fixtures.WaitPolicy) { p.Backoff = &fixtures.Backoff{Factor: 0.5} }, "backoff.factor >= 1"),
		g.Entry("zero timeout", func(p *fixtures.WaitPolicy) { p.Timeout = duration(0) }, "timeout must be positive"),
	)
})
