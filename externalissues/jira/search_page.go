package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type searchPage struct {
	Issues *[]json.RawMessage `json:"issues"`
	Next   string             `json:"nextPageToken"`
	Last   *bool              `json:"isLast"`
}

func (c *Client) searchPage(ctx context.Context, options SearchOptions) (searchPage, error) {
	var page searchPage
	body := struct {
		JQL           string   `json:"jql"`
		Fields        []string `json:"fields"`
		MaxResults    int      `json:"maxResults,omitempty"`
		NextPageToken string   `json:"nextPageToken,omitempty"`
	}{options.JQL, options.Fields, options.Limit, options.After}
	if err := c.json(ctx, requestOptions{Method: "POST", Path: "/search/jql", Body: body}, &page); err != nil {
		return page, err
	}
	if page.Issues == nil {
		return page, fmt.Errorf("Jira search response has no issues array")
	}
	return page, nil
}

func (c *Client) syncKeys(ctx context.Context, options SyncOptions, selection *syncSelection) ([]string, error) {
	keys := []string{}
	for _, target := range options.Keys {
		query, err := syncTargetQuery(target)
		if err != nil {
			return nil, err
		}
		if query == "" {
			keys = append(keys, strings.ToUpper(strings.TrimSpace(target)))
			continue
		}
		if !selection.since.IsZero() {
			query += fmt.Sprintf(" AND updated >= %d", selection.since.UnixMilli())
		}
		query += " ORDER BY key ASC"
		seen := map[string]bool{}
		cursor := ""
		for {
			page, err := c.searchPage(ctx, SearchOptions{JQL: query, Fields: []string{"key"}, Limit: 100, After: cursor})
			if err != nil {
				return nil, fmt.Errorf("discover Jira target %s: %w", target, err)
			}
			for _, raw := range *page.Issues {
				var issue struct {
					Key string `json:"key"`
				}
				if err := json.Unmarshal(raw, &issue); err != nil {
					return nil, err
				}
				if !issueTarget.MatchString(issue.Key) {
					return nil, fmt.Errorf("Jira discovery returned invalid key %q", issue.Key)
				}
				keys = append(keys, issue.Key)
			}
			if page.Last != nil && *page.Last {
				if page.Next != "" {
					return nil, fmt.Errorf("Jira search final page has a continuation token")
				}
				break
			}
			if page.Next == "" || seen[page.Next] || len(*page.Issues) == 0 {
				return nil, fmt.Errorf("Jira search pagination stalled for %s", target)
			}
			seen[page.Next] = true
			cursor = page.Next
		}
	}
	return keys, nil
}
