package taskhistory_test

import (
	"path/filepath"
	"time"

	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/internal/taskhistory"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task history spool", func() {
	It("retains terminal snapshots for thirty days and excludes expired runs", func() {
		root := GinkgoT().TempDir()
		// Write prunes with the wall clock, so fixtures must be relative to it.
		now := time.Now().UTC().Truncate(time.Second)
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

		records, err := taskhistory.LoadRoots([]string{root}, taskhistory.LoadOptions{Now: now})

		Expect(err).NotTo(HaveOccurred())
		Expect(records).To(ConsistOf(fresh))
	})

	It("lets a sweep skip spool files it has already imported", func() {
		root := GinkgoT().TempDir()
		// Write prunes with the wall clock, so fixtures must be relative to it.
		now := time.Now().UTC().Truncate(time.Second)
		record := func(id string) taskhistory.Record {
			return taskhistory.Record{
				Run:        clickytask.RunMeta{ID: id, Name: id, StartedAt: now.Add(-time.Hour).Format(time.RFC3339Nano)},
				Snapshots:  []clickytask.TaskSnapshot{{ID: id, GroupID: id, Type: "group", Status: string(clickytask.StatusSuccess)}},
				ArchivedAt: now.Add(-time.Hour),
			}
		}
		Expect(taskhistory.Write(root, record("imported-run"))).To(Succeed())
		Expect(taskhistory.Write(root, record("new-run"))).To(Succeed())

		var offered []string
		records, err := taskhistory.LoadRoots([]string{root}, taskhistory.LoadOptions{
			Now: now,
			Skip: func(path string, modTime time.Time) bool {
				offered = append(offered, filepath.Base(path))
				Expect(modTime).NotTo(BeZero())
				return filepath.Base(path) == "imported-run.json"
			},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(records).To(ConsistOf(record("new-run")))
		Expect(offered).To(ConsistOf("imported-run.json", "new-run.json"), "every spool file is offered to Skip, even the ones it declines")
	})

	It("refuses to load without a reference time", func() {
		_, err := taskhistory.LoadRoots([]string{GinkgoT().TempDir()}, taskhistory.LoadOptions{})
		Expect(err).To(MatchError(ContainSubstring("reference time")))
	})

	It("rejects a run id that could escape the spool directory", func() {
		root := GinkgoT().TempDir()
		err := taskhistory.Write(root, taskhistory.Record{Run: clickytask.RunMeta{ID: "../outside"}, ArchivedAt: time.Now()})
		Expect(err).To(MatchError(ContainSubstring("invalid task run id")))
	})
})
