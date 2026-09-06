package git

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("prepared commit prompt runtime", func() {
	It("keeps summary configuration independent of commit message configuration", func() {
		capture := &runtimeCaptureAgent{result: "name: Runtime summary\ndescription: Preserve the requested settings"}
		options := SummaryOptions{
			Prompt: verify.PromptSpec{Spec: api.Spec{Model: api.Model{Name: "api:gpt-4o"}, Budget: api.Budget{Cost: 3}, Memory: api.Memory{SkipHooks: true}}},
			AgentFactory: func(config ai.AgentConfig) (ai.Agent, error) {
				Expect(config.Budget.Cost).To(Equal(float64(3)))
				return capture, nil
			},
		}
		name, _, err := GenerateGroupSummary(context.Background(), models.ScopeType("api"), "this week", models.CommitAnalyses{sampleCommit()}, options)
		Expect(err).NotTo(HaveOccurred())
		Expect(name).To(Equal("Runtime summary"))
		Expect(capture.request.Spec.Memory.SkipHooks).To(BeTrue())
		Expect(capture.request.Spec.Prompt.User).To(ContainSubstring("this week"))
		Expect(capture.closed).To(BeTrue())
	})

	It("fails a summary with invalid authored configuration before constructing its agent", func() {
		commit := sampleCommit()
		commit.Author.Date = time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
		_, err := Summarize(models.CommitAnalyses{commit}, SummaryOptions{
			MaxCategories: 3, Context: context.Background(),
			Prompt: verify.PromptSpec{Spec: api.Spec{Budget: api.Budget{Cost: -1}}},
			AgentFactory: func(ai.AgentConfig) (ai.Agent, error) {
				Fail("invalid configuration reached the factory")
				return nil, nil
			},
		})
		Expect(err).To(MatchError(SatisfyAll(ContainSubstring(".gavel.yaml commit.summary"), ContainSubstring("invalid budget cost -1"))))
	})

	It("constructs the caller's agent from the exact prepared request runtime", func() {
		capture := &runtimeCaptureAgent{}
		var config ai.AgentConfig
		options := AnalyzeOptions{
			Prompt:       verify.PromptSpec{Spec: api.Spec{Model: api.Model{Name: "api:gpt-4o"}, Budget: api.Budget{Cost: 2}}},
			AgentFactory: func(cfg ai.AgentConfig) (ai.Agent, error) { config = cfg; return capture, nil },
		}
		_, err := AnalyzeWithAI(context.Background(), sampleCommit(), nil, options)
		Expect(err).NotTo(HaveOccurred())
		Expect(config.Model).To(Equal(capture.request.Spec.Model))
		Expect(config.Budget.Cost).To(Equal(float64(2)))
		Expect(capture.closed).To(BeTrue())
	})

	It("transports the final full spec without rereading saved configuration", func() {
		dir := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", dir)
		temperature := 0.25
		options := AnalyzeOptions{
			Prompt: verify.PromptSpec{Spec: api.Spec{
				Model:   api.Model{Temperature: &temperature},
				Budget:  api.Budget{Cost: 2},
				Memory:  api.Memory{SkipMemory: true},
				CLIArgs: map[string]any{"verbose": true},
			}},
			PromptOptions: verify.PromptResolveOptions{Dir: dir},
			Saved:         captainconfig.Config{AI: captainconfig.AIDefaults{DefaultModel: "api:gpt-4o", BudgetUSD: 1}},
		}
		prepared, err := PrepareCommitMessage(sampleCommit(), options)
		Expect(err).NotTo(HaveOccurred())
		Expect(prepared.Request).To(SatisfyAll(
			HaveField("Model.Name", "gpt-4o"),
			HaveField("Model.Temperature", &temperature),
			HaveField("Budget.Cost", float64(2)),
			HaveField("Memory.SkipMemory", true),
			HaveField("CLIArgs", map[string]any{"verbose": true}),
		))
		Expect(prepared.Config.Model).To(Equal(prepared.Request.Model))
		Expect(prepared.Config.Budget).To(Equal(prepared.Request.Budget))
		Expect(os.WriteFile(filepath.Join(dir, ".captain.yaml"), []byte("ai: [invalid mutation]\n"), 0o600)).To(Succeed())
		options.Prepared = &prepared
		capture := &runtimeCaptureAgent{}
		_, err = AnalyzeWithAI(context.Background(), sampleCommit(), capture, options)
		Expect(err).NotTo(HaveOccurred())
		Expect(capture.request.Spec).To(Equal(api.Spec(prepared.Request)))
	})
})

type runtimeCaptureAgent struct {
	request ai.PromptRequest
	closed  bool
	result  string
}

func (a *runtimeCaptureAgent) ExecutePrompt(_ context.Context, request ai.PromptRequest) (*ai.PromptResponse, error) {
	a.request = request
	if a.result != "" {
		return &ai.PromptResponse{Result: a.result}, nil
	}
	return &ai.PromptResponse{Result: `{"type":"fix","subject":"preserve runtime options"}`}, nil
}

func (*runtimeCaptureAgent) ExecuteBatch(context.Context, []ai.PromptRequest) (map[string]*ai.PromptResponse, error) {
	panic("unexpected batch")
}
func (*runtimeCaptureAgent) GetCosts() ai.Costs { return ai.Costs{} }
func (a *runtimeCaptureAgent) Close() error     { a.closed = true; return nil }
