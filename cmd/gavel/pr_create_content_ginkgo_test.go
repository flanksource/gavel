package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/captain/pkg/aiflags"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PR content configuration snapshot", func() {
	It("carries the worktree full specification and captured saved settings to generation", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("ai:\n  model: api:sonnet\n  memory: {skipHooks: true}\npr:\n  content:\n    file: review.prompt\n    budget: {cost: 2, maxTurns: 3}\n    setup: {envVars: [{name: REVIEW_SCOPE, value: pull-request}]}\n"), 0o600)).To(Succeed())
		saved := &captainconfig.Config{AI: captainconfig.AIDefaults{MaxTokens: 1700}}
		flags := aiflags.ModelFlags{Model: "api:gpt-5", Temperature: "0"}
		got, err := loadPRContentOptions(dir, prCreateOptions{Flags: flags, Saved: saved})
		Expect(err).NotTo(HaveOccurred())
		Expect(got.WorkDir).To(Equal(dir))
		Expect(got.Saved).To(BeIdenticalTo(saved))
		Expect(got.Flags).To(Equal(flags))
		Expect(got.AI.Memory.SkipHooks).To(BeTrue())
		Expect(got.PR.Content.File).To(Equal("review.prompt"))
		Expect(got.PR.Content.Spec.Budget).To(Equal(api.Budget{Cost: 2, MaxTurns: 3}))
		Expect(got.PR.Content.Spec.Setup.EnvVars).To(HaveLen(1))
		Expect(got.PR.Content.Spec.Setup.EnvVars[0].Name).To(Equal("REVIEW_SCOPE"))
		Expect(got.PR.Content.Spec.Setup.EnvVars[0].ValueStatic).To(Equal("pull-request"))
	})

	It("keeps the status prompt in the same commit invocation snapshot", func() {
		cfg := verify.GavelConfig{Status: verify.StatusConfig{Summary: verify.PromptSpec{File: "summary.prompt"}}}
		got := buildCommitOptions(CommitOptions{Summary: true}, "/work/review", cfg, nil)
		Expect(got.Status).To(Equal(cfg.Status))
	})

	It("reports malformed commit configuration before entering the commit pipeline", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		dir := GinkgoT().TempDir()
		output, err := exec.Command("git", "init", dir).CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))
		Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("commit: [invalid"), 0o600)).To(Succeed())
		_, err = runCommit(CommitOptions{WorkDir: dir, Message: "fix: explicit"})
		Expect(err).To(MatchError(ContainSubstring("load commit configuration")))
		Expect(err.Error()).To(ContainSubstring("yaml"))
	})
})
