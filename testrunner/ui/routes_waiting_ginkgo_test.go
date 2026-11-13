package testui

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
)

var _ = Describe("fixture benchmark progress nodes", func() {
	It("lists only finished leaves, so a fixture waiting on a form is not one", func() {
		leaf := func(name string, state fixtures.ExecutionState) *fixtures.ExecutionNode {
			return &fixtures.ExecutionNode{Name: name, Kind: "fixture", State: state}
		}
		state := &FixtureBenchmarkState{Progress: &fixtures.ExecutionSnapshot{Root: &fixtures.ExecutionNode{Children: []*fixtures.ExecutionNode{
			leaf("approval", fixtures.ExecutionWaiting),
			leaf("inspect", fixtures.ExecutionRunning),
			leaf("setup", fixtures.ExecutionPassed),
		}}}}

		var names []string
		for _, node := range buildFixtureBenchmarkViewNodes(state) {
			names = append(names, node.Name)
		}
		Expect(names).To(Equal([]string{"setup"}))
	})
})
