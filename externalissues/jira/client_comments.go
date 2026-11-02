package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/externalissues"
)

func (c *Client) Comments(ctx context.Context, key string) (Comments, error) {
	return runJiraTask(ctx, "Load Jira comments "+key, func(ctx context.Context, progress *task.Task) (Comments, error) {
		if key == "" {
			return nil, fmt.Errorf("jira issue key is required")
		}
		comments := make(Comments, 0)
		progress.SetDescription("Fetching comments")
		for start := 0; ; {
			var page struct {
				Total    *int               `json:"total"`
				Comments *[]json.RawMessage `json:"comments"`
			}
			if err := c.json(ctx, requestOptions{Path: "/issue/" + url.PathEscape(key) + "/comment", Query: url.Values{"startAt": {strconv.Itoa(start)}, "maxResults": {"100"}}}, &page); err != nil {
				return nil, err
			}
			if page.Total == nil || page.Comments == nil || *page.Total < 0 {
				return nil, fmt.Errorf("jira comments response requires total and comments")
			}
			for _, raw := range *page.Comments {
				comment, err := parseComment(raw)
				if err != nil {
					return nil, err
				}
				comments = append(comments, comment)
			}
			start += len(*page.Comments)
			progress.SetProgress(start, *page.Total)
			progress.SetDescription(fmt.Sprintf("Loaded %d/%d comments", start, *page.Total))
			if start >= *page.Total {
				return comments, nil
			}
			if len(*page.Comments) == 0 {
				return nil, fmt.Errorf("jira comments pagination stopped at %d of %d", start, *page.Total)
			}
		}
	})
}

type CommentOptions struct {
	Key  string
	Body string
}

func (c *Client) CreateComment(ctx context.Context, options CommentOptions) (externalissues.Comment, error) {
	body, err := MarkdownToADF(options.Body)
	if err != nil {
		return externalissues.Comment{}, err
	}
	if options.Key == "" {
		return externalissues.Comment{}, fmt.Errorf("jira issue key is required")
	}
	var raw json.RawMessage
	if err := c.json(ctx, requestOptions{Method: "POST", Path: "/issue/" + url.PathEscape(options.Key) + "/comment", Body: struct {
		Body ADFNode `json:"body"`
	}{body}}, &raw); err != nil {
		return externalissues.Comment{}, err
	}
	return parseComment(raw)
}
