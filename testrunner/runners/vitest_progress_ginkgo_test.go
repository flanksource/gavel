package runners

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sync"

	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// passSignal writes its marker file when the stream reports the first passed test, so a test
// still running in vitest can see that gavel got the result.
type passSignal struct {
	marker string
	once   sync.Once
}

func (s *passSignal) UpdatePackageProgress(_ string, _ parsers.Framework, progress parsers.Test) {
	if progress.Sum().Passed > 0 {
		s.once.Do(func() { _ = os.WriteFile(s.marker, []byte("passed\n"), 0o600) })
	}
}

// findTest finds the test called name among tests and their children.
func findTest(tests []parsers.Test, name string) (parsers.Test, bool) {
	for _, test := range tests {
		if test.Name == name {
			return test, true
		}
		if found, ok := findTest(test.Children, name); ok {
			return found, true
		}
	}
	return parsers.Test{}, false
}

// vitestProgressSpec is a fast test and a slow one that passes only once gavel has streamed the
// fast test's pass, which writes the marker file at %q. The slow test waits on that signal, so
// the spec holds however busy the machine is.
const vitestProgressSpec = `
import { existsSync } from "node:fs"

test("fast", () => { expect(1 + 1).toBe(2) })
test("slow", async () => {
  const deadline = Date.now() + 30000
  while (!existsSync(%[1]q)) {
    if (Date.now() > deadline) {
      throw new Error("gavel had not streamed the fast test's pass 30s after it finished")
    }
    await new Promise((resolve) => setTimeout(resolve, 20))
  }
}, 35000)
`

var _ = Describe("Vitest progress reporter", func() {
	It("loads from a nested package whose workspace file confines Vite's file access", func() {
		if _, err := exec.LookPath("node"); err != nil {
			Skip("node is not installed")
		}
		vitestBin, err := filepath.Abs("../ui/node_modules/vitest/vitest.mjs")
		Expect(err).NotTo(HaveOccurred())
		Expect(vitestBin).To(BeAnExistingFile(), "run `pnpm install` in testrunner/ui")

		root := GinkgoT().TempDir()
		pkg := filepath.Join(root, "web")
		Expect(os.Mkdir(pkg, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(pkg, "pnpm-workspace.yaml"), []byte("packages:\n  - .\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(pkg, "nested.test.js"), []byte(`test("nested", () => { expect(1).toBe(1) })`+"\n"), 0o600)).To(Succeed())
		run, err := NewVitest(root).BuildCommand(TestOptions{Cwd: root, Packages: []string{"web"}})
		Expect(err).NotTo(HaveOccurred())

		args := run.Process.Args[slices.Index(run.Process.Args, "vitest")+1:]
		cmd := exec.Command("node", append(append([]string{vitestBin}, args...), "--globals")...)
		cmd.Dir = pkg
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(out))
		Expect(run.ReportPath).To(BeAnExistingFile())
	})
	It("streams a finished test before the vitest process exits", func() {
		if _, err := exec.LookPath("node"); err != nil {
			Skip("node is not installed")
		}
		vitestBin, err := filepath.Abs("../ui/node_modules/vitest/vitest.mjs")
		Expect(err).NotTo(HaveOccurred())
		Expect(vitestBin).To(BeAnExistingFile(), "run `pnpm install` in testrunner/ui")

		dir := GinkgoT().TempDir()
		marker := filepath.Join(GinkgoT().TempDir(), "fast-passed")
		Expect(os.WriteFile(filepath.Join(dir, "progress.test.js"), []byte(fmt.Sprintf(vitestProgressSpec, marker)), 0o600)).To(Succeed())
		run, err := NewVitest(dir).BuildCommand(TestOptions{Cwd: dir, Packages: []string{"./."}})
		Expect(err).NotTo(HaveOccurred())

		// Run the built vitest arguments against the repo's vitest install,
		// since the temp package has no node_modules of its own.
		args := run.Process.Args[slices.Index(run.Process.Args, "vitest")+1:]
		cmd := exec.Command("node", append(append([]string{vitestBin}, args...), "--globals")...)
		cmd.Dir = dir
		stream := parsers.NewVitestStreamWriter(dir, ".", &passSignal{marker: marker})
		cmd.Stdout = stream
		cmd.Stderr = GinkgoWriter
		runErr := cmd.Run()
		stream.Flush()

		report, err := os.Open(run.ReportPath)
		Expect(err).NotTo(HaveOccurred())
		defer report.Close()
		tests, err := run.Parser.Parse(report)
		Expect(err).NotTo(HaveOccurred())
		slow, found := findTest(tests, "slow")
		Expect(found).To(BeTrue(), "the report lists the slow test")
		Expect(slow.Passed).To(BeTrue(), "the slow test passes only once the fast test's pass is streamed while it runs: %s", slow.Message)
		Expect(parsers.Tests(tests).Sum().Passed).To(Equal(2))
		Expect(runErr).NotTo(HaveOccurred())
	})
})
