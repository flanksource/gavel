package prwatch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("whole PR verification", func() {
	It("returns provider errors to the fixture without retrying inaccessible status", func() {
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			requests++
			w.WriteHeader(http.StatusUnauthorized)
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		result, code := Run(WatchOptions{Context: ctx, Options: github.Options{Token: "fixture-token", Repo: "acme/service"}, PRNumber: 27, WholePR: true, Follow: true, Interval: time.Millisecond})
		Expect(code).To(Equal(1))
		Expect(result).To(BeNil())
		Expect(requests).To(Equal(1))
	})
	DescribeTable("does not pass incomplete or stale PR evidence", func(head string, checks model.StatusChecks, unresolved bool, want int) {
		result := &PRWatchResult{PR: &model.PRInfo{State: "OPEN", Mergeable: "MERGEABLE", MergeState: "CLEAN", HeadRefOID: head, StatusCheckRollup: checks}}
		if unresolved {
			result.PR.ReviewThreads = []model.PRComment{{ThreadID: "native", ID: 1, IsReviewThread: true, IsOutdated: true}}
		}
		_, code := Run(WatchOptions{Context: context.Background(), ExpectedHead: "pushed-head", WholePR: true, FetchSnapshot: func(context.Context, WatchOptions) (*PRWatchResult, error) { return result, nil }})
		Expect(code).To(Equal(want))
	},
		Entry("old green head", "old-head", model.StatusChecks{{Status: "COMPLETED", Conclusion: "SUCCESS"}}, false, 1),
		Entry("checks not registered", "pushed-head", model.StatusChecks{}, false, 1),
		Entry("unresolved feedback", "pushed-head", model.StatusChecks{{Status: "COMPLETED", Conclusion: "SUCCESS"}}, true, 1),
		Entry("whole PR green", "pushed-head", model.StatusChecks{{Status: "COMPLETED", Conclusion: "SUCCESS"}}, false, 0),
	)
	DescribeTable("rejects unresolved merge policies and lint findings", func(state string, lint int) {
		result := &PRWatchResult{PR: &model.PRInfo{State: "OPEN", Mergeable: "MERGEABLE", MergeState: state, StatusCheckRollup: model.StatusChecks{{Status: "COMPLETED", Conclusion: "SUCCESS"}}}, GavelResults: []*GavelResultsSummary{{LintViolations: lint}}}
		_, code := Run(WatchOptions{WholePR: true, FetchSnapshot: func(context.Context, WatchOptions) (*PRWatchResult, error) { return result, nil }})
		Expect(code).To(Equal(1))
	}, Entry("blocked policy", "BLOCKED", 0), Entry("unknown merge state", "UNKNOWN", 0), Entry("lint findings", "CLEAN", 1))
	It("waits for checks and mergeability at the pushed head", func() {
		polls := 0
		result := &PRWatchResult{PR: &model.PRInfo{State: "OPEN", HeadRefOID: "pushed", Mergeable: "UNKNOWN", MergeState: "UNKNOWN"}}
		_, code := Run(WatchOptions{Context: context.Background(), ExpectedHead: "pushed", WholePR: true, Follow: true, Interval: time.Millisecond, FetchSnapshot: func(context.Context, WatchOptions) (*PRWatchResult, error) {
			polls++
			if polls == 2 {
				result.PR.Mergeable, result.PR.MergeState = "MERGEABLE", "CLEAN"
				result.PR.StatusCheckRollup = model.StatusChecks{{Status: "COMPLETED", Conclusion: "SUCCESS"}}
			}
			return result, nil
		}})
		Expect(code).To(Equal(0))
		Expect(polls).To(Equal(2))
	})
	It("discards artifacts from an earlier PR revision", func() {
		pr := &model.PRInfo{StatusCheckRollup: model.StatusChecks{{DetailsURL: "https://github.com/acme/service/actions/runs/92/job/1"}}}
		current := github.GavelArtifact{RunID: 92}
		Expect(currentPRArtifacts(pr, []github.GavelArtifact{{RunID: 91}, current})).To(Equal([]github.GavelArtifact{current}))
	})
})
