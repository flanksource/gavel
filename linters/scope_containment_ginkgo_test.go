package linters

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scope metadata confinement", func() {
	It("does not classify an escaping directory symlink as a scope directory", func() {
		root, outside := GinkgoT().TempDir(), GinkgoT().TempDir()
		alias := filepath.Join(root, "alias")
		Expect(os.Symlink(outside, alias)).To(Succeed())
		Expect(pathMatchesAnyScope(filepath.Join(alias, "file.go"), []string{alias})).To(BeFalse())
	})
})
