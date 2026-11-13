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

// deliveryPaths are the agent delivery operations RegisterDeliveryRoutes mounts.
func deliveryPaths(id map[string]any, parameter func(name, kind string, required bool) map[string]any) map[string]any {
	str := map[string]string{"type": "string"}
	target := map[string]any{"type": "object", "required": []string{"kind"}, "description": `{kind:"session", sessionId} or {kind:"run", prompt, runOptions?, groupKey?, fresh?}`, "properties": map[string]any{"kind": map[string]any{"type": "string", "enum": []string{"session", "run"}}, "sessionId": str, "prompt": str, "runOptions": map[string]string{"type": "object"}, "groupKey": str, "fresh": map[string]string{"type": "boolean", "description": "Run targets only: always start a new session instead of reusing groupKey's; it becomes the group's session"}}}
	body := func(schema map[string]any) map[string]any {
		return map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": schema}}}
	}
	ok := func(description string) map[string]any {
		return map[string]any{"200": map[string]any{"description": description}, "501": map[string]any{"description": "Agent delivery is not enabled on this host"}}
	}
	deliverResponses := ok("Delivered; the comment with its agentRun")
	deliverResponses["409"] = map[string]any{"description": "Stale expectedVersion, or a delivery is already active"}
	deliverResponses["502"] = map[string]any{"description": "Delivery ran and failed; the failure is recorded in the comment's agentRun"}
	agentResponses := map[string]any{
		"200": map[string]any{"description": "{settings, agent}: the target's settings layers and positions, and its resolved comment agent with each field traced to the layer that set it"},
		"400": map[string]any{"description": "The target is not one the host's settings cover"},
		"501": map[string]any{"description": "Comment agent settings are not enabled on this host"},
	}
	saveResponses := map[string]any{"403": map[string]any{"description": "Only a person configures comment agents"}}
	for status, response := range agentResponses {
		saveResponses[status] = response
	}
	saveBody := map[string]any{"type": "object", "required": []string{"target", "layer", "agent"}, "properties": map[string]any{
		"target": map[string]any{"type": "object", "required": []string{"scopeUri", "uri"}, "properties": map[string]any{"scopeUri": str, "uri": str}},
		"layer":  map[string]any{"type": "object", "required": []string{"kind"}, "description": "The settings layer to save at: {kind: user|parent|project|path, path?, file?} (no file saves the database layer)", "properties": map[string]any{"kind": str, "path": str, "file": str}},
		"page":   map[string]any{"type": "boolean", "description": "Save the layer's entry for the target instead of its default"},
		"agent":  map[string]any{"type": []string{"object", "null"}, "description": "The comments.agents entry to save; null removes it"},
	}}
	batchBody := map[string]any{"type": "object", "required": []string{"requestId", "target", "threads"}, "properties": map[string]any{
		"requestId": str, "target": target,
		"threads": map[string]any{"type": "array", "minItems": 1, "items": map[string]any{"type": "object", "required": []string{"id", "expectedVersion"}, "properties": map[string]any{"id": str, "expectedVersion": map[string]string{"type": "integer", "format": "int64"}}}},
	}}
	batchResponses := ok("Delivered; the threads' comments in request order, each with the same agentRun")
	batchResponses["400"] = map[string]any{"description": "No threads, a thread twice, a closed thread, or threads of different scopes; nothing is delivered"}
	batchResponses["409"] = map[string]any{"description": "A thread's expectedVersion is stale, or a delivery is already active; nothing is delivered"}
	batchResponses["502"] = deliverResponses["502"]
	return map[string]any{
		"/comments/agent": map[string]any{
			"get": map[string]any{"operationId": "comments-agent", "summary": "Read the comment agent a target's settings configure", "parameters": []any{parameter("scope", "string", true), parameter("uri", "string", true)}, "responses": agentResponses},
			"put": map[string]any{"operationId": "comments-save-agent", "summary": "Save or remove the comment agent at one settings layer of a target", "requestBody": body(saveBody), "responses": saveResponses},
		},
		"/comments/agent-sessions":       map[string]any{"get": map[string]any{"operationId": "comments-agent-sessions", "summary": "List the agent sessions a thread can be delivered to: live ones, and with a scope its managed sessions, ended ones included", "parameters": []any{parameter("scope", "string", false), parameter("all", "string", false)}, "responses": sessionsResponses(ok)}},
		"/comments/deliver":              map[string]any{"post": map[string]any{"operationId": "comments-deliver-batch", "summary": "Deliver several threads of one scope to one target in a single message", "requestBody": body(batchBody), "responses": batchResponses}},
		"/comments/{id}/deliver":         map[string]any{"post": map[string]any{"operationId": "comments-deliver", "summary": "Deliver a thread to a live agent session or a new managed run", "parameters": []any{id}, "requestBody": body(map[string]any{"type": "object", "required": []string{"expectedVersion", "requestId", "target"}, "properties": map[string]any{"expectedVersion": map[string]string{"type": "integer", "format": "int64"}, "requestId": str, "target": target}}), "responses": deliverResponses}},
		"/comments/{id}/deliver/preview": map[string]any{"post": map[string]any{"operationId": "comments-deliver-preview", "summary": "Render what a run target would start for a thread", "parameters": []any{id}, "requestBody": body(map[string]any{"type": "object", "required": []string{"target"}, "properties": map[string]any{"target": target}}), "responses": ok("Rendered run preview")}},
	}
}

// sessionsResponses describe GET /comments/agent-sessions: comments.AgentSession items.
func sessionsResponses(ok func(string) map[string]any) map[string]any {
	str, flag := map[string]string{"type": "string"}, func(description string) map[string]string {
		return map[string]string{"type": "boolean", "description": description}
	}
	session := map[string]any{"type": "object", "required": []string{"id", "source", "title", "inCmux"}, "properties": map[string]any{
		"id": str, "source": str, "title": str, "model": str, "cwd": str, "inCmux": map[string]string{"type": "boolean"},
		"status":    map[string]string{"type": "string", "description": `The live session's status, or "ended" for a managed session that is no longer live`},
		"managed":   flag("A session the host launched for the scope's threads"),
		"resumable": flag("An ended managed session: delivering to it resumes it with its recorded prompt and run options"),
		"context":   map[string]any{"type": "object", "properties": map[string]any{"usedTokens": map[string]string{"type": "integer"}, "windowTokens": map[string]string{"type": "integer"}}},
	}}
	responses := ok("Agent sessions")
	responses["200"] = map[string]any{"description": "Agent sessions", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "array", "items": session}}}}
	return responses
}

type documentResponse struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (r *documentResponse) Header() http.Header            { return r.header }
func (r *documentResponse) WriteHeader(status int)         { r.status = status }
func (r *documentResponse) Write(data []byte) (int, error) { return r.body.Write(data) }
