package lifecycle_test

import (
	"context"
	"fmt"
	"sync"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	// Registers gavel's fixture engine as captain's `fixture` verifier, which is
	// what the run step's definition of done dispatches to.
	_ "github.com/flanksource/gavel/fixtures/verifier"
)

// scriptedProvider is the streaming provider a step is dispatched through in
// place of a real agent: it emits one scripted turn and counts the requests it
// received, so a spec can read the prompt captain actually sent.
type scriptedProvider struct {
	mu       sync.Mutex
	events   []captainai.Event
	requests []api.Spec
}

func (p *scriptedProvider) GetModel() string { return "scripted-model" }
func (p *scriptedProvider) GetRuntime() api.Runtime {
	return api.RuntimeOf(api.Anthropic, api.ModeAgent)
}
func (p *scriptedProvider) Execute(context.Context, api.Spec) (*api.Response, error) {
	return nil, fmt.Errorf("the scripted provider only streams")
}
func (p *scriptedProvider) ExecuteStream(_ context.Context, req api.Spec) (<-chan captainai.Event, error) {
	p.mu.Lock()
	p.requests = append(p.requests, req)
	p.mu.Unlock()
	ch := make(chan captainai.Event, len(p.events))
	for _, ev := range p.events {
		ch <- ev
	}
	close(ch)
	return ch, nil
}

// scriptedTurn is one agent turn: the session announced, the envelope as the
// final text, and the provider's result event.
func scriptedTurn(envelope string, success bool) []captainai.Event {
	return []captainai.Event{
		{Kind: captainai.EventSystem, SessionID: "sess-scripted"},
		{Kind: captainai.EventText, Text: envelope},
		{Kind: captainai.EventResult, Success: success, Usage: &api.Usage{InputTokens: 7, OutputTokens: 3}, CostUSD: 0.25},
	}
}

