package prompt

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = Describe("Source context confinement", func() {
	It("rejects traversal, absolute external files and escaping symlinks", func() {
		root, outside := GinkgoT().TempDir(), GinkgoT().TempDir()
		external := filepath.Join(outside, "secret.go")
		gomega.Expect(os.WriteFile(external, []byte("private context"), 0600)).To(gomega.Succeed())
		gomega.Expect(os.Symlink(external, filepath.Join(root, "alias.go"))).To(gomega.Succeed())
		relative, err := filepath.Rel(root, external)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		for _, file := range []string{relative, external, "alias.go"} {
			text, err := ReadSourceLines(root, types.PathRef{File: file})
			gomega.Expect(err).To(gomega.HaveOccurred(), file)
			gomega.Expect(text).To(gomega.BeEmpty())
		}
	})
})
