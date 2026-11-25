package snapshots

import (
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CurrentProblems", func() {
	var workDir, guide, readme string
	start := time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)

	save := func(at time.Duration, snap testui.Snapshot) {
		if snap.Metadata == nil {
			snap.Metadata = &testui.SnapshotMetadata{}
		}
		snap.Metadata.Started = start.Add(at)
		snap.Metadata.Ended = start.Add(at + time.Minute)
		_, err := SavePerRun(workDir, &snap, snap.Metadata.Started, "")
		Expect(err).NotTo(HaveOccurred())
	}
	lintRun := func(files []string, results ...*linters.LinterResult) testui.Snapshot {
		return testui.Snapshot{
			Metadata: &testui.SnapshotMetadata{Kind: "lint", LintFiles: files},
			Git:      &testui.SnapshotGit{Root: workDir},
			Status:   testui.SnapshotStatus{LintRun: true},
			Lint:     results,
		}
	}
	result := func(linter string, files ...string) *linters.LinterResult {
		r := &linters.LinterResult{Linter: linter, WorkDir: workDir, Success: true}
		for _, file := range files {
			r.Violations = append(r.Violations, models.Violation{File: file, Line: 2, Message: new(linter + " finding in " + file)})
		}
		return r
	}
	messages := func(problems *Problems) []string {
		var out []string
		for _, r := range problems.Lint {
			for _, v := range r.Violations {
				out = append(out, *v.Message)
			}
		}
		return out
	}

	BeforeEach(func() {
		dir, err := filepath.EvalSymlinks(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		workDir = dir
		guide, readme = filepath.Join(workDir, "guide.md"), filepath.Join(workDir, "readme.md")
		for _, file := range []string{guide, readme, filepath.Join(workDir, "unopened.md")} {
			Expect(os.WriteFile(file, []byte("# doc\n"), 0o600)).To(Succeed())
		}
	})

	It("drops a violation a newer run limited to that file no longer reports, keeping other files' results", func() {
		save(0, lintRun(nil, result("markdown", "guide.md", "readme.md", "unopened.md")))
		save(time.Hour, lintRun([]string{guide}, result("markdown")))

		problems, err := CurrentProblems(workDir, []string{guide, readme})

		Expect(err).NotTo(HaveOccurred())
		Expect(messages(problems)).To(ConsistOf("markdown finding in readme.md"))
		Expect(problems.LintEnded).To(Equal(map[string]time.Time{
			guide:  start.Add(time.Hour + time.Minute),
			readme: start.Add(time.Minute),
		}))
	})

	It("keeps an older linter's violations when a newer run of the same file ran other linters only", func() {
		save(0, lintRun(nil, result("markdown", "guide.md"), result("vale", "guide.md")))
		save(time.Hour, lintRun([]string{guide}, result("markdown")))

		problems, err := CurrentProblems(workDir, []string{guide})

		Expect(err).NotTo(HaveOccurred())
		Expect(messages(problems)).To(ConsistOf("vale finding in guide.md"))
	})

	It("stops at the newest whole-tree run, ignoring older runs and skipped linters", func() {
		save(0, lintRun(nil, result("vale", "guide.md")))
		skipped := result("vale", "guide.md")
		skipped.Skipped = true
		save(time.Hour, lintRun(nil, result("markdown", "guide.md"), skipped))

		problems, err := CurrentProblems(workDir, []string{guide})

		Expect(err).NotTo(HaveOccurred())
		Expect(messages(problems)).To(ConsistOf("markdown finding in guide.md"))
	})

	It("ignores a run that is still in progress", func() {
		save(0, lintRun(nil, result("markdown", "guide.md")))
		running := lintRun(nil)
		running.Status.Running = true
		save(time.Hour, running)

		problems, err := CurrentProblems(workDir, []string{guide})

		Expect(err).NotTo(HaveOccurred())
		Expect(messages(problems)).To(ConsistOf("markdown finding in guide.md"))
	})

	It("returns the newest test run's failed tests located in the files, as absolute leaves", func() {
		save(0, testui.Snapshot{Tests: []parsers.Test{{Name: "old", File: "readme.md", Line: 1, Failed: true}}})
		save(time.Hour, testui.Snapshot{
			Git: &testui.SnapshotGit{Root: workDir},
			Tests: []parsers.Test{{Name: "guide", File: "guide.md", Failed: true, Children: []parsers.Test{
				{Name: "step one", File: "guide.md", Line: 3, Failed: true},
				{Name: "step two", File: "guide.md", Line: 5, Passed: true},
			}}, {Name: "elsewhere", File: "unopened.md", Line: 1, Failed: true}},
		})

		problems, err := CurrentProblems(workDir, []string{guide, readme})

		Expect(err).NotTo(HaveOccurred())
		Expect(problems.Tests).To(Equal([]parsers.Test{{Name: "step one", WorkDir: workDir, File: guide, Line: 3, Failed: true}}))
		Expect(problems.TestsEnded).To(Equal(start.Add(time.Hour + time.Minute)))
	})

	It("is empty for a workspace that never ran", func() {
		problems, err := CurrentProblems(workDir, []string{guide})

		Expect(err).NotTo(HaveOccurred())
		Expect(problems.Lint).To(BeEmpty())
		Expect(problems.Tests).To(BeEmpty())
	})
})