var _ = Describe("RunStep through a scripted provider", func() {
	var (
		provider *fakeProvider
		host     *lifecycle.Host
		run      lifecycle.Step
		ctx      context.Context
	)
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		provider = &fakeProvider{plan: todos.PlanState{Exists: true, Approved: true, Content: "# plan", Revision: 2}}
		host = newHost(provider)
		ctx = context.Background()

		// The built-in run step checks out a worktree and commits; neither has a
		// repository to act on here. The definition of done stays: it is what the
		// verified path proves.
		def := host.Def.Definition()
		for i := range def.Steps {
			if def.Steps[i].Name == "run" {
				def.Steps[i].Spec = &api.Spec{Workflow: &api.Workflow{Verify: &api.Verify{
					Fixture: "{{subject.verification.document}}", Scope: api.VerifyScopeAll,
				}}}
			}
		}
		engine, err := lifecycle.New(def)
		Expect(err).NotTo(HaveOccurred())
		host.Def = engine
		run = stepNamed(engine, "run")
	})

	It("lands a run whose definition of done passed as verified", func() {
		agent := &scriptedProvider{events: scriptedTurn(`{"summary":"Built it.","endStatus":"completed"}`, true)}
		todo := hostTodo()

		outcome, err := host.RunStep(ctx, todo, run, lifecycle.RunOptions{Provider: agent})

		Expect(err).NotTo(HaveOccurred())
		Expect(outcome.Status).To(Equal("verified"))
		Expect(outcome.Result.Run.State).To(Equal(lifecycle.RunSucceeded))
		Expect(outcome.Result.Run.StopReason).To(Equal("condition-met"))
		Expect(outcome.Result.Verify).NotTo(BeNil())
		Expect(outcome.Result.Verify.Ran).To(BeTrue())
		Expect(outcome.Result.Verify.Passed).To(BeTrue())
		Expect(outcome.Execution.Success).To(BeTrue())
		Expect(outcome.Execution.Summary).To(Equal("Built it."))
		Expect(outcome.Execution.DoD).NotTo(BeNil())
		Expect(outcome.Execution.DoD.Passed).To(BeTrue())
		Expect(outcome.Execution.CostUSD).To(Equal(0.25))
		Expect(agent.requests).To(HaveLen(1), "one generate turn, then the fixture verified it")
		Expect(agent.requests[0].Prompt.User).To(ContainSubstring("Implement the thing"))
		// The turn's row is what the attempt listing reads its verification from;
		// it is filed before the outcome so the report exists by the time the
		// status that depends on it does.
		Expect(provider.iterations).To(HaveLen(1))
		Expect(provider.iterations[0].Iteration).To(Equal(1))
		Expect(provider.iterations[0].State).To(Equal(captaindb.PromptRunIterationStateSucceeded))
		Expect(provider.iterations[0].VerificationResult).NotTo(BeNil())
		Expect(provider.iterations[0].VerificationResult.Passed).To(BeTrue())

		Expect(host.OnOutcome(ctx, todo, run, outcome, outcome.Status)).To(Succeed())

		Expect(todo.Status).To(Equal(types.StatusVerified))
		Expect(provider.attempts).To(HaveLen(1))
		Expect(*provider.states[0].Status).To(Equal(types.StatusVerified))
		Expect(*provider.states[0].Attempts).To(Equal(3))
		Expect(provider.events).To(HaveLen(1))
		Expect(provider.events[0].Payload).To(HaveKeyWithValue("status", "verified"))
	})

	It("lands a run whose provider result was not a success as failed", func() {
		agent := &scriptedProvider{events: scriptedTurn(`{"summary":"Built it.","endStatus":"completed"}`, false)}
		todo := hostTodo()

		outcome, err := host.RunStep(ctx, todo, run, lifecycle.RunOptions{Provider: agent})

		Expect(err).NotTo(HaveOccurred())
		Expect(outcome.Status).To(Equal("failed"))
		Expect(outcome.Result.Run.State).To(Equal(lifecycle.RunFailed))
		Expect(outcome.Result.Run.Error).NotTo(BeEmpty())
		Expect(outcome.Execution.Success).To(BeFalse())
		Expect(outcome.Execution.ErrorMessage).NotTo(BeEmpty())

		Expect(host.OnOutcome(ctx, todo, run, outcome, outcome.Status)).To(Succeed())

		Expect(todo.Status).To(Equal(types.StatusFailed))
		Expect(provider.events[0].Payload).To(HaveKeyWithValue("status", "failed"))
	})

	// A verify-only step runs no agent turn — captain's loop never starts — so
	// the only record of its verdict is the iteration row the host files. This
	// is the `gavel todos check` path: without that row the dashboard showed a
	// passed check as an errored attempt with no report.
	It("files a verify-only step's verdict as iteration 1", func() {
		agent := &scriptedProvider{}
		todo := hostTodo()
		verifyStep := stepNamed(host.Def, "verify")

		outcome, err := host.RunStep(ctx, todo, verifyStep, lifecycle.RunOptions{Provider: agent})

		Expect(err).NotTo(HaveOccurred())
		Expect(outcome.Status).To(Equal("verified"))
		Expect(outcome.Result.Run.State).To(Equal(lifecycle.RunSucceeded))
		Expect(outcome.Execution.DoD).NotTo(BeNil())
		Expect(outcome.Execution.DoD.Passed).To(BeTrue())
		Expect(agent.requests).To(BeEmpty(), "a verify-only step generates nothing")
		Expect(provider.iterations).To(HaveLen(1))
		Expect(provider.iterations[0].Iteration).To(Equal(1))
		Expect(provider.iterations[0].State).To(Equal(captaindb.PromptRunIterationStateSucceeded))
		Expect(provider.iterations[0].VerificationResult).NotTo(BeNil())
		Expect(provider.iterations[0].VerificationResult.Passed).To(BeTrue())
		Expect(provider.iterations[0].VerificationResult.Ran).To(BeTrue())
	})
})
