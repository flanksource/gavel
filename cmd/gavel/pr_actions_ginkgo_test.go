package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/flanksource/gavel/prwatch"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Actions URL status", func() {
	var requests []string
	var polls int
	BeforeEach(func() {
		requests, polls = nil, 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests = append(requests, r.URL.Path)
			switch r.URL.Path {
			case "/repos/acme/widgets/actions/runs/101":
				fmt.Fprint(w, `{"id":101,"name":"CI","status":"in_progress","head_sha":"abc123","head_branch":"main","pull_requests":[]}`)
			case "/repos/acme/widgets/actions/runs/101/jobs":
				polls++
				status, conclusion := "in_progress", ""
				if polls > 1 {
					status, conclusion = "completed", "success"
				}
				fmt.Fprintf(w, `{"jobs":[{"id":7,"name":"build","status":%q,"conclusion":%q},{"id":8,"name":"lint","status":"completed","conclusion":"failure"}]}`, status, conclusion)
			case "/repos/acme/widgets/actions/jobs/7":
				polls++
				status, conclusion := "in_progress", ""
				if polls > 1 {
					status, conclusion = "completed", "success"
				}
				fmt.Fprintf(w, `{"id":7,"run_id":101,"name":"build","status":%q,"conclusion":%q}`, status, conclusion)
			case "/repos/acme/widgets/actions/jobs/8":
				fmt.Fprint(w, `{"id":8,"run_id":101,"name":"lint","status":"completed","conclusion":"failure"}`)
			case "/repos/acme/widgets/actions/jobs/8/logs":
				fmt.Fprint(w, "lint failed: invalid configuration\n")
			default:
				http.Error(w, "unexpected request", http.StatusNotFound)
			}
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		GinkgoT().Setenv("GITHUB_TOKEN", "test-token")
		GinkgoT().Setenv("GAVEL_DB", "off")
		DeferCleanup(func() { exitCode = 0 })
	})

	It("follows a main-branch job without resolving a PR or waiting for siblings", func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		result, err := runPRStatus(PRStatusOptions{Context: ctx, Args: []string{"https://github.com/acme/widgets/actions/runs/101/job/7"}, Follow: true, Interval: "1ms"})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())
		status := result.(*prwatch.PRWatchResult)
		Expect(status.PR).To(BeNil())
		Expect(status.Runs[101].Jobs).To(HaveLen(1))
		Expect(status.Runs[101].Jobs[0].DatabaseID).To(Equal(int64(7)))
		Expect(polls).To(Equal(2))
		Expect(exitCode).To(BeZero())
		Expect(status.Pretty().String()).To(ContainSubstring("build"))
		Expect(requests).NotTo(ContainElement(ContainSubstring("/pulls")))
		Expect(requests).NotTo(ContainElement("/repos/acme/widgets/actions/runs/101/jobs"))
	})

	It("reports failure for a run URL containing a failed job", func() {
		result, err := runPRStatus(PRStatusOptions{Args: []string{"https://github.com/acme/widgets/actions/runs/101"}, Interval: "1ms"})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())
		Expect(exitCode).To(Equal(1))
	})

	It("fails when the URL's job is absent from the run", func() {
		result, err := runPRStatus(PRStatusOptions{Args: []string{"https://github.com/acme/widgets/actions/runs/101/job/999"}, Interval: "1ms"})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeNil())
		Expect(exitCode).To(Equal(1))
	})

	It("fetches and renders the linked job's logs", func() {
		result, err := runPRStatus(PRStatusOptions{Args: []string{"https://github.com/acme/widgets/actions/runs/101/job/8"}, Logs: true, TailLogs: 10, Interval: "1ms"})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())
		Expect(result.(*prwatch.PRWatchResult).Pretty().String()).To(ContainSubstring("lint failed: invalid configuration"))
		Expect(requests).To(ContainElement("/repos/acme/widgets/actions/jobs/8/logs"))
		Expect(exitCode).To(Equal(1))
	})

	It("does not fetch logs of jobs outside the URL's scope", func() {
		_, err := runPRStatus(PRStatusOptions{Args: []string{"https://github.com/acme/widgets/actions/runs/101/job/7"}, Logs: true, Interval: "1ms"})
		Expect(err).NotTo(HaveOccurred())
		Expect(requests).NotTo(ContainElement(ContainSubstring("/logs")))
	})

	It("checks the run's commit against --head-sha", func() {
		result, err := runPRStatus(PRStatusOptions{Args: []string{"https://github.com/acme/widgets/actions/runs/101"}, HeadSHA: "HEAD", Interval: "1ms"})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeNil())
		Expect(exitCode).To(Equal(1))
	})

	DescribeTable("rejects PR-only options before fetching a run", func(opts PRStatusOptions, flag string) {
		opts.Args = []string{"https://github.com/acme/widgets/actions/runs/101/job/7"}
		_, err := runPRStatus(opts)
		Expect(err).To(MatchError(ContainSubstring(flag)))
		Expect(requests).To(BeEmpty())
	},
		Entry("whole PR", PRStatusOptions{WholePR: true}, "--whole-pr"),
		Entry("comments", PRStatusOptions{Comments: []string{"@reviewer"}}, "--comments"),
		Entry("AI fix", PRStatusOptions{AIFix: true}, "--ai-fix"),
		Entry("merge rules", PRStatusOptions{Rules: true}, "--rules"),
	)
})
