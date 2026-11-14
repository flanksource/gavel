package library_test

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/prompts/library"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Render", func() {
	It("preserves the system role of branch context", func() {
		got, err := library.Render(context.Background(), library.RenderOptions{Directory: GinkgoT().TempDir(), Name: "todos-reuse-branch", System: true, Data: map[string]any{"branch": "fix/parser", "setup": "abc123", "head": "def456"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(got.User).To(BeEmpty())
		Expect(got.System).To(ContainSubstring("branch `fix/parser`"))
		Expect(got.System).To(ContainSubstring("abc123..def456"))
	})
	It("uses a workspace override and leaves feedback literal", func() {
		dir := GinkgoT().TempDir()
		prompts := filepath.Join(dir, ".captain", "prompts")
		Expect(os.MkdirAll(prompts, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(prompts, "todos-plan-revision.prompt"), []byte("Reviewer: {{{feedback}}}"), 0o600)).To(Succeed())
		got, err := library.Render(context.Background(), library.RenderOptions{Directory: dir, Name: "todos-plan-revision", Data: map[string]any{"feedback": "{{unexpanded}} <tag>"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(got.User).To(Equal("Reviewer: {{unexpanded}} <tag>"))
	})
	It("reports invalid workspace overrides", func() {
		dir := GinkgoT().TempDir()
		prompts := filepath.Join(dir, ".captain", "prompts")
		Expect(os.MkdirAll(prompts, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(prompts, "todos-plan-revision.prompt"), []byte("{{#if}"), 0o600)).To(Succeed())
		_, err := library.Render(context.Background(), library.RenderOptions{Directory: dir, Name: "todos-plan-revision"})
		Expect(err).To(HaveOccurred())
	})
})
