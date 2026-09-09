package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/flanksource/captain/pkg/aiflags"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/ai/aifix"
	"github.com/flanksource/gavel/ai/prfix"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AI fix full spec resolution", func() {
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
	})

	It("preserves the operation's complete native request fields", func() {
		dir := GinkgoT().TempDir()
		operation := api.Spec{
			Model: api.Model{Name: "agent:sonnet"},
			Prompt: api.Prompt{
				User: "Repair the reported failures.", AppendSystem: "Keep the change scoped.",
				SchemaJSON: json.RawMessage(`{"type":"object"}`), SchemaStrictness: api.SchemaStrictnessRetry,
			},
			Permissions:     api.Permissions{Mode: api.PermissionAcceptEdits, Tools: api.Tools{"Bash": api.ToolPolicyDeny}},
			Memory:          api.Memory{SkipHooks: true, SkipMemory: true},
			ToolPreferences: api.ToolPreferences{"review": api.ToolPolicyDeny},
			Setup:           &shell.Setup{Cwd: dir, DotEnv: []string{".env.test"}},
			Sandbox:         &api.SandboxRef{Mode: api.SandboxNative},
			SessionID:       "prior-review-session",
			CLIArgs:         map[string]any{"verbose": true},
		}
		resolved, err := buildAIFixRequest(aiFixRequestOptions{Layers: []api.SpecLayer{api.PromptSpecLayer("operation", operation)}, Dir: dir})
		Expect(err).NotTo(HaveOccurred())
		Expect(api.Spec(resolved.Request)).To(SatisfyAll(
			HaveField("Prompt", operation.Prompt), HaveField("Permissions.Tools", operation.Permissions.Tools),
			HaveField("Memory", operation.Memory), HaveField("ToolPreferences", operation.ToolPreferences),
			HaveField("Setup", &shell.Setup{Cwd: dir, BaseDir: filepath.Join(dir, ".shell"), DotEnv: []string{filepath.Join(dir, ".env.test")}}), HaveField("Sandbox", operation.Sandbox),
			HaveField("SessionID", operation.SessionID), HaveField("CLIArgs", operation.CLIArgs),
		))
		Expect(operation.Setup.DotEnv).To(Equal([]string{".env.test"}))
	})

	It("keeps the PR repair project budget above the saved Captain budget", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(os.Getenv("HOME"), ".captain.yaml"), []byte("ai:\n  budgetUSD: 1\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("ai:\n  budget:\n    cost: 2\n"), 0o600)).To(Succeed())
		config, err := verify.LoadGavelConfig(dir)
		Expect(err).NotTo(HaveOccurred())
		layers, err := prfix.Layers(prfix.ResolveOptions{Base: config.AI, Prompt: config.PR.Fix, Dir: dir, PR: prfix.PRContext{Number: 7, Title: "Repair checks", StatusText: "Lint failed"}})
		Expect(err).NotTo(HaveOccurred())
		saved, _, err := captainconfig.Load()
		Expect(err).NotTo(HaveOccurred())
		resolved, err := buildAIFixRequest(aiFixRequestOptions{Layers: layers, Saved: saved, Dir: dir})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Config.Budget.Cost).To(Equal(2.0))
		Expect(resolved.Request.Budget.Cost).To(Equal(2.0))
	})

	It("resolves CLI runtime repairs after authored capability policy", func() {
		layers := []api.SpecLayer{api.PromptSpecLayer("operation", api.Spec{
			Model: api.Model{Name: "gpt-5.6-sol", Mode: api.ModeAgent},
			// A deny, not an allow: captain treats an allow naming another agent's
			// built-in as inert, so only a deny still makes the openai agent runtime
			// refuse the policy — which is the refusal this spec turns on.
			Permissions: api.Permissions{Tools: api.Tools{"Read": api.ToolPolicyDeny}},
			Prompt:      api.Prompt{User: "Repair the lint failures"},
		})}
		options := aiFixRequestOptions{Layers: layers}
		_, err := buildAIFixRequest(options)
		Expect(err).To(HaveOccurred())
		options.Runtime = captaincli.AIRuntimeOptions{AIProviderOptions: captaincli.AIProviderOptions{ModelFlags: aiflags.ModelFlags{Model: "agent:sonnet"}}}
		resolved, err := buildAIFixRequest(options)
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Request.Name).To(Equal("claude-sonnet-5"))
		Expect(resolved.Request.Permissions.Tools).To(Equal(api.Tools{"Read": api.ToolPolicyDeny}))
	})

	It("refuses a tool policy the selected runtime cannot enforce", func() {
		_, err := buildAIFixRequest(aiFixRequestOptions{Layers: []api.SpecLayer{api.PromptSpecLayer("operation", api.Spec{
			Model: api.Model{Name: "agent:sonnet"}, Permissions: api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyAsk}},
		})}})
		Expect(err).To(MatchError(SatisfyAll(ContainSubstring("Bash"), ContainSubstring("not enforceable"))))
	})

	It("renders later lint results with the captured saved snapshot", func() {
		dir := GinkgoT().TempDir()
		message := "first failure"
		results := []*linters.LinterResult{{Linter: "example", Violations: []models.Violation{{File: "example.go", Message: &message}}}}
		runtime := lintFixRuntime{
			Prompt: aifix.ResolveOptions{Dir: dir, Prompt: verify.PromptSpec{Spec: api.Spec{Model: api.Model{Name: "agent:sonnet"}, Budget: api.Budget{Cost: 2}}}},
			Saved:  captainconfig.Config{AI: captainconfig.AIDefaults{BudgetUSD: 1, NoMemory: true}},
		}
		initial, err := runtime.Resolve(results)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(os.Getenv("HOME"), ".captain.yaml"), []byte("ai: [changed after capture]\n"), 0o600)).To(Succeed())
		message = "remaining failure"
		next, err := runtime.BuildRequest(results)
		Expect(err).NotTo(HaveOccurred())
		Expect(next.Prompt.User).To(ContainSubstring("remaining failure"))
		Expect(next.Prompt.User).NotTo(ContainSubstring("first failure"))
		Expect(next.Model).To(Equal(initial.Request.Model))
		Expect(next.Budget).To(Equal(initial.Request.Budget))
		Expect(next.Memory.SkipMemory).To(BeTrue())
		Expect(next.Budget.Cost).To(Equal(float64(2)))
	})
})
