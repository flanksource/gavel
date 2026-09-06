package commit

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/status"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type promptRuntimeProbe struct {
	request  clickyai.PromptRequest
	response *clickyai.PromptResponse
}

func (a *promptRuntimeProbe) ExecutePrompt(_ context.Context, request clickyai.PromptRequest) (*clickyai.PromptResponse, error) {
	a.request = request
	if a.response != nil {
		return a.response, nil
	}
	return &clickyai.PromptResponse{StructuredData: json.RawMessage(`{"groups":[{"label":"fix: preserve runtime","files":["main.go"]}],"ignore":[]}`)}, nil
}

func (*promptRuntimeProbe) ExecuteBatch(context.Context, []clickyai.PromptRequest) (map[string]*clickyai.PromptResponse, error) {
	return nil, errors.New("unexpected batch request")
}
func (*promptRuntimeProbe) GetCosts() clickyai.Costs { return nil }
func (*promptRuntimeProbe) Close() error             { return nil }

var _ = Describe("commit prompt runtime preservation", func() {
	It("rejects an explicitly cleared grouping model instead of recovering an inherited selection", func() {
		opts := promptTestOptions()
		opts.Flags = modelFlags("api:sonnet")
		opts.Config = msgGroupSpecCfg("api:haiku", "api:gpt-5")
		opts.GroupModel = (api.Model{}).WithExplicit("/model")
		_, err := renderGroupingPrompt(opts, groupingTable)
		Expect(err).To(MatchError(ContainSubstring("configure")))
	})

	DescribeTable("loads saved configuration only for operations that may call AI", func(opts Options, wantError bool) {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		path, err := captainconfig.Path()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(path, []byte("ai: [invalid"), 0o600)).To(Succeed())
		err = opts.LoadAIConfig()
		if wantError {
			Expect(err).To(HaveOccurred())
		} else {
			Expect(err).NotTo(HaveOccurred())
			Expect(opts.Saved).To(BeNil())
		}
	},
		Entry("explicit message", Options{Message: "fix: explicit"}, false),
		Entry("explicit fixup", Options{Fixup: "abc123"}, false),
		Entry("new PR after explicit message", Options{Message: "fix: explicit", Push: true}, true),
		Entry("automatic fixup may generate an unmatched message", Options{Fixup: FixupAuto}, true),
	)

	It("resolves grouping frontmatter before constructing the provider and carries all runtime groups", func() {
		workDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(workDir, "group.prompt"), []byte("---\nmodel: api:sonnet\nfallbacks: [api:gpt-5]\nbudget: {cost: 3, maxTokens: 1200}\nmemory: {skipUser: true}\npermissions: {mode: auto}\nsetup: {cwd: /work/group, envVars: [{name: REVIEW_SCOPE, value: staged}]}\nworkflow: {verify: {commands: ['true']}}\ncliArgs: {verbose: true}\n---\nGroup {{table}}"), 0o600)).To(Succeed())
		agent := &promptRuntimeProbe{}
		var config clickyai.AgentConfig
		previousFactory, previousGather := newAgentFunc, gatherStatusFunc
		DeferCleanup(func() { newAgentFunc, gatherStatusFunc = previousFactory, previousGather })
		newAgentFunc = func(cfg clickyai.AgentConfig) (clickyai.Agent, error) { config = cfg; return agent, nil }
		gatherStatusFunc = func(string) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{{Path: "main.go"}}}, nil
		}

		groups, err := groupChangesByAI(context.Background(), Options{
			WorkDir: workDir, MaxCommits: 2,
			Saved:  &captainconfig.Config{},
			AI:     api.Spec{Model: api.Model{Name: "api:gpt-5"}, Budget: api.Budget{Cost: 8}},
			Config: verify.CommitConfig{Grouping: verify.PromptSpec{File: "group.prompt"}},
		}, stagedSource{Changes: []stagedChange{{Path: "main.go", Status: "modified"}}})
		Expect(err).NotTo(HaveOccurred())
		Expect(groups).To(HaveLen(1))
		Expect(config.Model.Name).To(Equal("claude-sonnet-5"))
		Expect(config.Budget).To(Equal(api.Budget{Cost: 3, MaxTokens: 1200}))
		Expect(agent.request.Spec.Model).To(Equal(config.Model))
		Expect(agent.request.Spec.Budget).To(Equal(config.Budget))
		Expect(agent.request.Spec.Fallbacks).To(HaveLen(1))
		Expect(agent.request.Spec.Fallbacks[0].Name).To(Equal("gpt-5"))
		Expect(agent.request.Spec.Memory.SkipUser).To(BeTrue())
		Expect(agent.request.Spec.Permissions.Mode).To(Equal(api.PermissionMode("auto")))
		Expect(agent.request.Spec.Setup).NotTo(BeNil())
		Expect(agent.request.Spec.Setup.Cwd).To(Equal("/work/group"))
		Expect(agent.request.Spec.Setup.EnvVars).To(HaveLen(1))
		Expect(agent.request.Spec.Setup.EnvVars[0].ValueStatic).To(Equal("staged"))
		Expect(agent.request.Spec.Workflow.Verify.Commands).To(Equal([]string{"true"}))
		Expect(agent.request.Spec.CLIArgs).To(HaveKeyWithValue("verbose", true))
		Expect(agent.request.Spec.Prompt.User).To(ContainSubstring("main.go"))
		Expect(agent.request.Spec.Prompt.SchemaJSON).NotTo(BeEmpty())
		Expect(agent.request.Spec.Prompt.SchemaStrictness).To(Equal(api.SchemaStrictnessRetry))
	})

	It("keeps explicit false and zero PR settings above saved defaults in request and provider", func() {
		agent := &promptRuntimeProbe{response: &clickyai.PromptResponse{StructuredData: json.RawMessage(`{"title":"fix: preserve settings","body":"reviewed","branch":"fix/settings"}`)}}
		var config clickyai.AgentConfig
		previous := newAgentFunc
		DeferCleanup(func() { newAgentFunc = previous })
		newAgentFunc = func(cfg clickyai.AgentConfig) (clickyai.Agent, error) { config = cfg; return agent, nil }
		opts := promptTestOptions()
		opts.Saved.AI = captainconfig.AIDefaults{BudgetUSD: 9, MaxTokens: 9000, NoCache: true, NoUser: true}
		Expect(json.Unmarshal([]byte(`{"budget":{"cost":0,"maxTokens":1500},"noCache":false,"memory":{"skipUser":false},"prompt":{"user":"Summarize {{#each commits}}{{message}}{{/each}}","system":"Keep the rationale"}}`), &opts.PR.Content.Spec)).To(Succeed())
		content, err := GeneratePRContent(context.Background(), PRContentInput{Options: opts, Commits: []PRCommitInput{{Message: "fix: explicit settings"}}})
		Expect(err).NotTo(HaveOccurred())
		Expect(content.Title).To(Equal("fix: preserve settings"))
		Expect(config.NoCache).To(BeFalse())
		Expect(config.Budget).To(Equal(api.Budget{MaxTokens: 1500}))
		Expect(agent.request.Spec.Budget).To(Equal(config.Budget))
		Expect(agent.request.Spec.NoCache).To(BeFalse())
		Expect(agent.request.Spec.Memory.SkipUser).To(BeFalse())
		Expect(agent.request.Spec.Prompt.System).To(Equal("Keep the rationale"))
		Expect(agent.request.Spec.Prompt.User).To(Equal("Summarize fix: explicit settings"))
		Expect(agent.request.Spec.Prompt.SchemaStrictness).To(Equal(api.SchemaStrictnessRetry))
	})

	It("rejects an invalid operation setting before constructing a provider", func() {
		previous := newAgentFunc
		DeferCleanup(func() { newAgentFunc = previous })
		calls := 0
		newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) { calls++; return &promptRuntimeProbe{}, nil }
		opts := promptTestOptions()
		opts.PR.Content.Spec.Budget.Cost = -1
		_, err := GeneratePRContent(context.Background(), PRContentInput{Options: opts, Commits: []PRCommitInput{{Message: "fix: invalid settings"}}})
		Expect(err).To(HaveOccurred())
		Expect(calls).To(BeZero())
	})
})
