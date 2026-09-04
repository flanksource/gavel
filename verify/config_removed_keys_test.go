package verify

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGavelConfigRemovedKeys(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "gavel config removed keys")
}

// removedKeyWorkspace writes cfg as the only .gavel.yaml the loader can see.
// HOME is redirected because LoadGavelConfig merges ~/.gavel.yaml — the person
// running the suite has removed keys in theirs, which would otherwise be a
// hidden layer of every assertion below.
func removedKeyWorkspace(cfg string) string {
	GinkgoT().Setenv("HOME", GinkgoT().TempDir())
	dir := GinkgoT().TempDir()
	Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte(cfg), 0o644)).To(Succeed())
	return dir
}

var _ = Describe("removed .gavel.yaml keys", func() {
	DescribeTable("rejects a key that no longer exists, naming its replacement",
		func(yamlDoc, wantPath, wantReplacementFragment string) {
			_, err := LoadGavelConfig(removedKeyWorkspace(yamlDoc))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(wantPath))
			Expect(err.Error()).To(ContainSubstring(wantReplacementFragment))
		},
		Entry("todos.driver", "todos:\n  driver: cmux\n",
			"todos.driver", `ai.model: "cli:opus:high"`),
		Entry("todos.prompts", "todos:\n  prompts:\n    security:\n      class: run\n",
			"todos.prompts", "a lifecycle step under todos.<step>"),
		Entry("todos.groupBy", "todos:\n  groupBy: repo\n",
			"todos.groupBy", "grouping was removed"),
		Entry("todos.approvals", "todos:\n  approvals: true\n",
			"todos.approvals", "permissions.mode: default"),
		Entry("checks.maxIterations", "checks:\n  maxIterations: 7\n",
			"checks.maxIterations", "todos.run.workflow.verify.maxIterations"),
	)

	It("fails on a key whose value is null, because presence is what is rejected", func() {
		_, err := LoadGavelConfig(removedKeyWorkspace("todos:\n  prompts:\n"))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("todos.prompts"))
	})

	It("names every removed key in one error so all three are fixed in one pass", func() {
		_, err := LoadGavelConfig(removedKeyWorkspace(
			"checks:\n  maxIterations: 3\ntodos:\n  driver: cli\n  groupBy: repo\n"))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(SatisfyAll(
			ContainSubstring("todos.driver"),
			ContainSubstring("todos.groupBy"),
			ContainSubstring("checks.maxIterations"),
		))
	})

	It("loads a config carrying none of them, keeping its ordinary fields", func() {
		cfg, err := LoadGavelConfig(removedKeyWorkspace("todos:\n  timeout: 45m\n"))
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.Todos.Timeout).To(Equal("45m"))
	})

	It("matches by path, not substring, so the same names at another depth load", func() {
		cfg, err := LoadGavelConfig(removedKeyWorkspace(
			"commit:\n  groupBy: file\ntodos:\n  timeout: 45m\n  run:\n    prompts: x\n"))
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.Todos.Timeout).To(Equal("45m"))
	})
})
