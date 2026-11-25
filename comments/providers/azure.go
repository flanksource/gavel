package providers

import (
	"context"
	"fmt"
	"github.com/flanksource/gavel/azuredevops"
	"github.com/flanksource/gavel/comments"
)

type Azure struct {
	Resolve func(context.Context, comments.Principal, comments.Target) (*azuredevops.Client, error)
}

func (Azure) Name() string { return "azuredevops" }
func (p Azure) Fetch(ctx context.Context, actor comments.Principal, target comments.Target) ([]comments.RemoteComment, error) {
	if p.Resolve == nil {
		return nil, fmt.Errorf("Azure credential resolver is required")
	}
	client, err := p.Resolve(ctx, actor, target)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, fmt.Errorf("Azure credential resolver returned no client")
	}
	return client.FetchComments(ctx, target)
}
func (p Azure) Apply(ctx context.Context, actor comments.Principal, action string, publication comments.Publication) (comments.RemoteComment, error) {
	if p.Resolve == nil {
		return comments.RemoteComment{}, fmt.Errorf("Azure credential resolver is required")
	}
	client, err := p.Resolve(ctx, actor, publication.Comment.Target)
	if err != nil {
		return comments.RemoteComment{}, err
	}
	if client == nil {
		return comments.RemoteComment{}, fmt.Errorf("Azure credential resolver returned no client")
	}
	return client.ApplyComment(ctx, action, publication)
}
