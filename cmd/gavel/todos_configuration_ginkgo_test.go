package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TODO project configuration CLI", func() {
	It("identifies the invalid project layer during dry-run even when a flag overrides it", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		resetTodosRunFlags(GinkgoTB())
		dir := GinkgoT().TempDir()
		provider := testProviderFor(dir)
		todo, err := provider.Create(context.Background(), todos.CreateRequest{Title: "Preview configuration", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("ai:\n  budget:\n    maxTurns: -1\n"), 0o600)).To(Succeed())
		todosStep, maxTurns, dryRun = "plan", 5, true
		captureStdout(GinkgoTB(), func() {
			err = runTodoStep(context.Background(), dir, provider, todo, todosRunOptions())
		})
		Expect(err).To(MatchError(And(
			ContainSubstring("configuration"), ContainSubstring(".gavel.yaml"), ContainSubstring("maxTurns"),
		)))
	})
})
