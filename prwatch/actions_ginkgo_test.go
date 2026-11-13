package prwatch

import (
	"context"
	"time"

	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("standalone Actions snapshots", func() {
	DescribeTable("uses run conclusions even when no jobs started", func(conclusion string, code int) {
		result := &PRWatchResult{Runs: map[int64]*model.WorkflowRun{101: {DatabaseID: 101, Status: "completed", Conclusion: conclusion}}}
		Expect(followDone(newResultFilters(nil, nil), result, false)).To(BeTrue())
		Expect(statusExitCode(result)).To(Equal(code))
	},
		Entry("success", "success", 0),
		Entry("startup failure", "startup_failure", 1),
		Entry("cancelled", "cancelled", 1),
		Entry("action required", "action_required", 1),
		Entry("stale", "stale", 1),
	)

	It("follows standalone snapshots and renders their progress without a PR", func() {
		polls := 0
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		result, code := Run(WatchOptions{Context: ctx, Follow: true, Interval: time.Millisecond,
			FetchSnapshot: func(context.Context, WatchOptions) (*PRWatchResult, error) {
				polls++
				status := "in_progress"
				if polls == 2 {
					status = "completed"
				}
				return &PRWatchResult{Runs: map[int64]*model.WorkflowRun{101: {DatabaseID: 101, Name: "CI", Status: status, Conclusion: "success"}}}, nil
			},
		})
		Expect(result).NotTo(BeNil())
		Expect(code).To(BeZero())
		Expect(polls).To(Equal(2))
		Expect(result.Pretty().String()).To(ContainSubstring("CI"))
	})
})
