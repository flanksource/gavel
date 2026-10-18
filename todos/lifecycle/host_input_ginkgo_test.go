package lifecycle

import (
	"context"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/tools"
	"github.com/flanksource/gavel/todos/types"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("lifecycle actual run input", func() {
	g.It("preflights real commit hooks and carries approval opt-in without constructing a broker", func() {
		dir := g.GinkgoT().TempDir()
		host := &Host{}
		todo := &types.TODO{ID: "review-example"}
		exec := todos.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
		prepared := &preparedStep{agent: "claude", class: types.ModeRun, workDir: dir, timeout: time.Minute,
			request: api.Spec{Model: api.Model{Name: "claude-sonnet-5", Provider: api.Anthropic, Mode: api.ModeAgent, NoCache: true},
				Prompt: api.Prompt{User: "review the changes"}, Setup: &shell.Setup{Cwd: dir},
				ToolPreferences: api.ToolPreferences{"review": api.ToolPolicyAsk},
				Workflow:        &api.Workflow{Commits: []api.Commit{{On: api.CommitOnRun, DryRun: true}}}},
			trace:      []api.SpecLayer{{Name: "request", Scope: api.SpecLayerUser, Source: api.SpecLayerSourceRequest}},
			provenance: map[string]api.FieldProvenance{"/prompt/user": {Source: api.FieldSource{Kind: api.FieldSourceLayer, Name: "request"}}},
			warnings:   []string{"example warning"},
		}
		input := mustRunInput(host.runInput(exec, todo, prepared, RunOptions{Approvals: true}))
		warnings, err := promptrun.Preflight(input.input)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(warnings).To(o.BeEmpty())
		o.Expect(input.input.CallerOwnsCommits).To(o.BeTrue())
		o.Expect(input.input.Resolved).To(o.Equal(api.ResolvedSpec{
			Spec: prepared.request, Trace: prepared.trace, Provenance: prepared.provenance, Warnings: prepared.warnings,
		}))
		o.Expect(input.input.Config.Model).To(o.Equal(prepared.request.Model))
		o.Expect(input.input.Config.NoCache).To(o.BeTrue())
		o.Expect(input.input.Verify.Timeout).To(o.Equal(time.Minute))
		o.Expect(input.input.Verify.Progress).To(o.BeNil(), "Captain records verification progress on a recorded run")
		o.Expect(input.input.Record).To(o.BeNil(), "nothing is recorded before the runtime admits the run")
		o.Expect(hookNames(input.input.Hooks)).To(o.Equal([]string{"commit:run", "gavel-run-env", "gavel-run-labels"}))
		o.Expect(prepared.request.Setup.Env).To(o.BeEmpty())
		o.Expect(input.input.Approvals).To(o.Equal(&promptrun.ApprovalOptions{RequestedBy: "gavel-dashboard"}))
		o.Expect(input.input.Config.OnApproval).To(o.BeNil())
		input.start(exec, todo, prepared)
		o.Expect(input.input.Config.OnApproval).To(o.BeNil(), "Captain binds the callback after admitting the run")
	})

	// The step's declared sets are the run's tools; one that raises approvals
	// brings the broker with it, whatever the caller asked for.
	g.It("attaches the step's tools and the broker an ask-policy tool needs", func() {
		host := &Host{Provider: &inputProvider{}}
		todo := &types.TODO{ID: "triage-example"}
		exec := todos.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
		prepared := &preparedStep{agent: "claude", class: types.ModePlan, timeout: time.Minute, tools: []string{tools.SetTodos},
			request: api.Spec{Model: api.Model{Name: "claude-sonnet-5", Provider: api.Anthropic, Mode: api.ModeAgent}}}

		input := mustRunInput(host.runInput(exec, todo, prepared, RunOptions{}))

		names := make([]string, len(input.input.Config.Tools))
		for i, def := range input.input.Config.Tools {
			names[i] = def.Name
		}
		o.Expect(names).To(o.ContainElements("todo_get", "todo_update", "todo_merge", "todo_close"))
		o.Expect(names).NotTo(o.ContainElement(tools.AskToolName), "the ask set was not declared")
		o.Expect(input.tools).NotTo(o.BeNil())
		o.Expect(input.input.Approvals).To(o.Equal(&promptrun.ApprovalOptions{RequestedBy: "gavel-dashboard"}))
	})

	// A run started from a terminal answers its own approvals there: the
	// in-process callback replaces the durable broker the ask-policy tools would
	// otherwise bring, and preflight admits it as the run's broker.
	g.It("answers approvals through the caller's in-process callback instead of the durable broker", func() {
		host := &Host{Provider: &inputProvider{}}
		exec := todos.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
		prepared := &preparedStep{agent: "claude", class: types.ModePlan, timeout: time.Minute, tools: []string{tools.SetTodos, tools.SetAsk},
			request: api.Spec{Model: api.Model{Name: "claude-sonnet-5", Provider: api.Anthropic, Mode: api.ModeAgent}, Prompt: api.Prompt{User: "plan"}}}
		var asked []api.ApprovalRequest
		terminal := func(_ context.Context, req api.ApprovalRequest) (api.ApprovalDecision, error) {
			asked = append(asked, req)
			return api.ApprovalDecision{Allow: true}, nil
		}

		input := mustRunInput(host.runInput(exec, &types.TODO{ID: "plan-example"}, prepared, RunOptions{OnApproval: terminal}))

		o.Expect(input.input.Approvals).To(o.BeNil())
		o.Expect(input.input.Config.OnApproval).NotTo(o.BeNil())
		_, err := promptrun.Preflight(input.input)
		o.Expect(err).NotTo(o.HaveOccurred())
		_, err = input.input.Config.OnApproval(context.Background(), api.ApprovalRequest{Tool: "todo_close", Kind: api.ApprovalKindTool})
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(asked).To(o.HaveLen(1))
	})

	g.It("refuses a run asking for both the in-process callback and the durable broker", func() {
		host := &Host{}
		exec := todos.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
		prepared := &preparedStep{agent: "claude", class: types.ModeRun, timeout: time.Minute,
			request: api.Spec{Model: api.Model{Name: "claude-sonnet-5", Provider: api.Anthropic, Mode: api.ModeAgent}}}
		allow := func(context.Context, api.ApprovalRequest) (api.ApprovalDecision, error) {
			return api.ApprovalDecision{Allow: true}, nil
		}

		_, err := host.runInput(exec, &types.TODO{ID: "run-example"}, prepared, RunOptions{Approvals: true, OnApproval: allow})

		o.Expect(err).To(o.MatchError(o.ContainSubstring("one approval broker")))
	})

	g.It("leaves a run without tools and without an approval request unbrokered", func() {
		host := &Host{}
		exec := todos.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
		prepared := &preparedStep{agent: "claude", class: types.ModeRun, timeout: time.Minute,
			request: api.Spec{Model: api.Model{Name: "claude-sonnet-5", Provider: api.Anthropic, Mode: api.ModeAgent}}}

		input := mustRunInput(host.runInput(exec, &types.TODO{ID: "run-example"}, prepared, RunOptions{}))

		o.Expect(input.input.Config.Tools).To(o.BeEmpty())
		o.Expect(input.input.Approvals).To(o.BeNil())
	})
})

// inputProvider is a provider with no label store: the tool sets need one to
// be built, and the taxonomy falls back to the built-in labels.
type inputProvider struct{ todos.Provider }

func mustRunInput(input *stepInput, err error) *stepInput {
	g.GinkgoHelper()
	o.Expect(err).NotTo(o.HaveOccurred())
	return input
}
