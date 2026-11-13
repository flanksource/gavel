package comments_test

import (
	"github.com/flanksource/gavel/comments"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MatchRoute", func() {
	const page = "gavel://project/acme/page/team"
	routes := []comments.Route{
		{ID: "team", TargetPrefix: page, WorkflowID: "human-review", Enabled: true},
		{ID: "pages", TargetPrefix: "gavel://project/acme/page/", WorkflowID: "human-review", Enabled: true},
	}
	DescribeTable("matches a prefix only at a path boundary, preferring the longest",
		func(uri, want string) {
			r, ok := comments.MatchRoute(routes, uri, "human-review")
			Expect([]any{r.ID, ok}).To(Equal([]any{want, want != ""}))
		},
		Entry("the prefix itself", page, "team"),
		Entry("a path under the prefix", page+"/members", "team"),
		Entry("a sibling sharing the prefix's leading characters", page+"-x", "pages"),
		Entry("a path under a prefix that ends in a slash", "gavel://project/acme/page/about", "pages"),
		Entry("a sibling of a slash-terminated prefix", "gavel://project/acme/pages", ""),
	)
})
