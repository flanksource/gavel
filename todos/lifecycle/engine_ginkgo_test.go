package lifecycle_test

import (
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// subjectFor is a todo as the host projects it, with every field the default
// lifecycle declares present.
func subjectFor(status string, plan map[string]any, verification map[string]any) map[string]any {
	if plan == nil {
		plan = map[string]any{"exists": false, "approved": false, "content": "", "path": "", "revision": 0}
	}
	if verification == nil {
		verification = map[string]any{"exists": false, "document": ""}
	}
	return map[string]any{
		"id": "todo-1", "status": status, "priority": "medium", "labels": []string{"area/todos"},
		"body": "Implement the thing", "attempts": 2,
		"execution": map[string]any{"state": "idle"}, "plan": plan, "verification": verification,
	}
}

func at(offset time.Duration) *time.Time {
	t := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC).Add(offset)
	return &t
}

var (
	approvedPlan   = map[string]any{"exists": true, "approved": true, "content": "# plan", "path": "plans/1.md", "revision": 1}
	unapprovedPlan = map[string]any{"exists": true, "approved": false, "content": "# plan", "path": "plans/1.md", "revision": 1}
	withDoD        = map[string]any{"exists": true, "document": "# DoD"}
)

func defaultEngine() *lifecycle.Engine {
	def, err := lifecycle.Default()
	Expect(err).NotTo(HaveOccurred())
	engine, err := lifecycle.New(def)
	Expect(err).NotTo(HaveOccurred())
	return engine
}

func stepNamed(engine *lifecycle.Engine, name string) lifecycle.Step {
	step, ok := engine.Definition().Step(name)
	Expect(ok).To(BeTrue(), "step %s", name)
	return step
}

var _ = Describe("default lifecycle", func() {
	It("loads with a verify step and the built-in step order", func() {
		engine := defaultEngine()
		var names []string
		for _, step := range engine.Definition().Steps {
			names = append(names, step.Name)
		}
		Expect(names).To(Equal([]string{"triage", "plan", "verify", "run"}))
		Expect(engine.Definition().Name).To(Equal("todos"))
		_, hasVerify := engine.Definition().Step(lifecycle.StepVerify)
		Expect(hasVerify).To(BeTrue())
	})

	DescribeTable("Next picks the first applicable non-auxiliary step",
		func(c lifecycle.Context, want string) {
			step, ok, err := defaultEngine().Next(c)
			Expect(err).NotTo(HaveOccurred())
			if want == "" {
				Expect(ok).To(BeFalse(), "expected no step, got %s", step.Name)
				return
			}
			Expect(ok).To(BeTrue(), "expected %s, got nothing", want)
			Expect(step.Name).To(Equal(want))
		},
		Entry("draft with no plan → plan", lifecycle.Context{Subject: subjectFor("draft", nil, nil)}, "plan"),
		Entry("pending with an approved plan → run", lifecycle.Context{Subject: subjectFor("pending", approvedPlan, nil)}, "run"),
		Entry("pending with an unapproved plan → nothing", lifecycle.Context{Subject: subjectFor("pending", unapprovedPlan, nil)}, ""),
		Entry("landed but unchecked → verify", lifecycle.Context{
			Subject: subjectFor("pending", approvedPlan, withDoD),
			Runs:    []lifecycle.StepRun{{Step: "run", State: "succeeded"}},
		}, "verify"),
		Entry("landed and already verified this round → run", lifecycle.Context{
			Subject: subjectFor("unverified", approvedPlan, withDoD),
			Runs: []lifecycle.StepRun{
				{Step: "run", State: "succeeded", FinishedAt: at(0)},
				{Step: "verify", State: "succeeded", Outcome: "unverified", FinishedAt: at(time.Minute)},
			},
		}, "run"),
		Entry("landed after a stale verification → verify", lifecycle.Context{
			Subject: subjectFor("unverified", approvedPlan, withDoD),
			Runs: []lifecycle.StepRun{
				{Step: "run", State: "succeeded", FinishedAt: at(time.Hour)},
				{Step: "verify", State: "succeeded", Outcome: "unverified", FinishedAt: at(0)},
			},
		}, "verify"),
		Entry("failed with an approved plan → run", lifecycle.Context{Subject: subjectFor("failed", approvedPlan, nil)}, "run"),
		Entry("failed with no plan → plan", lifecycle.Context{Subject: subjectFor("failed", nil, nil)}, "plan"),
		Entry("review → nothing", lifecycle.Context{Subject: subjectFor("review", unapprovedPlan, nil)}, ""),
		Entry("ask → nothing", lifecycle.Context{Subject: subjectFor("ask", approvedPlan, nil)}, ""),
		Entry("verified → nothing", lifecycle.Context{Subject: subjectFor("verified", approvedPlan, withDoD)}, ""),
		Entry("completed → nothing", lifecycle.Context{Subject: subjectFor("completed", approvedPlan, withDoD)}, ""),
	)

	It("Applicable lists auxiliary steps that Next skips", func() {
		engine := defaultEngine()
		steps, err := engine.Applicable(lifecycle.Context{Subject: subjectFor("draft", nil, nil)})
		Expect(err).NotTo(HaveOccurred())
		var names []string
		for _, step := range steps {
			names = append(names, step.Name)
		}
		Expect(names).To(Equal([]string{"triage", "plan"}))
	})

	It("evaluates step inputs from the subject", func() {
		engine := defaultEngine()
		inputs, err := engine.Inputs(stepNamed(engine, "run"), lifecycle.Context{Subject: subjectFor("pending", approvedPlan, nil)})
		Expect(err).NotTo(HaveOccurred())
		Expect(inputs).To(Equal(map[string]any{"existingPlan": "# plan"}))
	})

	It("rejects a subject missing a declared field", func() {
		subject := subjectFor("draft", nil, nil)
		delete(subject, "verification")
		_, _, err := defaultEngine().Next(lifecycle.Context{Subject: subject})
		Expect(err).To(MatchError(ContainSubstring("missing declared field(s): verification")))
	})
})

