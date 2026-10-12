package prwatch

import (
	"context"
	"fmt"
	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
)

func azureResult(state string, checks model.StatusChecks) *PRWatchResult {
	return &PRWatchResult{PR: &model.PRInfo{Provider: "azuredevops", State: "OPEN", MergeReadiness: &model.MergeReadiness{State: state}, StatusCheckRollup: checks}}
}

var _ = Describe("Azure status watches", func() {
	It("fails an action selector on a settled PR with no pipelines", func() {
		Expect(newResultFilters(nil, []string{"missing"}).noActionMatch(0, 0, azureResult("ready", nil))).To(BeTrue())
	})
	It("does not leak a failed sibling into a selected successful job", func() {
		r := azureResult("ready", model.StatusChecks{{Name: "Test", RunID: 101, JobID: "test", Status: "COMPLETED", Conclusion: "SUCCESS"}, {Name: "Deploy", RunID: 101, JobID: "deploy", Status: "COMPLETED", Conclusion: "FAILURE"}})
		r.Runs = map[int64]*model.WorkflowRun{101: {DatabaseID: 101, Name: "CI", Conclusion: "FAILURE", Jobs: []model.Job{{NativeID: "test", Name: "Test", Status: "COMPLETED", Conclusion: "SUCCESS"}, {NativeID: "deploy", Name: "Deploy", Status: "COMPLETED", Conclusion: "FAILURE"}}}}
		newResultFilters(nil, []string{"Test"}).apply(r)
		Expect(statusExitCode(r)).To(BeZero())
		Expect(r.Runs[101].Conclusion).To(Equal("SUCCESS"))
	})
	It("finishes empty pipelines when merge readiness has settled", func() {
		r := azureResult("blocked", nil)
		Expect(followDone(newResultFilters(nil, nil), r, false)).To(BeTrue())
		Expect(statusExitCode(r)).To(Equal(1))
	})
	It("waits for queued build validation without waiting for human policies", func() {
		r := azureResult("pending", nil)
		r.PR.MergeReadiness.WaitingForBuilds = true
		Expect(followDone(newResultFilters(nil, nil), r, false)).To(BeFalse())
		r.PR.MergeReadiness.WaitingForBuilds = false
		r.PR.MergeReadiness.WaitingForMerge = true
		Expect(followDone(newResultFilters(nil, nil), r, false)).To(BeFalse())
	})
	It("returns failure for cancelled Azure checks", func() {
		r := azureResult("ready", model.StatusChecks{{Status: "COMPLETED", Conclusion: "CANCELLED"}})
		Expect(statusExitCode(r)).To(Equal(1))
	})
	It("scopes Azure jobs by native UUID and retains their checks", func() {
		r := azureResult("ready", model.StatusChecks{{Name: "Test", WorkflowName: "CI", RunID: 101, JobID: "job-uuid", Status: "COMPLETED", Conclusion: "SUCCESS"}})
		r.Runs = map[int64]*model.WorkflowRun{101: {DatabaseID: 101, Name: "CI", Jobs: []model.Job{{NativeID: "job-uuid", Name: "Test"}, {NativeID: "deploy-uuid", Name: "Deploy"}}}}
		newResultFilters(nil, []string{"job-uuid"}).apply(r)
		Expect(r.Runs[101].Jobs).To(HaveLen(1))
		Expect(r.PR.StatusCheckRollup).To(HaveLen(1))
	})
	It("follows injected snapshots until pipelines complete", func() {
		polls := 0
		opts := WatchOptions{Context: context.Background(), Follow: true, Interval: time.Millisecond, FetchSnapshot: func(context.Context, WatchOptions) (*PRWatchResult, error) {
			polls++
			status := "IN_PROGRESS"
			if polls == 2 {
				status = "COMPLETED"
			}
			return azureResult("ready", model.StatusChecks{{Name: "Test", Status: status, Conclusion: "SUCCESS"}}), nil
		}}
		result, code := Run(opts)
		Expect(result).NotTo(BeNil())
		Expect(code).To(BeZero())
		Expect(polls).To(Equal(2))
	})
	It("does not retry an injected provider permission failure", func() {
		polls := 0
		_, code := Run(WatchOptions{Context: context.Background(), Follow: true, FetchSnapshot: func(context.Context, WatchOptions) (*PRWatchResult, error) {
			polls++
			return nil, fmt.Errorf("HTTP 403")
		}})
		Expect(code).To(Equal(1))
		Expect(polls).To(Equal(1))
	})
})
