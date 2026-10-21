package verifier_test

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/verifier"
)

const specName = "session tab lists every attempt"

func resultChild(result fixtures.FixtureResult) *fixtures.FixtureNode {
	return &fixtures.FixtureNode{Name: result.Name, Type: fixtures.TestNode, Results: &result}
}

func playwrightStep(status task.Status, message string, children ...*fixtures.FixtureNode) fixtures.FixtureResult {
	return fixtures.FixtureResult{Name: "session-ui", Type: "playwright", Status: status, Error: message, Children: children}
}

type verdict struct {
	State   api.VerifyState
	Passed  bool
	Summary api.VerifySummary
	Tests   []api.VerifyNode
}

func verdictOf(report api.VerifyReport) verdict {
	return verdict{State: report.State, Passed: report.Passed, Summary: report.Summary, Tests: report.Tests}
}

var _ = Describe("Report of a step with children", func() {
	passingSpec := fixtures.FixtureResult{Name: specName, Type: "playwright", Status: task.StatusPASS}
	specNode := api.VerifyNode{Name: specName, Framework: "playwright", Passed: true}

	It("fails when the step failed although every child passed", func() {
		step := playwrightStep(task.StatusFAIL, "playwright exited 1", resultChild(passingSpec))

		report := verifier.Report([]fixtures.FixtureResult{step}, nil)

		summary := api.VerifySummary{Total: 2, Passed: 1, Failed: 1}
		Expect(verdictOf(report)).To(Equal(verdict{
			State: api.VerifyStateFailed, Summary: summary,
			Tests: []api.VerifyNode{{
				Name: "session-ui", Framework: "playwright", Message: "playwright exited 1",
				Summary: &summary, Children: []api.VerifyNode{specNode},
			}},
		}))
		Expect(report.Validate()).To(Succeed())
	})

	It("counts a folded teardown failure once beside the passing children", func() {
		background := fixtures.FixtureResult{Name: "server logs", Type: "background", Status: task.StatusFAIL, Error: "panic in log"}
		step := playwrightStep(task.StatusFAIL, "server logs: panic in log", resultChild(passingSpec), resultChild(background))

		report := verifier.Report([]fixtures.FixtureResult{step}, nil)

		Expect(verdictOf(report)).To(Equal(verdict{
			State: api.VerifyStateFailed, Summary: api.VerifySummary{Total: 2, Passed: 1, Failed: 1},
			Tests: []api.VerifyNode{{
				Name: "session-ui", Framework: "playwright", Message: "server logs: panic in log",
				Children: []api.VerifyNode{specNode, {Name: "server logs", Framework: "background", Message: "panic in log", Failed: true}},
			}},
		}))
		Expect(report.Validate()).To(Succeed())
	})

	It("passes a passing step with passing children from its children alone", func() {
		step := playwrightStep(task.StatusPASS, "", resultChild(passingSpec))

		report := verifier.Report([]fixtures.FixtureResult{step}, nil)

		Expect(verdictOf(report)).To(Equal(verdict{
			State: api.VerifyStatePassed, Passed: true, Summary: api.VerifySummary{Total: 1, Passed: 1},
			Tests: []api.VerifyNode{{Name: "session-ui", Framework: "playwright", Children: []api.VerifyNode{specNode}}},
		}))
		Expect(report.Validate()).To(Succeed())
	})
})
