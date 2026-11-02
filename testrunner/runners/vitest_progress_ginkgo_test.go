package runners

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type timedSink struct {
	mu        sync.Mutex
	firstPass time.Time
}

func (s *timedSink) UpdatePackageProgress(_ string, _ parsers.Framework, progress parsers.Test) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.firstPass.IsZero() && progress.Sum().Passed > 0 {
		s.firstPass = time.Now()
	}
}

const vitestProgressSpec = `
test("fast", () => { expect(1 + 1).toBe(2) })
test("slow", async () => {
  await new Promise((resolve) => setTimeout(resolve, 1500))
  expect(true).toBe(true)
})
`

var _ = Describe("Vitest progress reporter", func() {
	It("streams a finished test before the vitest process exits", func() {
		if _, err := exec.LookPath("node"); err != nil {
			Skip("node is not installed")
		}
		vitestBin, err := filepath.Abs("../ui/node_modules/vitest/vitest.mjs")
		Expect(err).NotTo(HaveOccurred())
		Expect(vitestBin).To(BeAnExistingFile(), "run `pnpm install` in testrunner/ui")

		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "progress.test.js"), []byte(vitestProgressSpec), 0o600)).To(Succeed())
		run, err := NewVitest(dir).BuildCommand(".")
		Expect(err).NotTo(HaveOccurred())

		// Run the built vitest arguments against the repo's vitest install,
		// since the temp package has no node_modules of its own.
		args := run.Process.Args[slices.Index(run.Process.Args, "vitest")+1:]
		cmd := exec.Command("node", append(append([]string{vitestBin}, args...), "--globals")...)
		cmd.Dir = dir
		sink := &timedSink{}
		stream := parsers.NewVitestStreamWriter(dir, ".", sink)
		cmd.Stdout = stream
		cmd.Stderr = GinkgoWriter
		Expect(cmd.Run()).To(Succeed())
		exited := time.Now()
		stream.Flush()

		Expect(sink.firstPass).NotTo(BeZero(), "no passed test was streamed")
		Expect(exited.Sub(sink.firstPass)).To(BeNumerically(">", 500*time.Millisecond),
			"the fast test should be reported while the slow one is still running")

		report, err := os.Open(run.ReportPath)
		Expect(err).NotTo(HaveOccurred())
		defer report.Close()
		tests, err := run.Parser.Parse(report)
		Expect(err).NotTo(HaveOccurred())
		Expect(parsers.Tests(tests).Sum().Passed).To(Equal(2))
	})
})
