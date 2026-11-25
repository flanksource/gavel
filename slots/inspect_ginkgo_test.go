package slots

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/sdk/status"
	"github.com/gofrs/flock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Inspecting slots", func() {
	var root, cache, repo, project, projectPool string
	limits := Limits{Global: 2, Project: 1}
	BeforeEach(func() {
		root = GinkgoT().TempDir()
		cache = filepath.Join(root, "cache")
		repo = slotRepo(root, "repo")
		var id string
		var err error
		project, id, err = ProjectDir(context.Background(), repo)
		Expect(err).NotTo(HaveOccurred())
		projectPool = "project/" + id
	})
	acquire := func(ctx context.Context, command string) (*Lease, error) {
		return Acquire(ctx, Request{Limits: limits, WorkDir: repo, CacheDir: cache, Command: command})
	}

	It("shows a held lease in both pools until it closes", func(ctx SpecContext) {
		lease, err := acquire(ctx, "go test ./a")
		Expect(err).NotTo(HaveOccurred())

		Expect(Leases()).To(ConsistOf(
			leaseAt(project, projectPool, 1, 0, status.LeaseHeld),
			leaseAt("", "global", 2, 0, status.LeaseHeld),
		))
		pools, err := Inspect(cache)
		Expect(err).NotTo(HaveOccurred())
		Expect(pools).To(HaveLen(2))
		Expect(pools[0].Name).To(Equal("global"))
		Expect(pools[0].Limit).To(Equal(2))
		Expect(pools[1]).To(SatisfyAll(
			HaveField("Name", projectPool), HaveField("Project", project), HaveField("Limit", 1),
		))
		for _, pool := range pools {
			Expect(pool.Held).To(HaveLen(1))
			Expect(pool.Held[0].Owner).To(SatisfyAll(
				HaveField("PID", os.Getpid()), HaveField("Command", "go test ./a"), HaveField("Started", Not(BeZero())),
			))
		}

		Expect(lease.Close()).To(Succeed())
		Expect(Leases()).To(BeEmpty())
		pools, err = Inspect(cache)
		Expect(err).NotTo(HaveOccurred())
		for _, pool := range pools {
			Expect(pool.Held).To(BeEmpty(), "the stale owner record of %s is not a holder", pool.Name)
		}
	})

	It("tracks a queued Acquire until it gives up", func(ctx SpecContext) {
		held, err := acquire(ctx, "go test ./a")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(held.Close)
		waitCtx, cancel := context.WithCancel(ctx)
		errs := make(chan error, 1)
		go func() {
			_, err := acquire(waitCtx, "go test ./b")
			errs <- err
		}()
		Eventually(Leases).Should(ContainElement(leaseAt(project, projectPool, 1, -1, status.LeaseWaiting)))
		cancel()
		Eventually(errs).Should(Receive(MatchError(context.Canceled)))
		Expect(Leases()).To(HaveLen(2))
		Expect(Leases()).NotTo(ContainElement(HaveField("State", status.LeaseWaiting)))
	})

	It("lists the waiters other processes publish under their pool", func() {
		Expect(os.MkdirAll(status.Dir(cache), 0o700)).To(Succeed())
		waiting := status.Lease{Pool: projectPool, Project: project, Limit: 3, Index: -1, State: status.LeaseWaiting, Command: "go test ./c"}
		data, err := json.Marshal(status.Status{PID: os.Getpid(), Slots: []status.Lease{waiting}})
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(status.Path(cache, os.Getpid()), data, 0o600)).To(Succeed())

		pools, err := Inspect(cache)
		Expect(err).NotTo(HaveOccurred())
		Expect(pools).To(HaveLen(1))
		Expect(pools[0]).To(SatisfyAll(
			HaveField("Name", projectPool), HaveField("Project", project), HaveField("Limit", 3),
			HaveField("Waiting", ConsistOf(Waiter{PID: os.Getpid(), Lease: waiting})),
		))
	})

	It("names a held slot whose owner record is corrupt", func() {
		dir := filepath.Join(cache, "slots", "global")
		Expect(os.MkdirAll(dir, 0o700)).To(Succeed())
		path := filepath.Join(dir, "0.lock")
		lock := flock.New(path)
		Expect(lock.Lock()).To(Succeed())
		DeferCleanup(lock.Close)
		Expect(os.WriteFile(path, []byte("{not json"), 0o600)).To(Succeed())

		_, err := Inspect(cache)
		Expect(err).To(MatchError(ContainSubstring(path)))
	})

	It("is empty before any slot was taken", func() {
		Expect(Inspect(cache)).To(BeEmpty())
	})
})

// leaseAt matches a lease by where it is and what state it is in.
func leaseAt(project, pool string, limit, index int, state string) OmegaMatcher {
	return SatisfyAll(
		HaveField("Project", project), HaveField("Pool", pool), HaveField("Limit", limit),
		HaveField("Index", index), HaveField("State", state),
	)
}
