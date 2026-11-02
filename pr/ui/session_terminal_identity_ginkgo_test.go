package ui

import (
	"context"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("terminal session identity", func() {
	g.DescribeTable("uses Captain's canonical provider identity", func(provider, expected string) {
		agent, err := terminalAgentFromProvider(provider)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(agent).To(o.Equal(expected))
	}, g.Entry("Anthropic", "anthropic", "claude"), g.Entry("OpenAI", "openai", "codex"), g.Entry("Google", "google", "gemini"))
	g.It("rejects an unknown provider", func() {
		_, err := terminalAgentFromProvider("unknown-provider")
		o.Expect(err).To(o.MatchError(o.ContainSubstring("unknown terminal provider")))
	})
	g.It("refuses a missing session instead of guessing a provider", func() {
		_, err := sessionTerminalAgent(context.Background(), g.GinkgoT().TempDir(), " ")
		o.Expect(err).To(o.MatchError(o.ContainSubstring("sessionId is required")))
	})
})
