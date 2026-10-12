package types

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestTriageNew(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Triage new") }

var _ = Describe("Triage new", func() {
	var env TriageNewEnvelope
	BeforeEach(func() {
		env = TriageNewEnvelope{ResultEnvelope: validEnvelope(), Title: "Repair parser crash", Labels: []string{"bug"}, Action: "keep"}
	})
	It("accepts complete keep output", func() { Expect(env.Validate()).To(Succeed()) })
	DescribeTable("rejects invalid completed output", func(mutate func(*TriageNewEnvelope)) {
		mutate(&env)
		Expect(env.Validate()).To(HaveOccurred())
	},
		Entry("missing title", func(e *TriageNewEnvelope) { e.Title = " " }),
		Entry("missing labels", func(e *TriageNewEnvelope) { e.Labels = nil }),
		Entry("blank label", func(e *TriageNewEnvelope) { e.Labels = []string{" "} }),
		Entry("duplicate labels", func(e *TriageNewEnvelope) { e.Labels = []string{"bug", "Bug"} }),
		Entry("unknown action", func(e *TriageNewEnvelope) { e.Action = "retire" }),
		Entry("relationship without target", func(e *TriageNewEnvelope) { e.Action = "child-of" }),
		Entry("relationship without rationale", func(e *TriageNewEnvelope) { e.Action = "duplicate-of"; e.Target = "other" }),
		Entry("merge without combined content", func(e *TriageNewEnvelope) { e.Action = "merge-into"; e.Target = "other"; e.Rationale = "Same work" }),
		Entry("keep with relationship target", func(e *TriageNewEnvelope) { e.Target = "other" }))
})
