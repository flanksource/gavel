package verify

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("raw configuration spec layers", func() {
	DescribeTable("rejects invalid home configuration before repository overrides",
		func(homeConfig, repositoryConfig, field string) {
			home, repository := GinkgoT().TempDir(), GinkgoT().TempDir()
			GinkgoT().Setenv("HOME", home)
			path := filepath.Join(home, ".gavel.yaml")
			Expect(os.WriteFile(path, []byte(homeConfig), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(repository, ".gavel.yaml"), []byte(repositoryConfig), 0600)).To(Succeed())
			_, err := LoadGavelConfig(repository)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(path))
			Expect(err.Error()).To(ContainSubstring(field))
			_, err = LoadGavelConfigTrace(repository)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(path))
			Expect(err.Error()).To(ContainSubstring(field))
		},
		Entry("model temperature", "ai: {temperature: 3}\n", "ai: {temperature: 1}\n", "temperature"),
		Entry("step budget", "todos: {run: {budget: {maxTurns: -1}}}\n", "todos: {run: {budget: {maxTurns: 5}}}\n", "maxTurns"),
		Entry("custom step budget", "todos: {steps: {handoff: {budget: {cost: -1}}}}\n", "todos: {steps: {handoff: {budget: {cost: 5}}}}\n", "cost"),
		Entry("timeout constraint", "todos: {timeout: invalid}\n", "todos: {timeout: 5m}\n", "timeout"),
	)

	It("keeps structurally valid partial runtimes available for later composition", func() {
		home, repository := GinkgoT().TempDir(), GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		Expect(os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte("ai: {mode: agent, permissions: {tools: {Read: allow}}}\n"), 0600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repository, ".gavel.yaml"), []byte("ai: {model: gpt-5.6-sol}\n"), 0600)).To(Succeed())
		config, err := LoadGavelConfig(repository)
		Expect(err).NotTo(HaveOccurred())
		Expect(config.AI.Name).To(Equal("gpt-5.6-sol"))
		Expect(string(config.AI.Permissions.Tools["Read"])).To(Equal("allow"))
	})
})
