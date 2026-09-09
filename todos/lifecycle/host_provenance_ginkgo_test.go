package lifecycle

import (
	"context"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

type provenanceLifecycleProvider struct {
	todos.Provider
	todos.RunLifecycleProvider
	preparations []todos.RunPreparation
}

func (p *provenanceLifecycleProvider) PrepareRun(_ context.Context, _ *types.TODO, preparation todos.RunPreparation) (todos.RunPreparationResult, error) {
	p.preparations = append(p.preparations, preparation)
	return todos.RunPreparationResult{}, nil
}

var _ = g.Describe("lifecycle admission provenance", func() {
	g.It("passes the selected profile and ordered input layers beside the prepared execution spec", func() {
		profile := &runtimeprofiles.Resolution{Profile: runtimeprofiles.Profile{ID: "file:project:reviewer", Name: "reviewer"}}
		trace := []api.SpecLayer{
			{Name: "reviewer", Scope: api.SpecLayerContext, Source: api.SpecLayerSourceProfile, Spec: api.Spec{Budget: api.Budget{MaxTurns: 12}}},
			{Name: "request", Scope: api.SpecLayerUser, Source: api.SpecLayerSourceRequest, Spec: api.Spec{Setup: &shell.Setup{Cwd: "/work/request"}}},
		}
		request := api.Spec{Model: api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent}, Setup: &shell.Setup{Cwd: "/work/prepared"}}
		provider := &provenanceLifecycleProvider{}
		host := &Host{Provider: provider}
		exec := todos.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
		_, err := host.admit(exec, &types.TODO{ID: "example"}, Step{Name: "plan"}, &preparedStep{
			class: types.ModePlan, agent: "claude", request: request, runtimeProfile: profile, trace: trace,
		}, RunOptions{})
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(provider.preparations).To(o.HaveExactElements(o.SatisfyAll(
			o.HaveField("RuntimeProfile", o.BeIdenticalTo(profile)),
			o.HaveField("SpecTrace", o.Equal(trace)),
			o.HaveField("Spec", o.Equal(request)),
		)))
		o.Expect(trace[1].Spec.Setup.Cwd).To(o.Equal("/work/request"))
	})
})
