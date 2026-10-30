package snapshots

import (
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Latest runs", func() {
	DescribeTable("selects the newest collision-suffixed run within the same second", func(count int) {
		dir := GinkgoT().TempDir()
		started := time.Date(2026, 10, 28, 8, 0, 0, 0, time.UTC)
		var last string
		for index := range count {
			timestamp := started.Add(time.Duration(index) * time.Millisecond)
			var err error
			last, err = SavePerRun(dir, &testui.Snapshot{
				Metadata: &testui.SnapshotMetadata{Started: timestamp, Ended: timestamp},
				Status:   testui.SnapshotStatus{LintRun: true},
				Tests:    []parsers.Test{{Name: "review", Passed: true}},
			}, timestamp, "")
			Expect(err).NotTo(HaveOccurred())
		}
		paths, err := LatestRuns(dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(Equal([]string{last}))
	},
		Entry("two runs", 2),
		Entry("double digit collision suffix", 10),
	)
})
