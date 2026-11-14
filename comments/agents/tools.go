package agents

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/comments"
	commentapi "github.com/flanksource/gavel/comments/clicky"
)

// threadTools are the caller tools every managed session gets, whatever prompt
// it runs: they read, reply to and resolve the threads on the session's target
// in-process, as the session's agent, so the workflow's human-only actions stay refused.
func threadTools(service *comments.Service, target comments.Target, agent comments.Actor) []api.ToolDefinition {
	threads := threadToolset{service: service, target: target, agent: agent}
	readOnly := true
	return []api.ToolDefinition{
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
		{
			Name:              "resolve_thread",
			Description:       "Resolve a comment thread once its change is made, verified and replied to. A human then reviews it.",
			InputSchema:       stringsSchema(map[string]string{"id": "The thread's comment id", "summary": "One line on what was done"}),
			DefaultPermission: api.ToolPolicyAllow,
			Handler:           threads.resolve,
		},
	}
}

// stringsSchema is a JSON Schema object whose properties are all required strings.
func stringsSchema(properties map[string]string) map[string]any {
	props := map[string]any{}
	for name, description := range properties {
		props[name] = map[string]any{"type": "string", "description": description}
	}
	return map[string]any{"type": "object", "properties": props, "required": slices.Sorted(maps.Keys(properties)), "additionalProperties": false}
}

type threadToolset struct {
	service *comments.Service
	target  comments.Target
	agent   comments.Actor
}

func (t threadToolset) list(ctx context.Context, _ map[string]any) (any, error) {
	return t.service.List(commentapi.WithActor(ctx, t.agent), comments.Filter{ScopeURI: t.target.ScopeURI, TargetURI: t.target.URI, Limit: 200})
}

func (t threadToolset) reply(ctx context.Context, input map[string]any) (any, error) {
	body, err := requiredString(input, "body")
	if err != nil {
		return nil, err
	}
	return t.act(ctx, input, comments.ActionRequest{Action: "reply", Body: body})
}

func (t threadToolset) resolve(ctx context.Context, input map[string]any) (any, error) {
	summary, err := requiredString(input, "summary")
	if err != nil {
		return nil, err
	}
	return t.act(ctx, input, comments.ActionRequest{Action: "resolve", Summary: summary})
}

// act applies req to the thread input names, at its current version, refusing
// a thread on any other target than the session's.
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
	if thread.Target.URI != t.target.URI {
		return comments.Comment{}, fmt.Errorf("thread %s is on %s, not this session's target %s", id, thread.Target.URI, t.target.URI)
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
