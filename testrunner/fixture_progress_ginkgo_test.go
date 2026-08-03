package testrunner

import (
	"time"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture progress adapter", func() {
	It("preserves the typed execution hierarchy and lifecycle states", func() {
		snapshot := fixtures.ExecutionSnapshot{
			Version: fixtures.ExecutionSnapshotVersion,
			State:   fixtures.ExecutionRunning,
			Root: &fixtures.ExecutionNode{
				Key: "fixtures", Name: "Fixtures", Kind: fixtures.ExecutionKindRoot, State: fixtures.ExecutionRunning,
				Children: []*fixtures.ExecutionNode{
					{Key: "setup", Name: "Setup", Kind: fixtures.ExecutionKindSetup, State: fixtures.ExecutionPassed, Duration: time.Second},
					{
						Key: "verification.md", Name: "verification.md", Kind: fixtures.ExecutionKindFile, State: fixtures.ExecutionRunning,
						Children: []*fixtures.ExecutionNode{
							{Key: "tests", Name: "Tests", Kind: fixtures.ExecutionKindTest, State: fixtures.ExecutionRunning, Done: 3, Total: 8},
							{Key: "lint", Name: "Lint", Kind: fixtures.ExecutionKindLint, State: fixtures.ExecutionQueued},
						},
					},
				},
			},
		}

		tests := executionSnapshotToTests(snapshot)

		Expect(tests).To(HaveLen(2))
		Expect(tests[0].Name).To(Equal("Setup"))
		Expect(tests[0].Passed).To(BeTrue())
		Expect(tests[1].Children).To(HaveLen(2))
		Expect(tests[1].Children[0].Running).To(BeTrue())
		Expect(tests[1].Children[0].Progress).To(Equal(&parsers.TestProgress{Phase: "test", Done: 3, Total: 8}))
		Expect(tests[1].Children[1].Pending).To(BeTrue())
		Expect(tests[1].Children[0].TaskID).To(Equal("tests"))
	})
})
