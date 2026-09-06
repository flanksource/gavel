package types

import (
	"encoding/json"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/gavel/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AI fixture final runtime resolution", func() {
	It("canonicalizes the final flat selector before provider construction", func() {
		var schema checklistResponse
		runtime, config, err := resolveAIStepSpec(checklistFixture(&fixtures.FixtureAIConfig{Model: "api:gpt-5:low"}), fixtures.RunOptions{}, &schema)
		Expect(err).NotTo(HaveOccurred())
		resolved := runtime.Spec
		Expect(resolved.Name).To(Equal("gpt-5"))
		Expect(resolved.Mode).To(Equal(api.ModeAPI))
		Expect(resolved.Effort).To(Equal(api.EffortLow))
		Expect(resolved.Provider).To(BeIdenticalTo(api.OpenAI))
		Expect(config.Model).To(Equal(resolved.Model))
	})

	It("fills saved defaults for the final fixture provider and preserves field ownership", func() {
		saved := captainconfig.AIDefaults{DefaultModel: "agent:claude-sonnet-4-6", BudgetUSD: 2, MaxTokens: 700, NoHooks: true,
			Providers: map[string]captainconfig.ProviderDefaults{
				"anthropic": {Mode: "agent", ReasoningEffort: "high"}, "openai": {Mode: "api", ReasoningEffort: "low"},
			}}
		layers := make([]api.SpecLayer, 1, 4)
		layers[0] = api.PromptSpecLayer("verify.profile", api.Spec{Model: api.Model{Name: "claude-sonnet-4-6"},
			Prompt: api.Prompt{System: "Review the exact change."}, Memory: api.Memory{SkipUser: true}})
		options := api.ResolveSpecOptions{Layers: layers, Saved: &saved}
		var schema checklistResponse
		resolved, config, err := resolveAIStepSpec(checklistFixture(&fixtures.FixtureAIConfig{Model: "gpt-5"}), fixtures.RunOptions{Runtime: &options}, &schema)
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Spec.Name).To(Equal("gpt-5"))
		Expect(resolved.Spec.Mode).To(Equal(api.ModeAPI))
		Expect(resolved.Spec.Effort).To(Equal(api.EffortLow))
		Expect(resolved.Spec.Budget).To(Equal(api.Budget{Cost: 2, MaxTokens: 700}))
		Expect(resolved.Spec.Memory).To(Equal(api.Memory{SkipUser: true, SkipHooks: true}))
		Expect(resolved.Spec.Prompt.System).To(Equal("Review the exact change."))
		Expect(resolved.Provenance["/model"].Source.Name).To(Equal("fixture.ai"))
		Expect(resolved.Provenance["/mode"].Source.Key).To(Equal("ai.providers.openai.mode"))
		Expect(resolved.Provenance["/memory/skipHooks"].Source.Key).To(Equal("ai.noHooks"))
		Expect(config.Model).To(Equal(resolved.Spec.Model))
		Expect(config.Budget).To(Equal(resolved.Spec.Budget))
		Expect(layers[:cap(layers)][1:]).To(Equal(make([]api.SpecLayer, 3)), "appending prompt layers must not overwrite the caller's backing array")
		Expect(layers[0].Spec.Mode).To(BeEmpty())
		Expect(saved.DefaultModel).To(Equal("agent:claude-sonnet-4-6"))
	})

	It("lets flat zero and false defeat saved values at final resolution", func() {
		var config fixtures.FixtureAIConfig
		Expect(json.Unmarshal([]byte(`{"model":"api:gpt-5","temperature":0,"noCache":false}`), &config)).To(Succeed())
		options := api.ResolveSpecOptions{Saved: &captainconfig.AIDefaults{Temperature: 0.7, NoCache: true}}
		resolved, agent, err := resolveAIStepSpec(checklistFixture(&config), fixtures.RunOptions{Runtime: &options}, &checklistResponse{})
		Expect(err).NotTo(HaveOccurred())
		zero := 0.0
		Expect(resolved.Spec.Temperature).To(Equal(&zero))
		Expect(resolved.Spec.NoCache).To(BeFalse())
		Expect(agent.NoCache).To(BeFalse())
		Expect(resolved.Provenance["/temperature"].Source.Name).To(Equal("fixture.ai"))
		Expect(resolved.Provenance["/noCache"].Source.Name).To(Equal("fixture.ai"))
	})

	DescribeTable("ignores mutable runtime defaults for an authoritative snapshot", func(embedded bool) {
		snapshot := api.Spec{Model: api.Model{Name: "gpt-5", Mode: api.ModeAPI}, Budget: api.Budget{MaxTurns: 3}}
		options := fixtures.RunOptions{Runtime: &api.ResolveSpecOptions{Saved: &captainconfig.AIDefaults{DefaultModel: "bad:selector", NoCache: true, NoHooks: true}}}
		fixture := checklistFixture(nil)
		if embedded {
			fixture.AI = &fixtures.FixtureAIConfig{Spec: &snapshot}
		} else {
			options.Spec = &snapshot
		}
		resolved, _, err := resolveAIStepSpec(fixture, options, &checklistResponse{})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Spec.Name).To(Equal("gpt-5"))
		Expect(resolved.Spec.Budget).To(Equal(api.Budget{MaxTurns: 3}))
		Expect(resolved.Spec.NoCache).To(BeFalse())
		Expect(resolved.Spec.Memory).To(BeZero())
		Expect(resolved.Warnings).To(BeEmpty())
	}, Entry("serialized ai.spec", true), Entry("explicit runner Spec", false))

	It("checks the final fixture override against the retained profile constraint", func() {
		options := api.ResolveSpecOptions{Layers: []api.SpecLayer{{Name: "review.catalog", Scope: api.SpecLayerGlobal,
			Constraints: api.RuntimeConstraints{Models: []string{"claude-sonnet-4-6"}},
		}}}
		_, _, err := resolveAIStepSpec(checklistFixture(&fixtures.FixtureAIConfig{Model: "api:gpt-5"}), fixtures.RunOptions{Runtime: &options}, &checklistResponse{})
		Expect(err).To(MatchError(ContainSubstring("outside the effective model catalog")))
	})

	It("rejects a snapshot whose fallback cannot honor its sandbox", func() {
		snapshot := api.Spec{Model: api.Model{Name: "claude-sonnet-4-6", Mode: api.ModeCLI,
			Fallbacks: []api.Model{{Name: "gpt-5", Mode: api.ModeAPI}}}, Sandbox: &api.SandboxRef{Mode: api.SandboxNative}}
		_, _, err := resolveAIStepSpec(checklistFixture(&fixtures.FixtureAIConfig{Spec: &snapshot}), fixtures.RunOptions{}, &checklistResponse{})
		Expect(err).To(MatchError(ContainSubstring(`sandbox mode "native" is not available for openai api`)))
	})
})
