package commit

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
)

func boolPtr(b bool) *bool { return &b }

var _ = Describe("resolveLintGates", func() {
	It("defaults to secrets-on, full-lint-off when nothing is configured", func() {
		gates, err := resolveLintGates("", "", verify.CommitLintConfig{})
		Expect(err).ToNot(HaveOccurred())
		Expect(gates).To(Equal(LintGates{FullLint: false, Secrets: true}))
	})

	It("honors .gavel.yaml commit.lint.enabled when set", func() {
		gates, err := resolveLintGates("", "", verify.CommitLintConfig{Enabled: boolPtr(true)})
		Expect(err).ToNot(HaveOccurred())
		Expect(gates.FullLint).To(BeTrue())
		Expect(gates.Secrets).To(BeTrue(), "secrets default stays on")
	})

	It("honors .gavel.yaml commit.lint.secrets=false to disable secrets", func() {
		gates, err := resolveLintGates("", "", verify.CommitLintConfig{Secrets: boolPtr(false)})
		Expect(err).ToNot(HaveOccurred())
		Expect(gates.Secrets).To(BeFalse())
	})

	It("--lint=true overrides .gavel.yaml commit.lint.enabled=false", func() {
		gates, err := resolveLintGates("true", "", verify.CommitLintConfig{Enabled: boolPtr(false)})
		Expect(err).ToNot(HaveOccurred())
		Expect(gates.FullLint).To(BeTrue())
	})

	It("--lint-secrets=false overrides default-on for secrets", func() {
		gates, err := resolveLintGates("", "false", verify.CommitLintConfig{})
		Expect(err).ToNot(HaveOccurred())
		Expect(gates.Secrets).To(BeFalse())
	})

	It("flags compose independently — --lint=false --lint-secrets=true runs only secrets", func() {
		gates, err := resolveLintGates("false", "true", verify.CommitLintConfig{
			Enabled: boolPtr(true),
			Secrets: boolPtr(false),
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(gates.FullLint).To(BeFalse())
		Expect(gates.Secrets).To(BeTrue())
	})

	It("rejects malformed flag values with a hard error", func() {
		_, err := resolveLintGates("maybe", "", verify.CommitLintConfig{})
		Expect(err).To(MatchError(ContainSubstring("invalid --lint value")))
		_, err = resolveLintGates("", "maybe", verify.CommitLintConfig{})
		Expect(err).To(MatchError(ContainSubstring("invalid --lint-secrets value")))
	})
})

var _ = Describe("applyLintGate", func() {
	var (
		previousRunner LintRunner
		captured       struct {
			workDir string
			linters []string
			files   []string
			called  bool
		}
	)

	BeforeEach(func() {
		previousRunner = lintRunnerImpl
		captured.called = false
		captured.workDir = ""
		captured.linters = nil
		captured.files = nil
	})

	AfterEach(func() {
		lintRunnerImpl = previousRunner
	})

	withFakeRunner := func(results []*linters.LinterResult, runErr error) {
		SetLintRunner(func(_ context.Context, workDir string, names, files []string) ([]*linters.LinterResult, error) {
			captured.called = true
			captured.workDir = workDir
			captured.linters = append([]string(nil), names...)
			captured.files = append([]string(nil), files...)
			return results, runErr
		})
	}

	It("is a no-op when both gates are off", func() {
		withFakeRunner(nil, nil)
		out, err := applyLintGate(context.Background(), "/repo", []string{"a.go"}, LintGates{})
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(BeNil())
		Expect(captured.called).To(BeFalse())
	})

	It("is a no-op when no files are staged even if gates are on", func() {
		withFakeRunner(nil, nil)
		_, err := applyLintGate(context.Background(), "/repo", nil, LintGates{Secrets: true})
		Expect(err).ToNot(HaveOccurred())
		Expect(captured.called).To(BeFalse())
	})

	It("requests only betterleaks when only Secrets is on", func() {
		withFakeRunner([]*linters.LinterResult{{Linter: "betterleaks"}}, nil)
		_, err := applyLintGate(context.Background(), "/repo", []string{"a.go"}, LintGates{Secrets: true})
		Expect(err).ToNot(HaveOccurred())
		Expect(captured.linters).To(Equal([]string{"betterleaks"}))
		Expect(captured.files).To(Equal([]string{"a.go"}))
	})

	It("requests every linter except betterleaks when only FullLint is on", func() {
		withFakeRunner([]*linters.LinterResult{{Linter: "golangci-lint"}}, nil)
		_, err := applyLintGate(context.Background(), "/repo", []string{"a.go"}, LintGates{FullLint: true})
		Expect(err).ToNot(HaveOccurred())
		Expect(captured.linters).To(Equal([]string{"!betterleaks"}))
	})

	It("requests every linter (nil) when both gates are on", func() {
		withFakeRunner([]*linters.LinterResult{{Linter: "betterleaks"}, {Linter: "golangci-lint"}}, nil)
		_, err := applyLintGate(context.Background(), "/repo", []string{"a.go"}, LintGates{FullLint: true, Secrets: true})
		Expect(err).ToNot(HaveOccurred())
		Expect(captured.linters).To(BeNil())
	})

	It("returns ErrLintFindings when violations exist", func() {
		msg := "secret detected"
		withFakeRunner([]*linters.LinterResult{
			{
				Linter:     "betterleaks",
				Violations: []models.Violation{{File: "config.yaml", Line: 3, Source: "betterleaks", Message: &msg}},
			},
		}, nil)
		out, err := applyLintGate(context.Background(), "/repo", []string{"config.yaml"}, LintGates{Secrets: true})
		Expect(err).To(MatchError(ErrLintFindings))
		Expect(out).ToNot(BeNil())
		Expect(out.Violations).To(Equal(1))
	})

	It("returns nil error and a clean LintGateResult when zero violations are found", func() {
		withFakeRunner([]*linters.LinterResult{{Linter: "betterleaks"}}, nil)
		out, err := applyLintGate(context.Background(), "/repo", []string{"a.go"}, LintGates{Secrets: true})
		Expect(err).ToNot(HaveOccurred())
		Expect(out).ToNot(BeNil())
		Expect(out.Violations).To(BeZero())
	})

	It("ignores skipped linter results when counting violations", func() {
		msg := "should not count"
		withFakeRunner([]*linters.LinterResult{
			{
				Linter:     "betterleaks",
				Skipped:    true,
				Violations: []models.Violation{{File: "x", Message: &msg}},
			},
		}, nil)
		out, err := applyLintGate(context.Background(), "/repo", []string{"x"}, LintGates{Secrets: true})
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Violations).To(BeZero())
	})

	It("propagates runner errors", func() {
		runErr := errors.New("boom")
		withFakeRunner(nil, runErr)
		_, err := applyLintGate(context.Background(), "/repo", []string{"a.go"}, LintGates{Secrets: true})
		Expect(err).To(MatchError(ContainSubstring("boom")))
	})
})

var _ = Describe("verify.MergeCommitConfig (lint)", func() {
	It("override commit.lint.enabled wins when set, base preserved otherwise", func() {
		base := verify.CommitConfig{Lint: verify.CommitLintConfig{Enabled: boolPtr(true)}}
		override := verify.CommitConfig{Lint: verify.CommitLintConfig{Enabled: boolPtr(false)}}
		merged := verify.MergeCommitConfig(base, override)
		Expect(merged.Lint.Enabled).ToNot(BeNil())
		Expect(*merged.Lint.Enabled).To(BeFalse())

		merged = verify.MergeCommitConfig(base, verify.CommitConfig{})
		Expect(merged.Lint.Enabled).ToNot(BeNil())
		Expect(*merged.Lint.Enabled).To(BeTrue(), "empty override must not clobber base")
	})

	It("override commit.lint.secrets wins when set, base preserved otherwise", func() {
		base := verify.CommitConfig{Lint: verify.CommitLintConfig{Secrets: boolPtr(true)}}
		override := verify.CommitConfig{Lint: verify.CommitLintConfig{Secrets: boolPtr(false)}}
		merged := verify.MergeCommitConfig(base, override)
		Expect(*merged.Lint.Secrets).To(BeFalse())

		merged = verify.MergeCommitConfig(base, verify.CommitConfig{})
		Expect(*merged.Lint.Secrets).To(BeTrue())
	})
})
