package ui

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("question answer validation", func() {
	request := map[string]any{"questions": []any{
		map[string]any{"id": "location", "question": "Where?"},
		map[string]any{"id": "format", "question": "How?"},
	}}

	It("accepts one nonblank answer per question id", func() {
		Expect(validateQuestionAnswers(request, map[string]any{
			"location": "Inline", "format": []any{"Markdown"},
		})).To(Succeed())
	})

	DescribeTable("rejects incomplete or extra answers",
		func(answers map[string]any) {
			Expect(validateQuestionAnswers(request, answers)).To(HaveOccurred())
		},
		Entry("missing id", map[string]any{"location": "Inline"}),
		Entry("blank answer", map[string]any{"location": "Inline", "format": " "}),
		Entry("extra id", map[string]any{"location": "Inline", "format": "Markdown", "other": "extra"}),
	)
})
