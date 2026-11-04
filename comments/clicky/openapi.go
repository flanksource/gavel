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
	return map[string]any{
		prefix + "/comments/events":                        map[string]any{"get": map[string]any{"operationId": "comments-events", "summary": "Follow commit-ordered comment events", "parameters": []any{parameter("scope", "string", true), parameter("after", "integer", false), map[string]any{"name": "Last-Event-ID", "in": "header", "schema": map[string]string{"type": "integer"}}}, "responses": response("text/event-stream")}},
		prefix + "/comments/{id}/attachments":              map[string]any{"post": map[string]any{"operationId": "comments-upload", "summary": "Attach bounded binary evidence", "parameters": []any{id, parameter("name", "string", true), parameter("expectedVersion", "integer", true), parameter("messageId", "string", false)}, "requestBody": map[string]any{"required": true, "content": map[string]any{"application/octet-stream": map[string]any{"schema": map[string]string{"type": "string", "format": "binary"}}}}, "responses": map[string]any{"200": map[string]any{"description": "Updated comment and attachment metadata"}}}},
		prefix + "/comments/{id}/attachments/{attachment}": map[string]any{"get": map[string]any{"operationId": "comments-download", "summary": "Download authorized attachment bytes", "parameters": []any{id, map[string]any{"name": "attachment", "in": "path", "required": true, "schema": map[string]string{"type": "string"}}}, "responses": response("application/octet-stream")}},
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