var _ = Describe("default outcomes", func() {
	ctx := lifecycle.Context{Subject: subjectFor("pending", approvedPlan, withDoD)}
	succeeded := lifecycle.RunFacts{State: "succeeded", Iterations: 3, CostUSD: 0.5}
	passed := &api.VerifyReport{Ran: true, Passed: true}
	failedReport := &api.VerifyReport{Ran: true, Passed: false}

	outcome := func(step string, result lifecycle.StepResult) (string, error) {
		engine := defaultEngine()
		return engine.Outcome(stepNamed(engine, step), ctx, result)
	}

	DescribeTable("run",
		func(result lifecycle.StepResult, want string) {
			Expect(outcome("run", result)).To(Equal(want))
		},
		Entry("asks", lifecycle.StepResult{Run: succeeded, Envelope: lifecycle.Envelope{EndStatus: "ask"}}, "ask"),
		Entry("cancelled → pending", lifecycle.StepResult{Run: lifecycle.RunFacts{State: "cancelled"}}, "pending"),
		Entry("run failed", lifecycle.StepResult{Run: lifecycle.RunFacts{State: "failed", Error: "boom"}}, "failed"),
		Entry("agent reported failure", lifecycle.StepResult{Run: succeeded, Envelope: lifecycle.Envelope{EndStatus: "failed"}}, "failed"),
		Entry("verifier passed", lifecycle.StepResult{Run: succeeded, Envelope: lifecycle.Envelope{EndStatus: "completed"}, Verify: passed}, "verified"),
		Entry("verifier failed", lifecycle.StepResult{Run: succeeded, Envelope: lifecycle.Envelope{EndStatus: "completed"}, Verify: failedReport}, "unverified"),
		Entry("no verifier → pending", lifecycle.StepResult{Run: succeeded, Envelope: lifecycle.Envelope{EndStatus: "completed"}}, "pending"),
	)

	DescribeTable("plan",
		func(result lifecycle.StepResult, want string) {
			Expect(outcome("plan", result)).To(Equal(want))
		},
		Entry("new plan → review", lifecycle.StepResult{Run: succeeded, Plan: &lifecycle.PlanFacts{Status: "new", Content: "# plan"}}, "review"),
		Entry("updated plan → review", lifecycle.StepResult{Run: succeeded, Plan: &lifecycle.PlanFacts{Status: "updated"}}, "review"),
		Entry("unchanged plan → pending", lifecycle.StepResult{Run: succeeded, Plan: &lifecycle.PlanFacts{Status: "unchanged"}}, "pending"),
		Entry("questions → ask", lifecycle.StepResult{Run: succeeded, Plan: &lifecycle.PlanFacts{Status: "new"}, Questions: []any{map[string]any{"question": "which db?"}}}, "ask"),
		Entry("run failed", lifecycle.StepResult{Run: lifecycle.RunFacts{State: "failed"}}, "failed"),
		Entry("agent reported failure", lifecycle.StepResult{Run: succeeded, Envelope: lifecycle.Envelope{EndStatus: "failed"}}, "failed"),
	)

	It("plan with no plan status is an unknown outcome", func() {
		_, err := outcome("plan", lifecycle.StepResult{Run: succeeded})
		Expect(err).To(MatchError(ContainSubstring("no outcome matched")))
	})

	It("resolves the outcomes from the compiled definition, not the caller's step value", func() {
		engine := defaultEngine()
		result := lifecycle.StepResult{Run: succeeded, Envelope: lifecycle.Envelope{EndStatus: "completed"}, Verify: passed}

		status, err := engine.Outcome(lifecycle.Step{Name: "run"}, ctx, result)

		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal("verified"))
	})

	DescribeTable("verify",
		func(result lifecycle.StepResult, want string) {
			Expect(outcome("verify", result)).To(Equal(want))
		},
		Entry("passed", lifecycle.StepResult{Run: succeeded, Verify: passed}, "verified"),
		Entry("failed", lifecycle.StepResult{Run: succeeded, Verify: failedReport}, "unverified"),
		Entry("never ran", lifecycle.StepResult{Run: lifecycle.RunFacts{State: "failed", Error: "fixture missing"}}, "failed"),
	)

	DescribeTable("triage",
		func(result lifecycle.StepResult, want string) {
			Expect(outcome("triage", result)).To(Equal(want))
		},
		Entry("done keeps the status", lifecycle.StepResult{Run: succeeded, Envelope: lifecycle.Envelope{EndStatus: "completed"}}, lifecycle.OutcomeKeep),
		Entry("asks", lifecycle.StepResult{Run: succeeded, Envelope: lifecycle.Envelope{EndStatus: "ask"}}, "ask"),
		Entry("run failed", lifecycle.StepResult{Run: lifecycle.RunFacts{State: "failed"}}, "failed"),
	)
})

