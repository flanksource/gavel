package types

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTypes(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Todo types") }

var _ = Describe("Envelope", func() {
	events := []string{"completed", "blocked"}

	It("accepts a declared event with a comment", func() {
		env := Envelope{Success: true, Event: "completed", Comment: "Fixed the parser crash."}
		Expect(env.Validate(events)).To(Succeed())
	})

	DescribeTable("rejects an envelope the step outcomes could not classify",
		func(env Envelope, declared []string, wantErr string) {
			Expect(env.Validate(declared)).To(MatchError(ContainSubstring(wantErr)))
		},
		Entry("blank comment", Envelope{Success: true, Event: "completed", Comment: "  "}, events, "comment"),
		Entry("undeclared event", Envelope{Event: "ask", Comment: "Which database?"}, events, `event "ask" is not one of completed, blocked`),
		Entry("step without events", Envelope{Event: "completed", Comment: "done"}, nil, "declares no events"),
	)
})
