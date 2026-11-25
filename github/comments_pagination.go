package github

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/flanksource/gavel/github/activity"
)

const commentsQuery = `query($id: ID!, $before: String!) {
  node(id: $id) { ... on PullRequest {
    comments(last: 100, before: $before) {
      pageInfo { hasPreviousPage startCursor }
      nodes { databaseId body author { login avatarUrl } url createdAt }
    }
  } }
}`

// completeComments prepends the comments older than the first page. Gavel's
// sticky comments are edited in place, so they are often a PR's oldest.
func completeComments(ctx context.Context, token string, pr *graphQLPR) error {
	page := pr.Comments.PageInfo
	for page.HasPreviousPage {
		if pr.ID == "" || page.StartCursor == "" {
			return fmt.Errorf("comment pagination requires a native node ID and cursor")
		}
		data, _, err := postGraphQL(ctx, token, graphqlEndpoint(), activity.KindGraphQL, commentsQuery, map[string]any{"id": pr.ID, "before": page.StartCursor})
		if err != nil {
			return err
		}
		var result struct {
			Data struct {
				Node *struct {
					Comments graphQLCommentList `json:"comments"`
				} `json:"node"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return fmt.Errorf("parse comments page: %w", err)
		}
		if result.Data.Node == nil {
			return fmt.Errorf("PR %s disappeared while paging comments", pr.ID)
		}
		next := result.Data.Node.Comments
		if next.PageInfo.HasPreviousPage && next.PageInfo.StartCursor == page.StartCursor {
			return fmt.Errorf("comment pagination did not advance")
		}
		pr.Comments.Nodes = slices.Concat(next.Nodes, pr.Comments.Nodes)
		page = next.PageInfo
	}
	return nil
}
