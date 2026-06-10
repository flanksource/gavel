package outline

import (
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/linters/jscpd"
	"github.com/flanksource/gavel/testrunner/parsers"
)

func jscpdRef(name string, start, end int) jscpd.JscpdFileRef {
	return jscpd.JscpdFileRef{
		Name:     name,
		StartLoc: jscpd.JscpdLoc{Line: start},
		EndLoc:   jscpd.JscpdLoc{Line: end},
	}
}

var _ = Describe("duplication", func() {
	Describe("coveredLines", func() {
		It("merges overlapping intervals and clips to the body span", func() {
			intervals := []lineInterval{{5, 12}, {10, 14}, {30, 40}}
			Expect(coveredLines(intervals, 8, 20)).To(Equal(7)) // lines 8-14
		})

		It("returns zero when no interval touches the span", func() {
			Expect(coveredLines([]lineInterval{{1, 3}}, 10, 20)).To(BeZero())
		})
	})

	Describe("annotateDuplication", func() {
		It("attributes both sides of a clone pair to their tests", func() {
			report := &Report{Entries: []*Entry{
				{Framework: parsers.GoTest, File: "a_test.go", Name: "TestA", Line: 10, EndLine: 19, SizeLines: 10},
				{Framework: parsers.GoTest, File: "b_test.go", Name: "TestB", Line: 1, EndLine: 20, SizeLines: 20},
				{Framework: parsers.GoTest, File: "c_test.go", Name: "TestClean", Line: 1, EndLine: 5, SizeLines: 5},
			}}
			jr := &jscpd.JscpdReport{Duplicates: []jscpd.JscpdDuplicate{
				{FirstFile: jscpdRef("a_test.go", 12, 16), SecondFile: jscpdRef("b_test.go", 3, 7)},
			}}
			annotateDuplication(report, intervalsByFile(jr, "/work"))
			Expect(report.Entries[0].DuplicationPct).To(BeNumerically("==", 50))
			Expect(report.Entries[1].DuplicationPct).To(BeNumerically("==", 25))
			Expect(report.Entries[2].DuplicationPct).To(BeZero())
		})

		It("normalizes absolute report paths against the workdir", func() {
			report := &Report{Entries: []*Entry{
				{Framework: parsers.GoTest, File: "pkg/a_test.go", Name: "TestA", Line: 1, EndLine: 4, SizeLines: 4},
			}}
			jr := &jscpd.JscpdReport{Duplicates: []jscpd.JscpdDuplicate{
				{FirstFile: jscpdRef("/work/pkg/a_test.go", 1, 2), SecondFile: jscpdRef("/elsewhere/b.go", 1, 2)},
			}}
			annotateDuplication(report, intervalsByFile(jr, "/work"))
			Expect(report.Entries[0].DuplicationPct).To(BeNumerically("==", 50))
		})
	})
})
