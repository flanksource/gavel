package github

import (
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/github/activity"
)

const mergePullRequestMutation = `mutation($prId: ID!, $method: PullRequestMergeMethod!, $head: GitObjectID) {
  mergePullRequest(input: {pullRequestId: $prId, mergeMethod: $method, expectedHeadOid: $head}) {
    pullRequest { number state }
  }
}`

type MergeOptions struct {
	NodeID, Method, HeadSHA string
}

// MergePullRequest merges the pull request identified by its GraphQL node ID
// immediately, using the given merge method (rebase|squash|merge). Unlike
// EnableAutoMerge this completes the merge now rather than waiting on required
// checks, so GitHub rejects it with a GraphQL error when the PR is not in a
// mergeable state (conflicts, failing required checks, missing approvals).
// postGraphQL surfaces those errors as Go errors, so any failure propagates
// rather than being swallowed.
func MergePullRequest(opts Options, input MergeOptions) error {
	if strings.TrimSpace(input.NodeID) == "" {
		return fmt.Errorf("MergePullRequest: PR node ID is required")
	}
	method, err := MergeMethodFor(input.Method)
	if err != nil {
		return err
	}
	token, err := opts.token()
	if err != nil {
		return err
	}

	variables := map[string]any{
		"prId":   input.NodeID,
		"method": method,
	}
	if input.HeadSHA != "" {
		variables["head"] = input.HeadSHA
	}
	_, _, err = postGraphQL(token, graphqlEndpoint(), activity.KindGraphQL, mergePullRequestMutation, variables)
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

const updatePullRequestBranchMutation = `mutation($prId: ID!, $head: GitObjectID) {
  updatePullRequestBranch(input: {pullRequestId: $prId, expectedHeadOid: $head, updateMethod: REBASE}) {
    pullRequest { number headRefOid }
  }
}`

// ErrUpdateBranchConflict marks an update-branch GitHub refused because the PR's
// commits do not rebase cleanly onto its base.
var ErrUpdateBranchConflict = errors.New("the branch does not rebase cleanly onto its base")

type UpdateBranchOptions struct {
	NodeID string
	// HeadSHA is the head the caller saw; GitHub refuses the update when the
	// branch has moved since, rather than rebasing commits nobody reviewed.
	HeadSHA string
}

// UpdatePullRequestBranch brings the PR's head branch up to date by rebasing it
// onto its base — "Update branch → Update with rebase" on github.com — so the
// branch never gains a merge commit. A refusal over conflicts wraps
// ErrUpdateBranchConflict, so callers can offer to resolve the rebase instead of
// only reporting it.
func UpdatePullRequestBranch(opts Options, input UpdateBranchOptions) error {
	if strings.TrimSpace(input.NodeID) == "" {
		return fmt.Errorf("UpdatePullRequestBranch: PR node ID is required")
	}
	token, err := opts.token()
	if err != nil {
		return err
	}

	variables := map[string]any{"prId": input.NodeID}
	if input.HeadSHA != "" {
		variables["head"] = input.HeadSHA
	}
	_, _, err = postGraphQL(token, graphqlEndpoint(), activity.KindGraphQL, updatePullRequestBranchMutation, variables)
	if err == nil {
		return nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "conflict") {
		return fmt.Errorf("update branch (rebase): %w: %w", ErrUpdateBranchConflict, err)
	}
	return fmt.Errorf("update branch (rebase): %w", err)
}

const closePullRequestMutation = `mutation($prId: ID!) {
  closePullRequest(input: {pullRequestId: $prId}) {
    pullRequest { number state }
  }
}`

// ClosePullRequest closes the pull request identified by its GraphQL node ID
// without merging it — the equivalent of gh pr close. GitHub rejects closing a
// PR that is already merged or closed with a GraphQL error, which postGraphQL
// surfaces as a Go error rather than swallowing it. Closing is reversible: the
// PR can be reopened on github.com.
func ClosePullRequest(opts Options, prNodeID string) error {
	if strings.TrimSpace(prNodeID) == "" {
		return fmt.Errorf("ClosePullRequest: PR node ID is required")
	}
	token, err := opts.token()
	if err != nil {
		return err
	}

	_, _, err = postGraphQL(token, graphqlEndpoint(), activity.KindGraphQL, closePullRequestMutation, map[string]any{
		"prId": prNodeID,
	})
	if err != nil {
		return fmt.Errorf("close pull request: %w", err)
	}
	return nil
}

const addCommentMutation = `mutation($subjectId: ID!, $body: String!) {
  addComment(input: {subjectId: $subjectId, body: $body}) {
    commentEdge { node { url } }
  }
}`

// CommentOnPullRequest posts an issue comment on the pull request identified by
// its GraphQL node ID. GitHub's addComment takes any commentable subject, so
// the PR node ID is the subject here. An empty body is rejected locally rather
// than posting a blank comment.
func CommentOnPullRequest(opts Options, prNodeID, body string) error {
	if strings.TrimSpace(prNodeID) == "" {
		return fmt.Errorf("CommentOnPullRequest: PR node ID is required")
	}
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("CommentOnPullRequest: comment body is required")
	}
	token, err := opts.token()
	if err != nil {
		return err
	}

	_, _, err = postGraphQL(token, graphqlEndpoint(), activity.KindGraphQL, addCommentMutation, map[string]any{
		"subjectId": prNodeID,
		"body":      body,
	})
	if err != nil {
		return fmt.Errorf("comment on pull request: %w", err)
	}
	return nil
}
