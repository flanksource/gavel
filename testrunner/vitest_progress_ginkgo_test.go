package testrunner

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("vitest live progress", func() {
	It("streams passed vitest tests while the package is still running", func() {
		if _, err := exec.LookPath("node"); err != nil {
			Skip("node is not installed")
		}
		nodeModules, err := filepath.Abs("ui/node_modules")
		Expect(err).NotTo(HaveOccurred())
		Expect(filepath.Join(nodeModules, "vitest")).To(BeADirectory(), "run `pnpm install` in testrunner/ui")

		dir := GinkgoT().TempDir()
		for name, content := range map[string]string{
			"package.json":     `{"name":"progress","devDependencies":{"vitest":"*"}}`,
			"vitest.config.js": `export default { test: { globals: true } }`,
			"progress.test.js": `
test("fast", () => { expect(1 + 1).toBe(2) })
test("slow", async () => {
  await new Promise((resolve) => setTimeout(resolve, 1500))
  expect(true).toBe(true)
})
`,
		} {
			Expect(os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600)).To(Succeed())
		}
		Expect(os.Symlink(nodeModules, filepath.Join(dir, "node_modules"))).To(Succeed())

		updates := make(chan []parsers.Test, 256)
		liveProgress := make(chan bool, 1)
		go func() {
			seen := false
			for tree := range updates {
				for _, node := range tree {
					seen = seen || node.Framework == parsers.Vitest && node.Pending && node.Sum().Passed > 0
				}
			}
			liveProgress <- seen
		}()

		streamer := NewTestStreamer(updates)
		o := &TestOrchestrator{
			RunOptions: RunOptions{WorkDir: dir, Recursive: true},
			registry:   DefaultRegistry(dir),
			streamer:   streamer,
		}
		results, err := o.detectAndRun([]Framework{parsers.Vitest}, nil, nil)
		streamer.Done()
		Expect(err).NotTo(HaveOccurred())
		Expect(results.All().Sum().Passed).To(Equal(2), "results: %+v", results)
		Expect(<-liveProgress).To(BeTrue(), "no pending vitest package carried a passed test before completion")
	})
})
