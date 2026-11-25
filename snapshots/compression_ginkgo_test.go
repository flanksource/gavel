package snapshots

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/commons/properties"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Snapshot compression", func() {
	DescribeTable("reads settled history and last-run results in the configured format", func(format, suffix string) {
		properties.Set("gavel.compress.default", format)
		DeferCleanup(properties.Unset, "gavel.compress.default")
		repo := initRepo(GinkgoTB())
		started := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
		snap := &testui.Snapshot{Error: strings.Repeat("failed assertion\n", 5_000), Tests: []parsers.Test{{Name: "TestFailed", Failed: true}}}

		path, err := SavePerRun(repo, snap, started, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(HaveSuffix(".json" + suffix))
		last, err := Save(repo, snap)
		Expect(err).NotTo(HaveOccurred())
		Expect(last).To(HaveSuffix(".json" + suffix))
		Expect(ResolveLast(repo)).To(Equal(last))
		data, err := ReadRun(repo, PointerLast)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("TestFailed"))
		Expect(LoadRunFile(path)).To(Equal(snap))

		runs, err := ListRuns(repo, ListRunsOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(runs).To(HaveLen(1))
		Expect(runs[0].RunID).To(Equal("run-2026-11-01T12-00-00Z"))
		Expect(runs[0].Path).To(Equal(path))
		data, err = ReadRun(repo, runs[0].RunID)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("TestFailed"))
		Expect(LatestRuns(repo)).To(Equal([]string{path}))

		small, err := SavePerRun(repo, sampleSnapshot(), started, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(small).To(HaveSuffix("-2.json"))
		Expect(strings.TrimSuffix(path, suffix)).NotTo(BeAnExistingFile())
	}, Entry("zstd", "zstd", ".zst"), Entry("gzip", "gzip", ".gz"))

	It("switches formats when a new result replaces the same commit's snapshot", func() {
		repo := initRepo(GinkgoTB())
		original, err := Save(repo, &testui.Snapshot{Error: strings.Repeat("failed assertion\n", 5_000)})
		Expect(err).NotTo(HaveOccurred())
		Expect(original).To(HaveSuffix(".zst"))
		oldPointer, err := LoadPointer(repo, PointerLast)
		Expect(err).NotTo(HaveOccurred())

		current, err := Save(repo, sampleSnapshot())
		Expect(err).NotTo(HaveOccurred())
		Expect(current).To(Equal(strings.TrimSuffix(original, ".zst")))
		Expect(filepath.Dir(current)).To(Equal(filepath.Join(repo, Dir)))
		Expect(LoadByPointer(repo, oldPointer)).To(Equal(sampleSnapshot()))
		Expect(ResolveLast(repo)).To(Equal(current))
		Expect(original).NotTo(BeAnExistingFile())
	})
})