var _ = Describe("strict predicates", func() {
	base := func() lifecycle.Lifecycle {
		def, err := lifecycle.Default()
		Expect(err).NotTo(HaveOccurred())
		return def
	}

	It("rejects an undeclared subject field at compile time", func() {
		def := base()
		def.Steps[1].When = "subject.nope == 'x'"
		_, err := lifecycle.New(def)
		Expect(err).To(MatchError(ContainSubstring("subject.nope is not declared")))
	})

	It("rejects an undeclared top-level variable at compile time", func() {
		def := base()
		def.Steps[1].When = "todo.status == 'x'"
		_, err := lifecycle.New(def)
		Expect(err).To(MatchError(ContainSubstring("undeclared reference to 'todo'")))
	})

	It("rejects an outcome variable used in a when predicate", func() {
		def := base()
		def.Steps[1].When = "verify.passed"
		_, err := lifecycle.New(def)
		Expect(err).To(MatchError(ContainSubstring("undeclared reference to 'verify'")))
	})

	It("rejects a predicate that does not evaluate to a bool", func() {
		def := base()
		def.Steps[3].Outcomes = []lifecycle.Outcome{{Status: "pending", When: "envelope.summary"}}
		engine, err := lifecycle.New(def)
		Expect(err).NotTo(HaveOccurred())
		_, err = engine.Outcome(stepNamed(engine, "run"), lifecycle.Context{Subject: subjectFor("pending", nil, nil)},
			lifecycle.StepResult{Envelope: lifecycle.Envelope{Summary: "done"}})
		Expect(err).To(MatchError(ContainSubstring("did not return a bool")))
	})

	It("rejects a subject type the declaration language does not know", func() {
		def := base()
		def.Subject["labels"] = "set<string>"
		_, err := lifecycle.New(def)
		Expect(err).To(MatchError(ContainSubstring(`type "set<string>" is not one of`)))
	})

	It("rejects an outcome naming a status that is not a todo status", func() {
		def := base()
		def.Steps[3].Outcomes = append(def.Steps[3].Outcomes, lifecycle.Outcome{Status: "done", When: "true"})
		_, err := lifecycle.New(def)
		Expect(err).To(MatchError(ContainSubstring(`status "done" is not a todo status`)))
	})

	It("rejects a verify step that names a prompt other than the definition of done", func() {
		def := base()
		def.Steps[2].Prompt = "file:prompts/custom-verify.prompt"
		_, err := lifecycle.New(def)
		Expect(err).To(MatchError(ContainSubstring(`step verify: prompt "file:prompts/custom-verify.prompt"`)))
		Expect(err).To(MatchError(ContainSubstring("todos.verify")))
	})
})
