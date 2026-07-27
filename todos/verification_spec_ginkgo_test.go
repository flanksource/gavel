package todos

import (
	"context"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TODO verification runtime spec", func() {
	It("passes one request spec through parsed and generated AI fixture steps", func() {
		original := fixtures.AIStepRunner
		DeferCleanup(func() { fixtures.AIStepRunner = original })

		spec := api.Spec{Model: api.Model{Name: "verification-model"}}
		var received []*api.Spec
		fixtureRunner := func(fixture fixtures.FixtureTest, opts fixtures.RunOptions) fixtures.FixtureResult {
			received = append(received, opts.Spec)
			return fixtures.FixtureResult{
				Name:     fixture.Name,
				Status:   task.StatusPASS,
				Metadata: map[string]interface{}{"checklist": []fixtures.ChecklistResult{}},
			}
		}
		fixtures.AIStepRunner = fixtureRunner
		verifier := &celVerifier{
			workDir: GinkgoT().TempDir(),
			spec:    &spec,
			nodes: []*fixtures.FixtureNode{{
				Name: "persisted fixture",
				Type: fixtures.TestNode,
				Test: &fixtures.FixtureTest{
					Name:   "persisted fixture",
					AIStep: &fixtures.AIStepSpec{Criteria: []fixtures.ChecklistItem{{Text: "passes"}}},
				},
			}},
			aiStep: &stepFixture{
				runner: fixtureRunner,
				fixture: fixtures.FixtureTest{
					Name:   "acceptance criteria",
					AIStep: &fixtures.AIStepSpec{Criteria: []fixtures.ChecklistItem{{Text: "passes"}}},
				},
			},
		}

		verifier.runDeterministic(nil)
		verifier.runChecklist(nil)

		Expect(received).To(HaveLen(2))
		Expect(received[0]).To(BeIdenticalTo(&spec))
		Expect(received[1]).To(BeIdenticalTo(&spec))
	})

	It("uses workflow verification from the request spec for the outer verifier", func() {
		spec := api.Spec{Workflow: &api.Workflow{Verify: &api.Verify{MaxIterations: 4}}}

		verifiers, maxIterations, err := BuildCheckVerifiers(
			GinkgoT().TempDir(),
			[]*types.TODO{{
				AcceptanceCriteria: []types.AcceptanceCriterion{{Text: "The implementation passes."}},
			}},
			&spec,
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(verifiers).To(HaveLen(1))
		Expect(maxIterations).To(Equal(5))
	})

	It("fails before executing when the request timeout is invalid", func() {
		spec := api.Spec{Budget: api.Budget{Timeout: "later"}}

		result := CheckTODO(context.Background(), &types.TODO{}, CheckOptions{
			WorkDir: GinkgoT().TempDir(),
			Spec:    &spec,
		})

		Expect(result.AllPassed).To(BeFalse())
		Expect(result.Error).To(MatchError(ContainSubstring("budget.timeout")))
	})
})
