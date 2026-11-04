package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/flanksource/gavel/github/cache"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Actions job targets", func() {
	DescribeTable("rejects zero IDs before PR resolution", func(url string) {
		_, _, _, err := ParseRunURL(url)
		Expect(err).To(MatchError(ContainSubstring("positive integer")))
	},
		Entry("run", "https://github.com/acme/widgets/actions/runs/0"),
		Entry("job", "https://github.com/acme/widgets/actions/runs/101/job/0"),
	)

	var requests []string
	var options Options
	BeforeEach(func() {
		GinkgoT().Setenv("GAVEL_DB", "off")
		store, err := cache.Open()
		Expect(err).NotTo(HaveOccurred())
		options = Options{Repo: "acme/widgets", Token: "test-token", Cache: store}
		requests = nil
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests = append(requests, r.URL.RequestURI())
			switch r.URL.Path {
			case "/repos/acme/widgets/actions/runs/101":
				fmt.Fprint(w, `{"id":101,"name":"CI","status":"completed","conclusion":"failure","path":".github/workflows/ci.yml","pull_requests":[]}`)
			case "/repos/acme/widgets/actions/jobs/7":
				fmt.Fprint(w, `{"id":7,"run_id":101,"name":"old attempt","status":"completed","conclusion":"success"}`)
			case "/repos/acme/widgets/actions/jobs/8":
				fmt.Fprint(w, `{"id":8,"run_id":999,"name":"other run","status":"completed","conclusion":"success"}`)
			case "/repos/acme/widgets/actions/runs/101/jobs":
				if r.URL.Query().Get("page") == "1" {
					fmt.Fprint(w, `{"total_count":2,"jobs":[{"id":10,"name":"latest attempt","status":"completed","conclusion":"success"}]}`)
				} else {
					fmt.Fprint(w, `{"total_count":2,"jobs":[{"id":11,"name":"later failure","status":"completed","conclusion":"failure"}]}`)
				}
			default:
				http.Error(w, "unexpected request", http.StatusNotFound)
			}
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
	})

	It("fetches an exact job from an earlier attempt without consulting latest jobs", func() {
		run, err := FetchRunJobs(options, 101, RunLogOptions{JobID: 7})
		Expect(err).NotTo(HaveOccurred())
		Expect(run.Jobs).To(HaveLen(1))
		Expect(run.Jobs[0].Name).To(Equal("old attempt"))
		Expect(run.WorkflowPath).To(Equal(".github/workflows/ci.yml"))
		Expect(requests).To(Equal([]string{"/repos/acme/widgets/actions/runs/101", "/repos/acme/widgets/actions/jobs/7"}))
	})

	It("rejects a job that belongs to another run", func() {
		_, err := FetchRunJobs(options, 101, RunLogOptions{JobID: 8})
		Expect(err).To(MatchError(ContainSubstring("belongs to run 999")))
	})

	It("retains failing jobs beyond the first page", func() {
		run, err := FetchRunJobs(options, 101, RunLogOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(run.Jobs).To(HaveLen(2))
		Expect(run.Jobs[1].Conclusion).To(Equal("failure"))
		Expect(requests).To(ContainElement("/repos/acme/widgets/actions/runs/101/jobs?per_page=100&page=2"))
	})
})
