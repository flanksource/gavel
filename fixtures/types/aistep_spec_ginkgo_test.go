package types

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AI fixture runtime spec", func() {
	It("layers request options over fixture defaults while retaining the fixture prompt and schema", func() {
		fixture := fixtures.FixtureTest{
			Name: "acceptance criteria",
			FrontMatter: fixtures.FrontMatter{
				AI: &fixtures.FixtureAIConfig{
					Model:     "fixture-model",
					MaxTokens: 100,
				},
			},
			AIStep: &fixtures.AIStepSpec{
				Description: "Inspect the implementation.",
				Criteria:    []fixtures.ChecklistItem{{Text: "The behavior is covered."}},
			},
		}
		override := api.Spec{
			Model:  api.Model{Name: "request-model"},
			Prompt: api.Prompt{User: "must not replace the fixture prompt", System: "verification system"},
			Budget: api.Budget{MaxTokens: 200},
		}
		var schema checklistResponse

		resolved := resolveAIStepSpec(fixture, fixtures.RunOptions{Spec: &override}, &schema)

		Expect(resolved.Model.Name).To(Equal("request-model"))
		Expect(resolved.Budget.MaxTokens).To(Equal(200))
		Expect(resolved.Prompt.System).To(Equal("verification system"))
		Expect(resolved.Prompt.User).To(ContainSubstring("The behavior is covered."))
		Expect(resolved.Prompt.User).NotTo(ContainSubstring("must not replace"))
		Expect(resolved.Prompt.Source).To(Equal("fixtures.ai-step"))
		Expect(resolved.Prompt.Schema).To(BeIdenticalTo(&schema))
	})
})
