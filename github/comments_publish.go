package github

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/comments"
)

func (a CommentAPI) Apply(ctx context.Context, action string, p comments.Publication) (comments.RemoteComment, error) {
	repo, number, err := a.target(p.Comment.Target)
	if err != nil {
		return comments.RemoteComment{}, err
	}
	if action == "publish" {
		return a.publish(ctx, p, repo, number)
	}
	if p.Binding == nil || p.Binding.ThreadID == "" {
		return comments.RemoteComment{}, fmt.Errorf("native GitHub thread binding required")
	}
	snapshots, err := a.Fetch(ctx, p.Comment.Target)
	if err != nil {
		return comments.RemoteComment{}, err
	}
	before, err := findNative(snapshots, p.Binding.RemoteKey, "")
	if err != nil {
		return before, err
	}
	if before.Binding.Revision != p.Binding.Revision {
		return before, comments.ErrConflict
	}
	var publishedReply int64
	switch action {
	case "publish-reply":
		if p.Message == nil || p.Message.ParentID != "" {
			return before, fmt.Errorf("GitHub replies must target the root thread")
		}
		id, err := strconv.ParseInt(p.Binding.RootID, 10, 64)
		if err != nil || id <= 0 {
			return before, comments.ErrInvalid
		}
		var receipt nativeComment
		err = a.request(ctx, "POST", fmt.Sprintf("/repos/%s/pulls/%d/comments/%d/replies", repo, number, id), map[string]any{"body": p.Message.Body}, &receipt)
		if err != nil {
			return before, err
		}
		if receipt.ID <= 0 {
			return before, fmt.Errorf("GitHub reply receipt has no native ID")
		}
		publishedReply = receipt.ID
		if before.Binding.MessageIDs == nil {
			before.Binding.MessageIDs = map[string]string{}
		}
		before.Binding.MessageIDs[strconv.FormatInt(receipt.ID, 10)] = p.Message.ID
	case "resolve-native", "reopen-native":
		mutation, input := "resolveReviewThread", "ResolveReviewThreadInput"
		if action == "reopen-native" {
			mutation, input = "unresolveReviewThread", "UnresolveReviewThreadInput"
		}
		var result struct {
			Data map[string]struct {
				Thread struct {
					ID       string `json:"id"`
					Resolved bool   `json:"isResolved"`
				} `json:"thread"`
			} `json:"data"`
		}
		err = a.graph(ctx, fmt.Sprintf("mutation($input:%s!){%s(input:$input){thread{id isResolved}}}", input, mutation), map[string]any{"input": map[string]any{"threadId": p.Binding.ThreadID}}, &result)
		if err != nil {
			return before, err
		}
		got := result.Data[mutation].Thread
		if got.ID != p.Binding.ThreadID || got.Resolved != (action == "resolve-native") {
			return before, fmt.Errorf("GitHub did not confirm thread resolution")
		}
	default:
		return before, comments.ErrInvalid
	}
	snapshots, err = a.Fetch(ctx, p.Comment.Target)
	if err != nil {
		return before, err
	}
	got, err := findNative(snapshots, p.Binding.RemoteKey, "")
	if err != nil {
		return before, err
	}
	if publishedReply > 0 {
		got.Binding.MessageIDs = map[string]string{strconv.FormatInt(publishedReply, 10): p.Message.ID}
	}
	return got, nil
}
func (a CommentAPI) publish(ctx context.Context, p comments.Publication, repo string, number int) (comments.RemoteComment, error) {
	if strings.TrimSpace(p.Comment.Body) == "" {
		return comments.RemoteComment{}, comments.ErrInvalid
	}
	if len(p.Comment.Target.Selector) == 0 {
		var receipt nativeComment
		err := a.request(ctx, "POST", fmt.Sprintf("/repos/%s/issues/%d/comments", repo, number), map[string]any{"body": p.Comment.Body}, &receipt)
		if err != nil {
			return comments.RemoteComment{}, err
		}
		return generalSnapshot(p.Comment.Target, receipt), nil
	}
	selector, err := comments.ParseCodeSelector(p.Comment.Target.Selector)
	if err != nil {
		return comments.RemoteComment{}, err
	}
	if selector.Kind != "diff-range" || p.Comment.Target.Revision == "" || selector.HeadCommit != p.Comment.Target.Revision {
		return comments.RemoteComment{}, fmt.Errorf("captured PR head revision required")
	}
	var pr struct {
		Head struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}
	if err = a.request(ctx, "GET", fmt.Sprintf("/repos/%s/pulls/%d", repo, number), nil, &pr); err != nil {
		return comments.RemoteComment{}, err
	}
	if pr.Head.SHA != selector.HeadCommit {
		return comments.RemoteComment{}, comments.ErrConflict
	}
	side := "RIGHT"
	if selector.Side == "old" {
		side = "LEFT"
	}
	body := map[string]any{"body": p.Comment.Body, "commit_id": selector.HeadCommit, "path": selector.Path, "side": side, "line": selector.EndLine}
	if selector.StartLine != selector.EndLine {
		body["start_line"] = selector.StartLine
		body["start_side"] = side
	}
	var receipt nativeComment
	if err = a.request(ctx, "POST", fmt.Sprintf("/repos/%s/pulls/%d/comments", repo, number), body, &receipt); err != nil {
		return comments.RemoteComment{}, err
	}
	partial := generalSnapshot(p.Comment.Target, receipt)
	snapshots, err := a.Fetch(ctx, p.Comment.Target)
	if err != nil {
		return partial, err
	}
	got, err := findNative(snapshots, "", strconv.FormatInt(receipt.ID, 10))
	if err != nil {
		return partial, err
	}
	return got, nil
}
func findNative(snapshots []comments.RemoteComment, key, root string) (comments.RemoteComment, error) {
	for _, s := range snapshots {
		if (key != "" && s.Binding.RemoteKey == key) || (root != "" && s.Binding.RootID == root) {
			return s, nil
		}
	}
	return comments.RemoteComment{}, fmt.Errorf("native GitHub comment not found in target PR")
}
