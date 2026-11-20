// Tests the registry's view of the runs this process owns: it is idle once every run it
// registered has released its claim.
package run

import (
	"github.com/google/uuid"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("the run registry", func() {
	ginkgo.It("is idle only while no run holds a claim", func() {
		registry := NewRegistry()
		Expect(registry.Idle()).To(BeTrue(), "a registry no run has claimed")

		first, err := registry.Register(RegisterOptions{IssueIDs: []uuid.UUID{uuid.New()}, Cancel: func(error) {}})
		Expect(err).NotTo(HaveOccurred())
		second, err := registry.Register(RegisterOptions{IssueIDs: []uuid.UUID{uuid.New()}, Cancel: func(error) {}})
		Expect(err).NotTo(HaveOccurred())
		Expect(registry.Idle()).To(BeFalse())

		first.Release()
		Expect(registry.Idle()).To(BeFalse(), "a run still holds its claim")
		second.Release()
		Expect(registry.Idle()).To(BeTrue())
	})
})
