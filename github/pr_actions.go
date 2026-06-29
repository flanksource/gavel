package github

import (
	"fmt"
	"strings"

	"github.com/flanksource/gavel/github/activity"
)

const mergePullRequestMutation = `mutation($prId: ID!, $method: PullRequestMergeMethod!) {
  mergePullRequest(input: {pullRequestId: $prId, mergeMethod: $method}) {
    pullRequest { number state }
  }
}`

// MergePullRequest merges the pull request identified by its GraphQL node ID
// immediately, using the given merge method (rebase|squash|merge). Unlike
// EnableAutoMerge this completes the merge now rather than waiting on required
// checks, so GitHub rejects it with a GraphQL error when the PR is not in a
// mergeable state (conflicts, failing required checks, missing approvals).
// postGraphQL surfaces those errors as Go errors, so any failure propagates
// rather than being swallowed.
func MergePullRequest(opts Options, prNodeID, mergeType string) error {
	if strings.TrimSpace(prNodeID) == "" {
		return fmt.Errorf("MergePullRequest: PR node ID is required")
	}
	method, err := MergeMethodFor(mergeType)
	if err != nil {
		return err
	}
	token, err := opts.token()
	if err != nil {
		return err
	}

	_, _, err = postGraphQL(token, graphqlEndpoint(), activity.KindGraphQL, mergePullRequestMutation, map[string]any{
		"prId":   prNodeID,
		"method": method,
	})
	if err != nil {
		return fmt.Errorf("merge pull request (%s): %w", method, err)
	}
	return nil
}

const approvePullRequestMutation = `mutation($prId: ID!, $body: String) {
  addPullRequestReview(input: {pullRequestId: $prId, event: APPROVE, body: $body}) {
    pullRequestReview { state }
  }
}`

// ApprovePullRequest submits an APPROVE review on the pull request identified by
// its GraphQL node ID. body is an optional review comment (empty for a bare
// approval). GitHub rejects self-approval and approvals without write access
// with a GraphQL error, which postGraphQL surfaces as a Go error.
func ApprovePullRequest(opts Options, prNodeID, body string) error {
	if strings.TrimSpace(prNodeID) == "" {
		return fmt.Errorf("ApprovePullRequest: PR node ID is required")
	}
	token, err := opts.token()
	if err != nil {
		return err
	}

	_, _, err = postGraphQL(token, graphqlEndpoint(), activity.KindGraphQL, approvePullRequestMutation, map[string]any{
		"prId": prNodeID,
		"body": body,
	})
	if err != nil {
		return fmt.Errorf("approve pull request: %w", err)
	}
	return nil
}
