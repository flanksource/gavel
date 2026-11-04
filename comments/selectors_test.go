package comments_test

import (
	"encoding/json"
	"github.com/flanksource/gavel/comments"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("captured code selectors", func() {
	DescribeTable("rejects unsafe or incomplete anchors", func(selector string) {
		_, err := comments.ParseCodeSelector(json.RawMessage(selector))
		Expect(err).To(HaveOccurred())
	},
		Entry("traversal", `{"kind":"code-range","version":1,"path":"../secret","revision":"a","startLine":1,"endLine":1}`),
		Entry("missing commits", `{"kind":"diff-range","version":1,"path":"a.go","side":"new","startLine":1,"endLine":1}`),
		Entry("reversed range", `{"kind":"code-range","version":1,"path":"a.go","revision":"a","startLine":3,"endLine":1}`),
		Entry("unsupported version", `{"kind":"code-range","version":2,"path":"a.go","revision":"a","startLine":1,"endLine":1}`),
	)
	It("round trips both captured commits and the selected side", func() {
		raw := json.RawMessage(`{"kind":"diff-range","version":1,"path":"src/a.go","baseCommit":"base","headCommit":"head","side":"old","startLine":2,"endLine":4}`)
		s, err := comments.ParseCodeSelector(raw)
		Expect(err).NotTo(HaveOccurred())
		got, err := json.Marshal(s)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(MatchJSON(raw))
	})
})
