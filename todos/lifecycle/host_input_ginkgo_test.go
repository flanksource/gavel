package lifecycle

import (
	"context"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("lifecycle actual run input", func() {
	g.It("preflights real commit hooks and defers the actual approval factory until admission is available", func() {
		dir := g.GinkgoT().TempDir()
		host := &Host{}
		todo := &types.TODO{ID: "review-example"}
		exec := todos.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
		prepared := &preparedStep{agent: "claude", class: types.ModeRun, workDir: dir, timeout: time.Minute,
			request: api.Spec{Model: api.Model{Name: "claude-sonnet-5", Provider: api.Anthropic, Mode: api.ModeAgent, NoCache: true},
				Prompt: api.Prompt{User: "review the changes"}, Setup: &shell.Setup{Cwd: dir},
				ToolPreferences: api.ToolPreferences{"review": api.ToolPolicyAsk},
				Workflow:        &api.Workflow{Commits: []api.Commit{{On: api.CommitOnRun, DryRun: true}}}},
		}
		factoryCalls, approvalCalls := 0, 0
		var seen api.PermissionRequest
		input := host.runInput(exec, todo, prepared, RunOptions{Broker: func(actual *todos.ExecutorContext) (api.PermissionFunc, error) {
			factoryCalls++
			o.Expect(actual).To(o.BeIdenticalTo(exec))
			return func(_ context.Context, request api.PermissionRequest) (api.PermissionDecision, error) {
				approvalCalls++
				seen = request
				return api.PermissionDecision{Allow: true, Message: "approved"}, nil
			}, nil
		}})
		warnings, err := promptrun.Preflight(input.input)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(warnings).To(o.BeEmpty())
		o.Expect(input.input.CallerOwnsCommits).To(o.BeTrue())
		o.Expect(input.input.Config.Model).To(o.Equal(prepared.request.Model))
		o.Expect(input.input.Config.NoCache).To(o.BeTrue())
		o.Expect(input.input.Verify.Timeout).To(o.Equal(time.Minute))
		o.Expect(input.input.Verify.Progress).NotTo(o.BeNil())
		var hookNames []string
		for _, hook := range input.input.Hooks {
			hookNames = append(hookNames, hook.(interface{ Name() string }).Name())
		}
		o.Expect(hookNames).To(o.Equal([]string{"commit:run", "gavel-run-env", "gavel-spec-recorder"}))
		o.Expect(prepared.request.Setup.Env).To(o.BeEmpty())
		o.Expect([]int{factoryCalls, approvalCalls}).To(o.Equal([]int{0, 0}))
		o.Expect(input.start(exec, todo, prepared)).To(o.Succeed())
		request := api.PermissionRequest{Tool: "review", SessionID: "admitted-session"}
		decision, err := input.input.Config.CanUseTool(exec, request)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(decision).To(o.Equal(api.PermissionDecision{Allow: true, Message: "approved"}))
		o.Expect(seen).To(o.Equal(request))
		o.Expect([]int{factoryCalls, approvalCalls}).To(o.Equal([]int{1, 1}))
	})
})
