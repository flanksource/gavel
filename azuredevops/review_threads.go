package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/flanksource/gavel/pr/model"
)

type reviewThread struct {
	ID      int             `json:"id"`
	Status  json.RawMessage `json:"status"`
	Deleted bool            `json:"isDeleted"`
	Context *struct {
		Path  string `json:"filePath"`
		Right *struct {
			Line int `json:"line"`
		} `json:"rightFileStart"`
	} `json:"threadContext"`
	Comments []struct {
		ID      int64     `json:"id"`
		Body    string    `json:"content"`
		Kind    string    `json:"commentType"`
		Deleted bool      `json:"isDeleted"`
		Created time.Time `json:"publishedDate"`
		Author  struct {
			Name   string `json:"displayName"`
			Login  string `json:"uniqueName"`
			Avatar string `json:"imageUrl"`
		} `json:"author"`
	} `json:"comments"`
}

func threadResolved(status json.RawMessage) (bool, error) {
	switch string(status) {
	case `1`, `6`, `"active"`, `"pending"`:
		return false, nil
	case `2`, `3`, `4`, `5`, `"fixed"`, `"wontFix"`, `"closed"`, `"byDesign"`:
		return true, nil
	default:
		return false, fmt.Errorf("unexpected Azure review thread status %s", status)
	}
}

func (c *Client) ReviewThreads(ctx context.Context, number int) ([]model.PRComment, error) {
	if number <= 0 {
		return nil, fmt.Errorf("azure PR number must be positive")
	}
	if err := c.Preflight(ctx); err != nil {
		return nil, err
	}
	threads, err := collection[reviewThread](ctx, c, c.repositoryPath()+"/pullrequests/"+strconv.Itoa(number)+"/threads", nil)
	if err != nil {
		return nil, err
	}
	var comments []model.PRComment
	for _, thread := range threads {
		if thread.Deleted {
			continue
		}
		for _, comment := range thread.Comments {
			if comment.Deleted || comment.Kind != "text" || comment.Body == "" {
				continue
			}
			resolved, err := threadResolved(thread.Status)
			if err != nil {
				return nil, err
			}
			if thread.ID <= 0 || comment.ID <= 0 {
				return nil, fmt.Errorf("azure review response is missing a native thread or comment ID")
			}
			item := model.PRComment{ID: comment.ID, ThreadID: strconv.Itoa(thread.ID), CanResolve: true, Body: comment.Body, Author: comment.Author.Login, AvatarURL: comment.Author.Avatar, CreatedAt: comment.Created, IsResolved: resolved, IsReviewThread: true, URL: fmt.Sprintf("%s/pullrequest/%d?discussionId=%d", c.repo.URL(), number, thread.ID)}
			if thread.Context != nil {
				item.Path = thread.Context.Path
				if thread.Context.Right != nil {
					item.Line = thread.Context.Right.Line
				}
			}
			comments = append(comments, item)
		}
	}
	return comments, nil
}

func (c *Client) ResolveThread(ctx context.Context, number int, id string) error {
	threadID, err := strconv.Atoi(id)
	if err != nil || threadID <= 0 {
		return fmt.Errorf("azure review thread ID must be a positive integer")
	}
	threads, err := c.ReviewThreads(ctx, number)
	if err != nil {
		return err
	}
	found := false
	for _, thread := range threads {
		if thread.ThreadID == id {
			if thread.IsResolved {
				return nil
			}
			found = true
		}
	}
	if !found {
		return fmt.Errorf("review thread %s does not belong to Azure PR #%d", id, number)
	}
	data, _, err := c.request(ctx, http.MethodPatch, c.repositoryPath()+"/pullrequests/"+strconv.Itoa(number)+"/threads/"+id, nil, map[string]string{"status": "fixed"})
	if err != nil {
		return err
	}
	var result reviewThread
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}
	resolved, err := threadResolved(result.Status)
	if err != nil {
		return err
	}
	if result.ID != threadID || !resolved {
		return fmt.Errorf("azure did not resolve review thread %s", id)
	}
	return nil
}
