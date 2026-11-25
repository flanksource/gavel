package comments_test

import (
	"github.com/flanksource/gavel/comments"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Patch anchors", func() {
	It("follows shifted hunks with repeated selected text using surrounding context", func() {
		source := "header\nfirst\nvalue\nseparator\nsecond\nvalue\nfooter\n"
		anchor, err := comments.CapturePatch(source, 6, 6)
		Expect(err).NotTo(HaveOccurred())
		Expect(comments.ResolvePatch("inserted\n"+source, anchor)).To(Equal(comments.PatchLocation{State: "matched", StartLine: 7, EndLine: 7}))
	})
	It("retains ambiguity and deleted targets without guessing", func() {
		anchor, err := comments.CapturePatch("value", 1, 1)
		Expect(err).NotTo(HaveOccurred())
		anchor.StartLine, anchor.EndLine = 5, 5
		Expect(comments.ResolvePatch("value\nvalue", anchor).State).To(Equal("ambiguous"))
		Expect(comments.ResolvePatch("changed", anchor).State).To(Equal("outdated"))
	})
	It("normalizes CRLF and clips context at boundaries", func() {
		anchor, err := comments.CapturePatch("first\r\nlast", 1, 2)
		Expect(err).NotTo(HaveOccurred())
		Expect(comments.ResolvePatch("first\nlast", anchor)).To(Equal(comments.PatchLocation{State: "matched", StartLine: 1, EndLine: 2}))
		_, err = comments.CapturePatch("one", 0, 2)
		Expect(err).To(HaveOccurred())
	})
})
