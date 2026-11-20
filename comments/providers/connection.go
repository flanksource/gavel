// Comments on Azure DevOps pull requests and GitLab merge requests: their threads, read and written
// through commons-db's thread writer on the repository a git connection reaches, as gavel's remote
// comments and bindings.
package providers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/forge"
)

// Connection publishes comments to the pull requests of one kind of provider a git connection
// reaches: azuredevops or gitlab. Resolve reaches a target's repository through the connection
// the actor may use.
type Connection struct {
	Kind    string
	Resolve func(context.Context, comments.Principal, comments.Target) (*git.Repository, error)
}

// connectionKinds are the connection types of each Connection kind's pull requests.
var connectionKinds = map[string]string{"azuredevops": models.ConnectionTypeAzureDevops, "gitlab": models.ConnectionTypeGitlab}

func (p Connection) Name() string { return p.Kind }

// threadsOf is the thread writer of target's repository and target's pull request number.
func (p Connection) threadsOf(ctx context.Context, actor comments.Principal, target comments.Target) (*git.Repository, git.PullRequestThreadWriter, git.PullRequestThreadReader, int, error) {
	if p.Resolve == nil {
		return nil, nil, nil, 0, fmt.Errorf("%s repository resolver is required", p.Kind)
	}
	pr, err := forge.ParsePullRequestURL(target.URI)
	if err != nil {
		return nil, nil, nil, 0, fmt.Errorf("%w: %v", comments.ErrInvalid, err)
	}
	if pr.Provider != connectionKinds[p.Kind] {
		return nil, nil, nil, 0, fmt.Errorf("%w: %s is not a %s pull request URL", comments.ErrInvalid, target.URI, p.Kind)
	}
	number := pr.Number
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

func (p Connection) Fetch(ctx context.Context, actor comments.Principal, target comments.Target) ([]comments.RemoteComment, error) {
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
		snapshot, err := p.snapshot(target, thread)
		if err != nil {
			return nil, err
		}
		fetched = append(fetched, snapshot)
	}
	return fetched, nil
}

func (p Connection) Apply(ctx context.Context, actor comments.Principal, action string, publication comments.Publication) (comments.RemoteComment, error) {
	target := publication.Comment.Target
	remote, writer, _, number, err := p.threadsOf(ctx, actor, target)
	if err != nil {
		return comments.RemoteComment{}, err
	}
	if action == "publish" {
		return p.publish(ctx, writer, remote, number, publication)
	}
	if publication.Binding == nil || publication.Binding.ThreadID == "" {
		return comments.RemoteComment{}, comments.ErrInvalid
	}
	binding := publication.Binding
	current, err := writer.PullRequestThread(ctx, remote.URL, number, binding.ThreadID)
	if err != nil {
		return comments.RemoteComment{}, err
	}
	before, err := p.snapshot(target, current)
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
		snapshot, err := p.snapshot(target, current)
		snapshot.Binding.MessageIDs = map[string]string{reply.ID: publication.Message.ID}
		return snapshot, err
	case "resolve-native", "reopen-native":
		changed, err := writer.ResolvePullRequestThread(ctx, remote.URL, number, binding.ThreadID, action == "resolve-native")
		if err != nil {
			return before, err
		}
		return p.snapshot(target, changed)
	default:
		return before, comments.ErrInvalid
	}
}

// publish opens a thread for publication: on the lines of its captured diff when it has a
// selector, refused as a conflict when that diff isn't the target's revision or the pull request's.
func (p Connection) publish(ctx context.Context, writer git.PullRequestThreadWriter, remote *git.Repository, number int, publication comments.Publication) (comments.RemoteComment, error) {
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
	return p.snapshot(target, thread)
}

// snapshot is thread as a remote comment of target: its first comment, the replies after it,
// and a binding whose revision changes with anything the thread holds. A thread without a status
// has no resolution to change.
func (p Connection) snapshot(target comments.Target, thread git.PullRequestThread) (comments.RemoteComment, error) {
	if thread.ID == "" || len(thread.Comments) == 0 {
		return comments.RemoteComment{}, fmt.Errorf("incomplete %s comment thread %q of %s", p.Kind, thread.ID, target.URI)
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
	snapshot := comments.RemoteComment{Binding: comments.Binding{Provider: p.Kind, RemoteKey: target.URI + "#thread=" + thread.ID,
		ThreadID: thread.ID, RootID: root.ID, URL: thread.URL, Revision: hex.EncodeToString(sum[:]),
		Resolvable: thread.Status != "", Native: native}, Body: root.Body, Author: actor(root), CreatedAt: root.CreatedAt, Resolved: thread.Resolved,
		ResolutionReason: thread.Status}
	for _, comment := range thread.Comments[1:] {
		snapshot.Messages = append(snapshot.Messages, comments.RemoteMessage{ID: comment.ID, ParentID: comment.ParentID, Body: comment.Body,
			Author: actor(comment), CreatedAt: comment.CreatedAt})
	}
	return snapshot, nil
}
