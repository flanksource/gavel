package fixtures

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fixture runtime isolation", func() {
	It("fails the invocation when setup cleanup fails", func() {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "cleanup.md")
		Expect(os.WriteFile(path, []byte("---\nsetup:\n  envVars: [{name: GAVEL_CLEANUP_TEST, value: child}]\n---\n\n| Name | Command |\n| --- | --- |\n| passing command | true |\n"), 0o600)).To(Succeed())
		var runner *Runner
		var err error
		runner, err = NewRunner(RunnerOptions{Paths: []string{path}, WorkDir: dir, ProgressSink: func(context.Context, ExecutionSnapshot) error {
			for _, setup := range runner.setups {
				setup.cleanup = func() error { return fmt.Errorf("cleanup failed") }
			}
			return nil
		}})
		Expect(err).NotTo(HaveOccurred())
		_, err = runner.Run()
		Expect(err).To(MatchError(ContainSubstring("cleanup failed")))
	})
})
