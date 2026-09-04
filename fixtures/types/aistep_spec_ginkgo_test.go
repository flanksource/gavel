package types

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func checklistFixture(ai *fixtures.FixtureAIConfig) fixtures.FixtureTest {
	return fixtures.FixtureTest{
		Name:        "acceptance criteria",
		FrontMatter: fixtures.FrontMatter{AI: ai},
		AIStep: &fixtures.AIStepSpec{
			Description: "Inspect the implementation.",
			Criteria:    []fixtures.ChecklistItem{{Text: "The behavior is covered."}},
		},
	}
}

var _ = Describe("AI fixture runtime spec", func() {
	// The caller's spec says how to run; the fixture — the most specific
	// statement about how it wants to be graded — overrides what it names. A todo
	// run's generated fixture names nothing, so the resolved verify spec stands.
	It("lets the fixture's ai: front matter override the caller's spec", func() {
		override := api.Spec{
			Model:  api.Model{Name: "caller-model"},
			Prompt: api.Prompt{User: "must not replace the fixture prompt", System: "verification system"},
			Budget: api.Budget{MaxTokens: 200, Cost: 3},
		}
		var schema checklistResponse

		resolved, _ := resolveAIStepSpec(
			checklistFixture(&fixtures.FixtureAIConfig{Model: "fixture-model", MaxTokens: 100}),
			fixtures.RunOptions{Spec: &override}, &schema)

		Expect(resolved.Model.Name).To(Equal("fixture-model"))
		Expect(resolved.Budget.MaxTokens).To(Equal(100))
		Expect(resolved.Budget.Cost).To(Equal(3.0), "and overrides only what it names")
		Expect(resolved.Prompt.System).To(Equal("verification system"))
		Expect(resolved.Prompt.User).To(ContainSubstring("The behavior is covered."))
		Expect(resolved.Prompt.User).NotTo(ContainSubstring("must not replace"))
		Expect(resolved.Prompt.Source).To(Equal("fixtures.ai-step"))
		Expect(resolved.Prompt.Schema).To(BeIdenticalTo(&schema))
	})

	// C1: the grader used to be built with the implementer's session id, so it
	// resumed into the very session it was judging — the candidate marking its own
	// exam. It inherits how to run, never what was already said.
	It("never resumes the session it is grading", func() {
		override := api.Spec{
			Model:     api.Model{Name: "caller-model"},
			SessionID: "the-implementer-session",
		}
		var schema checklistResponse

		resolved, config := resolveAIStepSpec(checklistFixture(nil), fixtures.RunOptions{Spec: &override}, &schema)

		Expect(resolved.SessionID).To(BeEmpty())
		Expect(config.SessionID).To(BeEmpty())
	})

	// No model is defaulted here. Which model grades a definition of done is a
	// configuration decision resolved by the caller (request > .gavel.yaml
	// todos.verify > ai:), so an unset one stays unset and captain says so.
	It("invents no model when neither the caller nor the fixture names one", func() {
		var schema checklistResponse

		resolved, config := resolveAIStepSpec(checklistFixture(nil), fixtures.RunOptions{}, &schema)

		Expect(resolved.Model.Name).To(BeEmpty())
		Expect(config.Model.Name).To(BeEmpty())
	})
})
