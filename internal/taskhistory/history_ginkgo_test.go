package taskhistory_test

import (
	"time"

	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/internal/taskhistory"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task history spool", func() {
	It("retains terminal snapshots for thirty days and excludes expired runs", func() {
		root := GinkgoT().TempDir()
		now := time.Date(2026, time.July, 21, 8, 0, 0, 0, time.UTC)
		fresh := taskhistory.Record{
			Run:        clickytask.RunMeta{ID: "fresh-run", Name: "Fresh", StartedAt: now.Add(-time.Hour).Format(time.RFC3339Nano)},
			Snapshots:  []clickytask.TaskSnapshot{{ID: "fresh-run", GroupID: "fresh-run", Type: "group", Status: string(clickytask.StatusSuccess)}},
			ArchivedAt: now.Add(-time.Hour),
		}
		expired := taskhistory.Record{
			Run:        clickytask.RunMeta{ID: "expired-run", Name: "Expired", StartedAt: now.Add(-31 * 24 * time.Hour).Format(time.RFC3339Nano)},
			Snapshots:  []clickytask.TaskSnapshot{{ID: "expired-run", GroupID: "expired-run", Type: "group", Status: string(clickytask.StatusSuccess)}},
			ArchivedAt: now.Add(-31 * 24 * time.Hour),
		}
		Expect(taskhistory.Write(root, fresh)).To(Succeed())
		Expect(taskhistory.Write(root, expired)).To(Succeed())

		records, err := taskhistory.LoadRoots([]string{root}, now)

		Expect(err).NotTo(HaveOccurred())
		Expect(records).To(ConsistOf(fresh))
	})

	It("rejects a run id that could escape the spool directory", func() {
		root := GinkgoT().TempDir()
		err := taskhistory.Write(root, taskhistory.Record{Run: clickytask.RunMeta{ID: "../outside"}, ArchivedAt: time.Now()})
		Expect(err).To(MatchError(ContainSubstring("invalid task run id")))
	})
})
