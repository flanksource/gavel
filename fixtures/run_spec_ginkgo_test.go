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
	var test fixtures.FixtureTest

	BeforeEach(func() {
		dispatched, received = false, nil
		test = fixtures.FixtureTest{Name: "review", AIStep: &fixtures.AIStepSpec{}}
		original := fixtures.AIStepRunner
		fixtures.AIStepRunner = func(test fixtures.FixtureTest, opts fixtures.RunOptions) fixtures.FixtureResult {
			dispatched, received = true, opts.Spec
			return fixtures.FixtureResult{Name: test.Name, Status: task.StatusPASS}
		}
		DeferCleanup(func() { fixtures.AIStepRunner = original })
	})

	It("passes the resolved runtime to the AI step", func() {
		expected := api.Spec{Budget: api.Budget{Cost: 3, MaxTurns: 12}}
		result := fixtures.RunNode(context.Background(), test, fixtures.RunOptions{
			ResolveSpec: func(context.Context) (api.Spec, error) { return expected, nil },
		})
		Expect(result.Status).To(Equal(task.StatusPASS))
		Expect(dispatched).To(BeTrue())
		Expect(received).To(Equal(&expected))
	})

	DescribeTable("does not resolve defaults for an authoritative fixture snapshot", func(snapshot api.Spec) {
		test.AI = &fixtures.FixtureAIConfig{Spec: &snapshot}
		called := false
		result := fixtures.RunNode(context.Background(), test, fixtures.RunOptions{
			ResolveSpec: func(context.Context) (api.Spec, error) {
				called = true
				return api.Spec{}, errors.New("selected profile is missing")
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
			ResolveSpec: func(context.Context) (api.Spec, error) {
				called = true
				return api.Spec{}, errors.New("runtime unavailable")
			},
		})
		Expect(result.Status).To(Equal(task.StatusSKIP))
		Expect(called).To(BeFalse())
		Expect(dispatched).To(BeFalse())
	})

	It("reports a resolution failure without calling the AI step", func() {
		result := fixtures.RunNode(context.Background(), test, fixtures.RunOptions{
			ResolveSpec: func(context.Context) (api.Spec, error) {
				return api.Spec{}, errors.New("selected profile is missing")
			},
		})
		Expect(result.Status).To(Equal(task.StatusERR))
		Expect(result.Error).To(Equal("resolve AI fixture runtime: selected profile is missing"))
		Expect(dispatched).To(BeFalse())
	})

	It("rejects competing eager and lazy runtime declarations", func() {
		result := fixtures.RunNode(context.Background(), test, fixtures.RunOptions{
			Spec: &api.Spec{},
			ResolveSpec: func(context.Context) (api.Spec, error) {
				Fail("the conflicting resolver must not run")
				return api.Spec{}, nil
			},
		})
		Expect(result.Status).To(Equal(task.StatusERR))
		Expect(result.Error).To(Equal("AI fixture cannot set both Spec and ResolveSpec"))
		Expect(dispatched).To(BeFalse())
	})
})
