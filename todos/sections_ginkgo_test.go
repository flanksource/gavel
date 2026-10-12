package todos

import (
	"strings"

	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TODO section insertion", func() {
	It("keeps heading-like lines inside fenced code when adding acceptance criteria", func() {
		body := "## Review comments\n\n### Review summary\n\n```markdown\n## Acceptance Criteria\n## Verification\n```\n\nFinal note."

		got := UpsertCriteriaSection(body, []types.AcceptanceCriterion{{Text: "Tests pass"}})

		Expect(got).To(ContainSubstring("```markdown\n## Acceptance Criteria\n## Verification\n```"))
		Expect(got).To(ContainSubstring("Final note."))
		Expect(strings.Count(got, "## Acceptance Criteria")).To(Equal(2))
		Expect(ParseAcceptanceCriteria(got)).To(ConsistOf(types.AcceptanceCriterion{Text: "Tests pass"}))
	})
})
