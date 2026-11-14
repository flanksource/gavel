package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/pr/blame"
	"github.com/flanksource/gavel/pr/model"
)

func (c *Client) RecommendReviewers(input blame.Request) (*blame.Recommendations, error) {
	if input.Limit <= 0 || input.HistoryLimit <= 0 {
		return nil, fmt.Errorf("reviewer limit and history limit must be positive")
	}
	if input.Number < 0 {
		return nil, fmt.Errorf("Azure PR number must be positive")
	}
	ctx := input.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if err := c.Preflight(ctx); err != nil {
		return nil, err
	}
	remote, err := c.blameRemote(ctx, input.Dir)
	if err != nil {
		return nil, err
	}
	pr, err := c.blamePR(ctx, input)
	if err != nil {
		return nil, err
	}
	info, err := c.prInfo(ctx, *pr)
	if err != nil {
		return nil, err
	}
	if info.BaseRefOID == "" || info.HeadRefOID == "" || pr.CreatedBy.ID == "" || pr.CreatedBy.Login == "" {
		return nil, fmt.Errorf("Azure PR #%d requires base and head commits and an author identity", pr.ID)
	}
	if pr.Status == "active" {
		base, head := info.BaseRefOID, info.HeadRefOID
		info.MergeReadiness = &model.MergeReadiness{}
		if err := c.refreshHead(ctx, pr, info); err != nil {
			return nil, err
		}
		if info.HeadRefOID != head || info.BaseRefOID != base {
			return nil, fmt.Errorf("Azure PR #%d advanced beyond its merge snapshot; wait for merge and policy evaluation to catch up", pr.ID)
		}
	}
	if err := c.blameCommits(ctx, blameCheckout{Dir: input.Dir, Remote: remote, PR: *pr}); err != nil {
		return nil, err
	}
	policies, err := c.reviewerPolicies(ctx, pr.ID)
	if err != nil {
		return nil, err
	}
	result, err := blame.Analyze(ctx, blame.Options{Dir: input.Dir, Base: info.BaseRefOID, Head: info.HeadRefOID,
		HistoryLimit: input.HistoryLimit, Exclude: append(append([]string{}, input.Exclude...), pr.CreatedBy.Login, pr.CreatedBy.ID), PolicyReviewers: policies, Since: input.Since})
	if err != nil {
		return nil, err
	}
	total := len(result.Reviewers)
	if total > input.Limit {
		result.Reviewers = result.Reviewers[:input.Limit]
	}
	return &blame.Recommendations{Result: result, Provider: c.Kind(), Number: pr.ID, URL: info.URL, Title: pr.Title, Author: pr.CreatedBy.Login, TotalReviewers: total}, nil
}

func (c *Client) blamePR(ctx context.Context, input blame.Request) (*pullRequest, error) {
	number := input.Number
	if number == 0 {
		branch, err := blameGit(ctx, input.Dir, "symbolic-ref", "--quiet", "HEAD")
		if err != nil {
			return nil, fmt.Errorf("specify a PR number when HEAD is detached: %w", err)
		}
		prs, err := collection[pullRequest](ctx, c, c.repositoryPath()+"/pullrequests", url.Values{"searchCriteria.status": {"active"}, "searchCriteria.sourceRefName": {branch}})
		if err != nil {
			return nil, err
		}
		if len(prs) != 1 || prs[0].ID <= 0 {
			return nil, fmt.Errorf("expected one active Azure PR for %s, found %d; specify a PR number", branch, len(prs))
		}
		number = prs[0].ID
	}
	if number <= 0 {
		return nil, fmt.Errorf("Azure PR number must be positive")
	}
	data, _, err := c.request(ctx, http.MethodGet, c.repositoryPath()+"/pullrequests/"+strconv.Itoa(number), nil, nil)
	if err != nil {
		return nil, err
	}
	var pr pullRequest
	if err := json.Unmarshal(data, &pr); err != nil {
		return nil, fmt.Errorf("decode Azure PR: %w", err)
	}
	if pr.ID != number || !strings.HasPrefix(pr.Source, "refs/heads/") || !strings.HasPrefix(pr.Target, "refs/heads/") {
		return nil, fmt.Errorf("Azure PR response has invalid identity or branch refs for #%d", number)
	}
	return &pr, nil
}
