package gitstate_test

import (
	"time"

	"github.com/flanksource/gavel/git/gitstate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("idle poll backoff", func() {
	const (
		idle    = time.Minute
		maxIdle = 30 * time.Minute
	)

	DescribeTable("doubles the idle interval per unchanged scan, up to the cap",
		func(streak int, expected time.Duration) {
			Expect(gitstate.Backoff(idle, maxIdle, streak)).To(Equal(expected))
		},
		Entry("after a change", 0, idle),
		Entry("after one unchanged scan", 1, 2*idle),
		Entry("after four unchanged scans", 4, 16*idle),
		Entry("when doubling would pass the cap", 5, maxIdle),
		Entry("after a streak long enough to overflow a shift", 100, maxIdle),
	)

	It("spreads an interval by at most a tenth either way", func() {
		const samples = 200
		for range samples {
			Expect(gitstate.Jitter(idle)).To(BeNumerically(">=", idle-idle/10))
			Expect(gitstate.Jitter(idle)).To(BeNumerically("<=", idle+idle/10))
		}
	})
})

var _ = Describe("focused repo refs", func() {
	It("rescans the refs of a repository at the hot cadence while one of its worktrees is focused", func(ctx SpecContext) {
		const hot = 200 * time.Millisecond
		r := newTrackedRepo(ctx, gitstate.Options{Idle: time.Hour, MaxIdle: time.Hour, Hot: hot})
		Expect(r.tracker.Start()).To(Succeed())
		_, err := r.tracker.State(ctx, r.root)
		Expect(err).NotTo(HaveOccurred())
		before := enqueued("refs", "cadence")

		Consistently(func() uint64 { return enqueued("refs", "cadence") }).
			WithTimeout(1500*time.Millisecond).Should(Equal(before), "unfocused refs wait out the idle interval")

		r.tracker.Focus(r.feature, time.Minute)

		Eventually(func() uint64 { return enqueued("refs", "cadence") }).
			WithTimeout(5 * time.Second).Should(BeNumerically(">=", before+2))
	})
})

var _ = Describe("poll schedule", func() {
	It("backs off after each unchanged scan and returns to the idle interval after a change", func() {
		const (
			idle    = time.Minute
			maxIdle = 4 * time.Minute
		)
		scans := []bool{false, false, false, false, true, false}
		expected := []time.Duration{2 * idle, 4 * idle, maxIdle, maxIdle, idle, 2 * idle}

		intervals := gitstate.IdleIntervals(scans, idle, maxIdle)

		Expect(intervals).To(HaveLen(len(expected)))
		for i, interval := range intervals {
			Expect(interval).To(BeNumerically("~", expected[i], expected[i]/10), "scan %d (changed=%t)", i, scans[i])
		}
	})
})
