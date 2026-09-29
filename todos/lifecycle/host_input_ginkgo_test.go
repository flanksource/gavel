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
		input := host.runInput(exec, todo, prepared, RunOptions{Approvals: true})
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
		o.Expect(hookNames(input.input.Hooks)).To(o.Equal([]string{"commit:run", "gavel-run-env"}))
		o.Expect(prepared.request.Setup.Env).To(o.BeEmpty())
		o.Expect(input.input.Approvals).To(o.Equal(&promptrun.ApprovalOptions{RequestedBy: "gavel-dashboard"}))
		o.Expect(input.input.Config.OnApproval).To(o.BeNil())
		input.start(exec, todo, prepared)
		o.Expect(input.input.Config.OnApproval).To(o.BeNil(), "Captain binds the callback after admitting the run")
	})
})
