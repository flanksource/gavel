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

var _ = g.Describe("readiness policy", func() {
	g.DescribeTable("layers an entry's wait over the readiness defaults field by field",
		func(wait *fixtures.WaitPolicy, expected fixtures.WaitPolicy) {
			Expect(readinessPolicy(wait)).To(Equal(expected))
		},
		g.Entry("no wait: every default", nil, fixtures.WaitPolicy{
			Interval: duration(defaultInterval), Timeout: duration(defaultTimeout),
			Backoff: &fixtures.Backoff{Factor: defaultFactor, Max: duration(defaultMax)},
		}),
		g.Entry("timeout and backoff.max only keep the default interval and factor",
			&fixtures.WaitPolicy{Timeout: duration(time.Second), Backoff: &fixtures.Backoff{Max: duration(500 * time.Millisecond)}},
			fixtures.WaitPolicy{
				Interval: duration(defaultInterval), Timeout: duration(time.Second),
				Backoff: &fixtures.Backoff{Factor: defaultFactor, Max: duration(500 * time.Millisecond)},
			}),
		g.Entry("interval, factor and jitter override",
			&fixtures.WaitPolicy{Interval: duration(50 * time.Millisecond), Jitter: "20%", Backoff: &fixtures.Backoff{Factor: 2}},
			fixtures.WaitPolicy{
				Interval: duration(50 * time.Millisecond), Timeout: duration(defaultTimeout), Jitter: "20%",
				Backoff: &fixtures.Backoff{Factor: 2, Max: duration(defaultMax)},
			}),
	)

	g.DescribeTable("rejects an exec entry whose wait cannot run, at config time",
		func(wait map[string]any, message string) {
			_, err := decodeEntry("api", map[string]any{"exec": "true", "ready": map[string]any{"exec": "true"}, "wait": wait})
			Expect(err).To(MatchError(And(ContainSubstring(`entry "api"`), ContainSubstring(message))))
		},
		g.Entry("jitter over 100%", map[string]any{"jitter": "150%"}, `jitter "150%"`),
		g.Entry("shrinking backoff", map[string]any{"backoff": map[string]any{"factor": 0.5}}, "backoff factor must be >= 1"),
		g.Entry("zero timeout", map[string]any{"timeout": "0s"}, "timeout must be a positive duration"),
	)
})
