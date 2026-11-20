// Tests a connection pull request's review decision as GitHub spells it, from its provider's
// reviewer votes and required approvals.
package provider

import (
	"github.com/flanksource/commons-db/git"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("A connection pull request's review decision", func() {
	approvals := func(status string) git.PullRequestPolicy {
		return git.PullRequestPolicy{Name: "Approvals", Status: status, Required: true, MinimumApproverCount: 2}
	}
	DescribeTable("is its reviewers' votes and required approvals", func(review git.PullRequestReview, want string) {
		Expect(reviewDecision(review)).To(Equal(want))
	},
		Entry("nobody reviewed", git.PullRequestReview{}, ""),
		Entry("someone approved", git.PullRequestReview{Reviewers: []git.PullRequestReviewer{{Name: "Lin", Vote: 10}}}, "APPROVED"),
		Entry("approved with suggestions", git.PullRequestReview{Reviewers: []git.PullRequestReviewer{{Name: "Lin", Vote: 5}}}, "APPROVED"),
		Entry("someone asked for changes, whoever approved",
			git.PullRequestReview{Reviewers: []git.PullRequestReviewer{{Name: "Lin", Vote: 10}, {Name: "Grace", Vote: -10}}}, "CHANGES_REQUESTED"),
		Entry("someone waits for the author", git.PullRequestReview{Reviewers: []git.PullRequestReviewer{{Name: "Grace", Vote: -5}}}, "CHANGES_REQUESTED"),
		Entry("a required reviewer hasn't approved",
			git.PullRequestReview{Reviewers: []git.PullRequestReviewer{{Name: "Lin", Vote: 10}, {Name: "Ada", Required: true}}}, "REVIEW_REQUIRED"),
		Entry("required approvals aren't met",
			git.PullRequestReview{Policies: []git.PullRequestPolicy{approvals("rejected")}, Reviewers: []git.PullRequestReviewer{{Name: "Lin", Vote: 10}}}, "REVIEW_REQUIRED"),
		Entry("required approvals are met",
			git.PullRequestReview{Policies: []git.PullRequestPolicy{approvals("approved")}, Reviewers: []git.PullRequestReviewer{{Name: "Lin", Vote: 10}}}, "APPROVED"),
		Entry("a build policy isn't a review",
			git.PullRequestReview{Policies: []git.PullRequestPolicy{{Name: "CI", Status: "rejected", Required: true, Build: true}}}, ""),
	)
})
