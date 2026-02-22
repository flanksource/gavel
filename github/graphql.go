package github

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/flanksource/commons/logger"
)

const prByNumberQuery = `query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $number) {
      ...prFields
    }
  }
}
` + prFragment

const prByBranchQuery = `query($owner: String!, $repo: String!, $branch: String!) {
  repository(owner: $owner, name: $repo) {
    pullRequests(headRefName: $branch, states: OPEN, first: 1) {
      nodes {
        ...prFields
      }
    }
  }
}
` + prFragment

const prFragment = `fragment prFields on PullRequest {
  number
  title
  author { login }
  headRefName
  baseRefName
  state
  isDraft
  reviewDecision
  mergeable
  url
  commits(last: 1) {
    nodes {
      commit {
        statusCheckRollup {
          contexts(first: 100) {
            nodes {
              __typename
              ... on CheckRun {
                name
                status
                conclusion
                detailsUrl
                checkSuite {
                  workflowRun {
                    workflow {
                      name
                    }
                  }
                }
              }
              ... on StatusContext {
                context
                state
                targetUrl
              }
            }
          }
        }
      }
    }
  }
}`

// GraphQL response types

type graphQLResponse struct {
	Data   graphQLData    `json:"data"`
	Errors []graphQLError `json:"errors"`
}

type graphQLError struct {
	Message string `json:"message"`
}

type graphQLData struct {
	Repository graphQLRepository `json:"repository"`
}

type graphQLRepository struct {
	PullRequest  *graphQLPR           `json:"pullRequest"`
	PullRequests *graphQLPRConnection `json:"pullRequests"`
}

type graphQLPRConnection struct {
	Nodes []graphQLPR `json:"nodes"`
}

type graphQLPR struct {
	Number         int            `json:"number"`
	Title          string         `json:"title"`
	Author         graphQLAuthor  `json:"author"`
	HeadRefName    string         `json:"headRefName"`
	BaseRefName    string         `json:"baseRefName"`
	State          string         `json:"state"`
	IsDraft        bool           `json:"isDraft"`
	ReviewDecision string         `json:"reviewDecision"`
	Mergeable      string         `json:"mergeable"`
	URL            string         `json:"url"`
	Commits        graphQLCommits `json:"commits"`
}

type graphQLAuthor struct {
	Login string `json:"login"`
}

type graphQLCommits struct {
	Nodes []graphQLCommitNode `json:"nodes"`
}

type graphQLCommitNode struct {
	Commit graphQLCommit `json:"commit"`
}

type graphQLCommit struct {
	StatusCheckRollup *graphQLStatusCheckRollup `json:"statusCheckRollup"`
}

type graphQLStatusCheckRollup struct {
	Contexts graphQLContexts `json:"contexts"`
}

type graphQLContexts struct {
	Nodes []graphQLCheckNode `json:"nodes"`
}

type graphQLCheckNode struct {
	Typename   string             `json:"__typename"`
	Name       string             `json:"name"`
	Status     string             `json:"status"`
	Conclusion *string            `json:"conclusion"`
	DetailsURL string             `json:"detailsUrl"`
	CheckSuite *graphQLCheckSuite `json:"checkSuite"`
	Context    string             `json:"context"`
	State      string             `json:"state"`
	TargetURL  string             `json:"targetUrl"`
}

type graphQLCheckSuite struct {
	WorkflowRun *graphQLWorkflowRun `json:"workflowRun"`
}

type graphQLWorkflowRun struct {
	Workflow graphQLWorkflow `json:"workflow"`
}

type graphQLWorkflow struct {
	Name string `json:"name"`
}

