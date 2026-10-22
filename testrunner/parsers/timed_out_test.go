package parsers

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type timeoutOutcome struct {
	Passed, Failed, TimedOut bool
}

func outcomesByName(tests []Test) map[string]timeoutOutcome {
	out := map[string]timeoutOutcome{}
	for _, t := range tests {
		out[t.Name] = timeoutOutcome{Passed: t.Passed, Failed: t.Failed, TimedOut: t.TimedOut}
	}
	return out
}

var _ = Describe("Timed-out tests", func() {
	It("marks the tests running when go test's own -timeout panic fired as failed and timed out", func() {
		data, err := os.ReadFile("testdata/gotest-timeout.jsonl")
		Expect(err).NotTo(HaveOccurred())

		tests, err := (&GoTestJSON{}).Parse(bytes.NewReader(data))
		Expect(err).NotTo(HaveOccurred())

		Expect(outcomesByName(tests)).To(Equal(map[string]timeoutOutcome{
			"TestFast":     {Passed: true},
			"TestSlow":     {Failed: true, TimedOut: true},
			"TestSlow/sub": {Failed: true, TimedOut: true},
		}))
		for _, t := range tests {
			if t.TimedOut {
				Expect(t.Message).To(Equal("panic: test timed out after 1s"), t.Name)
			}
		}
	})

	It("parses a Playwright timedOut result as failed and timed out, and interrupted as failed only", func() {
		report := `{"suites":[{"title":"a.spec.ts","specs":[
			{"title":"slow","tests":[{"results":[{"status":"timedOut","duration":30000}]}]},
			{"title":"stopped","tests":[{"results":[{"status":"interrupted","duration":10}]}]}
		]}]}`
		tests, err := NewPlaywrightJSON("").Parse(strings.NewReader(report))
		Expect(err).NotTo(HaveOccurred())

		Expect(outcomesByName(tests)).To(Equal(map[string]timeoutOutcome{
			"slow":    {Failed: true, TimedOut: true},
			"stopped": {Failed: true},
		}))
	})

	It("parses a ginkgo timedout spec as failed and timed out with its failure message", func() {
		const message = "A suite timeout occurred"
		data, err := json.Marshal([]ginkgoSuiteReport{{
			SuitePath: "/repo/pkg",
			SpecReports: []ginkgoSpecReport{{
				LeafNodeText: "waits forever",
				State:        "timedout",
				Failure:      &ginkgoFailure{Message: message},
			}},
		}})
		Expect(err).NotTo(HaveOccurred())

		tests, err := NewGinkgoJSON("").Parse(bytes.NewReader(data))
		Expect(err).NotTo(HaveOccurred())

		Expect(outcomesByName(tests)).To(Equal(map[string]timeoutOutcome{
			"waits forever": {Failed: true, TimedOut: true},
		}))
		Expect(tests[0].Message).To(Equal(message))
	})

	It("parses ginkgo panicked, aborted and interrupted specs as failures with their messages", func() {
		// SpecState strings as ginkgo v2 types.SpecState marshals them.
		report := `[{"SuitePath":"/repo/pkg","SuiteDescription":"Pkg Suite","SpecReports":[
			{"LeafNodeText":"panics","LeafNodeLocation":{"FileName":"/repo/pkg/a_test.go","LineNumber":10},"State":"panicked",
			 "Failure":{"Message":"Test Panicked\nruntime error: index out of range","Location":{"FileName":"/repo/pkg/a.go","LineNumber":42}}},
			{"LeafNodeText":"aborts","LeafNodeLocation":{"FileName":"/repo/pkg/a_test.go","LineNumber":20},"State":"aborted",
			 "Failure":{"Message":"database unreachable","Location":{"FileName":"/repo/pkg/a_test.go","LineNumber":21}}},
			{"LeafNodeText":"interrupted","LeafNodeLocation":{"FileName":"/repo/pkg/a_test.go","LineNumber":30},"State":"interrupted",
			 "Failure":{"Message":"Interrupted by User","Location":{"FileName":"/repo/pkg/a_test.go","LineNumber":30}}}
		]}]`
		tests, err := NewGinkgoJSON("/repo").Parse(strings.NewReader(report))
		Expect(err).NotTo(HaveOccurred())

		type outcome struct {
			Failed, Aborted, TimedOut bool
			Message, FailureLocation  string
		}
		got := map[string]outcome{}
		for _, t := range tests {
			got[t.Name] = outcome{Failed: t.Failed, Aborted: t.Aborted, TimedOut: t.TimedOut,
				Message: t.Message, FailureLocation: t.Context.(GinkgoContext).FailureLocation}
		}
		Expect(got).To(Equal(map[string]outcome{
			"panics":      {Failed: true, Message: "Test Panicked\nruntime error: index out of range", FailureLocation: "pkg/a.go:42"},
			"aborts":      {Failed: true, Aborted: true, Message: "database unreachable", FailureLocation: "pkg/a_test.go:21"},
			"interrupted": {Failed: true, Message: "Interrupted by User"},
		}))
	})

	It("counts timed-out leaves in TestSummary alongside Failed", func() {
		s := Test{
			Name: "pkg",
			Children: Tests{
				{Name: "ok", Passed: true},
				{Name: "slow", Failed: true, TimedOut: true},
				{Name: "broken", Failed: true},
			},
		}.Sum()
		Expect(s).To(Equal(TestSummary{Total: 3, Passed: 1, Failed: 2, TimedOut: 1}))
	})

	It("does not double-count a parent flagged TimedOut by propagation", func() {
		s := Test{
			Name:     "pkg",
			TimedOut: true,
			Children: Tests{{Name: "slow", Failed: true, TimedOut: true}},
		}.Sum()
		Expect(s).To(Equal(TestSummary{Total: 1, Failed: 1, TimedOut: 1}))
	})

	It("sums TimedOut across two summaries via Add", func() {
		Expect(TestSummary{TimedOut: 2}.Add(TestSummary{TimedOut: 1})).To(Equal(TestSummary{TimedOut: 3}))
	})
})
