package slots

import (
	"context"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Multi-slot leases", func() {
	var root, cache string
	BeforeEach(func() { root = GinkgoT().TempDir(); cache = filepath.Join(root, "cache") })

	request := func(limits Limits, count int) Request {
		return Request{Limits: limits, WorkDir: root, CacheDir: cache, Command: "multi", Count: count}
	}
	hold := func(limits Limits, count int) *Lease {
		lease, err := Acquire(context.Background(), request(limits, count))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(lease.Close)
		return lease
	}
	free := func(limits Limits) int {
		granted := 0
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			lease, err := Acquire(ctx, request(limits, 1))
			cancel()
			if err != nil {
				Expect(err).To(MatchError(context.DeadlineExceeded))
				return granted
			}
			DeferCleanup(lease.Close)
			granted++
		}
	}

	DescribeTable("grants what is free, capped by the limit",
		func(limit, heldByOthers, requested, granted int) {
			limits := Limits{Global: limit}
			if heldByOthers > 0 {
				hold(limits, heldByOthers)
			}
			Expect(hold(limits, requested).Count()).To(Equal(granted))
			Expect(free(limits)).To(Equal(limit - heldByOthers - granted))
		},
		Entry("within the limit", 3, 0, 2, 2),
		Entry("above the limit", 3, 0, 5, 3),
		Entry("downgraded to the free slots", 3, 2, 3, 1),
		Entry("zero means one", 3, 0, 0, 1),
	)

	It("waits while no slot is free and holds nothing meanwhile", func() {
		limits := Limits{Global: 2}
		blocker, err := Acquire(context.Background(), request(limits, 2))
		Expect(err).NotTo(HaveOccurred())
		time.AfterFunc(300*time.Millisecond, func() { _ = blocker.Close() })
		started := time.Now()
		lease := hold(limits, 2)
		Expect(time.Since(started)).To(BeNumerically(">=", 300*time.Millisecond))
		Expect(lease.Count()).To(Equal(2))
	})

	It("gives back project slots the global pool cannot match", func() {
		repo := slotRepo(root, "repo")
		limits := Limits{Global: 2, Project: 3}
		other, err := Acquire(context.Background(), Request{Limits: Limits{Global: 2}, WorkDir: repo, CacheDir: cache, Command: "other"})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(other.Close)
		lease, err := Acquire(context.Background(), Request{Limits: limits, WorkDir: repo, CacheDir: cache, Command: "batch", Count: 3})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(lease.Close)
		Expect(lease.Count()).To(Equal(1))
		project, err := Acquire(context.Background(), Request{Limits: Limits{Project: 3}, WorkDir: repo, CacheDir: cache, Command: "project", Count: 3})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(project.Close)
		Expect(project.Count()).To(Equal(2))
	})

	It("rejects a negative count", func() {
		_, err := Acquire(context.Background(), request(Limits{Global: 1}, -1))
		Expect(err).To(MatchError(ContainSubstring("negative count")))
	})
})
