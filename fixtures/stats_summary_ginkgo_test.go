package fixtures

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// runnerStepNode is a `yaml test` step that ran 40 tests and shows only the 3
// that did not pass: two failures and one timeout.
func runnerStepNode() *FixtureNode {
	summary := api.VerifySummary{Total: 40, Passed: 36, Failed: 2, TimedOut: 1, Warned: 1}
	step := &FixtureNode{Name: "unit tests", Type: TestNode}
	(&Runner{}).attachResult(step, FixtureResult{
		Name: "unit tests", Type: "test", Status: task.StatusFAIL, Summary: &summary,
		Children: []*FixtureNode{
			{Name: "TestA", Type: TestNode, Results: &FixtureResult{Name: "TestA", Status: task.StatusFAIL}},
			{Name: "TestB", Type: TestNode, Results: &FixtureResult{Name: "TestB", Status: task.StatusFAIL}},
			{Name: "TestSlow", Type: TestNode, Results: &FixtureResult{Name: "TestSlow", Status: task.StatusFAIL, TimedOut: true}},
		},
	})
	return step
}

var runnerStepStats = Stats{Total: 40, Passed: 36, Failed: 3, TimedOut: 1, Warned: 1}

var _ = Describe("Stats of a step that carries a summary", func() {
	It("counts the summary, not the visible rows", func() {
		Expect(runnerStepNode().GetStats()).To(Equal(runnerStepStats))
	})

	It("adds the summary to its siblings' counts in the file total", func() {
		file := &FixtureNode{Name: "steps.fixture.yaml", Type: FileNode}
		file.AddChild(runnerStepNode())
		file.AddChild(&FixtureNode{Name: "build", Type: TestNode, Results: &FixtureResult{Name: "build", Status: task.StatusPASS}})

		Expect(file.GetStats()).To(Equal(Stats{Total: 41, Passed: 37, Failed: 3, TimedOut: 1, Warned: 1}))
	})

	It("still counts the step's own failure once when its summary shows none", func() {
		unfinished := api.VerifySummary{Total: 2, Pending: 2}
		step := &FixtureNode{Name: "unit tests", Type: TestNode, Results: &FixtureResult{Status: task.StatusFAIL, Summary: &unfinished}}

		Expect(step.GetStats()).To(Equal(Stats{Total: 3, Failed: 1}))
	})

	It("prints timed out and warned beside the passed/failed counts", func() {
		Expect(runnerStepStats.Pretty().String()).To(Equal("36/3 1 timed out 1 warned"))
		Expect(runnerStepStats.String()).To(Equal("36/39 1 timed out 1 warned"))
	})
})

var _ = DescribeTable("Stats.Add of a leaf result",
	func(result FixtureResult, expected Stats) {
		Expect(Stats{}.Add(&result)).To(Equal(expected))
	},
	Entry("a timed-out failure counts as failed and timed out", FixtureResult{Status: task.StatusFAIL, TimedOut: true}, Stats{Total: 1, Failed: 1, TimedOut: 1}),
	Entry("a warning counts as warned only", FixtureResult{Status: task.StatusWarning}, Stats{Total: 1, Warned: 1}),
)
