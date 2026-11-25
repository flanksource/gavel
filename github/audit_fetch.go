package github

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/flanksource/gavel/pr/model"
	"strings"
	"time"
)

func (o Options) auditQuery(query string, variables map[string]any, target any) error {
	token, err := o.token()
	if err != nil {
		return err
	}
	ctx := o.Context
	if ctx == nil {
		ctx = context.Background()
	}
	response, err := newClient(token).R(ctx).Header("Content-Type", "application/json").Post(graphqlEndpoint(), map[string]any{"query": query, "variables": variables})
	if err != nil {
		return err
	}
	body, err := response.AsString()
	if err != nil {
		return err
	}
	if !response.IsOK() {
		return fmt.Errorf("GitHub audit HTTP %d: %s", response.StatusCode, body)
	}
	if err := graphQLErrors([]byte(body)); err != nil {
		return err
	}
	return json.Unmarshal([]byte(body), target)
}

type auditPR struct {
	ID, Title, URL, Mergeable string
	Number                    int
	UpdatedAt                 time.Time
	Commits                   struct {
		Nodes []struct{ Commit struct{ ID string } }
	}
}

func fetchAuditPRs(opts Options) ([]auditPR, error) {
	parts := strings.Split(opts.Repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid GitHub repository %q", opts.Repo)
	}
	var all []auditPR
	var after any
	seen := map[string]bool{}
	for {
		var response struct {
			Data struct {
				Repository *struct {
					PullRequests struct {
						Nodes    []auditPR
						PageInfo graphQLPageInfo
					}
				}
			}
		}
		err := opts.auditQuery(`query($owner:String!,$repo:String!,$after:String){repository(owner:$owner,name:$repo){pullRequests(states:OPEN,first:100,after:$after){pageInfo{hasNextPage endCursor} nodes{id number title url mergeable updatedAt commits(last:1){nodes{commit{id}}}}}}}`, map[string]any{"owner": parts[0], "repo": parts[1], "after": after}, &response)
		if err != nil {
			return all, err
		}
		if response.Data.Repository == nil {
			return all, fmt.Errorf("repository %s unavailable", opts.Repo)
		}
		page := response.Data.Repository.PullRequests
		all = append(all, page.Nodes...)
		if !page.PageInfo.HasNextPage {
			return all, nil
		}
		if page.PageInfo.EndCursor == "" || seen[page.PageInfo.EndCursor] {
			return all, fmt.Errorf("PR pagination did not advance")
		}
		seen[page.PageInfo.EndCursor] = true
		after = page.PageInfo.EndCursor
	}
}

func fetchAuditChecks(opts Options, pr auditPR) (model.StatusChecks, error) {
	if len(pr.Commits.Nodes) != 1 {
		return nil, fmt.Errorf("PR #%d has no head commit", pr.Number)
	}
	var all model.StatusChecks
	var after any
	seen := map[string]bool{}
	for {
		var response struct {
			Data struct {
				Node *struct {
					StatusCheckRollup *struct{ Contexts graphQLContexts }
				}
			}
		}
		err := opts.auditQuery(`query($id:ID!,$pr:ID!,$after:String){node(id:$id){... on Commit{statusCheckRollup{contexts(first:100,after:$after){pageInfo{hasNextPage endCursor} nodes{__typename ... on CheckRun{id name status conclusion detailsUrl isRequired(pullRequestId:$pr) checkSuite{app{slug}}} ... on StatusContext{context state targetUrl isRequired(pullRequestId:$pr)}}}}}}}`, map[string]any{"id": pr.Commits.Nodes[0].Commit.ID, "pr": pr.ID, "after": after}, &response)
		if err != nil {
			return nil, err
		}
		if response.Data.Node == nil {
			return nil, fmt.Errorf("PR #%d head unavailable", pr.Number)
		}
		if response.Data.Node.StatusCheckRollup == nil {
			return all, nil
		}
		page := response.Data.Node.StatusCheckRollup.Contexts
		for _, node := range page.Nodes {
			check := node.toStatusCheck()
			check.Required = node.Required
			all = append(all, check)
		}
		if !page.PageInfo.HasNextPage {
			return all, nil
		}
		if page.PageInfo.EndCursor == "" || seen[page.PageInfo.EndCursor] {
			return nil, fmt.Errorf("check pagination did not advance")
		}
		seen[page.PageInfo.EndCursor] = true
		after = page.PageInfo.EndCursor
	}
}
