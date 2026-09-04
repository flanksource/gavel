package verifier_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	capverify "github.com/flanksource/captain/pkg/ai/agent/verify"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/verifier"
)

func TestFixtureVerifier(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fixture Verifier Suite")
}

const greenDocument = `---
codeBlocks: [bash]
---

# Definition of done

### command: unit tests

` + "```bash\necho ok\n```" + `

- cel: exitCode == 0
`

const redDocument = `---
codeBlocks: [bash]
---

# Definition of done

### command: unit tests

` + "```bash\necho 'boom: assertion failed' >&2\n```" + `

- cel: stdout.contains('all tests passed')
`

func verify(document, cwd string) (capverify.Verdict, error) {
	GinkgoHelper()
	plugins, err := verifier.New(context.Background(), api.Verify{Fixture: document}, capverify.Options{})
	Expect(err).ToNot(HaveOccurred())
	Expect(plugins).To(HaveLen(1))
	Expect(plugins[0].Name()).To(Equal("fixture"))
	return plugins[0].Verifier().Verify(context.Background(), cwd, nil)
}

var _ = Describe("fixture verifier registration", func() {
	It("claims captain's fixture verifier kind", func() {
		Expect(capverify.Registered(capverify.KindFixture)).To(BeTrue())
	})

	It("contributes no hooks when the workflow declares no fixture", func() {
		plugins, err := verifier.New(context.Background(), api.Verify{}, capverify.Options{})
		Expect(err).ToNot(HaveOccurred())
		Expect(plugins).To(BeEmpty())
	})

	It("rejects a document whose retry predicate does not compile", func() {
		_, err := verifier.NewVerifier("---\nverify:\n  retry: \"nope(\"\n---\n\n# doc\n")
		Expect(err).To(MatchError(ContainSubstring("retry predicate")))
	})
})

var _ = Describe("fixture verifier verdicts", func() {
	It("passes a green document and produces a valid passed report", func() {
		verdict, err := verify(greenDocument, GinkgoT().TempDir())

		Expect(err).ToNot(HaveOccurred())
		Expect(verdict.OK).To(BeTrue())
		Expect(verdict.Report).ToNot(BeNil())
		Expect(verdict.Report.Validate()).To(Succeed())
		Expect(verdict.Report.Kind).To(Equal("fixture"))
		Expect(verdict.Report.Passed).To(BeTrue())
		Expect(verdict.Report.State).To(Equal(api.VerifyStatePassed))
		Expect(verdict.Report.Summary.Passed).To(Equal(1))
		Expect(verdict.Report.Feedback).To(BeEmpty())
	})

	It("fails a document whose CEL expectation does not hold, with the node's evidence", func() {
		verdict, err := verify(redDocument, GinkgoT().TempDir())

		Expect(err).ToNot(HaveOccurred())
		Expect(verdict.OK).To(BeFalse())
		Expect(verdict.Report.Validate()).To(Succeed())
		Expect(verdict.Report.Passed).To(BeFalse())
		Expect(verdict.Report.State).To(Equal(api.VerifyStateFailed))
		Expect(verdict.Report.Summary.Failed).To(Equal(1))

		Expect(verdict.Report.Tests).To(HaveLen(1))
		node := verdict.Report.Tests[0]
		Expect(node.Name).To(Equal("unit tests"))
		Expect(node.Failed).To(BeTrue())
		Expect(node.Context).ToNot(BeNil())
		Expect(node.Context.CELExpression).To(Equal("stdout.contains('all tests passed')"))
		Expect(node.Context.CELVars).To(HaveKey("stdout"))
		Expect(node.Context.Cwd).ToNot(BeEmpty())

		Expect(verdict.Feedback).To(ContainSubstring("unit tests"))
		Expect(verdict.Feedback).To(ContainSubstring("boom: assertion failed"))
	})

	It("forwards execution snapshots to the progress sink as running reports", func() {
		var mu sync.Mutex
		var reports []api.VerifyReport
		v, err := verifier.NewVerifier(greenDocument)
		Expect(err).ToNot(HaveOccurred())
		v.SetProgress(func(report api.VerifyReport) {
			mu.Lock()
			defer mu.Unlock()
			reports = append(reports, report)
		})

		_, err = v.Verify(context.Background(), GinkgoT().TempDir(), nil)
		Expect(err).ToNot(HaveOccurred())

		mu.Lock()
		defer mu.Unlock()
		Expect(reports).ToNot(BeEmpty())
		var states []api.VerifyState
		for _, report := range reports {
			Expect(report.Validate()).To(Succeed())
			Expect(report.Passed).To(BeFalse())
			states = append(states, report.State)
		}
		Expect(states[0]).To(Equal(api.VerifyStateQueued))
		Expect(states).To(ContainElement(api.VerifyStateRunning))
	})

	It("lets a retry predicate flip a green run to unverified, naming itself", func() {
		document := "---\ncodeBlocks: [bash]\nverify:\n  retry: \"verify.summary.total < 2\"\n---\n" +
			greenDocumentBody

		verdict, err := verify(document, GinkgoT().TempDir())

		Expect(err).ToNot(HaveOccurred())
		Expect(verdict.OK).To(BeFalse())
		Expect(verdict.Report.Validate()).To(Succeed())
		Expect(verdict.Report.State).To(Equal(api.VerifyStateFailed))
		Expect(verdict.Reason).To(ContainSubstring("verify.summary.total < 2"))
	})

	It("keeps a green run green under a predicate that does not hold", func() {
		document := "---\ncodeBlocks: [bash]\nverify:\n  retry: \"verify.summary.failed > 0\"\n---\n" +
			greenDocumentBody

		verdict, err := verify(document, GinkgoT().TempDir())

		Expect(err).ToNot(HaveOccurred())
		Expect(verdict.OK).To(BeTrue())
		Expect(verdict.Report.Validate()).To(Succeed())
	})
})

