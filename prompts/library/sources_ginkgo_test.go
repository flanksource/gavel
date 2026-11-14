package library_test

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/prompts/library"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Local prompt source containment", func() {
	DescribeTable("rejects prompt directories symlinked outside the workspace", func(component string) {
		root, outside := GinkgoT().TempDir(), GinkgoT().TempDir()
		if component == ".captain" {
			Expect(os.Mkdir(filepath.Join(outside, "prompts"), 0700)).To(Succeed())
		} else {
			Expect(os.Mkdir(filepath.Join(root, ".captain"), 0700)).To(Succeed())
		}
		Expect(os.Symlink(outside, filepath.Join(root, component))).To(Succeed())
		_, err := library.Context(context.Background(), root)
		Expect(err).To(HaveOccurred())
	}, Entry("parent", ".captain"), Entry("directory", ".captain/prompts"))

	It("accepts a workspace with no local prompt directory", func() {
		_, err := library.Context(context.Background(), GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
	})
})
