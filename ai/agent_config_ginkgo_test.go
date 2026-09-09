package ai

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/api/registry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAgentConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Agent config")
}

var _ = Describe("resolved provider configuration", func() {
	It("does not reread mutable saved defaults when constructing the resolved provider", func() {
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		model, err := registry.ResolveModel(api.Model{Name: "gpt-4o", Mode: api.ModeAPI})
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(home, ".captain.yaml"), []byte("ai: [invalid saved configuration]\n"), 0o600)).To(Succeed())
		provider, err := NewProvider(AgentConfig{Model: model, APIKey: "example-key", APIURL: "http://127.0.0.1:1"})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(CloseProvider(provider)).To(Succeed()) })
		Expect(provider.GetModel()).To(Equal("gpt-4o"))
		Expect(provider.GetRuntime()).To(Equal(api.RuntimeOf(registry.OpenAI, api.ModeAPI)))
	})
})
