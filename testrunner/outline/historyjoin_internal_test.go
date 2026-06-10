package outline

import (
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

var _ = Describe("joinHistory", func() {
	var workDir string
	var report *Report

	ran := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)

	writeRun := func(tests []parsers.Test) {
		_, err := snapshots.SavePerRun(workDir, &testui.Snapshot{
			Metadata: &testui.SnapshotMetadata{Started: ran},
			Git:      &testui.SnapshotGit{Root: workDir, Repo: "repo", SHA: "abc"},
			Tests:    tests,
		}, ran)
		Expect(err).NotTo(HaveOccurred())
	}

	// Run snapshots record package paths but rarely source files, mirroring
	// what gavel test writes in practice.
	BeforeEach(func() {
		workDir = ginkgo.GinkgoT().TempDir()
		writeRun([]parsers.Test{
			{
				Framework:   parsers.GoTest,
				PackagePath: "./pkg/foo",
				Name:        "TestFoo",
				Passed:      true,
				Duration:    20 * time.Millisecond,
			},
			{
				Framework:   parsers.Ginkgo,
				PackagePath: "./pkg/calc",
				Suite:       []string{"Calculator", "when adding"},
				Name:        "adds two positive numbers",
				Failed:      true,
				Duration:    5 * time.Millisecond,
			},
			{
				Framework:   parsers.Vitest,
				PackagePath: "./site",
				Suite:       []string{"WaitlistForm"},
				Name:        "rejects an invalid email",
				Passed:      true,
				Duration:    time.Millisecond,
			},
		})

		report = &Report{Entries: []*Entry{
			{Framework: parsers.GoTest, File: "pkg/foo/foo_test.go", Name: "TestFoo"},
			{
				Framework: parsers.Ginkgo, File: "pkg/calc/calc_test.go", Name: "Calculator", Container: true,
				Children: []*Entry{
					{Framework: parsers.Ginkgo, File: "pkg/calc/calc_test.go", Suite: []string{"Calculator", "when adding"}, Name: "adds two positive numbers"},
					{Framework: parsers.Ginkgo, File: "pkg/calc/calc_test.go", Suite: []string{"Calculator"}, Name: dynamicName, Dynamic: true},
					{Framework: parsers.Ginkgo, File: "pkg/calc/calc_test.go", Suite: []string{"Calculator"}, Name: "never executed"},
				},
			},
			{Framework: parsers.Vitest, File: "site/src/components/WaitlistForm.test.tsx", Suite: []string{"WaitlistForm"}, Name: "rejects an invalid email"},
		}}
	})

	It("joins gotest leaves on the package containing the file", func() {
		runs, err := joinHistory(report, Options{WorkDir: workDir})
		Expect(err).NotTo(HaveOccurred())
		Expect(runs).To(Equal(1))
		Expect(report.Entries[0].History).NotTo(BeNil())
		Expect(report.Entries[0].History.PassCount).To(Equal(1))
	})

	It("joins ginkgo specs on package, suite chain, and text", func() {
		_, err := joinHistory(report, Options{WorkDir: workDir})
		Expect(err).NotTo(HaveOccurred())
		spec := report.Entries[1].Children[0]
		Expect(spec.History).NotTo(BeNil())
		Expect(spec.History.FailCount).To(Equal(1))
	})

	It("joins vitest leaves by walking up to the npm package root", func() {
		_, err := joinHistory(report, Options{WorkDir: workDir})
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Entries[2].History).NotTo(BeNil())
	})

	It("leaves dynamic and never-executed specs without history", func() {
		_, err := joinHistory(report, Options{WorkDir: workDir})
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Entries[1].Children[1].History).To(BeNil())
		Expect(report.Entries[1].Children[2].History).To(BeNil())
	})

	It("returns zero runs without error when no history exists", func() {
		empty := ginkgo.GinkgoT().TempDir()
		runs, err := joinHistory(report, Options{WorkDir: empty})
		Expect(err).NotTo(HaveOccurred())
		Expect(runs).To(BeZero())
	})
})
