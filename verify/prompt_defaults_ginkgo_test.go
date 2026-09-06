package verify

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("authored prompt runtime layers", func() {
	It("fills saved defaults after project, prompt, operation and request layers", func() {
		resolved, err := (PromptSpec{Spec: api.Spec{Budget: api.Budget{Cost: 3}}}).Resolve(PromptResolveOptions{
			Base:          api.Spec{Budget: api.Budget{Cost: 2}},
			DefaultPrompt: "---\nbudget:\n  cost: 2.5\n---\nReview changes",
			Request:       api.Spec{Budget: api.Budget{Cost: 4}},
			Saved:         &captainconfig.AIDefaults{DefaultModel: "api:claude-haiku-4-5", BudgetUSD: 1},
			RequireModel:  true,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Spec.Budget.Cost).To(Equal(float64(4)))
		Expect(resolved.Spec.Name).To(Equal("claude-haiku-4-5"))
		Expect(resolved.Spec.Mode).To(Equal(api.ModeAPI))
	})

	It("retains dotprompt model options without deriving a mode", func() {
		source := "---\nmodel: sonnet\nconfig:\n  temperature: 0\n  reasoning: high\n---\nReview {{name}}"
		result, err := RenderPromptSpec(source, map[string]any{"name": "changes"}, PromptSpecOptions{Declared: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Spec.Name).To(Equal("sonnet"))
		Expect(result.Spec.Mode).To(BeEmpty())
		Expect(result.Spec.Temperature).To(Equal(new(float64)))
		Expect(result.Spec.Effort).To(Equal(api.EffortHigh))
		Expect(result.Spec.Prompt.User).To(Equal("Review changes"))
	})
})
