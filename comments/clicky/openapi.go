package clicky

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// OpenAPI adds the streaming and binary routes to the generated entity document.
func OpenAPI(next http.Handler, prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := &documentResponse{header: http.Header{}, status: 200}
		next.ServeHTTP(response, r)
		if response.status >= 400 {
			for key, value := range response.header {
				w.Header()[key] = value
			}
			w.WriteHeader(response.status)
			if _, err := w.Write(response.body.Bytes()); err != nil {
				return
			}
			return
		}
		var document map[string]any
		if err := json.Unmarshal(response.body.Bytes(), &document); err != nil {
			writeError(w, fmt.Errorf("read generated comments OpenAPI: %w", err))
			return
		}
		paths, ok := document["paths"].(map[string]any)
		if !ok {
			writeError(w, fmt.Errorf("generated document has no paths"))
			return
		}
		for path, item := range HTTPPaths(prefix) {
			paths[path] = item
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(document); err != nil {
			return
		}
	})
}
func HTTPPaths(prefix string) map[string]any {
	id := map[string]any{"name": "id", "in": "path", "required": true, "schema": map[string]string{"type": "string"}}
	parameter := func(name, kind string, required bool) map[string]any {
		return map[string]any{"name": name, "in": "query", "required": required, "schema": map[string]string{"type": kind}}
	}
	response := func(media string) map[string]any {
		return map[string]any{"200": map[string]any{"description": "Authorized comment response", "content": map[string]any{media: map[string]any{"schema": map[string]string{"type": "string"}}}}}
	}
	paths := map[string]any{
		prefix + "/comments/batch":                         map[string]any{"post": map[string]any{"operationId": "comments-batch", "summary": "Atomically apply versioned comment actions", "requestBody": map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "array", "minItems": 1, "maxItems": 1000, "items": map[string]any{"type": "object", "required": []string{"id", "expectedVersion", "action"}, "properties": map[string]any{"id": map[string]string{"type": "string"}, "expectedVersion": map[string]string{"type": "integer", "format": "int64"}, "action": map[string]string{"type": "string"}, "bodyMarkdown": map[string]string{"type": "string"}, "messageId": map[string]string{"type": "string"}, "parentId": map[string]string{"type": "string"}, "content": map[string]any{}, "target": map[string]string{"type": "object"}, "assignee": map[string]string{"type": "object"}, "summary": map[string]string{"type": "string"}, "evidence": map[string]any{"type": "array", "items": map[string]string{"type": "string"}}}}}}}}, "responses": map[string]any{"200": map[string]any{"description": "Updated comment aggregates in action order"}, "409": map[string]any{"description": "Stale version; no changes committed"}}}},
		prefix + "/comments/events":                        map[string]any{"get": map[string]any{"operationId": "comments-events", "summary": "Follow commit-ordered comment events", "parameters": []any{parameter("scope", "string", true), parameter("after", "integer", false), map[string]any{"name": "Last-Event-ID", "in": "header", "schema": map[string]string{"type": "integer"}}}, "responses": response("text/event-stream")}},
		prefix + "/comments/{id}/attachments":              map[string]any{"post": map[string]any{"operationId": "comments-upload", "summary": "Attach bounded binary evidence", "parameters": []any{id, parameter("name", "string", true), parameter("expectedVersion", "integer", true), parameter("messageId", "string", false)}, "requestBody": map[string]any{"required": true, "content": map[string]any{"application/octet-stream": map[string]any{"schema": map[string]string{"type": "string", "format": "binary"}}}}, "responses": map[string]any{"200": map[string]any{"description": "Updated comment and attachment metadata"}, "409": map[string]any{"description": "Thread is final, or expectedVersion is newer than the thread; changes since expectedVersion (an agent delivery claiming the thread) do not conflict"}}}},
		prefix + "/comments/{id}/attachments/{attachment}": map[string]any{"get": map[string]any{"operationId": "comments-download", "summary": "Download authorized attachment bytes", "parameters": []any{id, map[string]any{"name": "attachment", "in": "path", "required": true, "schema": map[string]string{"type": "string"}}}, "responses": response("application/octet-stream")}},
	}
	for path, item := range deliveryPaths(id, parameter) {
		paths[prefix+path] = item
	}
	return paths
}

