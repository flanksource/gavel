package status

import (
	"os"
	"path/filepath"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("status summary runtime", func() {
	It("resolves actual file data with captured prompt and saved settings before provider creation", func() {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "summary.prompt")
		Expect(os.WriteFile(path, []byte("---\nmodel: '{{details}}'\nmemory:\n  skipSkills: true\n---\nSummarize the change"), 0o600)).To(Succeed())
		prepared, err := ResolveSummaryPrompt(SummaryPromptOptions{Dir: dir, Override: verify.PromptSpec{File: path},
			Saved: captainconfig.Config{AI: captainconfig.AIDefaults{Timeout: "17m", MaxTokens: 3000,
				Providers: map[string]captainconfig.ProviderDefaults{"anthropic": {Mode: "agent", ReasoningEffort: "high"}},
			}}, Request: (api.Spec{}).WithExplicit("/noCache")})
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(path, []byte("broken: ["), 0o600)).To(Succeed())
		resolved, err := prepared.resolve("claude-sonnet-5")
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Request.Name).To(Equal("claude-sonnet-5"))
		Expect(resolved.Request.Memory.SkipSkills).To(BeTrue())
		Expect(resolved.Request.Budget.Timeout).To(Equal("17m"))
		Expect(resolved.Config.Budget).To(Equal(resolved.Request.Budget))
		Expect(resolved.Config.Model).To(Equal(resolved.Request.Model))
		Expect(resolved.Resolution.Provenance["/budget/timeout"].Source.Kind).To(Equal(api.FieldSourceSaved))
	})
})
