package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/clicky"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// entityLinter is an in-process linter over virtual entity keys, as an
// embedder (oipa-cli) registers one: no binary, no files on disk.
type entityLinter struct {
	includes []string
	opts     *linters.RunOptions
	runs     int
	finding  string
}

func (l *entityLinter) Name() string                    { return "fake-entities" }
func (l *entityLinter) InProcess() bool                 { return true }
func (l *entityLinter) DefaultIncludes() []string       { return l.includes }
func (l *entityLinter) DefaultExcludes() []string       { return nil }
func (l *entityLinter) SupportsJSON() bool              { return true }
func (l *entityLinter) JSONArgs() []string              { return nil }
func (l *entityLinter) SupportsFix() bool               { return false }
func (l *entityLinter) FixArgs() []string               { return nil }
func (l *entityLinter) SetOptions(o linters.RunOptions) { l.opts = &o }
func (l *entityLinter) ValidateConfig(*models.LinterConfig) error {
	return nil
}

func (l *entityLinter) Run(commonsContext.Context, *clicky.Task) ([]models.Violation, error) {
	l.runs++
	message := "copybook missing"
	return []models.Violation{{File: l.finding, Line: 3, Severity: models.SeverityError, Message: &message}}, nil
}

var _ = Describe("In-process linters", func() {
	const entityFile = "entities/transaction/Fee"
	var (
		root   string
		linter *entityLinter
		reg    *linters.Registry
	)
	BeforeEach(func() {
		clicky.ClearGlobalTasks()
		DeferCleanup(clicky.ClearGlobalTasks)
		root = GinkgoT().TempDir()
		Expect(os.Mkdir(filepath.Join(root, ".git"), 0o755)).To(Succeed())
		git, err := exec.LookPath("git")
		Expect(err).NotTo(HaveOccurred())
		bin := GinkgoT().TempDir()
		Expect(os.Symlink(git, filepath.Join(bin, "git"))).To(Succeed())
		GinkgoT().Setenv("PATH", bin)
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		linter = &entityLinter{includes: []string{"**/*.entity.xml"}, finding: entityFile}
		reg = linters.NewRegistry()
		reg.Register(linter)
	})

	It("runs when named with no executable on PATH and no matching files", func() {
		results, err := Execute(Options{WorkDir: root, Registry: reg, Linters: []string{linter.Name()}, Timeout: "5s"})

		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveExactElements(And(
			HaveField("Linter", linter.Name()),
			HaveField("Success", true),
			HaveField("Command", linter.Name()),
			HaveField("Violations", HaveExactElements(HaveField("File", entityFile))),
		)))
	})

	It("is prepared, not skipped, by a dry run", func() {
		Expect(DryRun(Options{WorkDir: root, Registry: reg, Linters: []string{linter.Name()}})).To(Succeed())

		Expect(linter.opts).NotTo(BeNil(), "dry run must prepare the in-process linter's command")
		cmd, args, err := linters.PrepareCommand(linter, *linter.opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(append([]string{cmd}, args...)).To(Equal([]string{linter.Name()}))
	})

	It("narrows its Files to the failed entities on --failed", func() {
		previous := testui.Snapshot{Lint: []*linters.LinterResult{
			{Linter: linter.Name(), WorkDir: root, Violations: []models.Violation{{File: entityFile}}},
			{Linter: "typescript", WorkDir: root},
		}}
		raw, err := json.Marshal(previous)
		Expect(err).NotTo(HaveOccurred())
		failed := filepath.Join(root, "last.json")
		Expect(os.WriteFile(failed, raw, 0o600)).To(Succeed())

		results, err := Run(context.Background(), Options{WorkDir: root, Registry: reg, Failed: failed, Timeout: "5s"})

		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveExactElements(HaveField("Linter", linter.Name())))
		Expect(linter.opts.Files).To(Equal([]string{entityFile}))
	})

	It("is never auto-detected, even with matching files", func() {
		writeLintFixture(root, map[string]string{"Fee.entity.xml": "<Transaction/>\n"})

		results, err := Execute(Options{WorkDir: root, Registry: reg, Timeout: "5s"})

		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(BeEmpty())
		Expect(linter.runs).To(BeZero())
	})
})

// namedEntityLinter is an entityLinter registered under its own name, so a run
// can schedule several.
type namedEntityLinter struct {
	entityLinter
	name string
}

func (l *namedEntityLinter) Name() string { return l.name }

var _ = Describe("lint progress", func() {
	var (
		root  string
		reg   *linters.Registry
		names = []string{"fake-copybooks", "fake-entities", "fake-rules"}
	)
	BeforeEach(func() {
		clicky.ClearGlobalTasks()
		DeferCleanup(clicky.ClearGlobalTasks)
		root = GinkgoT().TempDir()
		Expect(os.Mkdir(filepath.Join(root, ".git"), 0o755)).To(Succeed())
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		reg = linters.NewRegistry()
		for _, name := range names {
			reg.Register(&namedEntityLinter{entityLinter: entityLinter{finding: "entities/" + name}, name: name})
		}
	})

	It("reports done of total as each linter finishes, then the final count", func() {
		var reports [][2]int
		_, err := Run(context.Background(), Options{WorkDir: root, Registry: reg, Linters: names, Timeout: "5s",
			Progress: func(done, total int) error {
				reports = append(reports, [2]int{done, total})
				return nil
			}})

		Expect(err).NotTo(HaveOccurred())
		Expect(reports).To(Equal([][2]int{{1, 3}, {2, 3}, {3, 3}, {3, 3}}))
	})

	It("fails the run with the first progress report that failed", func() {
		_, err := Run(context.Background(), Options{WorkDir: root, Registry: reg, Linters: names, Timeout: "5s",
			Progress: func(done, total int) error { return fmt.Errorf("sink closed at %d/%d", done, total) }})

		Expect(err).To(MatchError("report lint progress: sink closed at 1/3"))
	})
})

var _ = Describe("ReportView", func() {
	message := "unused"
	results := []*linters.LinterResult{{Linter: "golangci-lint", Violations: []models.Violation{{File: "a.go", Line: 1, Message: &message}}}}

	It("renders the summary view the CLI shows for --summary", func() {
		Expect(ReportView(results, ReportOptions{Summary: true, SummaryLimit: 2})).To(Equal(&linters.SummaryView{Results: results, Limit: 2}))
	})

	It("renders the results themselves without --summary", func() {
		Expect(ReportView(results, ReportOptions{SummaryLimit: 2})).To(Equal(results))
	})
})
