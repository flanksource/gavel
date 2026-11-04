package azuredevops

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/gavel/comments"
)

type nativeAzureMessage struct {
	ID        int64           `json:"id"`
	ParentID  int64           `json:"parentCommentId"`
	Body      string          `json:"content"`
	Kind      json.RawMessage `json:"commentType"`
	Deleted   bool            `json:"isDeleted"`
	CreatedAt time.Time       `json:"publishedDate"`
	UpdatedAt string          `json:"lastUpdatedDate"`
	Author    struct {
		ID    string `json:"id"`
		Name  string `json:"displayName"`
		Login string `json:"uniqueName"`
	} `json:"author"`
}
type nativeAzureThread struct {
	ID        int                  `json:"id"`
	Status    json.RawMessage      `json:"status"`
	Deleted   bool                 `json:"isDeleted"`
	UpdatedAt string               `json:"lastUpdatedDate"`
	Context   json.RawMessage      `json:"threadContext"`
	Iteration json.RawMessage      `json:"pullRequestThreadContext"`
	Comments  []nativeAzureMessage `json:"comments"`
}

func (c *Client) commentTarget(t comments.Target) (int, error) {
	repo, number, err := ParsePRURL(t.URI)
	if err != nil {
		return 0, err
	}
	if repo != c.repo {
		return 0, fmt.Errorf("native Azure credentials are not scoped to target repository")
	}
	return number, nil
}
func azureSnapshot(target comments.Target, t nativeAzureThread) (comments.RemoteComment, error) {
	if t.ID <= 0 || len(t.Comments) == 0 {
		return comments.RemoteComment{}, fmt.Errorf("incomplete Azure comment thread")
	}
	resolved, err := threadResolved(t.Status)
	if err != nil {
		return comments.RemoteComment{}, err
	}
	raw, err := json.Marshal(t)
	if err != nil {
		return comments.RemoteComment{}, err
	}
	sum := sha256.Sum256(raw)
	root := t.Comments[0]
	actor := func(m nativeAzureMessage) comments.Actor {
		return comments.Actor{Kind: "user", ID: m.Author.ID, Label: m.Author.Name}
	}
	r := comments.RemoteComment{Binding: comments.Binding{Provider: "azuredevops", RemoteKey: target.URI + "#thread=" + strconv.Itoa(t.ID), ThreadID: strconv.Itoa(t.ID), RootID: strconv.FormatInt(root.ID, 10), URL: target.URI + "?discussionId=" + strconv.Itoa(t.ID), Revision: hex.EncodeToString(sum[:]), Resolvable: true, Native: raw}, Body: root.Body, Author: actor(root), CreatedAt: root.CreatedAt, Resolved: resolved, ResolutionReason: string(t.Status)}
	for _, m := range t.Comments[1:] {
		if m.Deleted {
			continue
		}
		parent := ""
		if m.ParentID > 0 {
			parent = strconv.FormatInt(m.ParentID, 10)
		}
		r.Messages = append(r.Messages, comments.RemoteMessage{ID: strconv.FormatInt(m.ID, 10), ParentID: parent, Body: m.Body, Author: actor(m), CreatedAt: m.CreatedAt})
	}
	return r, nil
}
func (c *Client) FetchComments(ctx context.Context, target comments.Target) ([]comments.RemoteComment, error) {
	number, err := c.commentTarget(target)
	if err != nil {
		return nil, err
	}
	if err = c.Preflight(ctx); err != nil {
		return nil, err
	}
	rows, err := collection[nativeAzureThread](ctx, c, c.repositoryPath()+"/pullrequests/"+strconv.Itoa(number)+"/threads", nil)
	if err != nil {
		return nil, err
	}
	out := make([]comments.RemoteComment, 0, len(rows))
	for _, row := range rows {
		if row.Deleted {
			continue
		}
		if len(row.Comments) == 0 {
			continue
		}
		if string(row.Comments[0].Kind) != "\"text\"" && string(row.Comments[0].Kind) != "1" {
			continue
		}
		snapshot, err := azureSnapshot(target, row)
		if err != nil {
			return nil, err
		}
		out = append(out, snapshot)
	}
	return out, nil
}
func (c *Client) ApplyComment(ctx context.Context, action string, p comments.Publication) (comments.RemoteComment, error) {
	number, err := c.commentTarget(p.Comment.Target)
	if err != nil {
		return comments.RemoteComment{}, err
	}
	if err = c.Preflight(ctx); err != nil {
		return comments.RemoteComment{}, err
	}
	path := c.repositoryPath() + "/pullrequests/" + strconv.Itoa(number) + "/threads"
	if action == "publish" {
		return c.publishComment(ctx, p, path, number)
	}
	if p.Binding == nil {
		return comments.RemoteComment{}, comments.ErrInvalid
	}
	id, err := strconv.Atoi(p.Binding.ThreadID)
	if err != nil || id <= 0 {
		return comments.RemoteComment{}, comments.ErrInvalid
	}
	data, _, err := c.request(ctx, http.MethodGet, path+"/"+strconv.Itoa(id), nil, nil)
	if err != nil {
		return comments.RemoteComment{}, err
	}
	var thread nativeAzureThread
	if err = json.Unmarshal(data, &thread); err != nil {
		return comments.RemoteComment{}, err
	}
	before, err := azureSnapshot(p.Comment.Target, thread)
	if err != nil {
		return before, err
	}
	if before.Binding.RemoteKey != p.Binding.RemoteKey || before.Binding.Revision != p.Binding.Revision {
		return before, comments.ErrConflict
	}
	switch action {
	case "publish-reply":
		if p.Message == nil {
			return before, comments.ErrInvalid
		}
		parent, err := strconv.ParseInt(p.Binding.RootID, 10, 64)
		if err != nil {
			return before, err
		}
		if p.Message.ParentID != "" {
			found := false
			for remote, local := range p.Binding.MessageIDs {
				if local == p.Message.ParentID {
					parent, err = strconv.ParseInt(remote, 10, 64)
					found = true
					break
				}
			}
			if err != nil || !found {
				return before, comments.ErrInvalid
			}
		}
		data, _, err = c.request(ctx, http.MethodPost, path+"/"+strconv.Itoa(id)+"/comments", nil, map[string]any{"content": p.Message.Body, "parentCommentId": parent, "commentType": 1})
		if err != nil {
			return before, err
		}
		var message nativeAzureMessage
		if err = json.Unmarshal(data, &message); err != nil {
			return before, err
		}
		thread.Comments = append(thread.Comments, message)
		snapshot, err := azureSnapshot(p.Comment.Target, thread)
		if message.ID <= 0 {
			return before, fmt.Errorf("native Azure reply receipt has no native ID")
		}
		snapshot.Binding.MessageIDs = map[string]string{strconv.FormatInt(message.ID, 10): p.Message.ID}
		return snapshot, err
	case "resolve-native", "reopen-native":
		status := "fixed"
		if action == "reopen-native" {
			status = "active"
		}
		data, _, err = c.request(ctx, http.MethodPatch, path+"/"+strconv.Itoa(id), nil, map[string]string{"status": status})
		if err != nil {
			return before, err
		}
		var result nativeAzureThread
		if err = json.Unmarshal(data, &result); err != nil {
			return before, err
		}
		resolved, err := threadResolved(result.Status)
		if err != nil || result.ID != id || resolved != (action == "resolve-native") {
			return before, fmt.Errorf("Azure did not confirm requested thread state")
		}
		thread.Status = result.Status
		thread.UpdatedAt = result.UpdatedAt
		return azureSnapshot(p.Comment.Target, thread)
	default:
		return before, comments.ErrInvalid
	}
}
func (c *Client) publishComment(ctx context.Context, p comments.Publication, path string, number int) (comments.RemoteComment, error) {
	if strings.TrimSpace(p.Comment.Body) == "" {
		return comments.RemoteComment{}, comments.ErrInvalid
	}
	payload := map[string]any{"comments": []any{map[string]any{"content": p.Comment.Body, "parentCommentId": 0, "commentType": 1}}, "status": "active"}
	if len(p.Comment.Target.Selector) > 0 {
		selector, err := comments.ParseCodeSelector(p.Comment.Target.Selector)
		if err != nil {
			return comments.RemoteComment{}, err
		}
		if selector.Kind != "diff-range" || selector.HeadCommit != p.Comment.Target.Revision {
			return comments.RemoteComment{}, comments.ErrConflict
		}
		var pr struct {
			Head struct {
				ID string `json:"commitId"`
			} `json:"lastMergeSourceCommit"`
		}
		data, _, err := c.request(ctx, http.MethodGet, c.repositoryPath()+"/pullrequests/"+strconv.Itoa(number), nil, nil)
		if err != nil {
			return comments.RemoteComment{}, err
		}
		if err = json.Unmarshal(data, &pr); err != nil {
			return comments.RemoteComment{}, err
		}
		if pr.Head.ID != selector.HeadCommit {
			return comments.RemoteComment{}, comments.ErrConflict
		}
		prefix := "rightFile"
		if selector.Side == "old" {
			prefix = "leftFile"
		}
		payload["threadContext"] = map[string]any{"filePath": "/" + selector.Path, prefix + "Start": map[string]int{"line": selector.StartLine, "offset": 0}, prefix + "End": map[string]int{"line": selector.EndLine, "offset": 0}}
		iteration, err := c.commentIteration(ctx, number, selector)
		if err != nil {
			return comments.RemoteComment{}, err
		}
		payload["pullRequestThreadContext"] = iteration
	}
	data, _, err := c.request(ctx, http.MethodPost, path, nil, payload)
	if err != nil {
		return comments.RemoteComment{}, err
	}
	var thread nativeAzureThread
	if err = json.Unmarshal(data, &thread); err != nil {
		return comments.RemoteComment{}, err
	}
	return azureSnapshot(p.Comment.Target, thread)
}
