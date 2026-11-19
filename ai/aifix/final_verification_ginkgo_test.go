package aifix

import (
	"context"
	"testing"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"
	commonscontext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/linters"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRepairVerification(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Repair verification")
}

type taskRecordingProvider struct {
	*fakeStreaming
	started chan *task.Task
}

func (p taskRecordingProvider) ExecuteStream(ctx context.Context, req captainai.Request) (<-chan captainai.Event, error) {
	p.started <- task.TaskFromContext(ctx)
	return p.fakeStreaming.ExecuteStream(ctx, req)
}

var _ = Describe("Final repair verification", func() {
	It("schedules the repair turn and releases its worker for re-linting", func(ctx SpecContext) {
		task.SetMaxConcurrent(1)
		DeferCleanup(func() { task.SetMaxConcurrent(4) })
		started := make(chan *task.Task, 1)
		provider := taskRecordingProvider{&fakeStreaming{model: "gpt-5.5", runtime: fakeRuntime}, started}
		captainai.RegisterProvider(fakeRuntime, func(captainai.Config) (captainai.Provider, error) { return provider, nil })
		result, err := Run(ctx, Request{
			Initial:        resultsWith("fakelint", violation("sample.go", "missing comma", "syntax", 1)),
			MaxIterations:  1,
			AIConfig:       captainai.Config{Model: api.Model{Name: "gpt-5.5", Mode: fakeRuntime.Mode}, APIKey: "example-key"},
			AIRequestProto: captainai.Request{Prompt: api.Prompt{User: "Repair the syntax error"}},
			BuildRequest:   unexpectedRequestBuild,
			ReLint: func(ctx context.Context) ([]*linters.LinterResult, error) {
				return task.StartTask("re-lint", func(commonscontext.Context, *task.Task) ([]*linters.LinterResult, error) {
					return nil, nil
				}, task.WithContext(ctx), task.WithCancellationDrain()).GetResult()
			},
		})
		Expect(err).NotTo(HaveOccurred())
		var repair *task.Task
		Expect(started).To(Receive(&repair))
		Expect(repair).NotTo(BeNil())
		Expect(result.FinalResults).To(BeEmpty())
	}, SpecTimeout(5*time.Second))

	It("rechecks edits from the final allowed iteration", func() {
		provider := &fakeStreaming{model: "gpt-5.5", runtime: fakeRuntime}
		captainai.RegisterProvider(fakeRuntime, func(captainai.Config) (captainai.Provider, error) { return provider, nil })
		checks := 0
		result, err := Run(context.Background(), Request{
			Initial:        resultsWith("fakelint", violation("sample.go", "missing comma", "syntax", 1)),
			MaxIterations:  1,
			AIConfig:       captainai.Config{Model: api.Model{Name: "gpt-5.5", Mode: fakeRuntime.Mode}, APIKey: "example-key"},
			AIRequestProto: captainai.Request{Prompt: api.Prompt{User: "Repair the syntax error"}},
			BuildRequest:   unexpectedRequestBuild,
			ReLint:         func(context.Context) ([]*linters.LinterResult, error) { checks++; return nil, nil },
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(checks).To(Equal(1))
		Expect(result.FinalResults).To(BeEmpty())
	})
})
