package github

import (
	"fmt"
	"strings"

	"github.com/flanksource/gavel/github/activity"
)

const enableAutoMergeMutation = `mutation($prId: ID!, $method: PullRequestMergeMethod!) {
  enablePullRequestAutoMerge(input: {pullRequestId: $prId, mergeMethod: $method}) {
    pullRequest { number }
  }
}`

// MergeMethodFor maps a user-facing merge-type (rebase|squash|merge) to GitHub's
// PullRequestMergeMethod enum value. It is case-insensitive and trims surrounding
// whitespace; any other value is an error naming the offending input.
func MergeMethodFor(mergeType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(mergeType)) {
	case "rebase":
		return "REBASE", nil
	case "squash":
		return "SQUASH", nil
	case "merge":
		return "MERGE", nil
	default:
		return "", fmt.Errorf("invalid merge type %q: want rebase, squash, or merge", mergeType)
	}
}

// EnableAutoMerge turns on GitHub auto-merge for the pull request identified by
// its GraphQL node ID, using the given merge method. GitHub auto-merge cannot be
// set via the REST create-PR call, so this runs the enablePullRequestAutoMerge
// GraphQL mutation. postGraphQL surfaces HTTP-level and GraphQL errors[] (e.g.
// "auto merge is not allowed for this repository") as Go errors, so any failure
// propagates rather than being swallowed.
func EnableAutoMerge(opts Options, prNodeID, mergeType string) error {
	if strings.TrimSpace(prNodeID) == "" {
		return fmt.Errorf("EnableAutoMerge: PR node ID is required")
	}
	method, err := MergeMethodFor(mergeType)
	if err != nil {
		return err
	}
	token, err := opts.token()
	if err != nil {
		return err
	}

	_, _, err = postGraphQL(token, graphqlEndpoint(), activity.KindGraphQL, enableAutoMergeMutation, map[string]any{
		"prId":   prNodeID,
		"method": method,
	})
	if err != nil {
		return fmt.Errorf("enable auto-merge (%s): %w", method, err)
	}
	return nil
}