const greenDocumentBody = `
# Definition of done

### command: unit tests

` + "```bash\necho ok\n```" + `

- cel: exitCode == 0
`

var _ = Describe("fixture verifier report mapping", func() {
	It("reports a step that never ran as errored rather than failed", func() {
		snapshot := &fixtures.ExecutionSnapshot{
			State: fixtures.ExecutionQueued,
			Root: &fixtures.ExecutionNode{
				Key: "fixtures", Name: "Fixtures", Kind: fixtures.ExecutionKindRoot,
				Children: []*fixtures.ExecutionNode{
					{Key: "a", Name: "ran", Kind: fixtures.ExecutionKindCommand, State: fixtures.ExecutionPassed},
					{Key: "b", Name: "never scheduled", Kind: fixtures.ExecutionKindCommand, State: fixtures.ExecutionQueued},
				},
			},
		}

		report := verifier.Report([]fixtures.FixtureResult{{Name: "ran", Status: task.StatusPASS}}, snapshot)

		Expect(report.State).To(Equal(api.VerifyStateErrored))
		Expect(report.Passed).To(BeFalse())
		Expect(report.Ran).To(BeFalse())
		Expect(report.Reason).To(ContainSubstring("never scheduled"))
		Expect(report.Tests).To(HaveLen(1), "the tree of what did run rides along with an errored report")
		Expect(report.Validate()).To(Succeed())
	})

	It("does not read a prose heading with no steps as a step that never ran", func() {
		snapshot := &fixtures.ExecutionSnapshot{
			State: fixtures.ExecutionPassed,
			Root: &fixtures.ExecutionNode{
				Key: "fixtures", Name: "Fixtures", Kind: fixtures.ExecutionKindRoot,
				Children: []*fixtures.ExecutionNode{
					{Key: "a", Name: "ran", Kind: fixtures.ExecutionKindCommand, State: fixtures.ExecutionPassed},
					{Key: "notes", Name: "Notes", Kind: fixtures.ExecutionKindSection, State: fixtures.ExecutionQueued},
				},
			},
		}

		report := verifier.Report([]fixtures.FixtureResult{{Name: "ran", Status: task.StatusPASS}}, snapshot)

		Expect(report.State).To(Equal(api.VerifyStatePassed))
		Expect(report.Passed).To(BeTrue())
		Expect(report.Validate()).To(Succeed())
	})

	It("carries an ai step's acceptance-criteria verdicts onto the checklist", func() {
		no, yes := false, true
		result := fixtures.FixtureResult{
			Name: "acceptance-criteria", Status: task.StatusFAIL,
			Error: "1/2 acceptance criteria not met",
			Metadata: map[string]any{"checklist": []fixtures.ChecklistResult{
				{Item: "the CLI exits 0", Passed: true, Message: "verified"},
				{Item: "docs updated", Passed: false, Message: "no doc change found"},
			}},
		}

		report := verifier.Report([]fixtures.FixtureResult{result}, nil)

		Expect(report.Checklist).To(Equal([]api.VerifyChecklistItem{
			{Item: "the CLI exits 0", Passed: &yes, Message: "verified"},
			{Item: "docs updated", Passed: &no, Message: "no doc change found"},
		}))
		Expect(report.Passed).To(BeFalse())
		Expect(report.Feedback).To(ContainSubstring("criterion not met: docs updated"))
		Expect(report.Validate()).To(Succeed())
	})

	It("maps a group result's children as children and counts only the leaves", func() {
		result := fixtures.FixtureResult{
			Name: "checks:test", Status: task.StatusFAIL,
			Children: []*fixtures.FixtureNode{
				{Name: "TestA", Results: &fixtures.FixtureResult{Name: "TestA", Status: task.StatusPASS}},
				{Name: "TestB", Results: &fixtures.FixtureResult{
					Name: "TestB", Status: task.StatusFAIL, Error: "want 2 got 3",
				}},
			},
		}

		report := verifier.Report([]fixtures.FixtureResult{result}, nil)

		Expect(report.Tests).To(HaveLen(1))
		Expect(report.Tests[0].Children).To(HaveLen(2))
		Expect(report.Summary).To(Equal(api.VerifySummary{Total: 2, Passed: 1, Failed: 1}))
		Expect(report.Feedback).To(ContainSubstring("checks:test › TestB"))
		Expect(report.Validate()).To(Succeed())
	})

	It("records the CEL trace and run artifact as node detail", func() {
		result := fixtures.FixtureResult{
			Name: "query", Status: task.StatusFAIL, CELTrace: "exitCode(1) == 0 -> false",
			Run: &fixtures.RunArtifact{RunID: "run-1", Kind: "test", Total: 3, Failed: 1},
		}

		report := verifier.Report([]fixtures.FixtureResult{result}, nil)

		var detail map[string]any
		Expect(json.Unmarshal(report.Tests[0].Detail, &detail)).To(Succeed())
		Expect(detail).To(HaveKeyWithValue("cel_trace", "exitCode(1) == 0 -> false"))
		Expect(detail).To(HaveKey("run"))
	})

	It("carries the run's wall clock from the execution snapshot", func() {
		started := time.Now().Add(-3 * time.Second)
		ended := started.Add(2 * time.Second)
		snapshot := &fixtures.ExecutionSnapshot{
			State: fixtures.ExecutionPassed, StartedAt: &started, EndedAt: &ended,
			Root: &fixtures.ExecutionNode{Key: "fixtures", Name: "Fixtures", State: fixtures.ExecutionPassed},
		}

		report := verifier.Report([]fixtures.FixtureResult{{Name: "ok", Status: task.StatusPASS}}, snapshot)

		Expect(report.Duration).To(Equal(2 * time.Second))
		Expect(report.Validate()).To(Succeed())
	})
})
