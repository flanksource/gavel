package testrunner

import (
	"path/filepath"

	"github.com/flanksource/gavel/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture benchmark run", func() {
	It("surfaces an invalid baseline to the command", func() {
		workDir := GinkgoT().TempDir()
		_, err := Run(RunOptions{WorkDir: workDir, FixturesOnly: true, FixtureRunner: &fixtures.RunnerOptions{
			WorkDir: workDir, Benchmark: true, Baseline: filepath.Join(workDir, "missing.json"),
		}})
		Expect(err).To(MatchError(ContainSubstring("--baseline")))
	})
})
