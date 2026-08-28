package ui

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/todos/drivers"
)

var _ = Describe("canonical TODO runtime selection", func() {
	DescribeTable("resolves the shared backend grammar through Captain",
		func(driver drivers.Kind, backend, model, wantBackend, wantModel string) {
			gotBackend, gotModel, err := resolveTodoRunBackendModel(driver, backend, model)

			Expect(err).ToNot(HaveOccurred())
			Expect(gotBackend).To(Equal(wantBackend))
			Expect(gotModel).To(Equal(wantModel))
		},
		Entry("agent Claude", drivers.Agent, "agent", "opus", "agent", "claude-opus-5"),
		Entry("CLI Codex", drivers.Cli, "cli", "gpt-5.6-sol", "cli", "gpt-5.6-sol"),
		Entry("model prefix wins", drivers.Cmux, "api", "agent:opus", "agent", "claude-opus-5"),
	)

	DescribeTable("rejects legacy values",
		func(backend string) {
			_, _, err := resolveTodoRunBackendModel(drivers.Agent, backend, "opus")
			Expect(err).To(MatchError(ContainSubstring("invalid model configuration")))
		},
		Entry("provider", "anthropic"),
		Entry("composite", "claude-agent"),
	)
})
