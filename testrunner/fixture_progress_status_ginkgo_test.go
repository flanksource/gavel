package testrunner

import (
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture tree progress labels", func() {
	DescribeTable("uses the reported status, then description, then execution kind",
		func(metadata map[string]any, phase string) {
			node := &fixtures.ExecutionNode{
				Key: "progress", Name: "Process batches", Kind: fixtures.ExecutionKindCommand,
				State: fixtures.ExecutionRunning, Done: 2, Total: 5, Metadata: metadata,
			}
			test, err := executionNodeToTest(node)
			Expect(err).NotTo(HaveOccurred())
			Expect(test.Progress).To(Equal(&parsers.TestProgress{Phase: phase, Done: 2, Total: 5}))
		},
		Entry("a stored routine's status", map[string]any{"status": "Completed batch 2 of 5", "description": "Processing batches"}, "Completed batch 2 of 5"),
		Entry("a description without status", map[string]any{"description": "Processing batches"}, "Processing batches"),
		Entry("an empty status", map[string]any{"status": "", "description": "Processing batches"}, "Processing batches"),
		Entry("no workflow messages", nil, "command"),
	)
})
