// A connection pull request's review decision as GitHub spells it, from its provider's reviewer
// votes (Azure DevOps' scale, which GitLab's review is read on too) and required approvals.
package provider

import "github.com/flanksource/commons-db/git"

// reviewDecision is CHANGES_REQUESTED when a reviewer asked for changes or waits for the author (a
// vote of -5 or lower); else REVIEW_REQUIRED while a required reviewer hasn't approved or required
// approvals aren't met; else APPROVED once someone approved; else "".
func reviewDecision(review git.PullRequestReview) string {
	approved, required := false, false
	for _, reviewer := range review.Reviewers {
		switch {
		case reviewer.Vote <= -5:
			return "CHANGES_REQUESTED"
		case reviewer.Vote >= 5:
			approved = true
		case reviewer.Required:
			required = true
		}
	}
	for _, policy := range review.Policies {
		if policy.Required && !policy.Build && policy.MinimumApproverCount > 0 && policy.Status != "approved" {
			required = true
		}
	}
	switch {
	case required:
		return "REVIEW_REQUIRED"
	case approved:
		return "APPROVED"
	}
	return ""
}
