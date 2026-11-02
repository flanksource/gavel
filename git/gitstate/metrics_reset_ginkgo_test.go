package gitstate

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
)

func recordGitMetrics() {
	refsScanDuration.WithLabelValues(resultChanged).Observe(0.01)
	statusScanDuration.WithLabelValues(resultUnchanged, "false", "true").Observe(0.02)
	gitCommandDuration.WithLabelValues("status", "false").Observe(0.03)
	scanQueueWait.WithLabelValues("status").Observe(0.04)
	readDuration.WithLabelValues("states").Observe(0.05)
	scansEnqueued.WithLabelValues("status", sourceCadence).Inc()
	rangesComputed.WithLabelValues().Inc()
}

var _ = Describe("ResetMetrics", func() {
	It("clears recorded activity while preserving the live tracker gauges and cache settings", func() {
		before, err := Metrics(prometheus.DefaultGatherer)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() {
			trackedRepos.Set(float64(before.TrackedRepos))
			trackedWorktrees.WithLabelValues(cadenceHot).Set(float64(before.HotWorktrees))
			trackedWorktrees.WithLabelValues(cadenceIdle).Set(float64(before.IdleWorktrees))
			trackedWorktrees.WithLabelValues(cadenceBackoff).Set(float64(before.BackoffWorktrees))
			scansInFlight.Set(float64(before.ScansInFlight))
			ResetMetrics()
		})
		const repos, hot, idle, backedOff, running = 3, 2, 5, 9, 1
		trackedRepos.Set(repos)
		trackedWorktrees.WithLabelValues(cadenceHot).Set(hot)
		trackedWorktrees.WithLabelValues(cadenceIdle).Set(idle)
		trackedWorktrees.WithLabelValues(cadenceBackoff).Set(backedOff)
		scansInFlight.Set(running)
		recordGitMetrics()

		ResetMetrics()
		ResetMetrics()

		snapshot, err := Metrics(prometheus.DefaultGatherer)
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshot).To(Equal(MetricsSnapshot{
			FSMonitor: before.FSMonitor, UntrackedCache: before.UntrackedCache,
			TrackedRepos: repos, HotWorktrees: hot, IdleWorktrees: idle,
			BackoffWorktrees: backedOff, ScansInFlight: running,
		}))
	})

	It("records new observations from zero after a reset", func() {
		DeferCleanup(ResetMetrics)
		recordGitMetrics()
		ResetMetrics()

		recordGitMetrics()

		snapshot, err := Metrics(prometheus.DefaultGatherer)
		Expect(err).NotTo(HaveOccurred())
		for _, stats := range [][]LatencyStat{snapshot.RefsScans, snapshot.StatusScans, snapshot.Commands, snapshot.QueueWait, snapshot.Reads} {
			Expect(stats).To(HaveLen(1))
			Expect(stats[0].Count).To(Equal(uint64(1)))
		}
		Expect(snapshot.ScanTriggers).To(Equal([]CountStat{{Labels: map[string]string{"kind": "status", "source": sourceCadence}, Count: 1}}))
		Expect(snapshot.RangesComputed).To(Equal(1))
	})
})
