package gitstate_test

import (
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/gavel/git/gitstate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tracker.Untrack", func() {
	// Far enough ahead that every tracked key is due.
	future := time.Now().Add(24 * time.Hour)

	It("stops tracking and scheduling a repository no other directory names", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		Expect(r.tracker.Start()).To(Succeed())
		_, err := r.tracker.Track(ctx, r.root)
		Expect(err).NotTo(HaveOccurred())
		Expect(r.tracker.DueRoots(future)).To(ContainElement(r.root))

		r.tracker.Untrack(r.root)

		_, tracked := r.tracker.Tracked(r.root)
		Expect(tracked).To(BeFalse())
		Expect(r.tracker.DueRoots(future)).To(BeEmpty())
	})

	It("keeps scheduling a repository another tracked directory still names", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		sub := filepath.Join(r.root, "sub")
		Expect(os.MkdirAll(sub, 0o755)).To(Succeed())
		_, err := r.tracker.Track(ctx, r.root)
		Expect(err).NotTo(HaveOccurred())
		_, err = r.tracker.Track(ctx, sub)
		Expect(err).NotTo(HaveOccurred())

		r.tracker.Untrack(r.root)

		_, rootTracked := r.tracker.Tracked(r.root)
		_, subTracked := r.tracker.Tracked(sub)
		Expect([]bool{rootTracked, subTracked}).To(Equal([]bool{false, true}))
		Expect(r.tracker.DueRoots(future)).To(ContainElement(r.root))
	})
})