// deliveryPaths are the agent delivery and route operations RegisterDeliveryRoutes mounts.
func deliveryPaths(id map[string]any, parameter func(name, kind string, required bool) map[string]any) map[string]any {
	str := map[string]string{"type": "string"}
	target := map[string]any{"type": "object", "required": []string{"kind"}, "description": `{kind:"session", sessionId} or {kind:"run", prompt, runOptions?, groupKey?, fresh?}`, "properties": map[string]any{"kind": map[string]any{"type": "string", "enum": []string{"session", "run"}}, "sessionId": str, "prompt": str, "runOptions": map[string]string{"type": "object"}, "groupKey": str, "fresh": map[string]string{"type": "boolean", "description": "Run targets only: always start a new session instead of reusing groupKey's; it becomes the group's session. On a route, every routed comment starts its own session"}}}
	route := map[string]any{"type": "object", "required": []string{"scopeUri", "targetPrefix", "workflowId", "target", "enabled"}, "properties": map[string]any{"id": str, "scopeUri": str, "targetPrefix": map[string]string{"type": "string", "description": "Matches a target URI equal to it or beneath it at a / boundary"}, "workflowId": str, "target": target, "enabled": map[string]string{"type": "boolean"}}}
	body := func(schema map[string]any) map[string]any {
		return map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": schema}}}
	}
	ok := func(description string) map[string]any {
		return map[string]any{"200": map[string]any{"description": description}, "501": map[string]any{"description": "Agent delivery is not enabled on this host"}}
	}
	deliverResponses := ok("Delivered; the comment with its agentRun")
	deliverResponses["409"] = map[string]any{"description": "Stale expectedVersion, or a delivery is already active"}
	deliverResponses["502"] = map[string]any{"description": "Delivery ran and failed; the failure is recorded in the comment's agentRun"}
	routeID := map[string]any{"name": "routeId", "in": "path", "required": true, "schema": str}
	return map[string]any{
		"/comments/agent-sessions":       map[string]any{"get": map[string]any{"operationId": "comments-agent-sessions", "summary": "List live agent sessions a thread can be delivered to", "parameters": []any{parameter("scope", "string", false), parameter("all", "string", false)}, "responses": ok("Live agent sessions")}},
		"/comments/{id}/deliver":         map[string]any{"post": map[string]any{"operationId": "comments-deliver", "summary": "Deliver a thread to a live agent session or a new managed run", "parameters": []any{id}, "requestBody": body(map[string]any{"type": "object", "required": []string{"expectedVersion", "requestId", "target"}, "properties": map[string]any{"expectedVersion": map[string]string{"type": "integer", "format": "int64"}, "requestId": str, "target": target}}), "responses": deliverResponses}},
		"/comments/{id}/deliver/preview": map[string]any{"post": map[string]any{"operationId": "comments-deliver-preview", "summary": "Render what a run target would start for a thread", "parameters": []any{id}, "requestBody": body(map[string]any{"type": "object", "required": []string{"target"}, "properties": map[string]any{"target": target}}), "responses": ok("Rendered run preview")}},
		"/comments/routes": map[string]any{
			"get": map[string]any{"operationId": "comments-routes", "summary": "List the scope's automatic delivery routes", "parameters": []any{parameter("scope", "string", true)}, "responses": ok("Routes")},
			"put": map[string]any{"operationId": "comments-set-route", "summary": "Create (no id) or replace an automatic delivery route", "requestBody": body(route), "responses": ok("Saved route")},
		},
		"/comments/routes/{routeId}": map[string]any{"delete": map[string]any{"operationId": "comments-delete-route", "summary": "Delete an automatic delivery route", "parameters": []any{routeID}, "responses": map[string]any{"204": map[string]any{"description": "Deleted"}, "501": map[string]any{"description": "Comment routes are not enabled on this host"}}}},
	}
}

type documentResponse struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (r *documentResponse) Header() http.Header            { return r.header }
func (r *documentResponse) WriteHeader(status int)         { r.status = status }
func (r *documentResponse) Write(data []byte) (int, error) { return r.body.Write(data) }
