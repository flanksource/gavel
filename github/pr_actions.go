package github

import (
	"context"
	"fmt"
	"io"
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

// UpdatePullRequestBranch updates the PR's head branch by merging or rebasing
// onto the base branch using GitHub's REST API
// (PUT /repos/{owner}/{repo}/pulls/{number}/update-branch). This is the
// equivalent of the "Update branch" button on github.com. GitHub returns 202
// on success, 422 when there are conflicts, and 403 for permission issues.
func UpdatePullRequestBranch(opts Options, prNumber int) error {
	token, err := opts.token()
	if err != nil {
		return err
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/repos/%s/pulls/%d/update-branch", apiBaseURL, repo, prNumber)
	client := newClient(token).
		Header("Content-Type", "application/json").
		// The update-branch endpoint requires the lydian-mass preview header.
		Header("Accept", "application/vnd.github.lydian-mass-preview+json")

	resp, err := client.R(context.Background()).Put(url, map[string]any{})
	if err != nil {
		return fmt.Errorf("update branch for PR #%d on %s: %w", prNumber, repo, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read update-branch response: %w", err)
	}

	activity.Shared().Record(activity.Entry{
		Method:     "PUT",
		URL:        fmt.Sprintf("/repos/%s/pulls/%d/update-branch", repo, prNumber),
		Kind:       activity.KindREST,
		StatusCode: resp.StatusCode,
		SizeBytes:  len(body),
	})

	// 202 Accepted means the branch update is queued/completed.
	if resp.StatusCode == 202 {
		return nil
	}
	return fmt.Errorf("update branch for PR #%d on %s: HTTP %d: %s", prNumber, repo, resp.StatusCode, string(body))
}
