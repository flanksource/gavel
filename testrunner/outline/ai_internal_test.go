package outline

import (
	"context"
	"encoding/json"
	"fmt"

	clickyai "github.com/flanksource/gavel/ai"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/testrunner/parsers"
)

// stubSummaryAgent answers every prompt with a summary per requested id,
// returning the structured JSON the real agent would produce from the
// frontmatter schema.
type stubSummaryAgent struct {
	requests []clickyai.PromptRequest
	fail     bool
	closed   bool
}

func (s *stubSummaryAgent) Close() error {
	s.closed = true
	return nil
}

func (s *stubSummaryAgent) ExecutePrompt(_ context.Context, req clickyai.PromptRequest) (*clickyai.PromptResponse, error) {
	s.requests = append(s.requests, req)
	if s.fail {
		return nil, fmt.Errorf("stub agent failure")
	}
	data, _ := json.Marshal(fileSummariesSchema{Tests: []testSummaryItem{
		{ID: "gotest/sample_test.go:8 TestAdd", Summary: "verifies add returns the arithmetic sum"},
	}})
	return &clickyai.PromptResponse{StructuredData: json.RawMessage(data)}, nil
}

var _ = Describe("applyAISummaries", func() {
	var report *Report

	BeforeEach(func() {
		report = &Report{Entries: []*Entry{
			{Framework: parsers.GoTest, File: "gotest/sample_test.go", Name: "TestAdd", Line: 8, Description: "add"},
			{Framework: parsers.GoTest, File: "gotest/sample_test.go", Name: "TestTable", Line: 14, Description: "table"},
		}}
	})

	It("attaches matched summaries and keeps static descriptions elsewhere", func() {
		stub := &stubSummaryAgent{}
		newSummaryAgent = func() (SummaryAgent, error) { return stub, nil }
		Expect(applyAISummaries(context.Background(), report, "testdata")).To(Succeed())
		Expect(stub.requests).To(HaveLen(1))
		Expect(report.Entries[0].AISummary).To(Equal("verifies add returns the arithmetic sum"))
		Expect(report.Entries[1].AISummary).To(BeEmpty())
		Expect(report.Entries[1].Description).To(Equal("table"))
		Expect(stub.closed).To(BeTrue(), "agent must be closed so process-backed backends release their child process")
	})

	It("keeps the outline alive when the agent fails for a file", func() {
		newSummaryAgent = func() (SummaryAgent, error) { return &stubSummaryAgent{fail: true}, nil }
		Expect(applyAISummaries(context.Background(), report, "testdata")).To(Succeed())
		Expect(report.Entries[0].AISummary).To(BeEmpty())
	})
})
