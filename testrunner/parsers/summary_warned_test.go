package parsers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test warned accounting", func() {
	It("counts a warned leaf in Total and Warned, not Passed or Failed", func() {
		s := Test{Name: "trace", Warned: true}.Sum()
		Expect(s.Total).To(Equal(1))
		Expect(s.Warned).To(Equal(1))
		Expect(s.Passed).To(Equal(0))
		Expect(s.Failed).To(Equal(0))
	})

	It("rolls a warned child up into the parent summary without failing it", func() {
		s := Test{
			Name: "step",
			Children: Tests{
				{Name: "activity", Passed: true},
				{Name: "trace", Warned: true},
			},
		}.Sum()
		Expect(s.Total).To(Equal(2))
		Expect(s.Passed).To(Equal(1))
		Expect(s.Warned).To(Equal(1))
		Expect(s.Failed).To(Equal(0))
	})

	It("sums Warned across two summaries via Add", func() {
		a := TestSummary{Warned: 2, Total: 3}
		b := TestSummary{Warned: 1, Total: 1}
		Expect(a.Add(b).Warned).To(Equal(3))
	})

	It("propagates a warned child to the parent when no child failed", func() {
		parent := Test{
			Name: "step",
			Children: Tests{
				{Name: "activity", Passed: true},
				{Name: "trace", Warned: true},
			},
		}
		propagateFailureStatusRecursive(&parent)
		Expect(parent.Warned).To(BeTrue())
		Expect(parent.Failed).To(BeFalse())
	})

	It("does not downgrade a failed parent: a failed child wins over a warned sibling", func() {
		parent := Test{
			Name: "step",
			Children: Tests{
				{Name: "trace", Warned: true},
				{Name: "assert", Failed: true},
			},
		}
		propagateFailureStatusRecursive(&parent)
		Expect(parent.Failed).To(BeTrue())
		Expect(parent.Warned).To(BeFalse(), "a warning must never mask a failure on the parent")
	})
})
