package lifecycle

import (
	"context"
	"time"

	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

type admissionLifecycleProvider struct {
	todos.Provider
	admission    todos.RunAdmission
	preparations []todos.RunPreparation
	admitted     []todos.RunPreparationResult
}

func (p *admissionLifecycleProvider) PrepareRun(_ context.Context, _ *types.TODO, preparation todos.RunPreparation) (todos.RunAdmission, error) {
	p.preparations = append(p.preparations, preparation)
	return p.admission, nil
}

func (p *admissionLifecycleProvider) RunAdmitted(_ context.Context, _ *types.TODO, admission todos.RunPreparationResult) error {
	p.admitted = append(p.admitted, admission)
	return nil
}

var _ = g.Describe("lifecycle admission", func() {
	var (
		provider *admissionLifecycleProvider
		host     *Host
		exec     *todos.ExecutorContext
		prepared *preparedStep
		identity todos.RunPreparationResult
	)
	g.BeforeEach(func() {
		identity = todos.RunPreparationResult{SessionID: uuid.NewString(), PromptRunID: uuid.New()}
		provider = &admissionLifecycleProvider{admission: todos.RunAdmission{
			RunPreparationResult: identity, Record: &promptrun.Recording{Origin: "gavel.todos"},
		}}
		host = &Host{Provider: provider}
		exec = todos.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
		prepared = &preparedStep{
			class: types.ModePlan, agent: "claude",
			request: api.Spec{
				Model: api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent}, Prompt: api.Prompt{User: "Draft the plan"},
				Setup: &shell.Setup{Cwd: "/work/prepared"},
			},
		}
	})

	g.It("asks the runtime for the step, its prompt and the requested runtime, and hands Captain its recording", func() {
		input := host.runInput(exec, &types.TODO{ID: "example"}, prepared, RunOptions{Provider: suppliedProvider{}})

		admission, err := host.admit(exec, &types.TODO{ID: "example"}, Step{Name: "plan"}, prepared, input, RunOptions{})

		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(admission).To(o.Equal(identity))
		o.Expect(provider.preparations).To(o.HaveExactElements(o.SatisfyAll(
			o.HaveField("Mode", types.ModePlan),
			o.HaveField("Prompt", "plan"),
			o.HaveField("ExecutorName", "agent-claude"),
			o.HaveField("PromptMarkdown", "Draft the plan"),
			o.HaveField("Requested.Model", "claude-sonnet-5"),
		)))
		o.Expect(input.input.Record).To(o.BeIdenticalTo(provider.admission.Record))
		o.Expect(todos.PromptRunFromContext(exec)).To(o.Equal(identity.PromptRunID))
		o.Expect(hookNames(input.input.Hooks)).To(o.Equal([]string{"gavel-run-env", "gavel-run-admitted", "setup"}))
	})

	g.It("tells the runtime and the dashboard of the run only once Captain has admitted it", func() {
		var notified []todos.RunPreparationResult
		exec.SetRunPreparedHook(func(result todos.RunPreparationResult) { notified = append(notified, result) })
		input := host.runInput(exec, &types.TODO{ID: "example"}, prepared, RunOptions{})
		_, err := host.admit(exec, &types.TODO{ID: "example"}, Step{Name: "plan"}, prepared, input, RunOptions{})
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect([]int{len(provider.admitted), len(notified)}).To(o.Equal([]int{0, 0}))

		admitted := input.input.Hooks[len(input.input.Hooks)-1].(agent.PreRun)
		o.Expect(admitted.PreRun(&agent.HookContext{})).To(o.Succeed())

		o.Expect(provider.admitted).To(o.Equal([]todos.RunPreparationResult{identity}))
		o.Expect(notified).To(o.Equal([]todos.RunPreparationResult{identity}))
	})

	g.It("reports a run Captain refused to admit as the dispatch's error, leaving no outcome to record", func() {
		provider.admission.Record = &promptrun.Recording{Origin: "gavel.todos"}
		prepared.timeout = time.Minute

		outcome, err := host.Dispatch(context.Background(), &types.TODO{ID: "example"},
			&Resolution{Step: Step{Name: "plan"}, prepared: prepared}, RunOptions{Exec: exec, Provider: suppliedProvider{}})

		o.Expect(err).To(o.MatchError(o.ContainSubstring("recording has no database")))
		o.Expect(outcome).To(o.BeNil())
		o.Expect(provider.admitted).To(o.BeEmpty(), "nothing was admitted for the runtime to drive")
	})

	g.It("refuses a runtime that cleared the run without a recording", func() {
		provider.admission.Record = nil
		input := host.runInput(exec, &types.TODO{ID: "example"}, prepared, RunOptions{})

		_, err := host.admit(exec, &types.TODO{ID: "example"}, Step{Name: "plan"}, prepared, input, RunOptions{})

		o.Expect(err).To(o.MatchError(o.ContainSubstring("returned no recording")))
		o.Expect(input.input.Record).To(o.BeNil())
	})
})

// suppliedProvider stands in for a caller-supplied provider: promptrun then
// builds no setup hook, so gavel adds its own.
type suppliedProvider struct{}

func (suppliedProvider) Execute(context.Context, api.Spec) (*api.Response, error) {
	return nil, context.Canceled
}
func (suppliedProvider) GetModel() string        { return "supplied-model" }
func (suppliedProvider) GetRuntime() api.Runtime { return api.RuntimeOf(api.Anthropic, api.ModeAgent) }

func hookNames(hooks []any) []string {
	names := make([]string, 0, len(hooks))
	for _, hook := range hooks {
		names = append(names, hook.(interface{ Name() string }).Name())
	}
	return names
}
