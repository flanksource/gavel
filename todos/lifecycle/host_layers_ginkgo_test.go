package lifecycle_test

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// frontmatterLayer is one .prompt document's contribution, named the way
// Resolve names it so a spec here reads like the stack a real run builds.
func frontmatterLayer(spec api.Spec) []api.SpecLayer {
	return []api.SpecLayer{api.PromptSpecLayer("todos-run.prompt", spec)}
}

func runLayerInput(cfg verify.GavelConfig, front api.Spec, request api.Spec) lifecycle.LayerInput {
	return lifecycle.LayerInput{
		Config:      cfg,
		Step:        "run",
		Frontmatter: frontmatterLayer(front),
		Request:     request,
	}
}

func resolveRun(cfg verify.GavelConfig, front api.Spec, request api.Spec) api.Spec {
	GinkgoHelper()
	resolved, err := lifecycle.ResolveLayers(runLayerInput(cfg, front, request))
	Expect(err).ToNot(HaveOccurred())
	return resolved.Spec
}

func layerNames(in lifecycle.LayerInput) []string {
	names := make([]string, 0)
	for _, layer := range lifecycle.Layers(in) {
		names = append(names, layer.Name)
	}
	return names
}

// askPolicies returns every tool a spec puts behind the per-tool "ask" policy.
// Captain refuses "ask" on every runtime, so a layer that emits one turns a run
// into a boundary error instead of a prompt.
func askPolicies(permissions api.Permissions) []string {
	var asked []string
	for tool, policy := range permissions.Tools {
		if policy == api.ToolPolicyAsk {
			asked = append(asked, tool)
		}
	}
	return asked
}

