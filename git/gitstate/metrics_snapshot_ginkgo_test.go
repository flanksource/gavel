package gitstate_test

import (
	"github.com/flanksource/gavel/git/gitstate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
)

var _ = Describe("Metrics", func() {
	It("summarises the gavel_git_* series with counts, averages and bucket-interpolated percentiles", func() {
		const (
			fastScanSeconds = 0.005
			slowScanSeconds = 0.05
			repos           = 3
			hot, idle       = 2, 5
			backedOff       = 9
		)
		registry := prometheus.NewRegistry()
		scans := prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "gavel_git_status_scan_duration_seconds",
			Buckets: []float64{0.01, 0.1, 1},
		}, []string{"result"})
		trackedRepos := prometheus.NewGauge(prometheus.GaugeOpts{Name: "gavel_git_tracked_repos"})
		worktrees := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "gavel_git_tracked_worktrees"}, []string{"cadence"})
		registry.MustRegister(scans, trackedRepos, worktrees)
		for _, seconds := range []float64{fastScanSeconds, fastScanSeconds, slowScanSeconds, slowScanSeconds} {
			scans.WithLabelValues("unchanged").Observe(seconds)
		}
		scans.WithLabelValues("changed")
		trackedRepos.Set(repos)
		worktrees.WithLabelValues("hot").Set(hot)
		worktrees.WithLabelValues("idle").Set(idle)
		worktrees.WithLabelValues("backoff").Set(backedOff)

		snapshot, err := gitstate.Metrics(registry)

		Expect(err).NotTo(HaveOccurred())
		Expect(snapshot.TrackedRepos).To(Equal(repos))
		Expect(snapshot.HotWorktrees).To(Equal(hot))
		Expect(snapshot.IdleWorktrees).To(Equal(idle))
		Expect(snapshot.BackoffWorktrees).To(Equal(backedOff))
		// The series that never observed a scan is left out.
		Expect(snapshot.StatusScans).To(HaveLen(1))
		stat := snapshot.StatusScans[0]
		Expect(stat.Labels).To(Equal(map[string]string{"result": "unchanged"}))
		Expect(stat.Count).To(Equal(uint64(4)))
		Expect(stat.TotalMs).To(BeNumerically("~", 110, 1e-9))
		Expect(stat.AvgMs).To(BeNumerically("~", 27.5, 1e-9))
		// Rank 2 of 4 ends the 0–10ms bucket; rank 3.8 is 90% into the 10–100ms one.
		Expect(stat.P50Ms).To(BeNumerically("~", 10, 1e-9))
		Expect(stat.P95Ms).To(BeNumerically("~", 91, 1e-9))
	})

	It("lists what requested scans, most frequent first", func() {
		const watched, cadenced = 7, 3
		registry := prometheus.NewRegistry()
		enqueued := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "gavel_git_scans_enqueued_total"}, []string{"kind", "source"})
		registry.MustRegister(enqueued)
		enqueued.WithLabelValues("status", "cadence").Add(cadenced)
		enqueued.WithLabelValues("status", "watch").Add(watched)

		snapshot, err := gitstate.Metrics(registry)

		Expect(err).NotTo(HaveOccurred())
		Expect(snapshot.ScanTriggers).To(Equal([]gitstate.CountStat{
			{Labels: map[string]string{"kind": "status", "source": "watch"}, Count: watched},
			{Labels: map[string]string{"kind": "status", "source": "cadence"}, Count: cadenced},
		}))
	})
})
