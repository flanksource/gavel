package git

import (
	"github.com/flanksource/captain/pkg/aimock"
	"github.com/flanksource/captain/pkg/aimock/anthropicmock"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Anthropic buffered commit analysis", func() {
	It("generates a structured commit message at high effort using internal streaming", func(ctx SpecContext) {
		scenario, err := aimock.Parse([]byte(`anthropic:
  - respond:
      text: '{"type":"fix","scope":"ai","subject":"stream buffered Anthropic requests"}'
      usage: {input: 14, output: 8}
`))
		Expect(err).NotTo(HaveOccurred())
		server, err := anthropicmock.Start(anthropicmock.Options{Scenario: scenario})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(server.Close)
		analysis, err := AnalyzeWithAI(ctx, sampleCommit(), nil, AnalyzeOptions{
			Prompt: verify.PromptSpec{Spec: api.Spec{
				Model:  api.Model{Name: "api:claude-sonnet-5-5:high", NoCache: true},
				Budget: api.Budget{MaxTokens: 4096},
			}},
			AgentFactory: func(config ai.AgentConfig) (ai.Agent, error) {
				Expect(config.Model.Effort).To(Equal(api.EffortHigh))
				config.APIKey = aimock.DummyKey
				config.APIURL = server.APIURL()
				return ai.NewAgent(config)
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(analysis.CommitType).To(Equal(models.CommitTypeFix))
		Expect(analysis.Scope).To(Equal(models.ScopeType("ai")))
		Expect(analysis.Subject).To(Equal("stream buffered Anthropic requests"))
		Expect(analysis.Trailers).To(HaveKeyWithValue("AI-Analyzed", "true"))
		Expect(server.Remaining()).To(BeEmpty())
		Eventually(server.Requests).Should(ContainElement(And(HaveField("Path", "/v1/messages"), HaveField("Stream", true), HaveField("Miss", ""))))
	})
})
