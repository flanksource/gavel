package types

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
)

var _ = Describe("exec step cleanup", func() {
	// statusesByName is every result's status keyed by its name; teardown
	// failures are reported under "teardown".
	statusesByName := func(results []fixtures.FixtureResult) map[string]task.Status {
		byName := map[string]task.Status{}
		for _, result := range results {
			byName[result.Name] = result.Status
		}
		return byName
	}

	It("runs at teardown with the step's env after a later step failed", func() {
		marker := filepath.Join(GinkgoT().TempDir(), "restored")
		nodes := parseYAMLDocument(`
kind: Fixture
steps:
  - id: scaleDown
    kind: exec
    env: {MARKER: "` + marker + `", REPLICAS: "3"}
    script: 'printf down'
    cleanup: 'printf "%s" "$REPLICAS" > "$MARKER"'
  - {id: purge, kind: exec, script: 'exit 1'}
  - {id: skipped, kind: exec, script: 'printf unreachable'}
`)
		results, _, err := fixtures.RunNodes(context.Background(), nodes, fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect(statusesByName(results)).To(Equal(map[string]task.Status{
			"scaleDown": task.StatusPASS, "purge": task.StatusFAIL, "skipped": task.StatusSKIP,
		}))
		Expect(os.ReadFile(marker)).To(Equal([]byte("3")))
	})

	It("fails the teardown, naming the step and its output, when the cleanup exits non-zero", func() {
		nodes := parseYAMLDocument(`
kind: Fixture
steps:
  - {id: scaleDown, kind: exec, script: 'printf down', cleanup: 'echo cannot-restore >&2; exit 3'}
`)
		results, _, err := fixtures.RunNodes(context.Background(), nodes, fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect(statusesByName(results)).To(HaveKeyWithValue("teardown", task.StatusERR))
		Expect(results[len(results)-1].Error).To(And(ContainSubstring(`exec step "scaleDown" cleanup exited 3`), ContainSubstring("cannot-restore")))
	})
})
