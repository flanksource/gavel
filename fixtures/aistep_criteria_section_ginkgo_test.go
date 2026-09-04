package fixtures

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// A TODO's definition of done is one document holding both an acceptance-criteria
// checklist and ordinary fixture steps. Those steps carry `- cel:` expectation
// bullets of their own; read whole-document, one of them becomes the AI step's
// CEL expectation and is evaluated against the checklist JSON, failing a step
// that graded perfectly well.
const mixedAIDocument = `# Definition of done

## checks

### command: it works

` + "```bash\ntrue\n```" + `

- cel: exitCode == 0

## Acceptance Criteria

- [ ] retries on 5xx
- [ ] docs updated
`

func aiStepOf(tree *FixtureNode) *FixtureTest {
	GinkgoHelper()
	var found *FixtureTest
	tree.Walk(func(node *FixtureNode) {
		if node.Test != nil && node.Test.IsAIStep() && found == nil {
			found = node.Test
		}
	})
	Expect(found).ToNot(BeNil(), "document produced no AI step")
	return found
}

var _ = Describe("ai step criteria scoping", func() {
	It("reads only the declared section's checklist and leaves other steps' bullets alone", func() {
		front := &FrontMatter{AI: &FixtureAIConfig{CriteriaSection: "Acceptance Criteria"}}

		tree, err := ParseMarkdownContentWithTree("dod", mixedAIDocument, GinkgoT().TempDir(), front)

		Expect(err).ToNot(HaveOccurred())
		step := aiStepOf(tree)
		Expect(step.AIStep.Criteria).To(HaveLen(2))
		Expect(step.AIStep.Criteria[0].Text).To(Equal("retries on 5xx"))
		Expect(step.Expected.CEL).To(BeEmpty(),
			"a shell step's expectation must not become the checklist's")
	})

	It("still reads the whole document when no section is declared", func() {
		front := &FrontMatter{AI: &FixtureAIConfig{}}

		tree, err := ParseMarkdownContentWithTree("dod", mixedAIDocument, GinkgoT().TempDir(), front)

		Expect(err).ToNot(HaveOccurred())
		step := aiStepOf(tree)
		Expect(step.AIStep.Criteria).To(HaveLen(2))
		Expect(step.Expected.CEL).To(Equal("exitCode == 0"))
	})

	It("yields no criteria when the declared section is absent", func() {
		front := &FrontMatter{AI: &FixtureAIConfig{CriteriaSection: "Nowhere"}}

		tree, err := ParseMarkdownContentWithTree("dod", mixedAIDocument, GinkgoT().TempDir(), front)

		Expect(err).ToNot(HaveOccurred())
		Expect(aiStepOf(tree).AIStep.Criteria).To(BeEmpty())
	})
})

var _ = Describe("markdownSection", func() {
	It("ends a section at the next heading of the same level", func() {
		Expect(markdownSection(mixedAIDocument, "checks")).To(ContainSubstring("### command: it works"))
		Expect(markdownSection(mixedAIDocument, "checks")).ToNot(ContainSubstring("retries on 5xx"))
	})

	It("runs a section to the end of the document when nothing closes it", func() {
		Expect(markdownSection(mixedAIDocument, "Acceptance Criteria")).
			To(Equal("- [ ] retries on 5xx\n- [ ] docs updated"))
	})

	It("returns nothing for a heading the document does not have", func() {
		Expect(markdownSection(mixedAIDocument, "Nowhere")).To(BeEmpty())
	})
})
