package todos

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ask attempt status", func() {
	It("records a question as waiting even when the execution is not successful", func() {
		result := &ExecutionResult{EndStatus: types.EndAsk}
		Expect(result.statusString()).To(Equal("waiting"))
		path := filepath.Join(GinkgoT().TempDir(), "question.md")
		Expect(os.WriteFile(path, []byte("# Question\n"), 0o600)).To(Succeed())
		todo := &types.TODO{FilePath: path}
		todo.Attempts = 1
		Expect(saveAttempt(todo, result)).To(Succeed())
		written, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(written)).To(ContainSubstring("| 1 | ask |"))
		transcript, err := os.ReadFile(filepath.Join(filepath.Dir(path), "question.attempts", "attempt-1.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(transcript)).To(ContainSubstring("**Status:** waiting"))
	})
})
