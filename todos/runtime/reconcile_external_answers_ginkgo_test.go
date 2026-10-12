package runtime

import (
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("a Captain turn as the lifecycle reads it", func() {
	const plainText = "Migrated the staging database."
	candidate := askCandidate{
		PromptRunID:        uuid.MustParse("0199f0aa-0000-7000-8000-00000000a001"),
		ExecutionSessionID: uuid.MustParse("0199f0aa-0000-7000-8000-00000000a002"),
	}
	const (
		askText       = `{"summary":"One more question.","endStatus":"ask","questions":[{"text":"Anonymise first?"}]}`
		completedText = `{"summary":"Done.","endStatus":"completed"}`
	)

	DescribeTable("maps Captain's run state onto a finished turn",
		func(state, want string) {
			turn, err := externalTurn(candidate, externalAnswer{TurnState: state, FinalText: plainText})

			Expect(err).NotTo(HaveOccurred())
			Expect(turn).To(Equal(lifecycle.ExternalTurn{
				Source: "captain", PromptRunID: candidate.PromptRunID, State: want, Summary: plainText,
			}))
		},
		Entry("succeeded", "succeeded", lifecycle.RunSucceeded),
		Entry("failed", "failed", lifecycle.RunFailed),
		Entry("cancelled", "cancelled", lifecycle.RunCancelled),
	)

	It("refuses a turn that has not finished", func() {
		_, err := externalTurn(candidate, externalAnswer{TurnState: "running"})

		Expect(err).To(MatchError(ContainSubstring(`unexpected state "running"`)))
	})

	DescribeTable("takes the structured result from the first source that carries one",
		func(answer externalAnswer, wantEndStatus string) {
			answer.TurnState = "succeeded"
			turn, err := externalTurn(candidate, answer)

			Expect(err).NotTo(HaveOccurred())
			Expect(turn.Envelope).To(HaveKeyWithValue("endStatus", wantEndStatus))
		},
		Entry("the run's result JSON before everything",
			externalAnswer{TurnResult: []byte(completedText), Envelope: []byte(askText), FinalText: askText}, "completed"),
		Entry("the StructuredOutput call before the reply",
			externalAnswer{TurnResult: []byte(`{"tokens":3}`), Envelope: []byte(completedText), FinalText: askText}, "completed"),
		Entry("the reply before the result text",
			externalAnswer{FinalText: askText, TurnText: completedText}, "ask"),
		Entry("the result text last", externalAnswer{FinalText: plainText, TurnText: askText}, "ask"),
	)

	DescribeTable("reads a reply that carries no endStatus as plain text",
		func(text string) {
			turn, err := externalTurn(candidate, externalAnswer{TurnState: "succeeded", FinalText: text})

			Expect(err).NotTo(HaveOccurred())
			Expect(turn.Envelope).To(BeNil())
			Expect(turn.Summary).To(Equal(text))
		},
		Entry("prose", plainText),
		Entry("a JSON object with no endStatus", `{"migrated":true}`),
		Entry("braces that are not JSON", "{staging} is migrated"),
		Entry("a JSON array", `[1,2]`),
	)

	It("prefers the run's result text as the summary", func() {
		turn, err := externalTurn(candidate, externalAnswer{TurnState: "succeeded", TurnText: plainText, FinalText: "partial"})

		Expect(err).NotTo(HaveOccurred())
		Expect(turn.Summary).To(Equal(plainText))
	})
})
