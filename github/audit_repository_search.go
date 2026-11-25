package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

func (o Options) searchAuditRepositories(ctx context.Context, query string) ([]string, error) {
	token, err := o.token()
	if err != nil {
		return nil, err
	}
	path := "/search/repositories?" + url.Values{"q": {query}, "per_page": {"100"}}.Encode()
	seen := map[string]bool{}
	var repos []string
	for path != "" {
		if seen[path] {
			return repos, fmt.Errorf("repository search pagination did not advance")
		}
		seen[path] = true
		response, err := newClient(token).R(ctx).Get(githubAPIBase() + path)
		if err != nil {
			return repos, err
		}
		body, err := response.AsString()
		if err != nil {
			return repos, err
		}
		if !response.IsOK() {
			return repos, fmt.Errorf("repository search HTTP %d: %s", response.StatusCode, body)
		}
		var page struct {
			TotalCount *int  `json:"total_count"`
			Incomplete *bool `json:"incomplete_results"`
			Items      []struct {
				FullName string `json:"full_name"`
			} `json:"items"`
		}
		if err := json.Unmarshal([]byte(body), &page); err != nil {
			return repos, err
		}
		if page.TotalCount == nil || page.Incomplete == nil || page.Items == nil || *page.TotalCount < 0 {
			return repos, fmt.Errorf("invalid repository search response")
		}
		for _, item := range page.Items {
			if item.FullName == "" {
				return repos, fmt.Errorf("repository search returned an item without full_name")
			}
			repos = append(repos, item.FullName)
		}
		if *page.Incomplete {
			return repos, fmt.Errorf("repository search results are incomplete; narrow the query")
		}
		if *page.TotalCount > 1000 {
			return repos, fmt.Errorf("repository search exceeds GitHub's 1,000-result limit; narrow the query")
		}
		path, err = auditNextPage(response.Header.Get("Link"))
		if err != nil {
			return repos, err
		}
		if path == "" && len(repos) < *page.TotalCount {
			return repos, fmt.Errorf("repository search pagination ended before all results were returned")
		}
	}
	return repos, nil
}
