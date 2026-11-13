// Package clicky registers the same comments service for CLI, HTTP and agent tools.
package clicky

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	framework "github.com/flanksource/clicky"
	"github.com/flanksource/clicky/entity"
	"github.com/flanksource/clicky/rpc"
	"github.com/flanksource/gavel/comments"
)

type actionFlags struct {
	Version   int64            `flag:"expectedVersion" required:"true" help:"Required current comment version"`
	Body      string           `flag:"bodyMarkdown" help:"Reply or updated Markdown"`
	Summary   string           `flag:"summary" help:"Resolution or closure summary"`
	ParentID  string           `flag:"parentId" help:"Local parent reply identity"`
	Assignee  *comments.Actor  `flag:"assignee" help:"Assignee as JSON; null clears it"`
	Evidence  []string         `flag:"evidence" help:"Outcome evidence URIs"`
	Content   json.RawMessage  `flag:"content" help:"Structured reply content as JSON"`
	Target    *comments.Target `flag:"target" help:"Replacement target within the same scope"`
	MessageID string           `flag:"messageId" help:"Reply to edit or delete"`
}

func (actionFlags) ClickyActionFlags() {}

type createFlags struct {
	Target    comments.Target `flag:"target" required:"true" help:"Captured host target as JSON"`
	Body      string          `flag:"bodyMarkdown" required:"true" help:"Comment Markdown"`
	Workflow  string          `flag:"workflowId" required:"true" help:"Registered workflow ID"`
	RequestID string          `flag:"requestId" required:"true" help:"Stable request identity"`
	Content   json.RawMessage `flag:"content" help:"Structured comment content"`
	Links     []comments.Link `flag:"links" help:"Typed resource links as JSON"`
}

func (createFlags) ClickyActionFlags() {}

// Register must be called once per command registry. Authentication remains request scoped.
type Resolver func(context.Context) (*comments.Service, error)
type listItem struct {
	comments.Comment
	ReadVersion int64 `json:"readVersion"`
	Unread      bool  `json:"unread"`
}

func Register(resolve Resolver) error {
	if resolve == nil {
		return fmt.Errorf("comments service is required")
	}
	fallback := resolve
	resolve = func(ctx context.Context) (*comments.Service, error) {
		if host, ok := ctx.Value(resolverKey{}).(Resolver); ok {
			return host(ctx)
		}
		return fallback(ctx)
	}
	builder := framework.NewEntity[listItem, comments.Filter, comments.Comment]("comments").
		ListPagedWithContext(func(ctx context.Context, f comments.Filter) (entity.PagedResult[listItem], error) {
			service, err := resolve(ctx)
			if err != nil {
				return entity.PagedResult[listItem]{}, StatusError(err)
			}
			page, err := service.Page(ctx, f)
			rows := make([]listItem, 0, len(page.Items))
			reads := map[string]int64{}
			for _, r := range page.Read {
				reads[r.CommentID] = r.Version
			}
			for _, c := range page.Items {
				rows = append(rows, listItem{Comment: c, ReadVersion: reads[c.ID], Unread: reads[c.ID] < c.Version})
			}
			return entity.NewPagedResult(rows, f.Limit, f.Offset, page.Total), StatusError(err)
		}).
		GetWithContext(func(ctx context.Context, id string) (comments.Comment, error) {
			service, err := resolve(ctx)
			if err != nil {
				return comments.Comment{}, StatusError(err)
			}
			v, err := service.Get(ctx, id)
			return v, StatusError(err)
		})
	builder = builder.WithAction(entity.TypedActionWithContext("create", createFlags{}, func(ctx context.Context, _ string, flags createFlags) (comments.Comment, error) {
		service, err := resolve(ctx)
		if err != nil {
			return comments.Comment{}, StatusError(err)
		}
		req := comments.CreateRequest{Target: flags.Target, Body: flags.Body, WorkflowID: flags.Workflow, RequestID: flags.RequestID, Content: flags.Content, Links: flags.Links}
		if _, ok := rpc.RequestFromContext(ctx); ok {
			if err := decode(ctx, nil, &req); err != nil {
				return comments.Comment{}, StatusError(err)
			}
		}
		v, err := service.Create(ctx, req)
		return v, StatusError(err)
	}).WithOptionalID().WithMethod("POST").WithPath("comments"))
	for _, action := range []string{"reply", "assign", "acknowledge", "edit", "resolve", "close", "reopen", "metadata", "retarget", "delete", "restore", "edit-reply", "delete-reply"} {
		builder = builder.WithAction(entity.TypedActionWithContext(action, actionFlags{}, func(ctx context.Context, id string, flags actionFlags) (comments.Comment, error) {
			service, err := resolve(ctx)
			if err != nil {
				return comments.Comment{}, StatusError(err)
			}
			req := comments.ActionRequest{ID: id, Action: action, ExpectedVersion: flags.Version, Body: flags.Body, Summary: flags.Summary, Assignee: flags.Assignee, ParentID: flags.ParentID, Evidence: flags.Evidence, Content: flags.Content, Target: flags.Target, MessageID: flags.MessageID}
			if _, ok := rpc.RequestFromContext(ctx); ok {
				if err := decode(ctx, nil, &req); err != nil {
					return comments.Comment{}, StatusError(err)
				}
			}
			req.ID = id
			req.Action = action
			if action == "edit" {
				req.Action = "update"
			}
			v, err := service.Act(ctx, req)
			return v, StatusError(err)
		}).WithMethod("POST"))
	}
	registerResources(builder, resolve)
	builder.Register()
	return nil
}
func decode(ctx context.Context, body map[string]any, out any) error {
	if r, ok := rpc.RequestFromContext(ctx); ok {
		decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(out); err != nil {
			return fmt.Errorf("%w: %v", comments.ErrInvalid, err)
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			return fmt.Errorf("%w: exactly one JSON value required", comments.ErrInvalid)
		}
		return nil
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}
func StatusError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, comments.ErrForbidden):
		return entity.NewStatusError(403, "forbidden", err.Error())
	case errors.Is(err, comments.ErrNotFound):
		return entity.NewStatusError(404, "not_found", err.Error())
	case errors.Is(err, comments.ErrConflict):
		return entity.NewStatusError(409, "conflict", err.Error())
	case errors.Is(err, comments.ErrInvalid):
		return entity.NewStatusError(400, "invalid_input", err.Error())
	default:
		return err
	}
}
