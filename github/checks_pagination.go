package github

import "fmt"

const checkContextsQuery = `query($id: ID!, $after: String!) {
  node(id: $id) { ... on Commit {
    statusCheckRollup {
      contexts(first: 100, after: $after) {
        pageInfo { hasNextPage endCursor }
        nodes {
          __typename
          ... on CheckRun { name status conclusion detailsUrl checkSuite { workflowRun { workflow { name } } } }
          ... on StatusContext { context state targetUrl }
        }
      }
    }
  } }
}`

func completeHeadChecks(token string, pr *graphQLPR) error {
	if len(pr.Commits.Nodes) == 0 {
		return nil
	}
	head := &pr.Commits.Nodes[len(pr.Commits.Nodes)-1].Commit
	if head.StatusCheckRollup == nil {
		return nil
	}
	page := head.StatusCheckRollup.Contexts.PageInfo
	for page.HasNextPage {
		var result struct {
			Data struct {
				Node *graphQLCommit `json:"node"`
			} `json:"data"`
		}
		if err := fetchReviewPage(token, checkContextsQuery, head.ID, page, &result); err != nil {
			return err
		}
		if result.Data.Node == nil || result.Data.Node.StatusCheckRollup == nil {
			return fmt.Errorf("PR head checks disappeared while paging")
		}
		next := result.Data.Node.StatusCheckRollup.Contexts
		if next.PageInfo.HasNextPage && next.PageInfo.EndCursor == page.EndCursor {
			return fmt.Errorf("PR check pagination did not advance")
		}
		head.StatusCheckRollup.Contexts.Nodes = append(head.StatusCheckRollup.Contexts.Nodes, next.Nodes...)
		page = next.PageInfo
	}
	return nil
}
