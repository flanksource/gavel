package types

import (
	"context"

	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
)

var _ = Describe("exec step workflow commands", func() {
	run := func(body string) fixtures.FixtureResult {
		GinkgoHelper()
		results, _, err := fixtures.RunNodes(context.Background(), parseYAMLDocument(body), fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})
		Expect(err).ToNot(HaveOccurred())
		Expect(results).To(HaveLen(1))
		return results[0]
	}
	children := func(result fixtures.FixtureResult) []string {
		var out []string
		for _, child := range result.Children {
			out = append(out, string(child.Results.Status)+" "+child.Name)
		}
		return out
	}

	It("fails a command that exits 0 after printing ::error, keeping only its plain output", func() {
		result := run(`
kind: Fixture
steps:
  - id: check
    kind: exec
    script: |
      echo building
      echo '::error file=main.go,line=7::undefined: foo'
`)
		Expect(result.Status).To(Equal(task.StatusFAIL))
		Expect(result.Error).To(Equal("1 error finding(s), 0 other: ::error: undefined: foo"))
		Expect(result.Stdout).To(Equal("building\n"))
		Expect(children(result)).To(Equal([]string{string(task.StatusFAIL) + " ::error: undefined: foo"}))
	})

	It("warns a command that printed ::warning, still running its cel and exports over the stripped stdout", func() {
		result := run(`
kind: Fixture
steps:
  - id: build
    kind: exec
    script: |
      echo '::group::Compile'
      echo '::warning title=Deprecated::flag -x'
      echo '::progress::1/1'
      echo 'version=1.2'
      echo '::endgroup::'
    cel: "!stdout.contains('::warning') && stdout.contains('::group::Compile')"
    export: { VERSION: "{{ .outputs.stdout | strings.TrimSpace }}" }
`)
		Expect(result.Status).To(Equal(task.StatusWarning), result.Error+result.CELTrace)
		Expect(result.Error).To(Equal("1 warning finding(s)"))
		Expect(result.CELResult).To(BeTrue())
		Expect(result.Stdout).To(Equal("::group::Compile\nversion=1.2\n::endgroup::\n"))
		Expect(children(result)).To(Equal([]string{string(task.StatusWarning) + " Deprecated: flag -x"}))
		Expect(result.Metadata["exports"]).To(HaveKey("VERSION"))
	})

	It("reads workflow commands from a pty's output", func() {
		result := run(`
kind: Fixture
steps:
  - id: tty
    kind: exec
    terminal: pty
    script: "echo '::notice::from a terminal'; echo done"
    cel: "stdout.contains('done') && !stdout.contains('::notice')"
`)
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error+result.CELTrace)
		Expect(children(result)).To(Equal([]string{string(task.StatusPASS) + " ::notice: from a terminal"}))
	})
})
