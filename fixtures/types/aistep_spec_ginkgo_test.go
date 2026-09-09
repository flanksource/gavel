package types

import (
	"time"

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
	It("applies flat fixture options over a canonical runtime snapshot", func() {
		canonical := api.Spec{
			Model: api.Model{Name: "claude-sonnet-4-6", Mode: api.ModeCLI, Effort: api.EffortHigh,
				Provider: api.Anthropic, Fallbacks: []api.Model{{Name: "gpt-5", Mode: api.ModeAPI}}},
			Budget:      api.Budget{Cost: 3, MaxTokens: 1000, Timeout: "2m"},
			Permissions: api.Permissions{Mode: api.PermissionDontAsk},
			Workflow:    &api.Workflow{Verify: &api.Verify{Fixture: "nested"}, Commits: []api.Commit{{On: "run"}}},
			SessionID:   "prior", ToolApproval: &api.ToolApprovalResume{}, Messages: []api.Message{{Role: api.RoleUser}},
		}
		temperature, noCache := 0.4, true
		fixture := checklistFixture(&fixtures.FixtureAIConfig{
			Spec: &canonical, Model: "api:gpt-5:low", Temperature: &temperature, MaxTokens: 500,
			MaxConcurrent: 2, CacheTTL: time.Minute, NoCache: &noCache,
		})
		var schema checklistResponse
		runtime, config, err := resolveAIStepSpec(fixture, fixtures.RunOptions{}, &schema)
		Expect(err).NotTo(HaveOccurred())
		resolved := runtime.Spec
		model, err := resolved.Model.Expand()
		Expect(err).NotTo(HaveOccurred())
		Expect(model.Name).To(Equal("gpt-5"))
		Expect(model.Mode).To(Equal(api.ModeAPI))
		Expect(model.Effort).To(Equal(api.EffortLow))
		Expect(resolved.Temperature).To(Equal(&temperature))
		Expect(resolved.NoCache).To(BeTrue())
		Expect(resolved.Fallbacks).To(HaveLen(1))
		Expect(resolved.Fallbacks[0].Name).To(Equal("gpt-5"))
		Expect(resolved.Fallbacks[0].Mode).To(Equal(api.ModeAPI))
		Expect(resolved.Budget).To(Equal(api.Budget{Cost: 3, MaxTokens: 500, Timeout: "2m"}))
		Expect(resolved.Permissions).To(Equal(canonical.Permissions))
		Expect(resolved.SessionID).To(BeEmpty())
		Expect(resolved.ToolApproval).To(BeNil())
		Expect(resolved.Messages).To(BeEmpty())
		Expect(resolved.Workflow).To(BeNil())
		Expect(config.NoCache).To(BeTrue())
		Expect(config.MaxConcurrent).To(Equal(2))
		Expect(config.CacheTTL).To(Equal(time.Minute))
		Expect(canonical.Model.Provider).To(BeIdenticalTo(api.Anthropic))
		Expect(canonical.Workflow.Verify).NotTo(BeNil())
	})

	It("lets the fixture's ai: front matter override the caller's spec", func() {
		override := api.Spec{
			Model:  api.Model{Name: "claude-sonnet-4-6", Mode: api.ModeAPI},
			Prompt: api.Prompt{User: "must not replace the fixture prompt", System: "verification system"},
			Budget: api.Budget{MaxTokens: 200, Cost: 3},
		}
		var schema checklistResponse

		runtime, _, err := resolveAIStepSpec(
			checklistFixture(&fixtures.FixtureAIConfig{Model: "gpt-5", MaxTokens: 100}),
			fixtures.RunOptions{Spec: &override}, &schema)
		Expect(err).NotTo(HaveOccurred())
		resolved := runtime.Spec
		Expect(resolved.Model.Name).To(Equal("gpt-5"))
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
			Model:     api.Model{Name: "gpt-5", Mode: api.ModeAPI},
			SessionID: "the-implementer-session",
		}
		var schema checklistResponse

		resolved, config, err := resolveAIStepSpec(checklistFixture(nil), fixtures.RunOptions{Spec: &override}, &schema)
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Spec.SessionID).To(BeEmpty())
		Expect(config.SessionID).To(BeEmpty())
	})

	// No model is defaulted here. Which model grades a definition of done is a
	// configuration decision resolved by the caller (request > .gavel.yaml
	// todos.verify > ai:), so an unset one stays unset and captain says so.
	It("invents no model when neither the caller nor the fixture names one", func() {
		var schema checklistResponse

		resolved, config, err := resolveAIStepSpec(checklistFixture(nil), fixtures.RunOptions{}, &schema)
		Expect(err).To(MatchError(ContainSubstring("configure")))
		Expect(resolved.Spec.Model.Name).To(BeEmpty())
		Expect(config.Model.Name).To(BeEmpty())
	})
})
