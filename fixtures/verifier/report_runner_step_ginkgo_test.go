package verifier_test

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/verifier"
)

var _ = Describe("Report of a runner step", func() {
	It("carries the step's summary over its elided children and a timed-out child's own bucket", func() {
		timedOut := fixtures.FixtureResult{Name: "TestSlow", Type: "go test", Status: task.StatusFAIL, TimedOut: true, Error: "panic: test timed out"}
		summary := &api.VerifySummary{Total: 40, Passed: 39, TimedOut: 1}
		step := fixtures.FixtureResult{
			Name: "unit tests", Type: "test", Status: task.StatusFAIL, Summary: summary,
			Children: []*fixtures.FixtureNode{{Name: timedOut.Name, Type: fixtures.TestNode, Results: &timedOut}},
		}

		report := verifier.Report([]fixtures.FixtureResult{step}, nil)

		Expect(report.Tests).To(Equal([]api.VerifyNode{{
			Name: "unit tests", Framework: "test", Message: "", Summary: summary,
			Children: []api.VerifyNode{{
				Name: "TestSlow", Framework: "go test", Message: "panic: test timed out", Failed: true, TimedOut: true,
			}},
		}}))
		Expect(report.Summary).To(Equal(*summary))
		Expect(report.State).To(Equal(api.VerifyStateTimedOut))
		Expect(report.Validate()).To(Succeed())
	})

	It("still passes when a skipped test sits beside passing ones", func() {
		summary := &api.VerifySummary{Total: 40, Passed: 39, Skipped: 1}
		step := fixtures.FixtureResult{Name: "unit tests", Type: "test", Status: task.StatusPASS, Summary: summary}

		report := verifier.Report([]fixtures.FixtureResult{step}, nil)

		Expect([]any{report.State, report.Passed, report.Summary}).To(Equal([]any{api.VerifyStatePassed, true, *summary}))
		Expect(report.Validate()).To(Succeed())
	})
})
