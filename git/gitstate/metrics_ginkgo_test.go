package gitstate_test

import (
	"github.com/flanksource/gavel/git/gitstate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// observations is how many times the registered metric name recorded a value
// under exactly labels: a histogram's sample count, or a counter's value.
func observations(name string, labels map[string]string) uint64 {
	GinkgoHelper()
	families, err := prometheus.DefaultGatherer.Gather()
	Expect(err).NotTo(HaveOccurred())
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, metric := range family.GetMetric() {
			if labelsEqual(metric.GetLabel(), labels) {
				if histogram := metric.GetHistogram(); histogram != nil {
					return histogram.GetSampleCount()
				}
				return uint64(metric.GetCounter().GetValue())
			}
		}
	}
	return 0
}

func labelsEqual(pairs []*dto.LabelPair, labels map[string]string) bool {
	if len(pairs) != len(labels) {
		return false
	}
	for _, pair := range pairs {
		if labels[pair.GetName()] != pair.GetValue() {
			return false
		}
	}
	return true
}

var _ = Describe("Tracker metrics", func() {
	It("records each status scan under its result and the git cache properties it ran with", func(ctx SpecContext) {
		setProperty(gitstate.PropertyUntrackedCache, "false")
		scans := func(result string) uint64 {
			return observations("gavel_git_status_scan_duration_seconds",
				map[string]string{"result": result, "fsmonitor": "false", "untracked_cache": "false"})
		}
		changed, unchanged := scans("changed"), scans("unchanged")
		gitStatus := observations("gavel_git_command_duration_seconds", map[string]string{"command": "status", "fsmonitor": "false"})
		ranges := observations("gavel_git_ranges_computed_total", map[string]string{})
		r := newTrackedRepo(ctx, gitstate.Options{})

		// Track scans both worktrees for the first time; the rescan finds them unchanged.
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())

		Expect(scans("changed") - changed).To(Equal(uint64(2)))
		Expect(scans("unchanged") - unchanged).To(Equal(uint64(2)))
		Expect(observations("gavel_git_command_duration_seconds",
			map[string]string{"command": "status", "fsmonitor": "false"}) - gitStatus).To(Equal(uint64(4)))
		Expect(observations("gavel_git_ranges_computed_total", map[string]string{}) - ranges).To(Equal(uint64(1)))
	})

	It("times reads of the stored state", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		reads := observations("gavel_git_read_duration_seconds", map[string]string{"op": "states"})

		_, err := r.tracker.State(ctx, r.root)

		Expect(err).NotTo(HaveOccurred())
		Expect(observations("gavel_git_read_duration_seconds", map[string]string{"op": "states"}) - reads).To(Equal(uint64(1)))
	})
})
