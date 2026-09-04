package testrunner

import (
	"os"
	"path/filepath"
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

	It("loads an exact direct fixture path without discovering ordinary tests", func() {
		workDir := GinkgoT().TempDir()
		fixturePath := filepath.Join(workDir, "direct.md")
		Expect(os.WriteFile(fixturePath, []byte(`# Direct fixture

| Name | Command | CEL |
|------|---------|-----|
| passes | printf direct | stdout == "direct" |
`), 0o600)).To(Succeed())
		nestedModule := filepath.Join(workDir, ".runtime", "worktrees", "nested")
		Expect(os.MkdirAll(nestedModule, 0o700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(nestedModule, "go.mod"), []byte("module example.com/nested\n"), 0o600)).To(Succeed())

		result, err := Run(RunOptions{
			WorkDir:      workDir,
			Fixtures:     true,
			FixturesOnly: true,
			FixtureRunner: &fixtures.RunnerOptions{
				Paths:   []string{fixturePath},
				WorkDir: workDir,
			},
		})

		Expect(err).NotTo(HaveOccurred())
		tests, ok := result.([]parsers.Test)
		Expect(ok).To(BeTrue())
		Expect(tests).To(HaveLen(1))
		Expect(tests[0].Name).To(Equal("direct.md"))
		Expect(tests[0].Children).To(HaveLen(1))
		Expect(tests[0].Children[0].Children).To(HaveLen(1))
		test := tests[0].Children[0].Children[0]
		Expect(test.Name).To(Equal("passes"))
		Expect(test.Framework).To(Equal(parsers.Fixture))
	})
})
