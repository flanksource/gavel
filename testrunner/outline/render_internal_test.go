package outline

import (
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/testrunner/parsers"
)

var _ = Describe("outline rendering", func() {
	It("summarizes each framework and badges fixture kinds", func() {
		report := &Report{Entries: []*Entry{
			{Framework: parsers.GoTest, File: "sample_test.go", Name: "TestSample"},
			{Framework: parsers.Jest, File: "sample.test.ts", Name: "renders"},
			{Framework: parsers.Fixture, File: "smoke.fixture.md", Name: "health", Labels: []string{"exec"}},
		}}

		Expect(report.Pretty().String()).To(ContainSubstring("(go test 1, jest 1, fixture 1)"))
		Expect((&entryNode{entry: report.Entries[2]}).Pretty().String()).To(And(
			ContainSubstring("[fixture]"),
			ContainSubstring("exec"),
		))
	})
})
