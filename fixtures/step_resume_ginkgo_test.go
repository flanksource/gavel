package fixtures

import (
	"context"
	"path/filepath"

	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("a run resumed from a step, whatever happened before it", func() {
	const (
		fromStep = "apply"
		earlier  = "lookup"
	)

	// resumeOver runs an earlier step, then fromStep, resumed over prior, and
	// returns both results, the steps that executed and what fromStep read.
	resumeOver := func(prior map[string]PriorStep, earlierSpec, fromSpec StepSpec) (results []FixtureResult, executed []string, seen string) {
		GinkgoHelper()
		useKind(fakeKind{name: "counted", execute: func(_ context.Context, step *StepContext) (StepOutcome, error) {
			executed = append(executed, step.Spec.ID)
			cfg, err := KindConfig[map[string]any](step)
			seen, _ = cfg["url"].(string)
			return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}}, err
		}})
		earlierSpec.ID, earlierSpec.Kind = earlier, "counted"
		fromSpec.ID, fromSpec.Kind = fromStep, "counted"
		resume := Resume{FromStep: fromStep, Prior: prior}
		results, _, err := RunNodes(context.Background(), []*FixtureNode{stepNode(earlierSpec), stepNode(fromSpec)},
			RunOptions{WorkDir: GinkgoT().TempDir(), Resume: &resume})
		Expect(err).ToNot(HaveOccurred())
		return results, executed, seen
	}

	DescribeTable("runs it and never the step before it",
		func(prior map[string]PriorStep, spec StepSpec, status task.Status, message string) {
			results, executed, _ := resumeOver(prior, spec, StepSpec{})

			Expect(executed).To(Equal([]string{fromStep}))
			Expect(results[0].Status).To(Equal(status), results[0].Error)
			Expect(results[0].Error).To(Equal(message))
			Expect(results[1].Status).To(Equal(task.StatusPASS), results[1].Error)
			Expect(metadataJSON(results[0])).To(HaveKeyWithValue("reused", map[string]any{"step": earlier, "from": fromStep}))
		},
		Entry("when it completed", map[string]PriorStep{earlier: {}}, StepSpec{}, task.StatusPASS, ""),
		Entry("when it failed", map[string]PriorStep{earlier: {Failed: true}}, StepSpec{}, task.StatusWarning, "failed in the resumed run"),
		Entry("when it was skipped", map[string]PriorStep{earlier: {Skipped: true}}, StepSpec{}, task.StatusSKIP, "skipped in the resumed run"),
		Entry("when it has no result there", map[string]PriorStep{}, StepSpec{}, task.StatusSKIP, "has no result in the resumed run"),
		Entry("when it ran in the background", map[string]PriorStep{earlier: {}}, StepSpec{Background: true}, task.StatusPASS, ""),
	)

	It("seeds what a failed step recorded", func() {
		prior := map[string]PriorStep{earlier: {Failed: true, Outputs: map[string]any{"cluster": "east"}}}

		results, _, seen := resumeOver(prior, StepSpec{}, StepSpec{Config: []byte(`{"url":"{{.steps.lookup.cluster}}"}`)})

		Expect(results[1].Status).To(Equal(task.StatusPASS), results[1].Error)
		Expect(seen).To(Equal("east"))
	})

	DescribeTable("fails a step that reads an output the resumed run does not have",
		func(prior map[string]PriorStep, from StepSpec) {
			results, executed, _ := resumeOver(prior, StepSpec{}, from)

			Expect(executed).To(BeEmpty())
			Expect(results[1].Status).To(Equal(task.StatusERR))
			Expect(results[1].Error).To(ContainSubstring(earlier))
		},
		Entry("in a template, of a failed step", map[string]PriorStep{earlier: {Failed: true}}, StepSpec{Config: []byte(`{"url":"{{.steps.lookup.cluster}}"}`)}),
		Entry("in a template, of a step with no result", map[string]PriorStep{}, StepSpec{Config: []byte(`{"url":"{{.steps.lookup.cluster}}"}`)}),
		Entry("in its if:, of a failed step", map[string]PriorStep{earlier: {Failed: true}}, StepSpec{If: "lookup.approved"}),
		Entry("in its if:, of a step with no result", map[string]PriorStep{}, StepSpec{If: "lookup.approved"}),
	)

	It("refuses a step the file does not have, running nothing", func() {
		executed := 0
		useKind(fakeKind{name: "counted", execute: func(context.Context, *StepContext) (StepOutcome, error) {
			executed++
			return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}}, nil
		}})
		path := writeDocument("unknown-step.fixture.yaml", `
kind: Fixture
steps:
  - { id: lookup, kind: counted }
`)
		runner, err := NewRunner(RunnerOptions{Paths: []string{path}, WorkDir: filepath.Dir(path), MaxWorkers: 1, Resume: &Resume{FromStep: fromStep}})
		Expect(err).ToNot(HaveOccurred())

		_, err = runner.Run()

		Expect(err).To(MatchError(ContainSubstring(`no step "apply" to resume from; steps: lookup`)))
		Expect(executed).To(BeZero())
	})

	It("gives a step skipped behind a failed one its step record, so it can be resumed from", func() {
		useKind(fakeKind{name: "counted", execute: func(context.Context, *StepContext) (StepOutcome, error) {
			return StepOutcome{Result: FixtureResult{Status: task.StatusFAIL}}, nil
		}})

		results, _, err := RunNodes(context.Background(),
			[]*FixtureNode{stepNode(StepSpec{ID: earlier, Kind: "counted"}), stepNode(StepSpec{ID: fromStep, Kind: "counted"})},
			RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect(results[1].Status).To(Equal(task.StatusSKIP))
		Expect(metadataJSON(results[1])).To(HaveKeyWithValue("step", map[string]any{"id": fromStep, "kind": "counted"}))
	})
})
