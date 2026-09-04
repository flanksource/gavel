package run_test

import (
	"encoding/json"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	priorCodexModel  = "gpt-5.6-sol"
	priorPlanSession = "sess-prior-plan"
	priorPlanCost    = 4.5
	priorPlanTurns   = 12
)

// persistedSpec projects a spec the way todos/runtime writes `rendered_spec`,
// so a spec's prior run carries exactly what a real one would.
func persistedSpec(spec api.Spec) map[string]any {
	GinkgoHelper()
	data, err := json.Marshal(spec)
	Expect(err).NotTo(HaveOccurred())
	var rendered map[string]any
	Expect(json.Unmarshal(data, &rendered)).To(Succeed())
	return rendered
}

// codexPlanRun is a finished plan turn as the native runtime persisted it: the
// spec it was dispatched with, plus the runtime it actually resolved.
func codexPlanRun() *captaindb.PromptRun {
	return &captaindb.PromptRun{
		State: captaindb.PromptRunStateWaiting,
		Runtime: captaindb.PromptRunRuntime{
			Mode: string(types.ModePlan),
			Resolved: captaindb.PromptRunRuntimeSelection{
				Provider: "openai", Mode: "agent", Model: priorCodexModel, Effort: "high",
			},
		},
		RenderedSpec: persistedSpec(api.Spec{
			SessionID: priorPlanSession,
			// The family alias the run was requested with; Runtime.Resolved carries
			// the concrete model it became.
			Model:  api.Model{Name: "codex", Mode: api.ModeAgent},
			Budget: api.Budget{Cost: priorPlanCost, MaxTurns: priorPlanTurns},
			// The turn that ran, the tree it ran in, and the fixture stamped on the
			// record — none of which is configuration.
			Prompt:   api.Prompt{User: "the previous turn's instructions"},
			Setup:    &shell.Setup{Cwd: "/previous/worktree"},
			Workflow: &api.Workflow{Verify: &api.Verify{Fixture: "```test\n```"}},
			Permissions: api.Permissions{
				Mode:  api.PermissionPlan,
				Tools: api.ToolsFromLists([]string{"Read", "Glob", "Grep"}, nil),
			},
		}),
	}
}

func layer(opts run.Options, name string) api.Spec {
	GinkgoHelper()
	for _, candidate := range opts.Prior {
		if candidate.Name == name {
			return candidate.Spec
		}
	}
	Fail("continuation carries no layer " + name)
	return api.Spec{}
}

var _ = Describe("Continue", func() {
	var dir string

	BeforeEach(func() {
		// An empty home and work dir: the default lifecycle, no authored layers.
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		dir = GinkgoT().TempDir()
	})

	It("continues a same-class run on the spec it dispatched and the runtime it resolved", func() {
		opts, err := run.Continue(run.Continuation{
			Dir: dir, Prior: codexPlanRun(), Override: run.Options{Host: lifecycle.HostCLI}, Step: "plan",
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Step).To(Equal("plan"))
		Expect(opts.Host).To(Equal(lifecycle.HostCLI))
		spec := layer(opts, "prior run spec")
		Expect(spec.Budget.Cost).To(Equal(priorPlanCost))
		Expect(spec.Budget.MaxTurns).To(Equal(priorPlanTurns))
		Expect(spec.Permissions.Mode).To(Equal(api.PermissionPlan))
		Expect(spec.Prompt.User).To(BeEmpty(), "the previous turn's instructions must not be re-sent")
		Expect(spec.Setup).To(BeNil(), "a consumed checkout must not pin the continuation to the old tree")
		Expect(spec.Workflow.Verify.Fixture).To(BeEmpty(), "the persistence stamp is re-stamped from the issue")
		Expect(spec.SessionID).To(BeEmpty(), "a continuation that does not resume inherits no session")
	})

	It("carries the resolved runtime, never the family alias, so a codex session is never continued by claude", func() {
		for _, step := range []string{"plan", "run"} {
			opts, err := run.Continue(run.Continuation{
				Dir: dir, Prior: codexPlanRun(), Override: run.Options{Host: lifecycle.HostDashboard}, Step: step,
			})
			Expect(err).NotTo(HaveOccurred(), step)

			runtime := layer(opts, "prior run runtime").Model
			Expect(runtime.Name).To(Equal(priorCodexModel), step)
			Expect(runtime.Mode).To(Equal(api.ModeAgent), step)
			Expect(string(runtime.Effort)).To(Equal("high"), step)
		}
	})

	It("inherits nothing but the runtime across a class change", func() {
		opts, err := run.Continue(run.Continuation{
			Dir: dir, Prior: codexPlanRun(), Override: run.Options{Host: lifecycle.HostCLI}, Step: "run",
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(layer(opts, "prior run spec")).To(Equal(api.Spec{}),
			"a plan's read-only posture and investigation budget belong to planning")
	})

	It("keeps the prior session and the message only when resuming", func() {
		opts, err := run.Continue(run.Continuation{
			Dir: dir, Prior: codexPlanRun(), Override: run.Options{Host: lifecycle.HostCLI},
			Step: "plan", Resume: true, Message: "bound the queue",
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Resume).To(BeTrue())
		Expect(opts.Message).To(Equal("bound the queue"))
		Expect(layer(opts, "prior run spec").SessionID).To(Equal(priorPlanSession))
	})

	It("resolves like a fresh run when the provider keeps no run history", func() {
		opts, err := run.Continue(run.Continuation{
			Dir: dir, Override: run.Options{Host: lifecycle.HostCLI, Concurrent: true}, Step: "run",
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(layer(opts, "prior run spec")).To(Equal(api.Spec{}))
		Expect(layer(opts, "prior run runtime")).To(Equal(api.Spec{}))
		Expect(opts.Concurrent).To(BeTrue(), "the caller's own decisions carry through")
	})

	DescribeTable("refuses a continuation it cannot place",
		func(c run.Continuation, want string) {
			c.Dir = dir
			_, err := run.Continue(c)
			Expect(err).To(MatchError(ContainSubstring(want)))
		},
		Entry("no step", run.Continuation{Override: run.Options{Host: lifecycle.HostCLI}}, "names no lifecycle step"),
		Entry("options naming another step",
			run.Continuation{Override: run.Options{Host: lifecycle.HostCLI, Step: "plan"}, Step: "run"}, "options.step"),
		Entry("no host", run.Continuation{Step: "run"}, "names no host"),
		Entry("a step the lifecycle does not declare",
			run.Continuation{Override: run.Options{Host: lifecycle.HostCLI}, Step: "ship-it"}, "not part of lifecycle"),
	)
})
