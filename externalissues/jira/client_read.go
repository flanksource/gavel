package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/flanksource/clicky/task"
)

type GetOptions struct {
	Key     string
	Full    bool
	NoCache bool
	Store   CatalogStore
}

type SearchOptions struct {
	JQL     string
	Limit   int
	After   string
	Fields  []string
	NoCache bool
	Store   CatalogStore
}

func (c *Client) Get(ctx context.Context, options GetOptions) (Issue, error) {
	return runJiraTask(ctx, "Get Jira issue "+options.Key, func(ctx context.Context, progress *task.Task) (Issue, error) {
		if strings.TrimSpace(options.Key) == "" {
			return Issue{}, fmt.Errorf("Jira issue key is required")
		}
		progress.SetProgress(0, 4)
		progress.SetDescription("Fetching issue")
		var raw json.RawMessage
		if err := c.json(ctx, requestOptions{Path: "/issue/" + url.PathEscape(options.Key), Query: url.Values{"fields": {"*all"}, "updateHistory": {"false"}}}, &raw); err != nil {
			return Issue{}, err
		}
		progress.SetProgress(1, 4)
		progress.SetDescription("Loading field catalog")
		fields, err := c.Catalog(ctx, CatalogOptions{NoCache: options.NoCache, Store: options.Store})
		if err != nil {
			return Issue{}, err
		}
		issue, err := c.parseIssue(raw, fields)
		if err != nil {
			return Issue{}, err
		}
		progress.SetProgress(2, 4)
		filter := ParseOptions{Catalog: fields, IgnoreFields: c.ignoreFields}
		if !filter.ignores("comment") {
			progress.SetDescription("Loading comments")
			comments, err := c.Comments(ctx, issue.Key)
			if err != nil {
				return Issue{}, err
			}
			issue.Comments = comments
			count := len(comments)
			issue.CommentCount = &count
		}
		progress.SetProgress(3, 4)
		if !filter.ignores("remoteLinks") {
			progress.SetDescription("Loading remote links")
			if err := c.loadRemoteLinks(ctx, &issue); err != nil {
				return Issue{}, err
			}
		}
		issue.Full = options.Full
		progress.SetProgress(4, 4)
		progress.SetDescription("Loaded " + issue.Key)
		if issue.CommentCount != nil {
			progress.SetDescription(fmt.Sprintf("Loaded %s with %d comments", issue.Key, *issue.CommentCount))
		}
		return issue, nil
	})
}

func (c *Client) loadRemoteLinks(ctx context.Context, issue *Issue) error {
	var links []json.RawMessage
	if err := c.json(ctx, requestOptions{Path: "/issue/" + url.PathEscape(issue.Key) + "/remotelink"}, &links); err != nil {
		return err
	}
	if links == nil {
		return fmt.Errorf("Jira remote links response requires an array")
	}
	value, err := json.Marshal(links)
	if err != nil {
		return err
	}
	mapped, err := MapFields(map[string]json.RawMessage{"remoteLinks": value}, map[string]FieldDefinition{"remoteLinks": {ID: "remoteLinks", Name: remoteLinksName}})
	if err != nil {
		return err
	}
	issue.Fields["remoteLinks"] = mapped["remoteLinks"]
	return nil
}

func (c *Client) List(ctx context.Context, options SearchOptions) ([]Issue, error) {
	return runJiraTask(ctx, "Search Jira issues", func(ctx context.Context, progress *task.Task) ([]Issue, error) {
		progress.SetDescription("Searching Jira")
		if strings.TrimSpace(options.JQL) == "" || options.Limit < 0 {
			return nil, fmt.Errorf("Jira search requires JQL and a nonnegative limit")
		}
		fields := options.Fields
		if len(fields) == 0 {
			fields = []string{"*all"}
		} else {
			fields = append(append([]string{}, fields...), "summary", "project", "issuetype")
		}
		var response struct {
			Issues *[]json.RawMessage `json:"issues"`
		}
		body := struct {
			JQL           string   `json:"jql"`
			Fields        []string `json:"fields"`
			MaxResults    int      `json:"maxResults,omitempty"`
			NextPageToken string   `json:"nextPageToken,omitempty"`
		}{options.JQL, fields, options.Limit, options.After}
		if err := c.json(ctx, requestOptions{Method: "POST", Path: "/search/jql", Body: body}, &response); err != nil {
			return nil, err
		}
		if response.Issues == nil {
			return nil, fmt.Errorf("Jira search response has no issues array")
		}
		issues := make([]Issue, 0, len(*response.Issues))
		progress.SetProgress(0, len(*response.Issues))
		if len(*response.Issues) == 0 && !options.NoCache {
			progress.SetDescription("No issues found")
			return issues, nil
		}
		progress.SetDescription("Loading field catalog")
		catalog, err := c.Catalog(ctx, CatalogOptions{NoCache: options.NoCache, Store: options.Store})
		if err != nil {
			return nil, err
		}
		for _, raw := range *response.Issues {
			issue, err := c.parseIssue(raw, catalog)
			if err != nil {
				return nil, err
			}
			issues = append(issues, issue)
			progress.SetProgress(len(issues), len(*response.Issues))
			progress.SetDescription(fmt.Sprintf("Loaded %d/%d issues", len(issues), len(*response.Issues)))
		}
		return issues, nil
	})
}

func (c *Client) parseIssue(raw json.RawMessage, catalog map[string]FieldDefinition) (Issue, error) {
	var payload struct {
		ID     string                     `json:"id"`
		Key    string                     `json:"key"`
		Fields map[string]json.RawMessage `json:"fields"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return Issue{}, fmt.Errorf("decode Jira issue: %w", err)
	}
	if payload.Fields == nil || payload.Key == "" {
		return Issue{}, fmt.Errorf("Jira issue response requires a key and fields object")
	}
	for key, value := range map[string]string{"id": payload.ID, "key": payload.Key, "url": c.Site + "/browse/" + url.PathEscape(payload.Key)} {
		encoded, err := json.Marshal(value)
		if err != nil {
			return Issue{}, err
		}
		payload.Fields[key] = encoded
	}
	flat, err := json.Marshal(payload.Fields)
	if err != nil {
		return Issue{}, err
	}
	return ParseIssue(flat, ParseOptions{Site: c.Site, Catalog: catalog, IgnoreFields: c.ignoreFields})
}
