package github

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Re-running a workflow keeps its run id, so the completed-run cache must notice.
var _ = Describe("Completed run cache", func() {
	const (
		runID     = int64(101)
		firstJob  = int64(7)
		secondJob = int64(8)
	)
	var (
		options    Options
		requests   int
		runBody    string
		jobsBody   string
		jobsStatus int
	)
	run := func(status, conclusion string) string {
		return fmt.Sprintf(`{"id":%d,"name":"CI","status":%q,"conclusion":%q}`, runID, status, conclusion)
	}
	jobs := func(id int64, status, conclusion string) string {
		return fmt.Sprintf(`{"total_count":1,"jobs":[{"id":%d,"run_id":%d,"name":"test","status":%q,"conclusion":%q}]}`, id, runID, status, conclusion)
	}
	check := func(job int64, status string) model.StatusChecks {
		return model.StatusChecks{{Name: "test", Status: status, DetailsURL: fmt.Sprintf("https://github.com/acme/widgets/actions/runs/%d/job/%d", runID, job)}}
	}
	fetch := func(logOpts RunLogOptions) *model.WorkflowRun {
		GinkgoHelper()
		result, err := FetchRunJobs(options, runID, logOpts)
		Expect(err).NotTo(HaveOccurred())
		return result
	}

	BeforeEach(func() {
		GinkgoT().Setenv(cache.EnvDatabaseDSN, dbtest.ForGinkgo(dbtest.Options{Name: "gavel_run_cache"}).DSN())
		GinkgoT().Setenv(cache.EnvDSN, "")
		GinkgoT().Setenv(cache.EnvDisable, "")
		GinkgoT().Setenv(cache.EnvDatabaseDisable, "")
		migrated, err := database.Open(GinkgoT().Context(), database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		Expect(migrated.Close()).To(Succeed())
		store, err := cache.Open()
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(store.Close)
		Expect(store.Disabled()).To(BeFalse())

		requests, jobsStatus = 0, http.StatusOK
		runBody, jobsBody = run("completed", "failure"), jobs(firstJob, "completed", "failure")
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests++
			if !strings.HasSuffix(r.URL.Path, "/jobs") {
				fmt.Fprint(w, runBody)
				return
			}
			w.WriteHeader(jobsStatus)
			fmt.Fprint(w, jobsBody)
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		options = Options{Repo: "acme/widgets", Token: "test-token", Cache: store}
	})

	It("serves a completed run its checks agree with without a request", func() {
		fetch(RunLogOptions{})
		requests = 0
		Expect(fetch(RunLogOptions{Checks: check(firstJob, "COMPLETED")}).Conclusion).To(Equal("failure"))
		Expect(requests).To(BeZero())
	})

	DescribeTable("fetches a re-run again", func(logOpts RunLogOptions) {
		fetch(RunLogOptions{})
		runBody, jobsBody = run("completed", "success"), jobs(secondJob, "completed", "success")
		rerun := fetch(logOpts)
		Expect(rerun.Conclusion).To(Equal("success"))
		Expect(rerun.Jobs[0].DatabaseID).To(Equal(secondJob))
		Expect(fetch(RunLogOptions{}).Conclusion).To(Equal("success"), "the cache holds the re-run")
	},
		Entry("when a check names a job of another attempt", RunLogOptions{Checks: check(secondJob, "COMPLETED")}),
		Entry("when a check is unfinished", RunLogOptions{Checks: check(firstJob, "IN_PROGRESS")}),
		Entry("when asked to revalidate", RunLogOptions{Revalidate: true}),
	)

	It("drops the cached run once it is running again", func() {
		fetch(RunLogOptions{})
		runBody, jobsBody = run("in_progress", ""), jobs(secondJob, "in_progress", "")
		fetch(RunLogOptions{Revalidate: true})
		Expect(fetch(RunLogOptions{}).Status).To(Equal("in_progress"))
	})

	It("marks a run GitHub could not confirm stale and never caches it as completed", func() {
		runBody, jobsBody = run("in_progress", ""), jobs(firstJob, "in_progress", "")
		fetch(RunLogOptions{})
		runBody, jobsStatus = run("completed", "success"), http.StatusBadGateway
		Expect(fetch(RunLogOptions{}).Stale).To(Equal("GitHub returned status 502"))

		jobsBody, jobsStatus = jobs(firstJob, "completed", "success"), http.StatusOK
		recovered := fetch(RunLogOptions{})
		Expect(recovered.Stale).To(BeEmpty())
		Expect(recovered.Jobs[0].Status).To(Equal("completed"))
	})

	It("refetches a run an earlier version cached with unfinished jobs", func() {
		persistCompletedRun(options.Cache, "acme/widgets", runID, &model.WorkflowRun{
			DatabaseID: runID, Status: "completed", Conclusion: "success",
			Jobs: []model.Job{{DatabaseID: firstJob, Status: "in_progress"}},
		})
		Expect(fetch(RunLogOptions{}).Jobs[0].Status).To(Equal("completed"))
	})

	It("fails a cancelled request instead of serving a stale body", func() {
		runBody, jobsBody = run("in_progress", ""), jobs(firstJob, "in_progress", "")
		fetch(RunLogOptions{})
		ctx, cancel := context.WithCancel(GinkgoT().Context())
		cancel()
		cancelled := options
		cancelled.Context = ctx
		_, err := FetchRunJobs(cancelled, runID, RunLogOptions{})
		Expect(err).To(MatchError(context.Canceled))
	})
})

var _ = Describe("PR comment pagination", func() {
	It("finds a sticky comment older than the last 100 comments", func() {
		const sticky = `<!-- sticky-comment:gavel --> [View full results](https://github.com/acme/widgets/actions/runs/101/artifacts/4242)`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprintf(w, `{"data":{"node":{"comments":{"pageInfo":{"hasPreviousPage":false},"nodes":[{"databaseId":1,"body":%q}]}}}}`, sticky)
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		pr := &graphQLPR{ID: "PR_node", Comments: graphQLCommentList{
			PageInfo: graphQLPageInfo{HasPreviousPage: true, StartCursor: "cursor"},
			Nodes:    []graphQLCommentNode{{DatabaseID: 2, Body: "later"}},
		}}
		Expect(completeComments(GinkgoT().Context(), "test-token", pr)).To(Succeed())
		Expect(FindGavelArtifacts(pr.toPRInfo().Comments)).To(Equal([]GavelArtifact{{
			StickyID: "gavel", RunID: 101, ArtifactID: 4242, CommentID: 1,
			ArtifactURL: "https://github.com/acme/widgets/actions/runs/101/artifacts/4242",
		}}))
	})
})
