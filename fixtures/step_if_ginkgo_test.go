package fixtures

import (
	"context"
	"encoding/json"

	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("step if: gates and step-id roots", func() {
	var executed []string
	recording := func(name string, outputs map[string]any) fakeKind {
		return fakeKind{name: name, execute: func(_ context.Context, step *StepContext) (StepOutcome, error) {
			executed = append(executed, step.Spec.ID)
			return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}, Outputs: outputs}, nil
		}}
	}
	BeforeEach(func() {
		executed = nil
		useKind(recording("approve", map[string]any{"approved": false, "count": int64(3)}))
		useKind(recording("act", nil))
	})

	It("skips a step whose if: is false without failing the steps after it, recording the expression", func() {
		results := runSteps(
			StepSpec{ID: "approval", Kind: "approve"},
			StepSpec{ID: "purge", Kind: "act", If: "approval.approved"},
			StepSpec{ID: "report", Kind: "act", If: "approval.count == 3 && steps.approval.count == 3"},
		)

		Expect(executed).To(Equal([]string{"approval", "report"}))
		Expect(results[1].Status).To(Equal(task.StatusSKIP))
		Expect(results[1].Error).To(Equal("skipped (if: approval.approved = false)"))
		Expect(metadataJSON(results[1])).To(HaveKeyWithValue("if", map[string]any{"expression": "approval.approved", "value": false}))
		Expect(results[2].Status).To(Equal(task.StatusPASS), results[2].Error)
	})

	It("binds a skipped step as {skipped, reason}, so a later && short-circuits and a bare read fails naming the key", func() {
		results := runSteps(
			StepSpec{ID: "approval", Kind: "approve", If: "1 == 2"},
			StepSpec{ID: "purge", Kind: "act", If: "false && approval.approved"},
			StepSpec{ID: "report", Kind: "act", If: "approval.skipped && steps.approval.reason == 'if: 1 == 2'"},
			StepSpec{ID: "strict", Kind: "act", If: "approval.approved"},
		)

		Expect(executed).To(Equal([]string{"report"}))
		Expect(results[1].Status).To(Equal(task.StatusSKIP))
		Expect(results[1].Error).To(Equal("skipped (if: false && approval.approved = false)"))
		Expect(results[2].Status).To(Equal(task.StatusPASS), results[2].Error)
		Expect(results[3].Status).To(Equal(task.StatusERR))
		Expect(results[3].Error).To(ContainSubstring("approved"))
	})

	It("fails a step whose if: does not evaluate, and skips the steps after it", func() {
		results := runSteps(
			StepSpec{ID: "purge", Kind: "act", If: "missing.approved"},
			StepSpec{ID: "report", Kind: "act"},
		)

		Expect(executed).To(BeEmpty())
		Expect(results[0].Status).To(Equal(task.StatusERR))
		Expect(results[0].Error).To(ContainSubstring("if: missing.approved"))
		Expect(results[1].Error).To(Equal("skipped: step purge failed"))
	})

	It("traces the values an if: read when it does not evaluate", func() {
		results := runSteps(
			StepSpec{ID: "approval", Kind: "approve", If: "1 == 2"},
			StepSpec{ID: "clear", Kind: "act", If: "approval.approved && approval.clear"},
		)

		Expect(results[1].Status).To(Equal(task.StatusERR))
		Expect(results[1].Error).To(ContainSubstring("no such overload"))
		Expect(results[1].CELExpression).To(Equal("approval.approved && approval.clear"))
		Expect(results[1].CELTrace).To(ContainSubstring("cel: approval.approved && approval.clear"))
		Expect(results[1].CELTrace).To(ContainSubstring("null"))
	})

	It("renders an earlier step's outputs under its id in a later step's config", func() {
		var rendered map[string]any
		useKind(fakeKind{name: "consume", execute: func(_ context.Context, step *StepContext) (StepOutcome, error) {
			Expect(json.Unmarshal(step.Spec.Config, &rendered)).To(Succeed())
			return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}}, nil
		}})

		results := runSteps(
			StepSpec{ID: "approval", Kind: "approve"},
			StepSpec{ID: "use", Kind: "consume", Config: json.RawMessage(`{"n":"{{.approval.count}} of {{.steps.approval.count}}"}`)},
		)

		Expect(results[1].Status).To(Equal(task.StatusPASS), results[1].Error)
		Expect(rendered).To(Equal(map[string]any{"n": "3 of 3"}))
	})

	DescribeTable("refuses step ids that cannot be roots at load",
		func(document, message string) {
			_, err := ParseFixtureFile(writeDocument("ids.fixture.yaml", document))
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		Entry("an id taking a reserved root", "kind: Fixture\nsteps:\n  - { id: env, kind: act }\n",
			`step id "env" collides with the reserved root "env"`),
		Entry("a non-identifier id used as a CEL root", "kind: Fixture\nsteps:\n  - { id: stage-approval, kind: approve }\n  - { id: stage, kind: act, if: stage-approval.approved }\n",
			`step id "stage-approval" is used as a root by step "stage" but is not a valid CEL identifier; rename it (e.g. stageApproval)`),
		Entry("a non-identifier id used as a template root", "kind: Fixture\nsteps:\n  - { id: stage-approval, kind: approve }\n  - { id: stage, kind: act, n: \"{{.stage-approval.count}}\" }\n",
			`step id "stage-approval" is used as a root`),
	)

	It("loads a non-identifier id that no step reads as a root", func() {
		_, err := ParseFixtureFile(writeDocument("ids.fixture.yaml",
			"kind: Fixture\nsteps:\n  - { id: local-app, kind: approve }\n  - { id: use, kind: act, url: \"http://local-app.example/x\", if: \"steps['local-app'].approved == false\" }\n"))
		Expect(err).ToNot(HaveOccurred())
	})

	It("gates a markdown table row through its If column", func() {
		useKind(recording("exec", nil))
		tree, err := ParseMarkdownDocument("rows", "| Name | Command | If |\n| --- | --- | --- |\n| gated | echo hi | 1 == 2 |\n", GinkgoT().TempDir())
		Expect(err).ToNot(HaveOccurred())
		var test *FixtureTest
		tree.Walk(func(n *FixtureNode) {
			if n.Test != nil {
				test = n.Test
			}
		})
		Expect(test.StepSpec().If).To(Equal("1 == 2"))
		result := RunNode(context.Background(), *test, RunOptions{WorkDir: GinkgoT().TempDir()})
		Expect(result.Status).To(Equal(task.StatusSKIP))
		Expect(result.Error).To(Equal("skipped (if: 1 == 2 = false)"))
	})

	It("publishes a fixture blocked on a person as waiting, then running again", func() {
		var states []ExecutionState
		var waitingCount int
		useKind(fakeKind{name: "ask", execute: func(ctx context.Context, step *StepContext) (StepOutcome, error) {
			if err := ReportWaiting(ctx, true); err != nil {
				return StepOutcome{}, err
			}
			return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}}, ReportWaiting(ctx, false)
		}})
		ctx := WithProgressSink(context.Background(), func(_ context.Context, snapshot ExecutionSnapshot) error {
			states = append(states, snapshot.State)
			waitingCount = max(waitingCount, snapshot.Summary.Waiting)
			return nil
		})

		_, _, err := RunNodes(ctx, []*FixtureNode{stepNode(StepSpec{ID: "approval", Kind: "ask"})}, RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect(states).To(Equal([]ExecutionState{ExecutionQueued, ExecutionRunning, ExecutionWaiting, ExecutionRunning, ExecutionPassed}))
		Expect(waitingCount).To(Equal(1))
	})
})
