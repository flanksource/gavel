package gitstate

import (
	"context"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service load", func() {
	It("reuses a state stored by an earlier flight instead of computing again", func() {
		var calls atomic.Int32
		now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
		service := New(Options{
			Compute: func(context.Context, string) (State, error) {
				calls.Add(1)
				return State{CurrentBranch: "main"}, nil
			},
			Now: func() time.Time { return now },
		})
		const dir = "/repo/acme"

		_, err := service.Get(context.Background(), dir)
		Expect(err).NotTo(HaveOccurred())

		// A caller that saw the miss but only reaches the flight after the first
		// one stored its result.
		state, err := service.load(context.Background(), dir, 0)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.CurrentBranch).To(Equal("main"))
		Expect(calls.Load()).To(BeEquivalentTo(1))
	})
})
