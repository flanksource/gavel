package main

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
)

var _ = Describe("Git CLI runtime preparation", func() {
	It("captures saved settings once and keeps explicit flag zero values", func() {
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		path := filepath.Join(home, ".captain.yaml")
		Expect(os.WriteFile(path, []byte("ai:\n  defaultModel: api:gpt-4o\n  budgetUSD: 1\n  maxTokens: 6144\n  noCache: true\n"), 0o600)).To(Succeed())
		flags := ai.DefaultConfig()
		flagSet := pflag.NewFlagSet("git", pflag.ContinueOnError)
		ai.BindFlags(flagSet, &flags)
		Expect(flagSet.Parse([]string{"--ai-model=api:haiku", "--ai-max-tokens=0", "--ai-no-cache=false"})).To(Succeed())
		config := verify.DefaultGavelConfig()
		config.AI.Budget.Cost = 2
		config.Commit.Message.Spec.Memory.SkipMemory = true
		options := git.AnalyzeOptions{HistoryOptions: git.HistoryOptions{Path: GinkgoT().TempDir()}}
		Expect(configureGitAI(&options, gitAIOptions{Config: config, Flags: flags, FlagSet: flagSet})).To(Succeed())
		Expect(os.WriteFile(path, []byte("ai: [changed after capture]\n"), 0o600)).To(Succeed())
		prepared, err := git.PrepareCommitMessage(models.CommitAnalysis{}, options)
		Expect(err).NotTo(HaveOccurred())
		Expect(prepared.Request).To(SatisfyAll(
			HaveField("Model.Name", "claude-haiku-4-5"), HaveField("Model.NoCache", false),
			HaveField("Budget.Cost", float64(2)), HaveField("Budget.MaxTokens", 0),
			HaveField("Memory.SkipMemory", true),
		))
		Expect(prepared.Config.Model).To(Equal(prepared.Request.Model))
		Expect(prepared.Config.Budget).To(Equal(prepared.Request.Budget))
		Expect(prepared.Config.NoCache).To(BeFalse())
	})
})
