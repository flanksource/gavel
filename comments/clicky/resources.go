package clicky

import (
	"context"

	"github.com/flanksource/clicky/entity"
	"github.com/flanksource/gavel/comments"
)

type publicationFlags struct {
	Version   int64  `flag:"expectedVersion" required:"true" help:"Captured comment version"`
	Provider  string `flag:"provider" required:"true" help:"Registered native provider"`
	RequestID string `flag:"requestId" required:"true" help:"Stable publication request identity"`
	MessageID string `flag:"messageId" help:"Local reply to publish"`
}

func (publicationFlags) ClickyActionFlags() {}

type syncFlags struct {
	Provider string          `flag:"provider" required:"true" help:"Registered native provider"`
	Target   comments.Target `flag:"target" required:"true" help:"Authorized PR target as JSON"`
}

func (syncFlags) ClickyActionFlags() {}

type operationFlags struct {
	OperationID string `flag:"operationId" help:"Admitted operation identity"`
}

func (operationFlags) ClickyActionFlags() {}

func registerResources(b *entity.EntityBuilder[listItem, comments.Filter, comments.Comment], resolve Resolver) {
	b.WithAction(entity.ActionWithContext("capabilities", func(ctx context.Context, id string, _ map[string]string) (comments.Capabilities, error) {
		s, err := resolve(ctx)
		if err != nil {
			return comments.Capabilities{}, err
		}
		v, err := s.Capabilities(ctx, id)
		return v, StatusError(err)
	}).WithMethod("GET"))
	b.WithAction(entity.TypedActionWithContext("read", actionFlags{}, func(ctx context.Context, id string, f actionFlags) (comments.ReadMarker, error) {
		s, err := resolve(ctx)
		if err != nil {
			return comments.ReadMarker{}, err
		}
		v, err := s.MarkRead(ctx, id, f.Version)
		return v, StatusError(err)
	}).WithMethod("POST"))
	for _, action := range []string{"publish", "publish-reply", "resolve-native", "reopen-native"} {
		b.WithAction(entity.TypedActionWithContext(action, publicationFlags{}, func(ctx context.Context, id string, f publicationFlags) (comments.Operation, error) {
			s, err := resolve(ctx)
			if err != nil {
				return comments.Operation{}, err
			}
			v, err := s.Publish(ctx, comments.PublicationRequest{CommentID: id, ExpectedVersion: f.Version, Provider: f.Provider, RequestID: f.RequestID, MessageID: f.MessageID, Action: action})
			return v, StatusError(err)
		}).WithMethod("POST"))
	}
	b.WithAction(entity.TypedActionWithContext("sync", syncFlags{}, func(ctx context.Context, _ string, f syncFlags) ([]comments.Comment, error) {
		s, err := resolve(ctx)
		if err != nil {
			return nil, err
		}
		v, err := s.Sync(ctx, f.Provider, f.Target)
		return v, StatusError(err)
	}).WithOptionalID().WithPath("comments/sync").WithMethod("POST"))
	b.WithAction(entity.TypedActionWithContext("operation", operationFlags{}, func(ctx context.Context, id string, f operationFlags) (comments.Operation, error) {
		s, err := resolve(ctx)
		if err != nil {
			return comments.Operation{}, err
		}
		v, err := s.Operation(ctx, f.OperationID)
		if err == nil && v.CommentID != id {
			return comments.Operation{}, StatusError(comments.ErrNotFound)
		}
		return v, StatusError(err)
	}).WithMethod("GET"))
}
