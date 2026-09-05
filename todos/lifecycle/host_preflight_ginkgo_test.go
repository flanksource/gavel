package lifecycle_test

import (
	"context"
	"os"
	"path/filepath"

	capverify "github.com/flanksource/captain/pkg/ai/agent/verify"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	fixtureverifier "github.com/flanksource/gavel/fixtures/verifier"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type preflightLifecycleProvider struct {
	*fakeProvider
	todos.RunLifecycleProvider
	admissions int
	starts     int
}

func (p *preflightLifecycleProvider) PrepareRun(context.Context, *types.TODO, todos.RunPreparation) (todos.RunPreparationResult, error) {
	p.admissions++
	return todos.RunPreparationResult{}, nil
}

func (p *preflightLifecycleProvider) RecordRunStart(context.Context, *types.TODO, todos.RunStartMetadata) error {
	p.starts++
	return nil
}

var _ = Describe("Host full-input preflight", func() {
	var host *lifecycle.Host
	var provider *preflightLifecycleProvider
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		provider = &preflightLifecycleProvider{fakeProvider: &fakeProvider{}}
		host = newHost(provider.fakeProvider)
		host.Provider = provider
	})

	DescribeTable("rejects invalid runs during resolution before admission",
		func(request api.Spec, message string) {
			_, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "plan"), lifecycle.RunOptions{Request: request})
			Expect(err).To(MatchError(ContainSubstring(message)))
			Expect(provider.admissions).To(BeZero())
		},
		Entry("unprepared attachment", api.Spec{Prompt: api.Prompt{Attachments: []api.AttachmentRef{{Path: "review.txt"}}}}, "not resolved"),
		Entry("missing judge file", api.Spec{Workflow: &api.Workflow{Verify: &api.Verify{Prompts: []string{"missing-review.prompt"}}}}, "missing-review.prompt"),
	)

	DescribeTable("retains restrictive constraints through prompt rendering",
		func(constraints api.RuntimeConstraints, message string) {
			_, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "plan"), lifecycle.RunOptions{
				Prior: []api.SpecLayer{{Name: "account limits", Scope: api.SpecLayerUser, Source: api.SpecLayerSourcePreset,
					Spec: api.Spec{Budget: api.Budget{MaxTurns: 7}}, Constraints: constraints}},
			})
			Expect(err).To(MatchError(ContainSubstring(message)))
			Expect(provider.admissions).To(BeZero())
		},
		Entry("rendered input exceeds its limit", api.RuntimeConstraints{Limits: api.RunLimits{MaxInputTokens: 1}}, "exceeding the configured limit"),
		Entry("usage quota is exhausted", api.RuntimeConstraints{Quotas: []api.UsageQuota{{Name: "daily", Scope: api.SpecLayerUser, TokenLimit: 100, TokensUsed: 100}}}, "quota"),
	)

	It("returns permission warnings and provenance without invoking a broker, provider, setup or persistence", func() {
		agent := &scriptedProvider{}
		brokerCalls := 0
		marker := filepath.Join(host.WorkDir, "setup-ran")
		request := api.Spec{Permissions: api.Permissions{Plugins: api.ResourcePolicies{"review-tools": api.ResourceEnabled}}}
		request.Setup = &shell.Setup{Cwd: marker}
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "plan"), lifecycle.RunOptions{
			Request: request, Provider: agent,
			Broker: func(*todos.ExecutorContext) (api.PermissionFunc, error) { brokerCalls++; return nil, nil },
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Warnings).To(ContainElement(ContainSubstring("plugins")))
		Expect(resolution.Trace).To(ContainElement(HaveField("Source", api.SpecLayerSourceRequest)))
		Expect([]int{brokerCalls, len(agent.requests), provider.admissions, len(provider.states)}).To(Equal([]int{0, 0, 0, 0}))
		_, err = os.Stat(marker)
		Expect(os.IsNotExist(err)).To(BeTrue())
	})

	It("rechecks the actual provider before admitting a previously resolved run", func() {
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "plan"), lifecycle.RunOptions{
			Request: api.Spec{Permissions: api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyDeny}}},
		})
		Expect(err).NotTo(HaveOccurred())
		agent := &apiPreflightProvider{scriptedProvider: &scriptedProvider{}}
		_, err = host.Dispatch(context.Background(), hostTodo(), resolution, lifecycle.RunOptions{Provider: agent})
		Expect(err).To(MatchError(ContainSubstring("cannot enforce a per-tool policy")))
		Expect(provider.admissions).To(BeZero())
		Expect(provider.starts).To(BeZero())
		Expect(agent.requests).To(BeEmpty())
	})

	It("previews a model-free fixture without invoking its factory and executes it only after admission", func() {
		factoryCalls := 0
		capverify.Unregister(capverify.KindFixture)
		capverify.Register(capverify.KindFixture, func(ctx context.Context, spec api.Verify, options capverify.Options) ([]*capverify.Plugin, error) {
			factoryCalls++
			return fixtureverifier.New(ctx, spec, options)
		})
		DeferCleanup(func() {
			capverify.Unregister(capverify.KindFixture)
			capverify.Register(capverify.KindFixture, fixtureverifier.New)
		})
		host.Config.AI = api.Spec{}
		todo := hostTodo()
		resolution, err := host.Resolve(context.Background(), todo, stepNamed(host.Def, "verify"), lifecycle.RunOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Name).To(BeEmpty())
		Expect(resolution.Warnings).To(BeEmpty())
		Expect([]int{factoryCalls, provider.admissions, provider.starts}).To(Equal([]int{0, 0, 0}))
		outcome, err := host.Dispatch(context.Background(), todo, resolution, lifecycle.RunOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(outcome.Execution.DoD.Passed).To(BeTrue())
		Expect(factoryCalls).To(Equal(1))
		Expect(provider.admissions).To(Equal(1))
	})
})

type apiPreflightProvider struct{ *scriptedProvider }

func (*apiPreflightProvider) GetModel() string        { return "gpt-5" }
func (*apiPreflightProvider) GetRuntime() api.Runtime { return api.RuntimeOf(api.OpenAI, api.ModeAPI) }
