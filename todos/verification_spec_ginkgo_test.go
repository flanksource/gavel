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

		reporter := verifier.executionReporter(context.Background())
		Expect(reporter.Publish(context.Background())).To(Succeed())
		_, err := verifier.runDeterministic(nil, reporter)
		Expect(err).NotTo(HaveOccurred())
		_, err = verifier.runChecklist(nil, reporter)
		Expect(err).NotTo(HaveOccurred())

		Expect(received).To(HaveLen(2))
		Expect(received[0]).To(BeIdenticalTo(&spec))
		Expect(received[1]).To(BeIdenticalTo(&spec))
	})

	It("uses workflow verification from the request spec for the outer verifier", func() {
		spec := api.Spec{Workflow: &api.Workflow{Verify: &api.Verify{MaxIterations: 4}}}

		verifiers, maxIterations, err := BuildCheckVerifiers(CheckVerifierOptions{
			WorkDir: GinkgoT().TempDir(),
			Todos: []*types.TODO{{
				AcceptanceCriteria: []types.AcceptanceCriterion{{Text: "The implementation passes."}},
			}},
			Run: &spec,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(verifiers).To(HaveLen(1))
		Expect(maxIterations).To(Equal(5))
	})

	// C1: the grader used to execute as the implementer — same model, same
	// backend, same session — so a cmux run spawned a TUI to mark its own work.
	// The run spec now decides only WHETHER to verify; the verify chain decides
	// what grades.
	It("grades on the verify spec, not the implementer's", func() {
		run := api.Spec{
			Model:     api.Model{Name: "claude-sonnet-5", Mode: api.ModeCmux},
			SessionID: "the-implementer-session",
			Workflow:  &api.Workflow{Verify: &api.Verify{MaxIterations: 2}},
		}
		grader := api.Spec{Model: api.Model{Name: "claude-code-sonnet"}}

		verifiers, _, err := BuildCheckVerifiers(CheckVerifierOptions{
			WorkDir: GinkgoT().TempDir(),
			Todos: []*types.TODO{{
				AcceptanceCriteria: []types.AcceptanceCriterion{{Text: "The implementation passes."}},
			}},
			Run:    &run,
			Grader: grader,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(verifiers).To(HaveLen(1))
		cel, ok := verifiers[0].(*celVerifier)
		Expect(ok).To(BeTrue())
		Expect(cel.spec.Name).To(Equal("claude-code-sonnet"))
		Expect(cel.spec.Mode).To(BeEmpty(), "the implementer's cmux mode must not grade")
		Expect(cel.spec.SessionID).To(BeEmpty(), "the grader never resumes the session it judges")
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
