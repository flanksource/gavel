package testrunner_test

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/fixtures"
	// The exec fixture type registers itself from fixtures/types, which imports
	// testrunner, so only an external test package can load it — exactly as the
	// gavel binary does.
	_ "github.com/flanksource/gavel/fixtures/types"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("direct fixture run", func() {
	It("runs an exact fixture path to a passing file/section/table/row tree without discovering ordinary tests", func() {
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

		result, err := testrunner.Run(testrunner.RunOptions{
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
		section := tests[0].Children[0]
		Expect(section.Name).To(Equal("Direct fixture"))
		Expect(section.Children).To(HaveLen(1))
		table := section.Children[0]
		Expect(table.Name).To(Equal("Table 1"))
		Expect(table.Children).To(HaveLen(1))
		test := table.Children[0]
		Expect(test.Name).To(Equal("passes"))
		Expect(test.Framework).To(Equal(parsers.Fixture))
		Expect(test.Passed).To(BeTrue())
		Expect(test.Stdout).To(Equal("direct"))
	})
})
