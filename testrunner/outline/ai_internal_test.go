package outline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	captaincli "github.com/flanksource/captain/pkg/cli"
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
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		Expect(os.WriteFile(filepath.Join(home, ".captain.yaml"), []byte("ai:\n  defaultModel: agent:claude-sonnet-5\n  timeout: 17m\n  maxTokens: 3000\n  noSkills: true\n"), 0o600)).To(Succeed())
		previous := newSummaryAgent
		GinkgoT().Cleanup(func() { newSummaryAgent = previous })
		report = &Report{Entries: []*Entry{
			{Framework: parsers.GoTest, File: "gotest/sample_test.go", Name: "TestAdd", Line: 8, Description: "add"},
			{Framework: parsers.GoTest, File: "gotest/sample_test.go", Name: "TestTable", Line: 14, Description: "table"},
		}}
	})

	It("attaches matched summaries and keeps static descriptions elsewhere", func() {
		stub := &stubSummaryAgent{}
		var projected captaincli.AIRuntimeResolved
		newSummaryAgent = func(runtime captaincli.AIRuntimeResolved) (SummaryAgent, error) {
			projected = runtime
			return stub, nil
		}
		Expect(applyAISummaries(context.Background(), report, "testdata")).To(Succeed())
		Expect(stub.requests).To(HaveLen(1))
		Expect(stub.requests[0].Spec.Budget.Timeout).To(Equal("17m"))
		Expect(stub.requests[0].Spec.Budget.MaxTokens).To(Equal(3000))
		Expect(stub.requests[0].Spec.Memory.SkipSkills).To(BeTrue())
		Expect(projected.Request).To(Equal(stub.requests[0].Spec))
		Expect(projected.Config.Budget).To(Equal(stub.requests[0].Spec.Budget))
		Expect(report.Entries[0].AISummary).To(Equal("verifies add returns the arithmetic sum"))
		Expect(report.Entries[1].AISummary).To(BeEmpty())
		Expect(report.Entries[1].Description).To(Equal("table"))
		Expect(stub.closed).To(BeTrue(), "agent must be closed so process-backed backends release their child process")
	})

	It("keeps the outline alive when the agent fails for a file", func() {
		newSummaryAgent = func(captaincli.AIRuntimeResolved) (SummaryAgent, error) { return &stubSummaryAgent{fail: true}, nil }
		Expect(applyAISummaries(context.Background(), report, "testdata")).To(Succeed())
		Expect(report.Entries[0].AISummary).To(BeEmpty())
	})
})