func (pr graphQLPR) toPRInfo() *PRInfo {
	info := &PRInfo{
		Number:         pr.Number,
		Title:          pr.Title,
		Author:         PRAuthor{Login: pr.Author.Login},
		HeadRefName:    pr.HeadRefName,
		BaseRefName:    pr.BaseRefName,
		State:          pr.State,
		IsDraft:        pr.IsDraft,
		ReviewDecision: pr.ReviewDecision,
		Mergeable:      pr.Mergeable,
		URL:            pr.URL,
	}

	if len(pr.Commits.Nodes) > 0 {
		rollup := pr.Commits.Nodes[0].Commit.StatusCheckRollup
		if rollup != nil {
			for _, node := range rollup.Contexts.Nodes {
				info.StatusCheckRollup = append(info.StatusCheckRollup, node.toStatusCheck())
			}
		}
	}
	return info
}

func (n graphQLCheckNode) toStatusCheck() StatusCheck {
	if n.Typename == "CheckRun" {
		return n.checkRunToStatusCheck()
	}
	return n.statusContextToStatusCheck()
}

func (n graphQLCheckNode) checkRunToStatusCheck() StatusCheck {
	sc := StatusCheck{
		Name:       n.Name,
		Status:     strings.ToUpper(n.Status),
		DetailsURL: n.DetailsURL,
	}
	if n.Conclusion != nil {
		sc.Conclusion = strings.ToUpper(*n.Conclusion)
	}
	if n.CheckSuite != nil && n.CheckSuite.WorkflowRun != nil {
		sc.WorkflowName = n.CheckSuite.WorkflowRun.Workflow.Name
	}
	return sc
}

func (n graphQLCheckNode) statusContextToStatusCheck() StatusCheck {
	sc := StatusCheck{
		Name:       n.Context,
		DetailsURL: n.TargetURL,
	}
	switch strings.ToUpper(n.State) {
	case "SUCCESS":
		sc.Status = "COMPLETED"
		sc.Conclusion = "SUCCESS"
	case "FAILURE", "ERROR":
		sc.Status = "COMPLETED"
		sc.Conclusion = "FAILURE"
	case "PENDING":
		sc.Status = "PENDING"
	case "EXPECTED":
		sc.Status = "QUEUED"
	default:
		sc.Status = strings.ToUpper(n.State)
	}
	return sc
}

const reviewThreadsQuery = `query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $number) {
      reviewThreads(first: 50) {
        nodes {
          isResolved
          isOutdated
          comments(first: 1) {
            nodes {
              databaseId
              body
              author { login }
              path
              line
              url
            }
          }
        }
      }
    }
  }
}`

type reviewThreadsResponse struct {
	Data struct {
		Repository struct {
			PullRequest struct {
				ReviewThreads struct {
					Nodes []struct {
						IsResolved bool `json:"isResolved"`
						IsOutdated bool `json:"isOutdated"`
						Comments   struct {
							Nodes []struct {
								DatabaseID int64         `json:"databaseId"`
								Body       string        `json:"body"`
								Author     graphQLAuthor `json:"author"`
								Path       string        `json:"path"`
								Line       int           `json:"line"`
								URL        string        `json:"url"`
							} `json:"nodes"`
						} `json:"comments"`
					} `json:"nodes"`
				} `json:"reviewThreads"`
			} `json:"pullRequest"`
		} `json:"repository"`
	} `json:"data"`
	Errors []graphQLError `json:"errors"`
}

