package azuredevops

import (
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Azure PR auto-complete", func() {
	const actor = "342bacfe-06de-43fc-83a4-b9732983482c"

	DescribeTable("verifies that Azure accepted auto-complete", func(response string, accepted bool) {
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			switch r.URL.Path {
			case "/acme/product/_apis/git/repositories/service":
				fmt.Fprint(w, fixtureRepository)
			case "/acme/_apis/connectionData":
				Expect(r.URL.Query().Get("api-version")).To(Equal("7.1-preview.1"))
				fmt.Fprintf(w, `{"authenticatedUser":{"id":%q}}`, actor)
			case "/acme/product/_apis/git/repositories/repo-id/pullrequests/27":
				Expect(r.Method).To(Equal(http.MethodPatch))
				fmt.Fprint(w, response)
			default:
				Fail("unexpected request: " + r.URL.Path)
			}
		})
		err := c.EnableAutoMerge(context.Background(), 27, "rebase")
		if accepted {
			Expect(err).NotTo(HaveOccurred())
		} else {
			Expect(err).To(MatchError(ContainSubstring("did not confirm auto-complete")))
		}
	},
		Entry("active with auto-complete enabled", `{"pullRequestId":27,"status":"active","autoCompleteSetBy":{"id":"`+actor+`"}}`, true),
		Entry("completed immediately", `{"pullRequestId":27,"status":"completed"}`, true),
		Entry("active without auto-complete", `{"pullRequestId":27,"status":"active"}`, false),
		Entry("wrong PR", `{"pullRequestId":28,"status":"completed"}`, false),
	)

	It("fails before updating a PR if Azure cannot identify the caller", func() {
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			switch r.URL.Path {
			case "/acme/product/_apis/git/repositories/service":
				fmt.Fprint(w, fixtureRepository)
			case "/acme/_apis/connectionData":
				fmt.Fprint(w, `{"authenticatedUser":{}}`)
			default:
				Fail("unexpected request: " + r.URL.Path)
			}
		})
		Expect(c.EnableAutoMerge(context.Background(), 27, "squash")).To(MatchError(ContainSubstring("missing authenticated user ID")))
	})
})
