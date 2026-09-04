package todos

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/todos/types"
)

var _ = Describe("verification fixture sections", func() {
	It("moves an H1 verification section while preserving its nested headings", func() {
		body := `Parser failures lose context.

# Verification

` + "```yaml test" + `
paths: [./pkg/parser]
` + "```" + `

## Focused checks

- parser errors include the source path

# Notes

Keep this note in the issue body.`

		cleanBody, verification, found := SplitVerificationFixture(body)

		Expect(found).To(BeTrue())
		Expect(cleanBody).To(Equal(`Parser failures lose context.

# Notes

Keep this note in the issue body.`))
		Expect(verification).To(Equal("```yaml test\npaths: [./pkg/parser]\n```\n\n## Focused checks\n\n- parser errors include the source path"))
	})

	It("moves all H2 verification sections in document order", func() {
		body := `Description.

## verification ##

first fixture

## Implementation

Keep this section.

## VERIFICATION

second fixture

## Verification Result

Keep the result.`

		cleanBody, verification, found := SplitVerificationFixture(body)

		Expect(found).To(BeTrue())
		Expect(cleanBody).To(Equal(`Description.

## Implementation

Keep this section.

## Verification Result

Keep the result.`))
		Expect(verification).To(Equal("first fixture\n\nsecond fixture"))
	})

	It("moves setext H1 and H2 verification headings", func() {
		body := `Description.

Verification
============

first fixture

Notes
=====

Keep this note.

Verification
------------

second fixture

Implementation
--------------

Keep this section.`

		cleanBody, verification, found := SplitVerificationFixture(body)

		Expect(found).To(BeTrue())
		Expect(cleanBody).To(Equal(`Description.

Notes
=====

Keep this note.

Implementation
--------------

Keep this section.`))
		Expect(verification).To(Equal("first fixture\n\nsecond fixture"))
	})

	It("ignores verification-looking content outside top-level H1 and H2 headings", func() {
		body := `Description.

` + "```markdown" + `
# Verification

not a fixture
` + "```" + `

### Verification

also not a fixture

> ## Verification
>
> quoted documentation`

		cleanBody, verification, found := SplitVerificationFixture(body)

		Expect(found).To(BeFalse())
		Expect(cleanBody).To(Equal(body))
		Expect(verification).To(BeEmpty())
	})

	It("keeps fixture frontmatter and its H2 steps inside an H2 verification section", func() {
		prose := `The step controls float over the source listing.

## Acceptance Criteria

- [ ] The step controls render in the AppShell chrome.

## Scope

- Only the remote debugger page and its colocated tests.`

		fixture := `---
cwd: .
timeout: 30m
codeBlocks: [test, lint]
ai: {}
verify:
  scope: diff
  threshold: 80
---

## Remote debugger chrome tests

` + "```yaml test" + `
work-dir: ./www
framework: [vitest]
` + "```" + `

## Changed-code lint

` + "```yaml lint" + `
linters:
  - eslint
files:
  - www/src/pages/RemoteDebugger*.tsx
` + "```"

		cleanBody, verification, found := SplitVerificationFixture(prose + "\n\n## Verification\n\n" + fixture)

		Expect(found).To(BeTrue())
		Expect(verification).To(Equal(fixture))
		Expect(cleanBody).To(Equal(prose))
	})

	It("ends a frontmatter fixture at the next operational section", func() {
		fixture := `---
cwd: .
codeBlocks: [test]
---

## Focused tests

` + "```yaml test" + `
paths: [./pkg/parser]
` + "```"

		body := "Description.\n\n## Verification\n\n" + fixture + "\n\n## Attempts\n\n| # |"

		cleanBody, verification, found := SplitVerificationFixture(body)

		Expect(found).To(BeTrue())
		Expect(verification).To(Equal(fixture))
		Expect(cleanBody).To(Equal("Description.\n\n## Attempts\n\n| # |"))
	})

	It("round-trips a frontmatter fixture with H2 steps through the body", func() {
		fixture := `---
cwd: .
codeBlocks: [test]
---

## Focused tests

` + "```yaml test" + `
paths: [./pkg/parser]
` + "```"

		body := UpsertVerificationFixture("Issue description.", fixture)

		Expect(ExtractVerificationFixture(body)).To(Equal(fixture))
	})

	It("parses dedicated verification with H2 subsections as one fixture document", func() {
		markdown := `## Focused tests

` + "```yaml test" + `
paths: [./pkg/parser]
` + "```" + `

## Assertions

` + "```cel" + `
results.all(r, r.passed)
` + "```"

		nodes, err := ParseVerificationMarkdown(VerificationMarkdownOptions{
			Name: "parser verification", Markdown: markdown, SourceDir: ".",
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(nodes).NotTo(BeEmpty())
		Expect(types.CountTests(nodes)).To(BeNumerically(">", 0))
	})
})
