package testrunner

import (
	"os"
	"path/filepath"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Go batch package identities", func() {
	It("maps module packages independently of overlapping suffixes and event order", func() {
		root := GinkgoT().TempDir()
		const module = "example.com/service"
		Expect(os.WriteFile(filepath.Join(root, "go.mod"), []byte("module "+module+"\n\ngo 1.26.1\n"), 0o600)).To(Succeed())
		imports := []string{module + "/examples/comments", module + "/comments/lint", module + "/comments", module + "/lint"}
		expected := map[string]string{
			module + "/comments":          "./comments",
			module + "/lint":              "./lint",
			module + "/examples/comments": "./examples/comments",
			module + "/comments/lint":     "./comments/lint",
		}
		for range 2 {
			paths, err := resolveGoBatchPaths(root, []string{"./comments", "./lint"}, imports, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(paths).To(Equal(expected))
			slices.Reverse(imports)
		}
	})

	It("reports an unreported module package instead of assigning another package to it", func() {
		root := GinkgoT().TempDir()
		const module = "example.com/service"
		Expect(os.WriteFile(filepath.Join(root, "go.mod"), []byte("module "+module+"\n\ngo 1.26.1\n"), 0o600)).To(Succeed())
		paths, err := resolveGoBatchPaths(root, []string{"./comments"}, []string{module + "/examples/comments"}, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(Equal(map[string]string{
			module + "/examples/comments": "./examples/comments",
			"gavel-unreported:./comments": "./comments",
		}))
	})
})
