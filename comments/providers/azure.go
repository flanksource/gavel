// Azure DevOps comments: a pull request's threads, read and written through commons-db's thread
// writer on the repository its connection reaches, as gavel's remote comments and bindings.
package providers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/gavel/comments"
)

// Azure publishes comments to Azure DevOps pull requests. Resolve reaches a target's repository
// through the connection the actor may use.
type Azure struct {
	Resolve func(context.Context, comments.Principal, comments.Target) (*git.Repository, error)
}

func (Azure) Name() string { return "azuredevops" }

// threadsOf is the thread writer of target's repository and target's pull request number.
func (p Azure) threadsOf(ctx context.Context, actor comments.Principal, target comments.Target) (*git.Repository, git.PullRequestThreadWriter, git.PullRequestThreadReader, int, error) {
	if p.Resolve == nil {
		return nil, nil, nil, 0, fmt.Errorf("Azure DevOps repository resolver is required")
	}
	_, numberText, found := strings.Cut(target.URI, "/pullrequest/")
	number, err := strconv.Atoi(strings.TrimSuffix(numberText, "/"))
	if !found || err != nil || number <= 0 {
		return nil, nil, nil, 0, fmt.Errorf("%w: %s is not an Azure DevOps pull request URL", comments.ErrInvalid, target.URI)
	}
	remote, err := p.Resolve(ctx, actor, target)
	if err != nil {
		return nil, nil, nil, 0, err
	}
	writer, writes := remote.Provider.(git.PullRequestThreadWriter)
	reader, reads := remote.Provider.(git.PullRequestThreadReader)
	if !writes || !reads {
		return nil, nil, nil, 0, fmt.Errorf("the provider of %s doesn't keep pull request threads", target.URI)
	}
	return remote, writer, reader, number, nil
}

func (p Azure) Fetch(ctx context.Context, actor comments.Principal, target comments.Target) ([]comments.RemoteComment, error) {
	remote, _, reader, number, err := p.threadsOf(ctx, actor, target)
	if err != nil {
		return nil, err
	}
	threads, err := reader.PullRequestThreads(ctx, remote.URL, number)
	if err != nil {
		return nil, err
	}
	fetched := make([]comments.RemoteComment, 0, len(threads))
	for _, thread := range threads {
		snapshot, err := azureSnapshot(target, thread)
		if err != nil {
			return nil, err
		}
		fetched = append(fetched, snapshot)
	}
	return fetched, nil
}

func (p Azure) Apply(ctx context.Context, actor comments.Principal, action string, publication comments.Publication) (comments.RemoteComment, error) {
	target := publication.Comment.Target
	remote, writer, _, number, err := p.threadsOf(ctx, actor, target)
	if err != nil {
		return comments.RemoteComment{}, err
	}
	if action == "publish" {
		return publishAzure(ctx, writer, remote, number, publication)
	}
	if publication.Binding == nil || publication.Binding.ThreadID == "" {
		return comments.RemoteComment{}, comments.ErrInvalid
	}
	binding := publication.Binding
	current, err := writer.PullRequestThread(ctx, remote.URL, number, binding.ThreadID)
	if err != nil {
		return comments.RemoteComment{}, err
	}
	before, err := azureSnapshot(target, current)
	if err != nil {
		return before, err
	}
	if before.Binding.RemoteKey != binding.RemoteKey || before.Binding.Revision != binding.Revision {
		return before, comments.ErrConflict
	}
	switch action {
	case "publish-reply":
		if publication.Message == nil {
			return before, comments.ErrInvalid
		}
		parent := binding.RootID
		if publication.Message.ParentID != "" {
			parent = ""
			for remoteID, local := range binding.MessageIDs {
				if local == publication.Message.ParentID {
					parent = remoteID
				}
			}
			if parent == "" {
				return before, comments.ErrInvalid
			}
		}
		reply, err := writer.ReplyPullRequestThread(ctx, remote.URL, number, binding.ThreadID, parent, publication.Message.Body)
		if err != nil {
			return before, err
		}
		current.Comments = append(current.Comments, reply)
		snapshot, err := azureSnapshot(target, current)
		snapshot.Binding.MessageIDs = map[string]string{reply.ID: publication.Message.ID}
		return snapshot, err
	case "resolve-native", "reopen-native":
		changed, err := writer.ResolvePullRequestThread(ctx, remote.URL, number, binding.ThreadID, action == "resolve-native")
		if err != nil {
			return before, err
		}
		return azureSnapshot(target, changed)
	default:
		return before, comments.ErrInvalid
	}
}

// publishAzure opens a thread for publication: on the lines of its captured diff when it has a
// selector, refused as a conflict when that diff isn't the target's revision or the pull request's.
func publishAzure(ctx context.Context, writer git.PullRequestThreadWriter, remote *git.Repository, number int, publication comments.Publication) (comments.RemoteComment, error) {
	target := publication.Comment.Target
	if strings.TrimSpace(publication.Comment.Body) == "" {
		return comments.RemoteComment{}, comments.ErrInvalid
	}
	request := git.PullRequestThreadRequest{Body: publication.Comment.Body}
	if len(target.Selector) > 0 {
		selector, err := comments.ParseCodeSelector(target.Selector)
		if err != nil {
			return comments.RemoteComment{}, err
		}
		if selector.Kind != "diff-range" || selector.HeadCommit != target.Revision {
			return comments.RemoteComment{}, comments.ErrConflict
		}
		request.Path, request.StartLine, request.EndLine, request.Old = selector.Path, selector.StartLine, selector.EndLine, selector.Side == "old"
		request.HeadSHA, request.BaseSHA = selector.HeadCommit, selector.BaseCommit
	}
	thread, err := writer.CreatePullRequestThread(ctx, remote.URL, number, request)
	if errors.Is(err, git.ErrThreadTargetMoved) {
		return comments.RemoteComment{}, fmt.Errorf("%w: %v", comments.ErrConflict, err)
	}
	if err != nil {
		return comments.RemoteComment{}, err
	}
	return azureSnapshot(target, thread)
}

// azureSnapshot is thread as a remote comment of target: its first comment, the replies after it,
// and a binding whose revision changes with anything the thread holds.
func azureSnapshot(target comments.Target, thread git.PullRequestThread) (comments.RemoteComment, error) {
	if thread.ID == "" || len(thread.Comments) == 0 {
		return comments.RemoteComment{}, fmt.Errorf("incomplete Azure DevOps comment thread")
	}
	native, err := json.Marshal(thread)
	if err != nil {
		return comments.RemoteComment{}, err
	}
	sum := sha256.Sum256(native)
	actor := func(comment git.PullRequestThreadComment) comments.Actor {
		return comments.Actor{Kind: "user", ID: comment.AuthorID, Label: comment.Author}
	}
	root := thread.Comments[0]
	snapshot := comments.RemoteComment{Binding: comments.Binding{Provider: "azuredevops", RemoteKey: target.URI + "#thread=" + thread.ID,
		ThreadID: thread.ID, RootID: root.ID, URL: target.URI + "?discussionId=" + thread.ID, Revision: hex.EncodeToString(sum[:]),
		Resolvable: true, Native: native}, Body: root.Body, Author: actor(root), CreatedAt: root.CreatedAt, Resolved: thread.Resolved,
		ResolutionReason: thread.Status}
	for _, comment := range thread.Comments[1:] {
		snapshot.Messages = append(snapshot.Messages, comments.RemoteMessage{ID: comment.ID, ParentID: comment.ParentID, Body: comment.Body,
			Author: actor(comment), CreatedAt: comment.CreatedAt})
	}
	return snapshot, nil
}
