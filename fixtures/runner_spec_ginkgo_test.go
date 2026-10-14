package fixtures

import (
	"context"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"
	flanksourceContext "github.com/flanksource/commons/context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture runner AI spec", func() {
	It("passes the runner spec to an embedded AI fixture", func() {
		spec := api.Spec{Model: api.Model{Name: "verification-model"}}
		var received *api.Spec
		useKind(fakeKind{name: "ai", execute: func(_ context.Context, step *StepContext) (StepOutcome, error) {
			received = step.Opts.Spec
			return StepOutcome{Result: FixtureResult{Name: step.Test.Name, Status: task.StatusPASS}}, nil
		}})
		runner := &Runner{options: RunnerOptions{WorkDir: GinkgoT().TempDir(), Spec: &spec}}

		_, err := runner.executeFixture(
			flanksourceContext.NewContext(context.Background()),
			FixtureTest{Name: "embedded prompt", AIStep: &AIStepSpec{Criteria: []ChecklistItem{{Text: "passes"}}}},
			fixtureEnv{},
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(received).To(BeIdenticalTo(&spec))
	})
})
