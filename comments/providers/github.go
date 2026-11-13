// Package providers adapts Gavel's provider clients to the shared comments service.
package providers

import (
	"context"
	"fmt"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/github"
)

type GitHub struct {
	Resolve func(context.Context, comments.Principal, comments.Target) (github.Options, error)
}

func (GitHub) Name() string { return "github" }
func (p GitHub) Fetch(ctx context.Context, actor comments.Principal, target comments.Target) ([]comments.RemoteComment, error) {
	if p.Resolve == nil {
		return nil, fmt.Errorf("GitHub credential resolver is required")
	}
	opts, err := p.Resolve(ctx, actor, target)
	if err != nil {
		return nil, err
	}
	return (github.CommentAPI{Options: opts}).Fetch(ctx, target)
}
func (p GitHub) Apply(ctx context.Context, actor comments.Principal, action string, publication comments.Publication) (comments.RemoteComment, error) {
	if p.Resolve == nil {
		return comments.RemoteComment{}, fmt.Errorf("GitHub credential resolver is required")
	}
	opts, err := p.Resolve(ctx, actor, publication.Comment.Target)
	if err != nil {
		return comments.RemoteComment{}, err
	}
	return (github.CommentAPI{Options: opts}).Apply(ctx, action, publication)
}
