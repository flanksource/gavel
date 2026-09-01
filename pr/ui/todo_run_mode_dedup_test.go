package ui

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/todos/drivers"
)

var _ = Describe("canonical TODO runtime selection", func() {
	DescribeTable("resolves the shared runtime grammar through Captain",
		func(driver drivers.Kind, mode, model, wantMode, wantModel string) {
			gotMode, gotModel, err := resolveTodoRunRuntime(driver, mode, model)

			Expect(err).ToNot(HaveOccurred())
			Expect(gotMode).To(Equal(wantMode))
			Expect(gotModel).To(Equal(wantModel))
		},
		Entry("agent Claude", drivers.Agent, "agent", "opus", "agent", "claude-opus-5"),
		Entry("CLI Codex", drivers.Cli, "cli", "gpt-5.6-sol", "cli", "gpt-5.6-sol"),
		Entry("model prefix wins", drivers.Cmux, "api", "agent:opus", "agent", "claude-opus-5"),
	)

	DescribeTable("rejects a mode that names a provider or a composite adapter",
		func(mode string) {
			_, _, err := resolveTodoRunRuntime(drivers.Agent, mode, "opus")
			Expect(err).To(MatchError(ContainSubstring("invalid model configuration")))
		},
		Entry("provider", "anthropic"),
		Entry("composite", "claude-agent"),
	)
})
