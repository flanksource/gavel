package status

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/gavel/snapshots"
	testui "github.com/flanksource/gavel/testrunner/ui"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("status result enrichment", func() {
	It("skips snapshot enrichment when results are disabled", func() {
		repo := GinkgoT().TempDir()
		for _, args := range [][]string{
			{"init"},
			{"config", "user.email", "test@example.com"},
			{"config", "user.name", "Test User"},
			{"config", "commit.gpgsign", "false"},
		} {
			command := exec.Command("git", args...)
			command.Dir = repo
			output, err := command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))
		}
		Expect(os.WriteFile(filepath.Join(repo, "README.md"), []byte("# test\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repo, "a.go"), []byte("package x\n"), 0o644)).To(Succeed())
		for _, args := range [][]string{{"add", "README.md"}, {"commit", "-m", "initial commit"}, {"add", "a.go"}} {
			command := exec.Command("git", args...)
			command.Dir = repo
			output, err := command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))
		}

		restore := stubSnapshot(func(context.Context, string) (string, string, error) {
			Fail("snapshot identity should not be read when results are disabled")
			return "", "", nil
		}, func(string, string) (*snapshots.Pointer, error) {
			Fail("snapshot pointer should not be read when results are disabled")
			return nil, nil
		}, func(string, *snapshots.Pointer) (*testui.Snapshot, error) {
			Fail("snapshot should not be read when results are disabled")
			return nil, nil
		})
		DeferCleanup(restore)

		result, err := Gather(repo, Options{NoRepomap: true, NoResults: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Files).To(HaveLen(1))
		Expect(result.ResultsSHA).To(BeEmpty())
		Expect(result.Files[0].TestStatus).To(BeZero())
		Expect(result.Files[0].LintStatus).To(BeZero())
	})
})
