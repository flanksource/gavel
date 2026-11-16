package comments_test

import (
	"github.com/flanksource/gavel/comments"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Text quotes", func() {
	source := "intro\nThe quick brown fox\njumps over\nthe lazy dog\noutro"
	patch, err := comments.CapturePatch(source, 2, 3)
	It("captures the patch", func() { Expect(err).NotTo(HaveOccurred()) })
	DescribeTable("lie inside the patch's selected lines",
		func(quote comments.TextQuote, valid bool) {
			err := quote.Within(patch)
			if valid {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(MatchError(comments.ErrInvalid))
			}
		},
		Entry("words within one line", comments.TextQuote{Prefix: "The ", Exact: "quick brown", Suffix: " fox"}, true),
		Entry("spanning the selected lines", comments.TextQuote{Exact: "fox\njumps"}, true),
		Entry("text from the context lines only", comments.TextQuote{Exact: "lazy dog"}, false),
		Entry("text not in the file", comments.TextQuote{Exact: "slow tortoise"}, false),
		Entry("an empty exact", comments.TextQuote{Prefix: "The"}, false),
		Entry("a CRLF quote across the selected lines", comments.TextQuote{Exact: "fox\r\njumps"}, true),
	)
	It("matches a quote against lines a CRLF file kept their carriage returns in", func() {
		crlf := comments.PatchAnchor{Version: 1, StartLine: 1, EndLine: 2, Lines: []string{"first\r", "second"}}
		Expect(comments.TextQuote{Exact: "first\nsecond"}.Within(crlf)).To(Succeed())
	})
	It("requires exact text of a quote with no patch to check it against", func() {
		Expect(comments.TextQuote{Prefix: "web "}.Validate()).To(MatchError(comments.ErrInvalid))
	})
})

var _ = Describe("Image regions", func() {
	DescribeTable("are boxes of fractions inside the image",
		func(region comments.Region, valid bool) {
			err := region.Validate()
			if valid {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(MatchError(comments.ErrInvalid))
			}
		},
		Entry("the whole image", comments.Region{X: 0, Y: 0, W: 1, H: 1}, true),
		Entry("a box in the lower right", comments.Region{X: 0.5, Y: 0.25, W: 0.5, H: 0.75}, true),
		Entry("zero width", comments.Region{X: 0.1, Y: 0.1, W: 0, H: 0.2}, false),
		Entry("negative origin", comments.Region{X: -0.1, Y: 0.1, W: 0.2, H: 0.2}, false),
		Entry("past the right edge", comments.Region{X: 0.6, Y: 0.1, W: 0.5, H: 0.2}, false),
		Entry("past the bottom edge", comments.Region{X: 0.1, Y: 0.9, W: 0.2, H: 0.2}, false),
	)
})
