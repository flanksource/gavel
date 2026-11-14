package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/comments"
	commentapi "github.com/flanksource/gavel/comments/clicky"
)

// threadTools are the caller tools every managed session gets, whatever prompt
// it runs: they read, reply to and resolve the threads on the targets of the
// threads it was given, in-process, as the session's agent, so the workflow's
// human-only actions stay refused. In propose mode (edits) the session gets
// propose_change in place of resolve_thread: a person accepts its proposal.
func threadTools(service *comments.Service, given []comments.Comment, agent comments.Actor, edits string) []api.ToolDefinition {
	threads := threadToolset{service: service, agent: agent}
	for _, thread := range given {
		if !slices.ContainsFunc(threads.targets, func(t comments.Target) bool { return t.URI == thread.Target.URI }) {
			threads.targets = append(threads.targets, thread.Target)
		}
	}
	readOnly := true
	tools := []api.ToolDefinition{
		{
			Name:              "list_threads",
			Description:       "List the comment threads on this session's target, with their replies and current versions.",
			ReadOnlyHint:      &readOnly,
			DefaultPermission: api.ToolPolicyAllow,
			Handler:           threads.list,
		},
		{
			Name:              "reply_to_thread",
			Description:       "Reply to a comment thread as this session's agent. Say what you changed and how you verified it.",
			InputSchema:       stringsSchema(map[string]string{"id": "The thread's comment id", "body": "The reply in Markdown"}),
			DefaultPermission: api.ToolPolicyAllow,
			Handler:           threads.reply,
		},
	}
	if edits == comments.EditsPropose {
		return append(tools, api.ToolDefinition{
			Name:              "propose_change",
			Description:       "Reply to a comment thread with a proposed change instead of editing files: gavel commits the edits (or patch) onto HEAD itself. Edits work like the Edit tool: oldString must match the file exactly once, unless replaceAll. A person accepts, rejects or asks for revisions.",
			InputSchema:       proposeSchema(),
			DefaultPermission: api.ToolPolicyAllow,
			Handler:           threads.propose,
		})
	}
	return append(tools, api.ToolDefinition{
		Name:              "resolve_thread",
		Description:       "Resolve a comment thread once its change is made, verified and replied to. A human then reviews it.",
		InputSchema:       stringsSchema(map[string]string{"id": "The thread's comment id", "summary": "One line on what was done"}),
		DefaultPermission: api.ToolPolicyAllow,
		Handler:           threads.resolve,
	})
}

// stringsSchema is a JSON Schema object whose properties are all required strings.
func stringsSchema(properties map[string]string) map[string]any {
	props := map[string]any{}
	for name, description := range properties {
		props[name] = map[string]any{"type": "string", "description": description}
	}
	return map[string]any{"type": "object", "properties": props, "required": slices.Sorted(maps.Keys(properties)), "additionalProperties": false}
}

// proposeSchema is propose_change's input: a thread, a reply body, and edits
// or a unified diff patch.
func proposeSchema() map[string]any {
	schema := stringsSchema(map[string]string{"id": "The thread's comment id", "body": "What the change does and why, in Markdown"})
	props := schema["properties"].(map[string]any)
	props["edits"] = comments.FileEditsSchema()
	props["patch"] = map[string]any{"type": "string", "description": "Instead of edits: a unified diff with a/ and b/ paths relative to the repository root"}
	return schema
}

type threadToolset struct {
	service *comments.Service
	targets []comments.Target
	agent   comments.Actor
}

func (t threadToolset) list(ctx context.Context, _ map[string]any) (any, error) {
	all := []comments.Comment{}
	for _, target := range t.targets {
		threads, err := t.service.List(commentapi.WithActor(ctx, t.agent), comments.Filter{ScopeURI: target.ScopeURI, TargetURI: target.URI, Limit: 200})
		if err != nil {
			return nil, err
		}
		all = append(all, threads...)
	}
	return all, nil
}

func (t threadToolset) reply(ctx context.Context, input map[string]any) (any, error) {
	body, err := requiredString(input, "body")
	if err != nil {
		return nil, err
	}
	return t.act(ctx, input, comments.ActionRequest{Action: "reply", Body: body})
}

func (t threadToolset) propose(ctx context.Context, input map[string]any) (any, error) {
	body, err := requiredString(input, "body")
	if err != nil {
		return nil, err
	}
	req := comments.ActionRequest{Action: "reply", Body: body}
	if raw, ok := input["patch"]; ok {
		if req.Patch, ok = raw.(string); !ok {
			return nil, fmt.Errorf("patch must be a string, not %T", raw)
		}
	}
	if req.Edits, err = fileEdits(input["edits"]); err != nil {
		return nil, err
	}
	if (req.Patch == "") == (len(req.Edits) == 0) {
		return nil, fmt.Errorf("propose_change needs edits or a patch, exactly one of them")
	}
	return t.act(ctx, input, req)
}

// fileEdits decodes the edits a tool call passes, refusing unknown fields.
func fileEdits(raw any) ([]comments.FileEdit, error) {
	if raw == nil {
		return nil, nil
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("edits: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	var edits []comments.FileEdit
	if err := decoder.Decode(&edits); err != nil {
		return nil, fmt.Errorf("edits must be an array of {path, oldString, newString, replaceAll?} | {path, content} | {path, delete}: %w", err)
	}
	return edits, nil
}

func (t threadToolset) resolve(ctx context.Context, input map[string]any) (any, error) {
	summary, err := requiredString(input, "summary")
	if err != nil {
		return nil, err
	}
	return t.act(ctx, input, comments.ActionRequest{Action: "resolve", Summary: summary})
}

// act applies req to the thread input names, at its current version, refusing
// a thread on a target other than those of the threads the session was given.
func (t threadToolset) act(ctx context.Context, input map[string]any, req comments.ActionRequest) (comments.Comment, error) {
	id, err := requiredString(input, "id")
	if err != nil {
		return comments.Comment{}, err
	}
	ctx = commentapi.WithActor(ctx, t.agent)
	thread, err := t.service.Get(ctx, id)
	if err != nil {
		return comments.Comment{}, fmt.Errorf("read thread %s: %w", id, err)
	}
	if !slices.ContainsFunc(t.targets, func(target comments.Target) bool { return target.URI == thread.Target.URI }) {
		return comments.Comment{}, fmt.Errorf("thread %s is on %s, not one of this session's targets", id, thread.Target.URI)
	}
	req.ID, req.ExpectedVersion = id, thread.Version
	updated, err := t.service.Act(ctx, req)
	if err != nil {
		return comments.Comment{}, fmt.Errorf("%s thread %s: %w", req.Action, id, err)
	}
	return updated, nil
}

func requiredString(input map[string]any, key string) (string, error) {
	value, ok := input[key].(string)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s is required and must be a non-empty string", key)
	}
	return value, nil
}
