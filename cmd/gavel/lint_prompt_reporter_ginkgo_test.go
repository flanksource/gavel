package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/workspace"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type promptReportFailure struct{ err error }

func (w promptReportFailure) Write([]byte) (int, error) { return 0, w.err }

var _ = Describe("Prompt lint reporter", func() {
	var root string
	var output bytes.Buffer
	var ctx context.Context
	var cancel context.CancelCauseFunc
	var reporter *promptLintReporter
	var finding models.Violation

	BeforeEach(func() {
		root = GinkgoT().TempDir()
		output.Reset()
		ctx, cancel = context.WithCancelCause(context.Background())
		DeferCleanup(func() { cancel(nil) })
		var err error
		reporter, err = newPromptLintReporter(promptLintReportOptions{
			Options: LintOptions{WorkDir: root, Dirs: workspace.Dirs{Invocation: root, StateRoot: root}, Exclude: []string{"prompt:excluded"}},
			Writer:  &output, Cancel: cancel,
		})
		Expect(err).NotTo(HaveOccurred())
		finding = models.Violation{File: filepath.Join(root, "proposal.md"), Line: 1, Source: "prompt", Message: models.StringPtr("Explain the decision"), Severity: models.SeverityError}
	})

	It("saves each batch immediately and prints only new findings", func() {
		result := &linters.LinterResult{Linter: "review", WorkDir: root, Status: linters.LinterStatusRunning, Violations: []models.Violation{finding, finding}}
		reporter.UpsertLintResult(result)
		live, err := snapshots.LoadRunFile(reporter.path)
		Expect(err).NotTo(HaveOccurred())
		Expect(live.Status.Running).To(BeTrue())
		Expect(live.Lint[0].Violations).To(HaveLen(1))
		Expect(strings.Count(output.String(), *finding.Message)).To(Equal(1))
		reporter.UpsertLintResult(result)
		finalOnly := finding
		finalOnly.Line, finalOnly.Message = 2, models.StringPtr("Document the alternative")
		result.Status, result.Success, result.Violations = "", true, []models.Violation{finding, finalOnly}
		final, err := reporter.Finish([]*linters.LinterResult{result}, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(final.Status.Running).To(BeFalse())
		Expect(strings.Count(output.String(), *finding.Message)).To(Equal(1))
		Expect(strings.Count(output.String(), *finalOnly.Message)).To(Equal(1))
		paths, err := filepath.Glob(filepath.Join(root, ".gavel", "run-*.json"))
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(Equal([]string{reporter.path}))
	})

	It("filters exclusions before writing snapshots or stderr", func() {
		excluded := finding
		excluded.Rule, excluded.Message = &models.Rule{Method: "excluded"}, models.StringPtr("Excluded finding")
		reporter.UpsertLintResult(&linters.LinterResult{Linter: "review", WorkDir: root, Violations: []models.Violation{excluded, finding}})
		snapshot, err := snapshots.LoadRunFile(reporter.path)
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshot.Lint[0].Violations).To(Equal([]models.Violation{finding}))
		Expect(output.String()).NotTo(ContainSubstring(*excluded.Message))
	})

	It("preserves findings when concurrent cumulative updates arrive out of order", func() {
		second := finding
		second.Line, second.Message = 2, models.StringPtr("Explain the alternative")
		result := &linters.LinterResult{Linter: "review", WorkDir: root, Status: linters.LinterStatusRunning, Violations: []models.Violation{finding, second}}
		reporter.UpsertLintResult(result)
		result.Violations = []models.Violation{finding}
		reporter.UpsertLintResult(result)
		snapshot, err := snapshots.LoadRunFile(reporter.path)
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshot.Lint[0].Violations).To(Equal([]models.Violation{finding, second}))
		Expect(strings.Count(output.String(), *second.Message)).To(Equal(1))
	})

	It("records nested linters and removes replaced placeholders", func() {
		reporter.UpsertLintResult(&linters.LinterResult{Linter: "review/lint", WorkDir: root, Status: linters.LinterStatusRunning})
		reporter.UpsertLintResult(&linters.LinterResult{Linter: "review/json", WorkDir: root, Success: true, Violations: []models.Violation{finding}})
		reporter.RemoveLintResult("review/lint", root)
		snapshot, err := snapshots.LoadRunFile(reporter.path)
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshot.Lint).To(HaveExactElements(HaveField("Linter", "review/json")))
		Expect(output.String()).To(SatisfyAll(ContainSubstring("review/json"), ContainSubstring(*finding.Message)))
	})

	It("does not repeat shared tool findings in final model variants", func() {
		reporter.UpsertLintResult(&linters.LinterResult{Linter: "review", WorkDir: root, Status: linters.LinterStatusRunning, Violations: []models.Violation{finding}})
		sonnet := &linters.LinterResult{Linter: "review:api:sonnet", WorkDir: root, Success: true, Violations: []models.Violation{finding}}
		haiku := &linters.LinterResult{Linter: "review:api:haiku", WorkDir: root, Success: true, Violations: []models.Violation{finding}}
		_, err := reporter.Finish([]*linters.LinterResult{sonnet, haiku}, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Count(output.String(), *finding.Message)).To(Equal(1))
	})

	It("retains accepted findings and the failure when a run is interrupted", func() {
		reporter.UpsertLintResult(&linters.LinterResult{Linter: "review", WorkDir: root, Status: linters.LinterStatusRunning, Violations: []models.Violation{finding}})
		final, err := reporter.Finish(nil, context.Canceled)
		Expect(err).NotTo(HaveOccurred())
		Expect(final.Status.Stopped).To(BeTrue())
		Expect(final.Error).To(ContainSubstring("context canceled"))
		Expect(final.Lint[0].Violations).To(Equal([]models.Violation{finding}))
		persisted, err := snapshots.LoadRunFile(reporter.path)
		Expect(err).NotTo(HaveOccurred())
		Expect(persisted).To(Equal(final))
	})

	It("cancels the run on stderr failure and preserves the accepted finding on disk", func() {
		reporter.options.Writer = promptReportFailure{err: io.ErrClosedPipe}
		reporter.UpsertLintResult(&linters.LinterResult{Linter: "review", WorkDir: root, Violations: []models.Violation{finding}})
		Expect(errors.Is(context.Cause(ctx), io.ErrClosedPipe)).To(BeTrue())
		_, err := reporter.Finish(nil, context.Cause(ctx))
		Expect(errors.Is(err, io.ErrClosedPipe)).To(BeTrue())
		snapshot, err := snapshots.LoadRunFile(reporter.path)
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshot.Status.Running).To(BeFalse())
		Expect(snapshot.Lint[0].Violations).To(Equal([]models.Violation{finding}))
	})

	It("surfaces a disk failure without corrupting the last valid snapshot", func() {
		reporter.UpsertLintResult(&linters.LinterResult{Linter: "review", WorkDir: root, Violations: []models.Violation{finding}})
		previous, err := os.ReadFile(reporter.path)
		Expect(err).NotTo(HaveOccurred())
		stateDir := filepath.Dir(reporter.path)
		Expect(os.Chmod(stateDir, 0o500)).To(Succeed())
		DeferCleanup(func() { Expect(os.Chmod(stateDir, 0o700)).To(Succeed()) })
		reporter.RemoveLintResult("review", root)
		Expect(context.Cause(ctx)).To(MatchError(ContainSubstring("remove prompt lint result")))
		_, err = reporter.Finish(nil, context.Cause(ctx))
		Expect(err).To(HaveOccurred())
		persisted, err := os.ReadFile(reporter.path)
		Expect(err).NotTo(HaveOccurred())
		Expect(persisted).To(Equal(previous))
	})
})
