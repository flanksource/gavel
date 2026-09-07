package utils

import (
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
