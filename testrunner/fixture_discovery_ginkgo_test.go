package testrunner_test

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/types"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("discovered fixtures", func() {
	const fixture = `# Discovered fixture

| Name | Command | CEL |
|------|---------|-----|
| passes | printf discovered | stdout == "discovered" |
`
	// A shared template that fixtures reference: matched by a broad glob such
	// as pkg/**/*.md, but it declares no fixtures of its own.
	const template = "# Shared template\n\nRendered by the fixtures that include it.\n"

	writeFile := func(dir, name, content string) string {
		path := filepath.Join(dir, name)
		Expect(os.WriteFile(path, []byte(content), 0o600)).To(Succeed())
		return path
	}

	It("skips a glob-matched markdown file that declares no fixtures and runs the rest", func() {
		workDir := GinkgoT().TempDir()
		writeFile(workDir, "checks.md", fixture)
		writeFile(workDir, "template.md", template)

		result, err := testrunner.Run(testrunner.RunOptions{
			WorkDir:      workDir,
			Fixtures:     true,
			FixturesOnly: true,
			FixtureFiles: []string{"*.md"},
		})

		Expect(err).NotTo(HaveOccurred())
		tests, ok := result.([]parsers.Test)
		Expect(ok).To(BeTrue())
		Expect(tests).To(ConsistOf(And(
			HaveField("Name", "checks.md"),
			HaveField("Children", HaveLen(1)),
		)))
		Expect(parsers.Tests(tests).Sum().Passed).To(Equal(1))
	})

	It("reports an empty run, not a fixture error, when every glob-matched file declares no fixtures", func() {
		workDir := GinkgoT().TempDir()
		writeFile(workDir, "template.md", template)

		_, err := testrunner.Run(testrunner.RunOptions{
			WorkDir:      workDir,
			Fixtures:     true,
			FixturesOnly: true,
			FixtureFiles: []string{"*.md"},
		})

		Expect(err).To(MatchError("no tests were executed"))
	})

	It("still rejects an explicitly named file that declares no fixtures", func() {
		workDir := GinkgoT().TempDir()
		templatePath := writeFile(workDir, "template.md", template)

		_, err := testrunner.Run(testrunner.RunOptions{
			WorkDir:      workDir,
			Fixtures:     true,
			FixturesOnly: true,
			FixtureRunner: &fixtures.RunnerOptions{
				Paths:   []string{templatePath},
				WorkDir: workDir,
			},
		})

		Expect(err).To(MatchError(fixtures.ErrNoFixtures))
	})
})
