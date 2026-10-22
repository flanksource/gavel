package parsers

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ginkgo spec states", func() {
	// A state gavel does not know must never read as passed or vanish from the
	// counts: it fails, and the message names the state ahead of any failure.
	It("fails an unknown spec state loudly while leaving known non-failure states alone", func() {
		// SpecState strings as ginkgo v2 types.SpecState marshals them;
		// "INVALID SPEC STATE" is SpecStateInvalid, "flaked" stands in for a
		// state a future ginkgo adds.
		report := `[{"SuitePath":"/repo/pkg","SuiteDescription":"Pkg Suite","SpecReports":[
			{"LeafNodeText":"passes","LeafNodeLocation":{"FileName":"/repo/pkg/a_test.go","LineNumber":10},"State":"passed"},
			{"LeafNodeText":"skips","LeafNodeLocation":{"FileName":"/repo/pkg/a_test.go","LineNumber":20},"State":"skipped"},
			{"LeafNodeText":"pends","LeafNodeLocation":{"FileName":"/repo/pkg/a_test.go","LineNumber":30},"State":"pending"},
			{"LeafNodeText":"invalid","LeafNodeLocation":{"FileName":"/repo/pkg/a_test.go","LineNumber":40},"State":"INVALID SPEC STATE",
			 "Failure":{"Message":"report was truncated","Location":{"FileName":"/repo/pkg/a_test.go","LineNumber":40}}},
			{"LeafNodeText":"future","LeafNodeLocation":{"FileName":"/repo/pkg/a_test.go","LineNumber":50},"State":"flaked"}
		]}]`
		tests, err := NewGinkgoJSON("/repo").Parse(strings.NewReader(report))
		Expect(err).NotTo(HaveOccurred())

		type outcome struct {
			Passed, Failed, Skipped bool
			Message                 string
		}
		got := map[string]outcome{}
		for _, t := range tests {
			got[t.Name] = outcome{Passed: t.Passed, Failed: t.Failed, Skipped: t.Skipped, Message: t.Message}
		}
		Expect(got).To(Equal(map[string]outcome{
			"passes":  {Passed: true},
			"skips":   {Skipped: true},
			"pends":   {Skipped: true},
			"invalid": {Failed: true, Message: `unknown ginkgo spec state "INVALID SPEC STATE": report was truncated`},
			"future":  {Failed: true, Message: `unknown ginkgo spec state "flaked"`},
		}))
	})
})
