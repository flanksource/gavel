package launch

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/resultstore"
	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("native launch execution", func() {
	It("resolves prompt change selectors before preparing the prompt", func() {
		_, err := (NativeExecutor{Schemas: schemaFixture{}}).Prepare(context.Background(), PreparedWorkspace{Directory: GinkgoT().TempDir()}, Configuration{Lint: &LintConfig{Since: "HEAD", Linters: []LinterConfig{{ID: "review", Linter: "prompt", Options: map[string]any{"prompt": "missing-launch-prompt", "promptVersion": "v1"}}}}})
		Expect(err).To(MatchError(ContainSubstring("git")))
	})

	It("runs real Go tests with the prepared workspace environment", func() {
		root := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/launch-proof\n\ngo 1.24\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, ".gavel.yaml"), []byte("test:\n  runtime:\n    env:\n      LAUNCH_CONFIG: configured-value\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "proof_test.go"), []byte(`package proof

import (
  "os"
  "testing"
)

func TestPreparedEnvironment(t *testing.T) {
  if os.Getenv("LAUNCH_PROOF") != "prepared-value" { t.Fatal("prepared environment was not passed to the native runner") }
  if os.Getenv("LAUNCH_CONFIG") != "configured-value" { t.Fatal("workspace runtime configuration was not passed to the native runner") }
}
`), 0o600)).To(Succeed())
		executor := NativeExecutor{Schemas: schemaFixture{}}
		execution, err := executor.Prepare(context.Background(), PreparedWorkspace{Directory: root, Env: []string{"GOWORK=off", "LAUNCH_PROOF=prepared-value"}}, Configuration{Test: map[string]any{"paths": []string{"."}, "framework": []string{"go test"}}})
		Expect(err).NotTo(HaveOccurred())
		var tests []parsers.Test
		Expect(execution.Run(context.Background(), func(update resultstore.Run) error {
			if update.Tests != nil {
				tests = update.Tests
			}
			return nil
		})).To(Succeed())
		Expect(parsers.Tests(tests).Sum().Passed).To(BeNumerically(">", 0))
		Expect(parsers.Tests(tests).Sum().Failed).To(BeZero())
	})
})
