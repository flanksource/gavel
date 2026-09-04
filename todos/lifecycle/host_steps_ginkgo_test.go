package lifecycle_test

import (
	"context"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func stateOf(states []lifecycle.StepState, name string) lifecycle.StepState {
	GinkgoHelper()
	for _, state := range states {
		if state.Step.Name == name {
			return state
		}
	}
	Fail("no step state for " + name)
	return lifecycle.StepState{}
}

var _ = Describe("Host step selection", func() {
	var (
		provider *fakeProvider
		host     *lifecycle.Host
		ctx      context.Context
	)
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		provider = &fakeProvider{plan: todos.PlanState{Exists: true, Approved: true, Content: "# plan", Revision: 2}}
		host = newHost(provider)
		ctx = context.Background()
	})

	Describe("Steps", func() {
		It("reports every declared step with its applicability and the suggested one", func() {
			todo := hostTodo()

			states, err := host.Steps(ctx, todo)

			Expect(err).NotTo(HaveOccurred())
			names := make([]string, len(states))
			suggested := ""
			byName := map[string]lifecycle.StepState{}
			for i, state := range states {
				names[i] = state.Step.Name
				byName[state.Step.Name] = state
				if state.Suggested {
					Expect(suggested).To(BeEmpty(), "exactly one step is suggested")
					suggested = state.Step.Name
				}
			}
			// Definition order, auxiliary steps included.
			Expect(names).To(Equal([]string{"triage", "plan", "verify", "run"}))
			// A pending todo with an approved plan and no prior run implements next;
			// verify needs a succeeded run to judge, and plan is satisfied.
			Expect(suggested).To(Equal("run"))
			Expect(byName["run"].Applicable).To(BeTrue())
			Expect(byName["plan"].Applicable).To(BeFalse())
			Expect(byName["verify"].Applicable).To(BeFalse())
			// Triage always applies but is auxiliary, so it is never suggested.
			Expect(byName["triage"].Applicable).To(BeTrue())
			Expect(byName["triage"].Suggested).To(BeFalse())
		})

		It("explains why a step does not apply by naming its predicate", func() {
			states, err := host.Steps(ctx, hostTodo())

			Expect(err).NotTo(HaveOccurred())
			for _, state := range states {
				Expect(state.Reason).NotTo(BeEmpty(), "step %s", state.Step.Name)
			}
			for _, state := range states {
				if state.Step.Name == "plan" {
					Expect(state.Reason).To(ContainSubstring("subject.plan.exists"))
				}
				if state.Step.Name == "triage" {
					Expect(state.Reason).To(ContainSubstring("always"))
				}
			}
		})

		It("carries the latest run per step", func() {
			provider.runs = []todos.StepRunRecord{
				{Step: "run", State: "failed", Outcome: "failed", PromptRunID: "run-1"},
				{Step: "run", State: "succeeded", Outcome: "verified", PromptRunID: "run-2"},
			}

			states, err := host.Steps(ctx, hostTodo())

			Expect(err).NotTo(HaveOccurred())
			for _, state := range states {
				if state.Step.Name != "run" {
					Expect(state.LastRun).To(BeNil(), "step %s", state.Step.Name)
					continue
				}
				Expect(state.LastRun).NotTo(BeNil())
				Expect(state.LastRun.State).To(Equal("succeeded"))
				Expect(state.LastRun.Outcome).To(Equal("verified"))
				Expect(state.LastRun.PromptRunID).To(Equal("run-2"), "the newest run of the step, not the first")
				Expect(state.Done).To(BeTrue())
			}
		})

		It("reads a custom step's runs under its own name", func() {
			// A project step that escalates once two runs have happened, whatever
			// they were, and never twice: it reads size(runs) and last.<itself>.
			def, err := lifecycle.LoadWith(verify.LifecycleConfig{Steps: []map[string]any{{
				"name": "escalate", "prompt": "file:prompts/escalate.prompt",
				"when":     "size(runs) >= 2 && !has(last.escalate)",
				"outcomes": []map[string]any{{"status": "keep", "when": "true"}},
			}}}, host.WorkDir)
			Expect(err).NotTo(HaveOccurred())
			host.Def, err = lifecycle.New(def)
			Expect(err).NotTo(HaveOccurred())
			provider.runs = []todos.StepRunRecord{{Step: "run", State: "failed"}}

			states, err := host.Steps(ctx, hostTodo())
			Expect(err).NotTo(HaveOccurred())
			Expect(stateOf(states, "escalate").Applicable).To(BeFalse(), "one run is not enough")

			provider.runs = append(provider.runs, todos.StepRunRecord{Step: "run", State: "failed"})
			states, err = host.Steps(ctx, hostTodo())
			Expect(err).NotTo(HaveOccurred())
			Expect(stateOf(states, "escalate").Applicable).To(BeTrue(), "size(runs) counts every recorded run")
			Expect(stateOf(states, "escalate").LastRun).To(BeNil())

			provider.runs = append(provider.runs, todos.StepRunRecord{Step: "escalate", State: "succeeded", PromptRunID: "run-3"})
			states, err = host.Steps(ctx, hostTodo())
			Expect(err).NotTo(HaveOccurred())
			Expect(stateOf(states, "escalate").Applicable).To(BeFalse(), "last.escalate is the step's own newest run")
			Expect(stateOf(states, "escalate").LastRun.PromptRunID).To(Equal("run-3"))
			Expect(stateOf(states, "escalate").Done).To(BeTrue())
		})
	})

	Describe("StepFor", func() {
		It("resolves a step the caller named, whether or not it applies", func() {
			step, reason, err := host.StepFor(ctx, hostTodo(), "plan")

			Expect(err).NotTo(HaveOccurred())
			Expect(step.Name).To(Equal("plan"))
			Expect(reason).To(ContainSubstring("requested"))
		})

		It("falls back to the step the lifecycle would run next", func() {
			step, reason, err := host.StepFor(ctx, hostTodo(), "")

			Expect(err).NotTo(HaveOccurred())
			Expect(step.Name).To(Equal("run"))
			Expect(reason).To(ContainSubstring("subject.status"))
		})

		It("fails loudly on a step the lifecycle does not declare", func() {
			_, _, err := host.StepFor(ctx, hostTodo(), "nonesuch")

			Expect(err).To(MatchError(ContainSubstring(`step "nonesuch" is not part of lifecycle todos`)))
			Expect(err).To(MatchError(ContainSubstring("triage, plan, verify, run")))
		})

		It("fails when no step applies rather than picking one anyway", func() {
			todo := hostTodo()
			todo.Status = types.StatusReview

			_, _, err := host.StepFor(ctx, todo, "")

			Expect(err).To(MatchError(ContainSubstring("no lifecycle step applies")))
		})
	})
})
