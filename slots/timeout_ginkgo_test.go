package slots

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Slot timeout budgets", func() {
	It("propagates deadline errors and causes to child contexts", func() {
		ctx, cancel := Timeout(context.Background(), TimeoutOptions{Limits: Limits{Global: 1}, Duration: 20 * time.Millisecond})
		defer cancel()
		child, stop := context.WithCancel(ctx)
		defer stop()
		Eventually(child.Done()).WithTimeout(time.Second).Should(BeClosed())
		Expect(child.Err()).To(MatchError(context.DeadlineExceeded))
		Expect(context.Cause(child)).To(MatchError(context.DeadlineExceeded))
	})
	It("excludes a slot wait longer than the subprocess budget", func() {
		request := Request{Limits: Limits{Global: 1}, CacheDir: GinkgoT().TempDir(), Command: "budget test"}
		lease, err := Acquire(context.Background(), request)
		Expect(err).NotTo(HaveOccurred())
		defer lease.Close()
		ctx, cancel := Timeout(context.Background(), TimeoutOptions{Limits: request.Limits, Duration: 150 * time.Millisecond})
		defer cancel()
		released := make(chan error, 1)
		time.AfterFunc(350*time.Millisecond, func() { released <- lease.Close() })
		started := time.Now()
		second, err := Acquire(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		defer second.Close()
		Expect(<-released).To(Succeed())
		Expect(time.Since(started)).To(BeNumerically(">=", 350*time.Millisecond))
		Expect(ctx.Err()).NotTo(HaveOccurred())
		Eventually(ctx.Done()).WithTimeout(time.Second).Should(BeClosed())
		Expect(ctx.Err()).To(MatchError(context.DeadlineExceeded))
	})
	It("acquires inside a deferred budget that has not started without arming it", func() {
		request := Request{Limits: Limits{Global: 1}, CacheDir: GinkgoT().TempDir()}
		ctx, cancel, _ := DeferredTimeout(context.Background(), TimeoutOptions{Limits: request.Limits, Duration: 20 * time.Millisecond})
		defer cancel()
		lease, err := Acquire(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(lease.Close()).To(Succeed())
		Consistently(ctx.Done()).WithTimeout(100 * time.Millisecond).ShouldNot(BeClosed())
	})
	It("runs a process nested under a held lease on the parent's slot", func() {
		request := Request{Limits: Limits{Global: 1}, CacheDir: GinkgoT().TempDir(), Command: "parent"}
		parent, err := Acquire(context.Background(), request)
		Expect(err).NotTo(HaveOccurred())
		defer parent.Close()
		Expect(parent.Env()).To(Equal(map[string]string{HeldEnv: "1"}))
		GinkgoT().Setenv(HeldEnv, parent.Env()[HeldEnv])
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		child, err := Acquire(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(child.Env()).To(Equal(map[string]string{HeldEnv: "1"}))
		Expect(child.Close()).To(Succeed())
	})
	It("marks no environment when slots are disabled", func() {
		lease, err := Acquire(context.Background(), Request{})
		Expect(err).NotTo(HaveOccurred())
		Expect(lease.Env()).To(BeNil())
	})
	It("keeps caller cancellation active while waiting and releases partial acquisitions", func() {
		request := Request{Limits: Limits{Global: 1}, CacheDir: GinkgoT().TempDir()}
		lease, err := Acquire(context.Background(), request)
		Expect(err).NotTo(HaveOccurred())
		defer lease.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		_, err = Acquire(ctx, request)
		Expect(err).To(MatchError(context.DeadlineExceeded))
		Expect(lease.Close()).To(Succeed())
		lease, err = Acquire(context.Background(), request)
		Expect(err).NotTo(HaveOccurred())
		Expect(lease.Close()).To(Succeed())
	})
})
