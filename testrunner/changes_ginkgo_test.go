package testrunner

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func changeRepository(files map[string]string) string {
	dir := GinkgoT().TempDir()
	for path, content := range files {
		Expect(os.MkdirAll(filepath.Dir(filepath.Join(dir, path)), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, path), []byte(content), 0o600)).To(Succeed())
	}
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.name", "Test"}, {"config", "user.email", "test@example.com"}, {"config", "commit.gpgsign", "false"}, {"add", "--", "."}, {"commit", "-qm", "initial"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(out))
	}
	GinkgoT().Setenv("GOWORK", "off")
	return dir
}

var _ = Describe("Shared change selection", func() {
	It("selects changed Go packages and their reverse dependencies", func() {
		dir := changeRepository(map[string]string{
			"go.mod": "module example.com/selection\n\ngo 1.22\n",
			"a/a.go": "package a\nimport \"example.com/selection/b\"\nvar Value = b.Value\n",
			"b/b.go": "package b\nvar Value = 1\n", "c/c.go": "package c\n",
		})
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: dir, ChangedFiles: []string{"b/b.go"}}}
		packages := map[Framework][]string{parsers.GoTest: {"./a", "./b", "./c"}}
		Expect(o.selectChanges(packages)).To(Succeed())
		Expect(packages[parsers.GoTest]).To(Equal([]string{"./a", "./b"}))
	})
	It("keeps non-Go projects without requiring a Go module", func() {
		dir := changeRepository(map[string]string{"sample.js": "export const value = 1\n"})
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: dir, ChangedFiles: []string{"sample.js"}}}
		packages := map[Framework][]string{parsers.Vitest: {"."}}
		Expect(o.selectChanges(packages)).To(Succeed())
		Expect(packages[parsers.Vitest]).To(Equal([]string{"."}))
	})
	It("clears all frameworks for an empty changeset", func() {
		dir := changeRepository(map[string]string{"sample.js": "export const value = 1\n"})
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: dir, Since: "HEAD"}}
		packages := map[Framework][]string{parsers.Vitest: {"."}}
		Expect(o.selectChanges(packages)).To(Succeed())
		Expect(packages[parsers.Vitest]).To(BeEmpty())
	})
})
