package parsers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test aborted accounting", func() {
	It("counts an aborted child once as aborted even when Skipped is also set", func() {
		s := Test{
			Name: "plan",
			Children: Tests{
				{Name: "setup", Passed: true},
				{Name: "after-failure", Aborted: true, Skipped: true},
				{Name: "gated", Skipped: true},
			},
		}.Sum()
		Expect(s).To(Equal(TestSummary{Total: 3, Passed: 1, Aborted: 1, Skipped: 1}))
	})

	It("sums Aborted across two summaries via Add", func() {
		a := TestSummary{Aborted: 2, Total: 2}
		b := TestSummary{Aborted: 1, Total: 1}
		Expect(a.Add(b)).To(Equal(TestSummary{Aborted: 3, Total: 3}))
	})

	It("is not a folder when only Aborted is set", func() {
		Expect(Test{Name: "step", Aborted: true}.IsFolder()).To(BeFalse())
	})

	It("renders the aborted count in the summary line", func() {
		Expect(TestSummary{Aborted: 2, Total: 2}.Pretty().String()).To(ContainSubstring("aborted"))
	})
})
