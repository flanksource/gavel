package gitstate_test

import (
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/gavel/git/gitstate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func enqueued(kind, source string) uint64 {
	GinkgoHelper()
	return observations("gavel_git_scans_enqueued_total", map[string]string{"kind": kind, "source": source})
}

func statusScans() uint64 {
	GinkgoHelper()
	var total uint64
	for _, result := range []string{"changed", "unchanged", "skipped", "error"} {
		total += observations("gavel_git_status_scan_duration_seconds",
			map[string]string{"result": result, "fsmonitor": "false", "untracked_cache": "true"})
	}
	return total
}

var _ = Describe("metadata watcher", func() {
	const debounce = 20 * time.Millisecond

	It("ignores an attribute-only change to an index, as a scan reading it causes, and rescans on a real index write", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{Debounce: debounce, Idle: time.Hour, Hot: time.Hour})
		Expect(r.tracker.Start()).To(Succeed())
		_, err := r.tracker.State(ctx, r.root)
		Expect(err).NotTo(HaveOccurred())
		index := filepath.Join(r.root, ".git", "index")
		info, err := os.Stat(index)
		Expect(err).NotTo(HaveOccurred())
		before := enqueued("status", "watch")

		Expect(os.Chtimes(index, time.Now(), info.ModTime())).To(Succeed())
		Expect(os.Chmod(index, 0o600)).To(Succeed())

		Consistently(func() uint64 { return enqueued("status", "watch") }).
			WithTimeout(20 * debounce).WithPolling(debounce).Should(Equal(before))

		write(r.root, "staged.txt", "s\n")
		git(r.root, "add", "staged.txt")

		Eventually(func() uint64 { return enqueued("status", "watch") }).
			WithTimeout(10 * time.Second).WithPolling(debounce).Should(BeNumerically(">", before))
	})
})

var _ = Describe("scan queue", func() {
	It("runs a scan requested again while it waits for a slot once", func(ctx SpecContext) {
		const requests = 3
		r := newTrackedRepo(ctx, gitstate.Options{Concurrency: 1})
		_, err := r.tracker.State(ctx, r.root)
		Expect(err).NotTo(HaveOccurred())
		before := statusScans()

		release := r.tracker.HoldSlot()
		for range requests {
			r.tracker.Touch(r.feature)
		}
		release()

		Eventually(statusScans).WithTimeout(10 * time.Second).Should(Equal(before + 1))
		Consistently(statusScans).WithTimeout(500 * time.Millisecond).Should(Equal(before + 1))
	})
})
