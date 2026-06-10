package sample

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Calculator", func() {
	Context("when adding", func() {
		It("adds two positive numbers", func() {
			Expect(add(1, 2)).To(Equal(3))
		})

		It("handles negatives", func() {
			result := add(-1, -2)
			if result > 0 {
				Fail("unexpected sign")
			}
			Expect(result).To(Equal(-3))
		})

		PIt("rounds floats someday", func() {})
	})

	It(fmt.Sprintf("dynamic case %d", 1), func() {
		Expect(true).To(BeTrue())
	})

	DescribeTable("doubling",
		func(in, want int) {
			Expect(double(in)).To(Equal(want))
		},
		Entry("zero", 0, 0),
		Entry("two", 2, 4),
	)
})

func add(a, b int) int { return a + b }

func double(n int) int { return n * 2 }
