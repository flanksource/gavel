package main

import (
	"errors"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/baseline"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/workspace"
)

var _ = Describe("persisting a completed run", func() {
	It("moves last.json to the run so a bare --failed re-runs its failures", func() {
		workDir, err := filepath.EvalSymlinks(GinkgoT().TempDir())
		Expect(err).ToNot(HaveOccurred())
		for _, args := range [][]string{{"init", "-q"}, {"-c", "user.email=t@example.com", "-c", "user.name=t", "commit", "-q", "--allow-empty", "-m", "init"}} {
			Expect(exec.Command("git", append([]string{"-C", workDir}, args...)...).Run()).To(Succeed())
		}
		failing := parsers.Test{Framework: parsers.GoTest, PackagePath: "./pkg/foo", Name: "TestFoo", Failed: true}
		opts := testrunner.RunOptions{WorkDir: workDir, Dirs: workspace.Dirs{Invocation: workDir, StateRoot: workDir}}
		snapshot := buildTestSnapshot(opts, []parsers.Test{failing}, nil, time.Now().UTC(), time.Now().UTC(), nil)

		persistRunSnapshot(workDir, &snapshot, snapshot.Metadata.Started)

		last, err := snapshots.ResolveLast(workDir)
		Expect(err).ToNot(HaveOccurred())
		loaded, err := baseline.LoadSnapshot(last)
		Expect(err).ToNot(HaveOccurred())
		Expect(baseline.ExtractFailedTestPackages(loaded.Tests)).To(Equal(map[parsers.Framework]map[string][]string{
			parsers.GoTest: {failing.PackagePath: {failing.Name}},
		}))
		runs, err := snapshots.LatestRuns(workDir)
		Expect(err).ToNot(HaveOccurred())
		Expect(runs).To(HaveLen(1))
	})
})

var _ = Describe("saving a run the runner failed", func() {
	It("records the partial tests and the failure where `gavel ui serve` finds the latest run", func() {
		workDir := GinkgoT().TempDir()
		started := time.Date(2026, time.October, 23, 11, 59, 0, 0, time.UTC)
		completed := parsers.Test{Framework: parsers.GoTest, PackagePath: "./pkg/foo", Name: "TestFoo", Passed: true}
		runErr := errors.New("test execution failed: read the process samples of pid 3916: session ended")

		saved := saveFailedRun(failedRun{
			Opts: testrunner.RunOptions{WorkDir: workDir, Dirs: workspace.Dirs{Invocation: workDir, StateRoot: workDir}}, Tests: []parsers.Test{completed},
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
