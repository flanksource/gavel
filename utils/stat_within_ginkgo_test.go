package utils

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Confined filesystem metadata", func() {
	It("rejects file and directory symlinks outside the selected root", func() {
		root, outside := GinkgoT().TempDir(), GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(outside, "data"), []byte("outside"), 0600)).To(Succeed())
		Expect(os.Symlink(outside, filepath.Join(root, "directory"))).To(Succeed())
		Expect(os.Symlink(filepath.Join(outside, "data"), filepath.Join(root, "file"))).To(Succeed())
		for _, name := range []string{"file", "directory/data"} {
			_, err := StatWithin(root, name)
			Expect(err).To(HaveOccurred(), name)
		}
	})
	It("accepts an internal relative symlink and preserves missing-file errors", func() {
		root := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(root, "data"), []byte("inside"), 0600)).To(Succeed())
		Expect(os.Symlink("data", filepath.Join(root, "alias"))).To(Succeed())
		info, err := StatWithin(root, "alias")
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Size()).To(Equal(int64(len("inside"))))
		_, err = StatWithin(root, "missing")
		Expect(os.IsNotExist(err)).To(BeTrue())
	})
	It("does not identify an escaping git marker as a repository", func() {
		root, outside := GinkgoT().TempDir(), GinkgoT().TempDir()
		Expect(os.Symlink(outside, filepath.Join(root, ".git"))).To(Succeed())
		Expect(FindGitRoot(root)).To(BeEmpty())
	})
})
