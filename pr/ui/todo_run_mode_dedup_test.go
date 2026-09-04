package ui

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("canonical TODO runtime selection", func() {
	DescribeTable("resolves the shared runtime grammar through Captain",
		func(mode, model, wantMode, wantModel string) {
			gotMode, gotModel, err := resolveTodoRunRuntime(mode, model)

			Expect(err).ToNot(HaveOccurred())
			Expect(gotMode).To(Equal(wantMode))
			Expect(gotModel).To(Equal(wantModel))
		},
		Entry("agent Claude", "agent", "opus", "agent", "claude-opus-5"),
		Entry("CLI Codex", "cli", "gpt-5.6-sol", "cli", "gpt-5.6-sol"),
		Entry("the model's own prefix names the runtime", "api", "agent:opus", "agent", "claude-opus-5"),
		// A fold that names no model offers no runtime: a default invented here
		// would come back from the dialog as the request layer, outranking the
		// configuration it was meant to defer to.
		Entry("a spec naming neither offers no default", "", "", "", ""),
		Entry("a mode without a model keeps the mode and offers no model", "agent", "", "agent", ""),
	)

	DescribeTable("rejects a mode that names a provider or a composite adapter",
		func(mode string) {
			_, _, err := resolveTodoRunRuntime(mode, "opus")
			Expect(err).To(MatchError(ContainSubstring("invalid model configuration")))
		},
		Entry("provider", "anthropic"),
		Entry("composite", "claude-agent"),
	)
})
