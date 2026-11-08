package models

import (
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Source excerpts", func() {
	It("preserves absolute line numbers, indentation, and the caret column", func() {
		excerpt := NewSourceExcerpt([]byte("a\nb\nc\nd\ne\nf\n  cleanup: true\nh\ni\nj\nk\n"), 7, 3)
		Expect(excerpt.StartLine).To(Equal(4))
		Expect(excerpt.Pretty().String()).To(ContainSubstring(">  7 |   cleanup: true\n     |   ^"))
		Expect(excerpt.Content).NotTo(ContainSubstring("k"))
		data, err := json.Marshal(excerpt)
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Contains(string(data), "\\u001b")).To(BeFalse())
	})
})
