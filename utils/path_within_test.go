package utils

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResolveWithin", func() {
	const (
		nested       = "prompts/grouping.prompt"
		escapingName = "../../etc/passwd"
	)
	var base string

	BeforeEach(func() {
		base = GinkgoT().TempDir()
	})

	It("resolves a relative name to an absolute path under the base", func() {
		Expect(ResolveWithin(base, nested)).To(Equal(filepath.Join(base, nested)))
	})

	It("accepts an absolute name that already lives under the base", func() {
		Expect(ResolveWithin(base, filepath.Join(base, nested))).To(Equal(filepath.Join(base, nested)))
	})

	// A workspace naming its own root — the dashboard's commonest dir param —
	// resolves to the root, not to an escape.
	It("accepts the base itself", func() {
		Expect(ResolveWithin(base, base)).To(Equal(filepath.Clean(base)))
		Expect(ResolveWithin(base, ".")).To(Equal(filepath.Clean(base)))
	})

	It("rejects a relative name that walks out of the base", func() {
		_, err := ResolveWithin(base, escapingName)

		Expect(err).To(MatchError(ErrPathEscapesBase))
		Expect(err).To(MatchError(ContainSubstring(escapingName)))
		Expect(err).To(MatchError(ContainSubstring(base)))
	})

	It("rejects an absolute name that points outside the base", func() {
		outside := filepath.Join(GinkgoT().TempDir(), nested)

		_, err := ResolveWithin(base, outside)

		Expect(err).To(MatchError(ErrPathEscapesBase))
		Expect(err).To(MatchError(ContainSubstring(outside)))
	})

	It("rejects a name that climbs above the base and back down into a sibling", func() {
		sibling := filepath.Join("..", filepath.Base(base)+"-other", nested)

		_, err := ResolveWithin(base, sibling)

		Expect(err).To(MatchError(ErrPathEscapesBase))
	})

	It("requires both a base and a name", func() {
		_, err := ResolveWithin(base, "")
		Expect(err).To(MatchError(ContainSubstring("path is required")))

		_, err = ResolveWithin("", nested)
		Expect(err).To(MatchError(ContainSubstring("base directory is required")))
	})
})

var _ = Describe("ReadWithin", func() {
	DescribeTable("reads files and internal relative symlinks", func(absoluteBase, absoluteName, symlink bool) {
		base := GinkgoT().TempDir()
		const name = "workflow.yaml"
		want := []byte("order: [run, plan]\n")
		Expect(os.WriteFile(filepath.Join(base, name), want, 0600)).To(Succeed())
		path := name
		if symlink {
			path = "linked.yaml"
			Expect(os.Symlink(name, filepath.Join(base, path))).To(Succeed())
		}
		if absoluteName {
			path = filepath.Join(base, path)
		}
		if !absoluteBase {
			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			base, err = filepath.Rel(cwd, base)
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(ReadWithin(base, path)).To(Equal(want))
	},
		Entry("absolute base, relative file", true, false, false),
		Entry("absolute base and file", true, true, false),
		Entry("relative base, relative file", false, false, false),
		Entry("relative base, absolute file", false, true, false),
		Entry("relative symlink within the base", true, false, true),
	)
})