var _ = Describe("todo spec layers", func() {
	Describe("precedence", func() {
		It("lets ai: supply what no later layer names", func() {
			cfg := verify.GavelConfig{AI: api.Spec{Budget: api.Budget{Cost: 9, MaxTurns: 4}}}

			spec := resolveRun(cfg, api.Spec{}, api.Spec{})

			Expect(spec.Budget.Cost).To(Equal(9.0))
			Expect(spec.Budget.MaxTurns).To(Equal(4))
		})

		It("lets the prompt frontmatter override ai:", func() {
			cfg := verify.GavelConfig{AI: api.Spec{Budget: api.Budget{Cost: 9, MaxTurns: 4}}}

			spec := resolveRun(cfg, api.Spec{Budget: api.Budget{MaxTurns: 11}}, api.Spec{})

			Expect(spec.Budget.MaxTurns).To(Equal(11))
			Expect(spec.Budget.Cost).To(Equal(9.0), "a layer overrides only the fields it names")
		})

		It("lets the lifecycle step override the prompt frontmatter", func() {
			in := runLayerInput(verify.GavelConfig{}, api.Spec{Budget: api.Budget{MaxTurns: 11}}, api.Spec{})
			in.StepSpec = api.Spec{Budget: api.Budget{MaxTurns: 20}}

			resolved, err := lifecycle.ResolveLayers(in)

			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Spec.Budget.MaxTurns).To(Equal(20))
		})

		It("lets todos.<step> override the lifecycle step and the prompt frontmatter", func() {
			cfg := verify.GavelConfig{Todos: verify.TodosConfig{
				Run: verify.PromptSpec{Spec: api.Spec{Budget: api.Budget{MaxTurns: 30}}},
			}}
			in := runLayerInput(cfg, api.Spec{Budget: api.Budget{MaxTurns: 11}}, api.Spec{})
			in.StepSpec = api.Spec{Budget: api.Budget{MaxTurns: 20}}

			resolved, err := lifecycle.ResolveLayers(in)

			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Spec.Budget.MaxTurns).To(Equal(30))
		})

		It("lets the todo's own llm: override todos.<step>", func() {
			cfg := verify.GavelConfig{Todos: verify.TodosConfig{
				Run: verify.PromptSpec{Spec: api.Spec{Budget: api.Budget{MaxTurns: 30}}},
			}}
			in := runLayerInput(cfg, api.Spec{}, api.Spec{})
			in.Todos = []*types.TODO{{TODOFrontmatter: types.TODOFrontmatter{
				Title: "widen the budget",
				LLM:   &types.LLM{MaxTurns: 44, MaxCost: 21},
			}}}

			resolved, err := lifecycle.ResolveLayers(in)

			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Spec.Budget.MaxTurns).To(Equal(44))
			Expect(resolved.Spec.Budget.Cost).To(Equal(21.0))
		})

		It("lets the request override every authored layer", func() {
			cfg := verify.GavelConfig{
				AI:    api.Spec{Budget: api.Budget{MaxTurns: 4}},
				Todos: verify.TodosConfig{Run: verify.PromptSpec{Spec: api.Spec{Budget: api.Budget{MaxTurns: 30}}}},
			}

			spec := resolveRun(cfg, api.Spec{Budget: api.Budget{MaxTurns: 11}}, api.Spec{Budget: api.Budget{MaxTurns: 7}})

			Expect(spec.Budget.MaxTurns).To(Equal(7))
		})
	})

	Describe("todos.timeout", func() {
		It("lowers a longer budget the prompt declared", func() {
			cfg := verify.GavelConfig{Todos: verify.TodosConfig{Timeout: "45m"}}

			spec := resolveRun(cfg, api.Spec{Budget: api.Budget{Timeout: "90m"}}, api.Spec{})

			Expect(spec.Budget.Timeout).To(Equal("45m"))
		})

		It("never raises a shorter budget the prompt declared", func() {
			cfg := verify.GavelConfig{Todos: verify.TodosConfig{Timeout: "45m"}}

			spec := resolveRun(cfg, api.Spec{Budget: api.Budget{Timeout: "10m"}}, api.Spec{})

			Expect(spec.Budget.Timeout).To(Equal("10m"))
		})

		It("caps a request that asked for longer", func() {
			cfg := verify.GavelConfig{Todos: verify.TodosConfig{Timeout: "45m"}}

			spec := resolveRun(cfg, api.Spec{}, api.Spec{Budget: api.Budget{Timeout: "3h"}})

			Expect(spec.Budget.Timeout).To(Equal("45m"))
		})

		It("leaves the rest of the frontmatter's budget alone", func() {
			cfg := verify.GavelConfig{Todos: verify.TodosConfig{Timeout: "45m"}}

			spec := resolveRun(cfg, api.Spec{Budget: api.Budget{Cost: 12, MaxTurns: 11, MaxTokens: 4096}}, api.Spec{})

			Expect(spec.Budget.Cost).To(Equal(12.0))
			Expect(spec.Budget.MaxTurns).To(Equal(11))
			Expect(spec.Budget.MaxTokens).To(Equal(4096))
			Expect(spec.Budget.Timeout).To(Equal("45m"))
		})
	})

	Describe("permissions", func() {
		It("carries a frontmatter dontAsk mode through to captain", func() {
			spec := resolveRun(verify.GavelConfig{},
				api.Spec{Permissions: api.Permissions{Mode: api.PermissionDontAsk}}, api.Spec{})

			Expect(spec.Permissions.Mode).To(Equal(api.PermissionDontAsk))
		})

		It("keeps the CLI host out of the prompt's declared posture", func() {
			in := runLayerInput(verify.GavelConfig{},
				api.Spec{Permissions: api.Permissions{Mode: api.PermissionAcceptEdits}}, api.Spec{})
			in.Host = lifecycle.HostCLI

			resolved, err := lifecycle.ResolveLayers(in)

			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Spec.Permissions.Mode).To(Equal(api.PermissionAcceptEdits))
		})

		It("puts the dashboard host on permissions.mode default so the broker is consulted", func() {
			in := runLayerInput(verify.GavelConfig{},
				api.Spec{Permissions: api.Permissions{Mode: api.PermissionAcceptEdits}}, api.Spec{})
			in.Host = lifecycle.HostDashboard

			resolved, err := lifecycle.ResolveLayers(in)

			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Spec.Permissions.Mode).To(Equal(api.PermissionDefault))
		})

		It("still lets the request override the dashboard's posture", func() {
			in := runLayerInput(verify.GavelConfig{}, api.Spec{},
				api.Spec{Permissions: api.Permissions{Mode: api.PermissionPlan}})
			in.Host = lifecycle.HostDashboard

			resolved, err := lifecycle.ResolveLayers(in)

			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Spec.Permissions.Mode).To(Equal(api.PermissionPlan))
		})

		// The regression: an approval-capable host used to rewrite the tool map to
		// `Bash: ask`, which captain refuses on every runtime — so the run failed at
		// the policy gate instead of prompting. Approval is a permission MODE plus a
		// broker now, never a per-tool policy.
		DescribeTable("never emits a per-tool ask on any layer",
			func(host lifecycle.HostKind, step string, cfg verify.GavelConfig, front api.Spec, request api.Spec) {
				in := lifecycle.LayerInput{
					Config: cfg, Step: step, Host: host, Request: request,
					Frontmatter: []api.SpecLayer{api.PromptSpecLayer("todos-"+step+".prompt", front)},
					Todos:       []*types.TODO{{TODOFrontmatter: types.TODOFrontmatter{Title: "t", LLM: &types.LLM{Model: "opus"}}}},
				}

				for _, layer := range lifecycle.Layers(in) {
					Expect(askPolicies(layer.Spec.Permissions)).To(BeEmpty(),
						"layer %q must not put a tool behind ask", layer.Name)
				}
				resolved, err := lifecycle.ResolveLayers(in)
				Expect(err).ToNot(HaveOccurred())
				Expect(askPolicies(resolved.Spec.Permissions)).To(BeEmpty())
			},
			Entry("cli run", lifecycle.HostCLI, "run", verify.GavelConfig{}, api.Spec{}, api.Spec{}),
			Entry("dashboard run", lifecycle.HostDashboard, "run", verify.GavelConfig{}, api.Spec{}, api.Spec{}),
			Entry("dashboard plan", lifecycle.HostDashboard, "plan", verify.GavelConfig{}, api.Spec{}, api.Spec{}),
			Entry("dashboard verify", lifecycle.HostDashboard, "verify", verify.GavelConfig{}, api.Spec{}, api.Spec{}),
			Entry("dashboard run with an allowlist",
				lifecycle.HostDashboard, "run", verify.GavelConfig{},
				api.Spec{Permissions: api.Permissions{Tools: api.ToolsFromLists([]string{"Read", "Edit"}, nil)}},
				api.Spec{}),
		)
	})

	Describe("a continuation's prior layers", func() {
		It("inherits the prior run's configuration over the authored layers", func() {
			cfg := verify.GavelConfig{Todos: verify.TodosConfig{
				Run: verify.PromptSpec{Spec: api.Spec{Budget: api.Budget{MaxTurns: 30}}},
			}}
			in := runLayerInput(cfg, api.Spec{}, api.Spec{})
			in.Prior = []api.SpecLayer{api.RequestSpecLayer("prior run", api.Spec{Budget: api.Budget{MaxTurns: 5}})}

			resolved, err := lifecycle.ResolveLayers(in)

			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Spec.Budget.MaxTurns).To(Equal(5))
		})

		It("yields to the host's posture and to the caller's own request", func() {
			in := runLayerInput(verify.GavelConfig{}, api.Spec{},
				api.Spec{Budget: api.Budget{MaxTurns: 2}})
			in.Host = lifecycle.HostDashboard
			in.Prior = []api.SpecLayer{api.RequestSpecLayer("prior run", api.Spec{
				Budget:      api.Budget{MaxTurns: 5},
				Permissions: api.Permissions{Mode: api.PermissionBypass},
			})}

			resolved, err := lifecycle.ResolveLayers(in)

			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Spec.Budget.MaxTurns).To(Equal(2))
			Expect(resolved.Spec.Permissions.Mode).To(Equal(api.PermissionDefault),
				"a continuation runs where it is continued from, not where it started")
		})

		It("contributes no layer when the prior run configured nothing", func() {
			in := runLayerInput(verify.GavelConfig{}, api.Spec{}, api.Spec{})
			in.Prior = []api.SpecLayer{api.RequestSpecLayer("prior run", api.Spec{})}

			Expect(layerNames(in)).ToNot(ContainElement("prior run"))
		})
	})

	Describe("the layer stack", func() {
		It("names every layer, lowest precedence first", func() {
			cfg := verify.GavelConfig{
				AI:    api.Spec{Budget: api.Budget{Cost: 1}},
				Todos: verify.TodosConfig{Timeout: "45m", Run: verify.PromptSpec{Spec: api.Spec{Budget: api.Budget{MaxTurns: 3}}}},
			}
			in := runLayerInput(cfg, api.Spec{Budget: api.Budget{MaxTokens: 10}},
				api.Spec{Model: api.Model{Effort: api.EffortHigh}})
			in.StepSpec = api.Spec{Permissions: api.Permissions{Mode: api.PermissionAcceptEdits}}
			in.Host = lifecycle.HostDashboard
			in.Todos = []*types.TODO{{TODOFrontmatter: types.TODOFrontmatter{Title: "widen", LLM: &types.LLM{MaxCost: 2}}}}

			Expect(layerNames(in)).To(Equal([]string{
				".gavel.yaml ai",
				".gavel.yaml todos.timeout",
				"todos-run.prompt",
				"lifecycle step run",
				".gavel.yaml todos.run",
				"todo widen",
				"host dashboard",
				"request",
			}))
		})

		It("omits a layer that configures nothing", func() {
			Expect(layerNames(runLayerInput(verify.GavelConfig{}, api.Spec{}, api.Spec{}))).To(
				Equal([]string{".gavel.yaml ai", "todos-run.prompt"}),
				"an empty todos.timeout, lifecycle step, todos.<step>, todo llm:, host and request contribute no layer")
		})

		It("records the resolved order as captain's trace", func() {
			cfg := verify.GavelConfig{AI: api.Spec{Budget: api.Budget{Cost: 1}}}

			resolved, err := lifecycle.ResolveLayers(runLayerInput(cfg, api.Spec{Budget: api.Budget{MaxTurns: 2}}, api.Spec{}))

			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Trace).To(HaveLen(2))
			Expect(resolved.Trace[0].Name).To(Equal(".gavel.yaml ai"))
			Expect(resolved.Trace[0].Scope).To(Equal(api.SpecLayerGlobal))
			Expect(resolved.Trace[1].Name).To(Equal("todos-run.prompt"))
			Expect(resolved.Trace[1].Scope).To(Equal(api.SpecLayerSurface))
		})
	})

	Describe("the prompt body", func() {
		It("never travels as a spec override", func() {
			cfg := verify.GavelConfig{Todos: verify.TodosConfig{
				Run: verify.PromptSpec{Spec: api.Spec{Prompt: api.Prompt{User: "{{{body}}}", System: "be terse"}}},
			}}

			spec := resolveRun(cfg, api.Spec{Prompt: api.Prompt{User: "raw template source"}}, api.Spec{})

			Expect(spec.Prompt.User).To(BeEmpty(), "the body is rendered by todos/prompt, not layered")
			Expect(spec.Prompt.System).To(Equal("be terse"), "system prompt is run configuration and survives")
		})

		It("is stripped from todos.verify like every other layer", func() {
			cfg := verify.GavelConfig{Todos: verify.TodosConfig{
				Verify: api.Spec{Prompt: api.Prompt{User: "grade it", System: "be strict"}, Budget: api.Budget{MaxTurns: 6}},
			}}

			resolved, err := lifecycle.ResolveLayers(lifecycle.LayerInput{Config: cfg, Step: lifecycle.StepVerify})

			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Spec.Prompt.User).To(BeEmpty(), "a verify step has no prompt body to override")
			Expect(resolved.Spec.Prompt.System).To(Equal("be strict"))
			Expect(resolved.Spec.Budget.MaxTurns).To(Equal(6))
		})
	})

	Describe("a custom step", func() {
		It("takes its project layer from todos.steps.<name>", func() {
			cfg := verify.GavelConfig{Todos: verify.TodosConfig{
				Steps: map[string]api.Spec{"handoff": {Budget: api.Budget{MaxTurns: 9}, Prompt: api.Prompt{User: "raw"}}},
			}}
			in := lifecycle.LayerInput{Config: cfg, Step: "handoff", StepSpec: api.Spec{Budget: api.Budget{MaxTurns: 2}}}

			resolved, err := lifecycle.ResolveLayers(in)

			Expect(err).ToNot(HaveOccurred())
			Expect(layerNames(in)).To(ContainElement(".gavel.yaml todos.handoff"))
			Expect(resolved.Spec.Budget.MaxTurns).To(Equal(9), "the project block outranks the step's own declaration")
			Expect(resolved.Spec.Prompt.User).To(BeEmpty())
		})

		It("contributes no layer when todos.steps does not name it", func() {
			in := lifecycle.LayerInput{Config: verify.GavelConfig{}, Step: "handoff"}

			Expect(layerNames(in)).ToNot(ContainElement(".gavel.yaml todos.handoff"))
		})
	})

	Describe("class invariants", func() {
		It("strips inherited commit policies from a plan-class spec", func() {
			spec := api.Spec{Workflow: &api.Workflow{Commits: []api.Commit{{On: "run"}}}}

			lifecycle.ApplyClassInvariants(&spec, types.ModePlan)

			Expect(spec.Workflow.Commits).To(BeEmpty())
		})

		It("keeps commit policies on a run-class spec", func() {
			spec := api.Spec{Workflow: &api.Workflow{Commits: []api.Commit{{On: "run"}}}}

			lifecycle.ApplyClassInvariants(&spec, types.ModeRun)

			Expect(spec.Workflow.Commits).To(HaveLen(1))
		})
	})
})
