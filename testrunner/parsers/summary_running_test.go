package parsers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test.Sum running accounting", func() {
	It("counts a running leaf in Total and Running, not Pending", func() {
		t := Test{Name: "step", Running: true}
		s := t.Sum()
		Expect(s.Total).To(Equal(1))
		Expect(s.Running).To(Equal(1))
		Expect(s.Pending).To(Equal(0))
	})

	It("counts a pending (queued) leaf in Total and Pending, not Running", func() {
		t := Test{Name: "step", Pending: true}
		s := t.Sum()
		Expect(s.Total).To(Equal(1))
		Expect(s.Pending).To(Equal(1))
		Expect(s.Running).To(Equal(0))
	})

	It("rolls up a mix of running, queued and passed children", func() {
		parent := Test{
			Name: "plan",
			Children: Tests{
				{Name: "setup", Passed: true},
				{Name: "step-1", Running: true},
				{Name: "step-2", Pending: true},
				{Name: "step-3", Pending: true},
			},
		}
		s := parent.Sum()
		Expect(s.Total).To(Equal(4))
		Expect(s.Passed).To(Equal(1))
		Expect(s.Running).To(Equal(1))
		Expect(s.Pending).To(Equal(2))
	})
})
