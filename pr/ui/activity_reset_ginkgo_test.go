package ui

import (
	"net/http"

	"github.com/flanksource/gavel/github/activity"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("activity reset", func() {
	It("clears HTTP and git activity while preserving tracked repositories", func() {
		f := newProjectGitFixture()
		tracked := newTrackedGitServer()
		_, err := tracked.tracker.Track(tracked.ctx, f.repo)
		Expect(err).NotTo(HaveOccurred())
		activity.Shared().Record(activity.Entry{Kind: "rest", Method: http.MethodGet, URL: "https://api.example.com/repos/acme/widgets", FromCache: true})
		DeferCleanup(activity.Shared().Reset)
		before := decodeProjectGit[gitMetricsResponse](serveProjectGit(tracked.server, http.MethodGet, "/api/git/metrics", nil))
		Expect(before.StatusScans).NotTo(BeEmpty())
		Expect(before.RangesComputed).To(BeNumerically(">", 0))

		response := serveProjectGit(tracked.server, http.MethodPost, "/api/activity/reset", nil)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(MatchJSON(`{"status":"reset"}`))
		Expect(snapshotActivity(100)).To(Equal(activitySnapshot{Entries: []activity.Entry{}, Stats: activity.Stats{ByKind: map[string]activity.KindStats{}}}))
		after := decodeProjectGit[gitMetricsResponse](serveProjectGit(tracked.server, http.MethodGet, "/api/git/metrics", nil))
		Expect(after.Tracking).To(BeTrue())
		Expect(after.TrackedRepos).To(Equal(before.TrackedRepos))
		Expect(after.RangesComputed).To(BeZero())
		Expect(after.ScanTriggers).To(BeEmpty())
		Expect(after.RefsScans).To(BeEmpty())
		Expect(after.StatusScans).To(BeEmpty())
		Expect(after.Commands).To(BeEmpty())
		Expect(after.QueueWait).To(BeEmpty())
		Expect(after.Reads).To(BeEmpty())
	})
})
