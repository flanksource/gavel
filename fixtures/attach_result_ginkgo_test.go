package fixtures

import (
	"time"

	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func specChild(name string, status task.Status) *FixtureNode {
	return &FixtureNode{Name: name, Type: TestNode, Results: &FixtureResult{Name: name, Status: status}}
}

func stepWithChildren(status task.Status, children ...*FixtureNode) FixtureResult {
	return FixtureResult{
		Name:     "session-ui",
		Type:     "playwright",
		Status:   status,
		Duration: 3 * time.Second,
		Stdout:   "1 passed",
		Command:  "npx playwright test",
		Children: children,
	}
}

var _ = Describe("attachResult on a step whose result has children", func() {
	It("keeps the step's own detail and status on the node and attaches its children", func() {
		spec := specChild("renders the session", task.StatusPASS)
		node := &FixtureNode{Name: "session-ui", Type: TestNode}

		(&Runner{}).attachResult(node, stepWithChildren(task.StatusPASS, spec))

		own := stepWithChildren(task.StatusPASS)
		own.Children = nil
		Expect(node.Results).To(Equal(&own))
		Expect(node.Children).To(Equal([]*FixtureNode{spec}))
		Expect(spec.Parent).To(BeIdenticalTo(node))
		Expect(node.GetStats()).To(Equal(Stats{Total: 1, Passed: 1}))
	})
})

var _ = DescribeTable("GetStats of a step node with children",
	func(status task.Status, children []*FixtureNode, expected Stats) {
		node := &FixtureNode{Name: "session-ui", Type: TestNode}
		(&Runner{}).attachResult(node, stepWithChildren(status, children...))
		Expect(node.GetStats()).To(Equal(expected))
	},
	Entry("a failed step with only passing children counts its own failure once",
		task.StatusFAIL, []*FixtureNode{specChild("a", task.StatusPASS), specChild("b", task.StatusPASS)},
		Stats{Total: 3, Passed: 2, Failed: 1}),
	Entry("an errored step with only passing children counts its own error once",
		task.StatusERR, []*FixtureNode{specChild("a", task.StatusPASS)},
		Stats{Total: 2, Passed: 1, Error: 1}),
	Entry("a failed runner step with failing children counts the children only",
		task.StatusFAIL, []*FixtureNode{specChild("a", task.StatusFAIL), specChild("b", task.StatusPASS)},
		Stats{Total: 2, Passed: 1, Failed: 1}),
	Entry("a passing step counts its children only",
		task.StatusPASS, []*FixtureNode{specChild("a", task.StatusPASS), specChild("b", task.StatusSKIP)},
		Stats{Total: 2, Passed: 1, Skipped: 1}),
)

var _ = Describe("ApplyDisplayOptions on a passing step with a failing child", func() {
	It("keeps the step so the failing child stays visible", func() {
		failing := specChild("a", task.StatusFAIL)
		root := &FixtureNode{Name: "file", Type: FileNode}
		step := &FixtureNode{Name: "session-ui", Type: TestNode}
		root.AddChild(step)
		(&Runner{}).attachResult(step, stepWithChildren(task.StatusPASS, failing))

		root.ApplyDisplayOptions(DisplayOptions{})

		Expect(root.Children).To(Equal([]*FixtureNode{step}))
		Expect(step.Children).To(Equal([]*FixtureNode{failing}))
	})
})
