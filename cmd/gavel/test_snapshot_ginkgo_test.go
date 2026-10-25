package main

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
)

var _ = Describe("saving a run the runner failed", func() {
	It("records the partial tests and the failure where `gavel ui serve` finds the latest run", func() {
		workDir := GinkgoT().TempDir()
		started := time.Date(2026, time.October, 23, 11, 59, 0, 0, time.UTC)
		completed := parsers.Test{Framework: parsers.GoTest, PackagePath: "./pkg/foo", Name: "TestFoo", Passed: true}
		runErr := errors.New("test execution failed: read the process samples of pid 3916: session ended")

		saved := saveFailedRun(failedRun{
			Opts: testrunner.RunOptions{WorkDir: workDir}, Tests: []parsers.Test{completed},
			Started: started, Err: runErr,
		})

		paths, err := snapshots.LatestRuns(workDir)
		Expect(err).ToNot(HaveOccurred())
		Expect(paths).To(HaveLen(1))
		loaded, err := snapshots.LoadRunFile(paths[0])
		Expect(err).ToNot(HaveOccurred())
		Expect(loaded.Error).To(Equal(runErr.Error()))
		Expect(*loaded.Metadata.ExitCode).To(Equal(1))
		Expect(loaded.Tests).To(HaveExactElements(HaveField("Name", completed.Name)))
		Expect(saved.Error).To(Equal(runErr.Error()))
	})
})
