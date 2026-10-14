package github

import (
	"encoding/json"
	"fmt"
	"github.com/flanksource/gavel/github/activity"
)

type graphQLPageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

const reviewThreadFields = `id viewerCanResolve isResolved isOutdated
comments(first: 100) {
  pageInfo { hasNextPage endCursor }
  nodes { databaseId body author { login avatarUrl } path line url createdAt }
}`

const reviewThreadsQuery = `query($id: ID!, $after: String!) {
  node(id: $id) { ... on PullRequest {
    reviewThreads(first: 100, after: $after) {
      pageInfo { hasNextPage endCursor }
      nodes { ` + reviewThreadFields + ` }
    }
  } }
}`

const threadCommentsQuery = `query($id: ID!, $after: String!) {
  node(id: $id) { ... on PullRequestReviewThread {
    comments(first: 100, after: $after) {
      pageInfo { hasNextPage endCursor }
      nodes { databaseId body author { login avatarUrl } path line url createdAt }
    }
  } }
}`

func completeReviewThreads(token string, pr *graphQLPR) error {
	page := pr.ReviewThreads.PageInfo
	for page.HasNextPage {
		var result struct {
			Data struct {
				Node *struct {
					ReviewThreads graphQLReviewThreads `json:"reviewThreads"`
				} `json:"node"`
			} `json:"data"`
		}
		if err := fetchReviewPage(token, reviewThreadsQuery, pr.ID, page, &result); err != nil {
			return err
		}
		if result.Data.Node == nil {
			return fmt.Errorf("PR %s disappeared while paging review threads", pr.ID)
		}
		next := result.Data.Node.ReviewThreads
		if next.PageInfo.HasNextPage && next.PageInfo.EndCursor == page.EndCursor {
			return fmt.Errorf("review thread pagination did not advance")
		}
		pr.ReviewThreads.Nodes = append(pr.ReviewThreads.Nodes, next.Nodes...)
		page = next.PageInfo
	}
	for i := range pr.ReviewThreads.Nodes {
		if err := completeThreadComments(token, &pr.ReviewThreads.Nodes[i]); err != nil {
			return err
		}
	}
	return nil
}

func completeThreadComments(token string, thread *graphQLReviewThreadNode) error {
	page := thread.Comments.PageInfo
	for page.HasNextPage {
		var result struct {
			Data struct {
				Node *graphQLReviewThreadNode `json:"node"`
			} `json:"data"`
		}
		if err := fetchReviewPage(token, threadCommentsQuery, thread.ID, page, &result); err != nil {
			return err
		}
		if result.Data.Node == nil {
			return fmt.Errorf("thread %s disappeared while paging comments", thread.ID)
		}
		next := result.Data.Node.Comments
		if next.PageInfo.HasNextPage && next.PageInfo.EndCursor == page.EndCursor {
			return fmt.Errorf("review comment pagination did not advance")
		}
		thread.Comments.Nodes = append(thread.Comments.Nodes, next.Nodes...)
		page = next.PageInfo
	}
	return nil
}

func fetchReviewPage(token, query, id string, page graphQLPageInfo, result any) error {
	if id == "" || page.EndCursor == "" {
		return fmt.Errorf("review pagination requires a native node ID and cursor")
	}
	data, _, err := postGraphQL(token, graphqlEndpoint(), activity.KindGraphQL, query, map[string]any{"id": id, "after": page.EndCursor})
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

func ResolveReviewThread(opts Options, threadID string) error {
	if threadID == "" {
		return fmt.Errorf("native review thread ID is required")
	}
	token, err := opts.token()
	if err != nil {
		return err
	}
	data, _, err := postGraphQL(token, graphqlEndpoint(), activity.KindGraphQL,
		`mutation($input: ResolveReviewThreadInput!) { resolveReviewThread(input: $input) { thread { id isResolved } } }`, map[string]any{"input": map[string]any{"threadId": threadID}})
	if err != nil {
		return err
	}
	var result struct {
		Data struct {
			Resolution struct {
				Thread struct {
					ID       string `json:"id"`
					Resolved bool   `json:"isResolved"`
				} `json:"thread"`
			} `json:"resolveReviewThread"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}
	if result.Data.Resolution.Thread.ID != threadID || !result.Data.Resolution.Thread.Resolved {
		return fmt.Errorf("GitHub did not resolve review thread %s", threadID)
	}
	return nil
}
