package safefs

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Contained stat", func() {
	It("accepts local absolute and relative paths and preserves missing-file errors", func() {
		root := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(root, "file.txt"), []byte("example"), 0600)).To(Succeed())
		for _, path := range []string{"file.txt", filepath.Join(root, "file.txt")} {
			info, err := Stat(root, path)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Name()).To(Equal("file.txt"))
		}
		_, err := Stat(root, "deleted/file.txt")
		Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue())
	})

	It("rejects missing roots, traversal, sibling prefixes and dangling symlink escapes", func() {
		root, outside := GinkgoT().TempDir(), GinkgoT().TempDir()
		Expect(os.Symlink(outside, filepath.Join(root, "link"))).To(Succeed())
		for _, path := range []string{"../file.txt", root + "-sibling/file.txt", "link/missing/file.txt"} {
			_, err := Stat(root, path)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, os.ErrNotExist)).To(BeFalse())
		}
		_, err := Stat("", "file.txt")
		Expect(err).To(HaveOccurred())
	})
})
