package types

import (
	"context"
	"os"

	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
)

var _ = Describe("exec step $VAR expansion", func() {
	It("expands exec args from the document env, an earlier step's export and the process env", func() {
		nodes := parseYAMLDocument(`
kind: Fixture
env:
  GREETING: hello
steps:
  - id: mint
    kind: exec
    script: 'printf "{\"token\":\"t-42\"}"'
    export: { TOKEN: "{{.outputs.json.token}}" }
  - id: use
    kind: exec
    exec: printf
    args: ["%s|%s|%s|%s", "$GREETING", "$TOKEN", "$HOME", "$EXEC_ENV_SPEC_UNSET"]
`)
		results, _, err := fixtures.RunNodes(context.Background(), nodes, fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect(results[1].Status).To(Equal(task.StatusPASS), results[1].Error)
		Expect(results[1].Stdout).To(Equal("hello|t-42|" + os.Getenv("HOME") + "|$EXEC_ENV_SPEC_UNSET"))
	})

	It("leaves a script body's environment variables to the shell, which may reassign them", func() {
		nodes := parseYAMLDocument(`
kind: Fixture
steps:
  - id: mint
    kind: exec
    script: 'printf "{\"token\":\"t-42\"}"'
    export: { TOKEN: "{{.outputs.json.token}}" }
  - id: reassign
    kind: exec
    script: 'TOKEN=bye; printf "%s" "$TOKEN"'
`)
		results, _, err := fixtures.RunNodes(context.Background(), nodes, fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect(results[1].Status).To(Equal(task.StatusPASS), results[1].Error)
		Expect(results[1].Stdout).To(Equal("bye"))
	})
})

var _ = Describe("exec step goProfiles", func() {
	It("appends the profile destination under --profile and ingests what the command wrote there", func() {
		nodes := parseYAMLDocument(`
kind: Fixture
steps:
  - id: profiled
    kind: exec
    script: 'printf "not a profile" > "${0#--cpuprofile=}"'
    goProfiles:
      cpu: { file: cpu.pprof, args: ["--cpuprofile={{.path}}"] }
`)
		results, _, err := fixtures.RunNodes(context.Background(), nodes, fixtures.RunOptions{WorkDir: GinkgoT().TempDir(), Profile: true})

		Expect(err).ToNot(HaveOccurred())
		Expect(results[0].Status).To(Equal(task.StatusFAIL))
		Expect(results[0].Error).To(ContainSubstring("goProfiles.cpu"), results[0].Stderr)
	})

	It("keeps one step's profile environment out of the next step that shares the document env", func() {
		nodes := parseYAMLDocument(`
kind: Fixture
env:
  GREETING: hello
steps:
  - {id: first, kind: exec, script: 'printf "%s" "$GREETING"'}
  - {id: second, kind: exec, script: 'printf "%s" "$GREETING"'}
`)
		results, _, err := fixtures.RunNodes(context.Background(), nodes, fixtures.RunOptions{WorkDir: GinkgoT().TempDir(), Profile: true})

		Expect(err).ToNot(HaveOccurred())
		for _, result := range results {
			Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
			Expect(result.Stdout).To(Equal("hello"))
		}
	})

	It("passes no profile arguments without --profile", func() {
		nodes := parseYAMLDocument(`
kind: Fixture
steps:
  - id: unprofiled
    kind: exec
    script: 'printf "%s" "$0"'
    args: [none]
    goProfiles:
      cpu: { file: cpu.pprof, args: ["--cpuprofile={{.path}}"] }
`)
		results, _, err := fixtures.RunNodes(context.Background(), nodes, fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect(results[0].Status).To(Equal(task.StatusPASS), results[0].Error)
		Expect(results[0].Stdout).To(Equal("none"))
	})
})
