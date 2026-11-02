package workflows

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
)

var _ = Describe("atomic lifecycle file writes", func() {
	It("preserves permissions and removes its temporary file", func() {
		dir := GinkgoT().TempDir()
		const name = "workflow.yaml"
		g.Expect(os.WriteFile(filepath.Join(dir, name), []byte("original"), 0640)).To(g.Succeed())
		root, err := os.OpenRoot(dir)
		g.Expect(err).NotTo(g.HaveOccurred())
		DeferCleanup(root.Close)
		g.Expect(writeAtomic(root, name, []byte("updated"))).To(g.Succeed())
		g.Expect(root.ReadFile(name)).To(g.Equal([]byte("updated")))
		info, err := root.Stat(name)
		g.Expect(err).NotTo(g.HaveOccurred())
		g.Expect(info.Mode().Perm()).To(g.Equal(os.FileMode(0640)))
		entries, err := os.ReadDir(dir)
		g.Expect(err).NotTo(g.HaveOccurred())
		g.Expect(entries).To(g.HaveLen(1))
	})

	DescribeTable("refuses to write through an escaping symlink", func(linkDirectory bool) {
		dir, outside := GinkgoT().TempDir(), GinkgoT().TempDir()
		const name = "workflow.yaml"
		original := []byte("outside")
		g.Expect(os.WriteFile(filepath.Join(outside, name), original, 0600)).To(g.Succeed())
		path := name
		if linkDirectory {
			g.Expect(os.Symlink(outside, filepath.Join(dir, "linked"))).To(g.Succeed())
			path = "linked/" + name
		} else {
			g.Expect(os.Symlink(filepath.Join(outside, name), filepath.Join(dir, name))).To(g.Succeed())
		}
		root, err := os.OpenRoot(dir)
		g.Expect(err).NotTo(g.HaveOccurred())
		DeferCleanup(root.Close)
		g.Expect(writeAtomic(root, path, []byte("updated"))).To(g.MatchError(g.ContainSubstring("path escapes")))
		g.Expect(os.ReadFile(filepath.Join(outside, name))).To(g.Equal(original))
	}, Entry("a file", false), Entry("a parent directory", true))
})