func FetchReviewThreads(opts Options, prNumber int) ([]PRComment, error) {
	token, err := opts.token()
	if err != nil {
		return nil, err
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return nil, err
	}
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo format %q", repo)
	}

	body := map[string]any{
		"query": reviewThreadsQuery,
		"variables": map[string]any{
			"owner": parts[0], "repo": parts[1], "number": prNumber,
		},
	}

	ctx := context.Background()
	client := newClient(token)

	logger.Tracef("fetching review threads via GraphQL for PR #%d", prNumber)
	resp, err := client.R(ctx).
		Header("Content-Type", "application/json").
		Post("https://api.github.com/graphql", body)
	if err != nil {
		return nil, fmt.Errorf("GraphQL request: %w", err)
	}
	if !resp.IsOK() {
		respBody, _ := resp.AsString()
		return nil, fmt.Errorf("GraphQL request: status %d: %s", resp.StatusCode, respBody)
	}

	var result reviewThreadsResponse
	if err := resp.Into(&result); err != nil {
		return nil, fmt.Errorf("parse review threads response: %w", err)
	}
	if len(result.Errors) > 0 {
		msgs := make([]string, len(result.Errors))
		for i, e := range result.Errors {
			msgs[i] = e.Message
		}
		return nil, fmt.Errorf("GraphQL errors: %s", strings.Join(msgs, "; "))
	}

	var comments []PRComment
	for _, thread := range result.Data.Repository.PullRequest.ReviewThreads.Nodes {
		if len(thread.Comments.Nodes) == 0 {
			continue
		}
		c := thread.Comments.Nodes[0]
		comments = append(comments, PRComment{
			ID:         c.DatabaseID,
			Body:       c.Body,
			Author:     c.Author.Login,
			URL:        c.URL,
			Path:       c.Path,
			Line:       c.Line,
			IsResolved: thread.IsResolved,
			IsOutdated: thread.IsOutdated,
		})
	}

	logger.Debugf("fetched %d review threads for PR #%d", len(comments), prNumber)
	return comments, nil
}

func FetchPR(opts Options, prNumber int) (*PRInfo, error) {
	token, err := opts.token()
	if err != nil {
		return nil, err
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo format %q, expected owner/repo", repo)
	}
	owner, name := parts[0], parts[1]

	var query string
	variables := map[string]any{"owner": owner, "repo": name}

	if prNumber > 0 {
		query = prByNumberQuery
		variables["number"] = prNumber
	} else {
		branch, err := opts.currentBranch()
		if err != nil {
			return nil, fmt.Errorf("no PR number and cannot determine branch: %w", err)
		}
		query = prByBranchQuery
		variables["branch"] = branch
	}

	body := map[string]any{"query": query, "variables": variables}

	ctx := context.Background()
	client := newClient(token)

	logger.Tracef("fetching PR via GraphQL (pr=%s, repo=%s)", formatPRArg(prNumber), repo)
	resp, err := client.R(ctx).
		Header("Content-Type", "application/json").
		Post("https://api.github.com/graphql", body)
	if err != nil {
		return nil, fmt.Errorf("GraphQL request: %w", err)
	}
	if !resp.IsOK() {
		respBody, _ := resp.AsString()
		return nil, fmt.Errorf("GraphQL request: status %d: %s", resp.StatusCode, respBody)
	}

	var result graphQLResponse
	if err := resp.Into(&result); err != nil {
		return nil, fmt.Errorf("parse GraphQL response: %w", err)
	}
	if len(result.Errors) > 0 {
		msgs := make([]string, len(result.Errors))
		for i, e := range result.Errors {
			msgs[i] = e.Message
		}
		return nil, fmt.Errorf("GraphQL errors: %s", strings.Join(msgs, "; "))
	}

	var gqlPR *graphQLPR
	if prNumber > 0 {
		gqlPR = result.Data.Repository.PullRequest
	} else if result.Data.Repository.PullRequests != nil && len(result.Data.Repository.PullRequests.Nodes) > 0 {
		gqlPR = &result.Data.Repository.PullRequests.Nodes[0]
	}
	if gqlPR == nil {
		return nil, fmt.Errorf("no PR found for %s in %s", formatPRArg(prNumber), repo)
	}

	pr := gqlPR.toPRInfo()
	logger.Debugf("fetched PR #%d %q (%s, %d checks)", pr.Number, pr.Title, pr.State, len(pr.StatusCheckRollup))
	return pr, nil
}

func formatPRArg(prNumber int) string {
	if prNumber > 0 {
		return "#" + strconv.Itoa(prNumber)
	}
	return "current branch"
}
