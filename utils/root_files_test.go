package utils

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("files within a base directory", func() {
	const (
		name    = "steps.yaml"
		content = "steps: []\n"
	)
	var base string

	BeforeEach(func() {
		base = GinkgoT().TempDir()
	})

	It("reads a file under the base", func() {
		Expect(os.WriteFile(filepath.Join(base, name), []byte(content), 0o600)).To(Succeed())
		Expect(ReadFileWithin(base, name)).To(Equal([]byte(content)))
	})

	It("refuses to read through a symlink that leaves the base", func() {
		outside := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(outside, name), []byte(content), 0o600)).To(Succeed())
		Expect(os.Symlink(outside, filepath.Join(base, "link"))).To(Succeed())
		_, err := ReadFileWithin(base, filepath.Join("link", name))
		Expect(err).To(HaveOccurred())
	})

	It("refuses a name that escapes the base", func() {
		_, err := ReadFileWithin(base, "../"+name)
		Expect(err).To(MatchError(ErrPathEscapesBase))
	})

	It("creates a new file with the given permissions", func() {
		Expect(WriteFileWithin(base, name, []byte(content), 0o640)).To(Succeed())
		info, err := os.Stat(filepath.Join(base, name))
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o640)))
	})

	It("replaces an existing file keeping its permissions and leaving no temp file", func() {
		path := filepath.Join(base, name)
		Expect(os.WriteFile(path, []byte("old"), 0o644)).To(Succeed())
		Expect(os.Chmod(path, 0o644)).To(Succeed())
		Expect(WriteFileWithin(base, name, []byte(content), 0o600)).To(Succeed())
		info, err := os.Stat(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o644)))
		entries, err := os.ReadDir(base)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(1))
		Expect(os.ReadFile(path)).To(Equal([]byte(content)))
	})

	It("refuses to write through a symlink that leaves the base", func() {
		outside := GinkgoT().TempDir()
		Expect(os.Symlink(outside, filepath.Join(base, "link"))).To(Succeed())
		Expect(WriteFileWithin(base, filepath.Join("link", name), []byte(content), 0o600)).NotTo(Succeed())
		Expect(filepath.Join(outside, name)).NotTo(BeAnExistingFile())
	})
})
