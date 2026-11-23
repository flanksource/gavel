// Package clicky registers the same comments service for CLI, HTTP and agent tools.
package clicky

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	framework "github.com/flanksource/clicky"
	"github.com/flanksource/clicky/entity"
	"github.com/flanksource/clicky/rpc"
	"github.com/flanksource/gavel/comments"
)

type actionFlags struct {
	Version   int64               `flag:"expectedVersion" required:"true" help:"Required current comment version"`
	Body      string              `flag:"bodyMarkdown" help:"Reply, edited Markdown or feedback saved with reopen; a reply naming @agent, @<runtime> or @<session> is delivered to it"`
	Summary   string              `flag:"summary" help:"Resolution or closure summary"`
	ParentID  string              `flag:"parentId" help:"Local parent reply identity"`
	Assignee  *comments.Actor     `flag:"assignee" help:"Assignee as JSON; null clears it"`
	Evidence  []string            `flag:"evidence" help:"Outcome evidence URIs"`
	Content   json.RawMessage     `flag:"content" help:"Structured reply content as JSON"`
	Target    *comments.Target    `flag:"target" help:"Replacement target within the same scope"`
	MessageID string              `flag:"messageId" help:"Reply to edit or delete"`
	Patch     string              `flag:"patch" help:"Unified diff file to propose ('-' reads stdin)"`
	Edits     []comments.FileEdit `flag:"edits" help:"Edits as JSON: [{path, oldString, newString, replaceAll?} | {path, content} | {path, delete:true}]"`
	Mode      string              `flag:"mode" help:"apply or stage (accept-proposal)"`
}

func (actionFlags) ClickyActionFlags() {}

type createFlags struct {
	Target    comments.Target         `flag:"target" required:"true" help:"Captured host target as JSON"`
	Body      string                  `flag:"bodyMarkdown" required:"true" help:"Comment Markdown; naming @agent, @<runtime> or @<session> delivers it unless --deliver is given"`
	Workflow  string                  `flag:"workflowId" required:"true" help:"Registered workflow ID"`
	RequestID string                  `flag:"requestId" required:"true" help:"Stable request identity"`
	Content   json.RawMessage         `flag:"content" help:"Structured comment content"`
	Links     []comments.Link         `flag:"links" help:"Typed resource links as JSON"`
	Deliver   *comments.DeliverTarget `flag:"deliver" help:"Delivery target as JSON; replaces the page's comment agent route"`
}

func (createFlags) ClickyActionFlags() {}

// Register must be called once per command registry. Authentication remains request scoped.
type Resolver func(context.Context) (*comments.Service, error)
type listItem struct {
	comments.Comment
	ReadVersion int64 `json:"readVersion"`
	Unread      bool  `json:"unread"`
	// Anchor is where the thread's patch anchor lies in its target now:
	// matched (with its lines), ambiguous or outdated. Only a thread whose
	// selector carries a patch, on a target the host has content for, has one.
	Anchor *comments.PatchLocation `json:"anchor,omitempty"`
}

// listPage is one page of threads as GET /api/v1/comments lists them.
func listPage(ctx context.Context, service *comments.Service, f comments.Filter) (entity.PagedResult[listItem], error) {
	page, err := service.Page(ctx, f)
	if err != nil {
		return entity.PagedResult[listItem]{}, StatusError(err)
	}
	anchors, err := service.Anchors(ctx, page.Items)
	if err != nil {
		return entity.PagedResult[listItem]{}, StatusError(err)
	}
	reads := map[string]int64{}
	for _, r := range page.Read {
		reads[r.CommentID] = r.Version
	}
	rows := make([]listItem, 0, len(page.Items))
	for _, c := range page.Items {
		item := listItem{Comment: c, ReadVersion: reads[c.ID], Unread: reads[c.ID] < c.Version}
		if anchor, ok := anchors[c.ID]; ok {
			item.Anchor = &anchor
		}
		rows = append(rows, item)
	}
	return entity.NewPagedResult(rows, f.Limit, f.Offset, page.Total), nil
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
			return listPage(ctx, service, f)
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
		req := comments.CreateRequest{Target: flags.Target, Body: flags.Body, WorkflowID: flags.Workflow, RequestID: flags.RequestID, Content: flags.Content, Links: flags.Links, Deliver: flags.Deliver}
		if _, ok := rpc.RequestFromContext(ctx); ok {
			if err := decode(ctx, createBodyLimit(service), &req); err != nil {
				return comments.Comment{}, StatusError(err)
			}
		}
		v, err := service.Create(ctx, req)
		return v, StatusError(err)
	}).WithOptionalID().WithMethod("POST").WithPath("comments"))
	for _, action := range []string{"reply", "assign", "acknowledge", "edit", "resolve", "close", "reopen", "metadata", "retarget", "delete", "restore", "edit-reply", "delete-reply", comments.ActionAcceptProposal, comments.ActionRejectProposal, comments.ActionRequestRevision} {
		builder = builder.WithAction(entity.TypedActionWithContext(action, actionFlags{}, func(ctx context.Context, id string, flags actionFlags) (comments.Comment, error) {
			service, err := resolve(ctx)
			if err != nil {
				return comments.Comment{}, StatusError(err)
			}
			req := comments.ActionRequest{ID: id, Action: action, ExpectedVersion: flags.Version, Body: flags.Body, Summary: flags.Summary, Assignee: flags.Assignee, ParentID: flags.ParentID, Evidence: flags.Evidence, Content: flags.Content, Target: flags.Target, MessageID: flags.MessageID, Edits: flags.Edits, Mode: flags.Mode}
			if _, ok := rpc.RequestFromContext(ctx); ok {
				if err := decode(ctx, bodyLimit, &req); err != nil {
					return comments.Comment{}, StatusError(err)
				}
			} else if req.Patch, err = readPatch(flags.Patch); err != nil {
				return comments.Comment{}, StatusError(err)
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

// readPatch reads the CLI's --patch file, or stdin for "-"; over HTTP the
// body's patch is the diff text itself.
func readPatch(path string) (string, error) {
	var data []byte
	var err error
	switch path {
	case "":
		return "", nil
	case "-":
		data, err = io.ReadAll(os.Stdin)
	default:
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return "", fmt.Errorf("%w: read --patch %s: %v", comments.ErrInvalid, path, err)
	}
	return string(data), nil
}

// bodyLimit bounds a JSON request body; a create's also carries its attachments.
const bodyLimit = 1 << 20

// createBodyLimit is a create's body bound: bodyLimit plus the base64 of the
// most attachments the service takes at its attachment limit.
func createBodyLimit(service *comments.Service) int64 {
	return int64(bodyLimit + comments.MaxCreateAttachments*base64.StdEncoding.EncodedLen(int(service.AttachmentLimit())))
}

// decode reads the request's JSON body, at most limit bytes, strictly into out.
func decode(ctx context.Context, limit int64, out any) error {
	r, ok := rpc.RequestFromContext(ctx)
	if !ok {
		return fmt.Errorf("decoding a comments request body needs the HTTP request in its context")
	}
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, limit))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("%w: %v", comments.ErrInvalid, err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("%w: exactly one JSON value required", comments.ErrInvalid)
	}
	return nil
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
