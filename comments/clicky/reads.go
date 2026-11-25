package clicky

import (
	"context"
	"net/http"

	"github.com/flanksource/clicky/route"
	"github.com/flanksource/gavel/comments"
)

// RegisterReadRoutes mounts the page read state: POST /api/v1/comments/read
// marks every thread on one target read for the caller.
func RegisterReadRoutes(router *route.Router, resolve Resolver) {
	router.Raw("POST /api/v1/comments/read", serve(resolve, markTargetRead), route.Meta{Entity: "comments", Verb: "read-target"})
}

func markTargetRead(ctx context.Context, service *comments.Service, _ *http.Request) (any, error) {
	var req struct {
		Target comments.ReadTarget `json:"target"`
	}
	if err := decode(ctx, bodyLimit, &req); err != nil {
		return nil, err
	}
	marked, err := service.MarkTargetRead(ctx, req.Target)
	if err != nil {
		return nil, err
	}
	return map[string]int{"marked": marked}, nil
}

// readPaths describe RegisterReadRoutes for the OpenAPI document.
func readPaths() map[string]any {
	str := map[string]string{"type": "string"}
	body := map[string]any{"type": "object", "required": []string{"target"}, "properties": map[string]any{
		"target": map[string]any{"type": "object", "required": []string{"scopeUri", "uri"}, "properties": map[string]any{"scopeUri": str, "uri": str}},
	}}
	marked := map[string]any{"type": "object", "required": []string{"marked"}, "properties": map[string]any{"marked": map[string]string{"type": "integer"}}}
	return map[string]any{"/comments/read": map[string]any{"post": map[string]any{
		"operationId": "comments-read-target",
		"summary":     "Mark every thread on a target read for the caller at its current version",
		"requestBody": map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": body}}},
		"responses": map[string]any{
			"200": map[string]any{"description": "How many threads were marked; ones already read are not counted", "content": map[string]any{"application/json": map[string]any{"schema": marked}}},
			"400": map[string]any{"description": "The target has no scopeUri or uri"},
		},
	}}}
}
