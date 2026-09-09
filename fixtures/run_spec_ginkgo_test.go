package fixtures_test

import (
	"context"
	"errors"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("lazy fixture runtime spec", func() {
	var dispatched bool
	var received *api.Spec
	var runtime *api.ResolveSpecOptions
	var test fixtures.FixtureTest

	BeforeEach(func() {
		dispatched, received = false, nil
		runtime = nil
		test = fixtures.FixtureTest{Name: "review", AIStep: &fixtures.AIStepSpec{Criteria: []fixtures.ChecklistItem{{Text: "The change has evidence."}}}}
		original := fixtures.AIStepRunner
		fixtures.AIStepRunner = func(test fixtures.FixtureTest, opts fixtures.RunOptions) fixtures.FixtureResult {
			dispatched, received = true, opts.Spec
			runtime = opts.Runtime
			return fixtures.FixtureResult{Name: test.Name, Status: task.StatusPASS}
		}
		DeferCleanup(func() { fixtures.AIStepRunner = original })
	})

	It("passes raw runtime layers to the AI step without pre-filling a spec", func() {
		expected := api.ResolveSpecOptions{Layers: []api.SpecLayer{api.PromptSpecLayer("profile", api.Spec{Budget: api.Budget{Cost: 3, MaxTurns: 12}})}}
		result := fixtures.RunNode(context.Background(), test, fixtures.RunOptions{
			ResolveSpec: func(context.Context) (api.ResolveSpecOptions, error) { return expected, nil },
		})
		Expect(result.Status).To(Equal(task.StatusPASS))
		Expect(dispatched).To(BeTrue())
		Expect(received).To(BeNil())
		Expect(runtime).To(Equal(&expected))
	})

	DescribeTable("does not resolve defaults for an authoritative fixture snapshot", func(snapshot api.Spec) {
		test.AI = &fixtures.FixtureAIConfig{Spec: &snapshot}
		called := false
		result := fixtures.RunNode(context.Background(), test, fixtures.RunOptions{
			ResolveSpec: func(context.Context) (api.ResolveSpecOptions, error) {
				called = true
				return api.ResolveSpecOptions{}, errors.New("selected profile is missing")
			},
		})
		Expect(result.Status).To(Equal(task.StatusPASS))
		Expect(called).To(BeFalse())
		Expect(dispatched).To(BeTrue())
		Expect(received).To(BeNil())
	},
		Entry("a configured snapshot", api.Spec{Budget: api.Budget{MaxTurns: 12}}),
		Entry("an explicitly empty snapshot", api.Spec{}),
	)

	It("does not resolve the runtime for a skipped AI step", func() {
		test.TestOS = "unsupported-test-os"
		called := false
		result := fixtures.RunNode(context.Background(), test, fixtures.RunOptions{
			ResolveSpec: func(context.Context) (api.ResolveSpecOptions, error) {
				called = true
				return api.ResolveSpecOptions{}, errors.New("runtime unavailable")
			},
		})
		Expect(result.Status).To(Equal(task.StatusSKIP))
		Expect(called).To(BeFalse())
		Expect(dispatched).To(BeFalse())
	})

	It("does not load runtime defaults when an AI step has no checklist to execute", func() {
		test.AIStep.Criteria = nil
		called := false
		result := fixtures.RunNode(context.Background(), test, fixtures.RunOptions{
			ResolveSpec: func(context.Context) (api.ResolveSpecOptions, error) {
				called = true
				return api.ResolveSpecOptions{}, errors.New("saved defaults are malformed")
			},
		})
		Expect(called).To(BeFalse())
		Expect(result.Status).To(Equal(task.StatusPASS))
		Expect(dispatched).To(BeTrue(), "the AI runner owns the empty-checklist skip result")
	})

	It("reports a resolution failure without calling the AI step", func() {
		result := fixtures.RunNode(context.Background(), test, fixtures.RunOptions{
			ResolveSpec: func(context.Context) (api.ResolveSpecOptions, error) {
				return api.ResolveSpecOptions{}, errors.New("selected profile is missing")
			},
		})
		Expect(result.Status).To(Equal(task.StatusERR))
		Expect(result.Error).To(Equal("resolve AI fixture runtime: selected profile is missing"))
		Expect(dispatched).To(BeFalse())
	})

	It("uses an explicit runner snapshot without consulting lazy defaults", func() {
		snapshot := api.Spec{Budget: api.Budget{MaxTurns: 12}}
		result := fixtures.RunNode(context.Background(), test, fixtures.RunOptions{
			Spec: &snapshot,
			ResolveSpec: func(context.Context) (api.ResolveSpecOptions, error) {
				Fail("the conflicting resolver must not run")
				return api.ResolveSpecOptions{}, nil
			},
		})
		Expect(result.Status).To(Equal(task.StatusPASS))
		Expect(dispatched).To(BeTrue())
		Expect(received).To(Equal(&snapshot))
	})
})
