package types_test

import (
	"encoding/json"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	dbtypes "github.com/flanksource/commons-db/types"
	"github.com/flanksource/gavel/fixtures"
	fixturetypes "github.com/flanksource/gavel/fixtures/types"
	"github.com/flanksource/gavel/todos"
	todotypes "github.com/flanksource/gavel/todos/types"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("serialized TODO grader runtime", func() {
	g.It("preserves the full grader through the generated document and fixture parser", func() {
		g.GinkgoT().Setenv("HOME", g.GinkgoT().TempDir())
		temperature := 0.0
		grader := api.Spec{
			Model: api.Model{Name: "claude-sonnet-4-6", Mode: api.ModeCLI, Effort: api.EffortHigh, Temperature: &temperature,
				Fallbacks: []api.Model{{Name: "claude-opus-4-6", Mode: api.ModeCLI, Effort: api.EffortLow, Temperature: &temperature, NoCache: true}}, NoCache: true},
			Budget:          api.Budget{Cost: 2, MaxTokens: 900, MaxTurns: 4, Timeout: "5m"},
			Permissions:     api.Permissions{Mode: api.PermissionDontAsk},
			Memory:          api.Memory{SkipUser: true, SkipHooks: true},
			ToolPreferences: api.ToolPreferences{"review.read": api.ToolPolicyAllow},
			Setup:           &shell.Setup{Cwd: "review-workspace", EnvVars: []dbtypes.EnvVar{{Name: "REVIEW_MODE", ValueStatic: "strict"}}},
			Sandbox:         &api.SandboxRef{Mode: api.SandboxNative},
			Prompt:          api.Prompt{System: "Be an independent reviewer.", User: "old task", Source: "old prompt", SchemaJSON: json.RawMessage(`{"type":"string"}`)},
			SessionID:       "previous-session", ToolApproval: &api.ToolApprovalResume{},
			Messages: []api.Message{{Role: api.RoleUser}},
			Workflow: &api.Workflow{Verify: &api.Verify{Fixture: "do not verify yourself"}, Commits: []api.Commit{{On: "run"}}},
		}
		dod, err := todos.BuildDefinitionOfDone(todos.DefinitionOfDoneOptions{
			WorkDir: g.GinkgoT().TempDir(), Grader: grader,
			Todos: []*todotypes.TODO{{AcceptanceCriteria: []todotypes.AcceptanceCriterion{{Text: "The parser preserves the grader."}}}},
		})
		o.Expect(err).NotTo(o.HaveOccurred())
		root, err := fixtures.ParseMarkdownDocument("grader", dod.Fixture, g.GinkgoT().TempDir())
		o.Expect(err).NotTo(o.HaveOccurred())
		var parsed *fixtures.FixtureTest
		root.Walk(func(node *fixtures.FixtureNode) {
			if node.Test != nil && node.Test.AIStep != nil {
				parsed = node.Test
			}
		})
		o.Expect(parsed).NotTo(o.BeNil())
		o.Expect(parsed.AI.Spec).NotTo(o.BeNil())
		want := grader
		want.SessionID, want.ToolApproval, want.Messages, want.Workflow = "", nil, nil, nil
		want.Prompt.User, want.Prompt.Source, want.Prompt.SchemaJSON = "", "", nil
		want = want.WithExplicit("/temperature", "/fallbacks/0/temperature")
		want.Fallbacks = []api.Model{want.Fallbacks[0].WithExplicit("/model", "/mode", "/effort", "/temperature", "/noCache")}
		o.Expect(*parsed.AI.Spec).To(o.Equal(want))
		o.Expect(parsed.AI.Model).To(o.BeEmpty())

		caller := api.Spec{Model: api.Model{Name: "implementer-model"}, Memory: api.Memory{SkipSkills: true}, CLIArgs: map[string]any{"implementation-only": true}}
		resolved, config, err := fixturetypes.ResolveAIStepSpecForTest(*parsed, fixtures.RunOptions{Spec: &caller})
		o.Expect(err).NotTo(o.HaveOccurred())
		want.Prompt.User = resolved.Prompt.User
		want.Prompt.Source = "fixtures.ai-step"
		want.Prompt.Schema = resolved.Prompt.Schema
		o.Expect(resolved).To(o.Equal(want))
		o.Expect(resolved.Prompt.User).To(o.ContainSubstring("The parser preserves the grader."))
		o.Expect(resolved.Prompt.Schema).NotTo(o.BeNil())
		o.Expect(config.Model).To(o.Equal(want.Model))
		o.Expect(config.Budget).To(o.Equal(grader.Budget))
		o.Expect(grader.SessionID).To(o.Equal("previous-session"))
		o.Expect(grader.Workflow.Verify).NotTo(o.BeNil())
	})
})
