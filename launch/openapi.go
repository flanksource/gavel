package launch

import (
	"fmt"
	"github.com/flanksource/commons-db/shell"
)

func (s *Service) AddOpenAPI(document map[string]any) error {
	document["openapi"] = "3.1.0"
	paths, ok := document["paths"].(map[string]any)
	if !ok {
		return fmt.Errorf("OpenAPI document has no paths object")
	}
	for _, kind := range []string{"", "test", "lint"} {
		path, suffix := "/api/launch", ""
		if kind != "" {
			path += "/" + kind
			suffix = "_" + kind
		}
		schema, err := Schema(kind, s.options.Schemas)
		if err != nil {
			return err
		}
		paths[path] = map[string]any{
			"get":  map[string]any{"operationId": "launch_schema" + suffix, "summary": "Get the launch request schema", "parameters": []any{map[string]any{"name": "schema", "in": "query", "required": true, "schema": map[string]any{"type": "string"}}}, "responses": map[string]any{"200": map[string]any{"description": "Route-specific JSON Schema"}}},
			"post": map[string]any{"operationId": "launch" + suffix, "summary": "Launch native tests and lint", "parameters": []any{map[string]any{"name": "Idempotency-Key", "in": "header", "required": true, "schema": map[string]any{"type": "string"}}}, "requestBody": map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": schema}}}, "responses": map[string]any{"202": map[string]any{"description": "Durably admitted execution"}, "409": map[string]any{"description": "Version or idempotency conflict"}, "422": map[string]any{"description": "Invalid launch input"}}},
		}
	}
	workspace, err := reflected(shell.Setup{})
	if err != nil {
		return err
	}
	paths["/api/launch/prepare"] = map[string]any{"post": map[string]any{
		"operationId": "launch_prepare", "summary": "Prepare a caller-scoped workspace for discovery and launch",
		"requestBody": map[string]any{"required": true, "content": jsonContent(object(map[string]any{"workspace": workspace}, "workspace"))},
		"responses":   map[string]any{"200": apiResponse("Prepared workspace with preparationId, directory, commit, patchDigest and expires", map[string]any{"type": "object"}), "422": apiResponse("Workspace rejected", nil)},
	}}
	paths["/api/launch/prompts"] = map[string]any{"get": map[string]any{
		"operationId": "launch_prompts", "summary": "Discover native Captain prompts or inspect one version and its input form",
		"parameters": []any{apiParameter("preparationId", "query", true), apiParameter("prompt", "query", false), apiParameter("query", "query", false), apiParameter("source", "query", false)},
		"responses":  map[string]any{"200": apiResponse("Prompt catalog, or prompt and variablesSchema for a prompt reference", map[string]any{"type": "object"}), "404": apiResponse("Unknown caller-scoped preparation", nil), "410": apiResponse("Preparation expired", nil)},
	}}
	paths["/api/runs"] = map[string]any{"get": map[string]any{
		"operationId": "launch_runs", "summary": "List caller-owned runs, newest first; label.<key> filters exact labels",
		"parameters": []any{map[string]any{"name": "limit", "in": "query", "schema": map[string]any{"type": "integer", "minimum": 1, "maximum": 200, "default": 50}}, apiParameter("before", "query", false)},
		"responses":  map[string]any{"200": apiResponse("Runs with metadata, status, native tests and lint results", map[string]any{"type": "object"})},
	}}
	for _, suffix := range []string{"", "/stream", "/stop"} {
		method, id, description, code := "get", "launch_run", "Exact persisted execution snapshot", "200"
		if suffix == "/stream" {
			id, description = "launch_run_stream", "Revisioned SSE snapshots; reconnect receives a full snapshot and terminal runs emit done"
		}
		if suffix == "/stop" {
			method, id, description, code = "post", "launch_run_stop", "Cancel this execution and its native children", "202"
		}
		response := apiResponse(description, map[string]any{"type": "object"})
		if suffix == "/stream" {
			response["content"] = map[string]any{"text/event-stream": map[string]any{"schema": map[string]any{"type": "string"}}}
		}
		paths["/api/runs/{executionId}"+suffix] = map[string]any{method: map[string]any{"operationId": id, "summary": description, "parameters": []any{apiParameter("executionId", "path", true)}, "responses": map[string]any{code: response, "404": apiResponse("Execution not found for this caller", nil)}}}
	}
	return nil
}

func apiParameter(name, location string, required bool) map[string]any {
	return map[string]any{"name": name, "in": location, "required": required, "schema": map[string]any{"type": "string"}}
}

func jsonContent(schema map[string]any) map[string]any {
	return map[string]any{"application/json": map[string]any{"schema": schema}}
}

func apiResponse(description string, schema map[string]any) map[string]any {
	result := map[string]any{"description": description}
	if schema != nil {
		result["content"] = jsonContent(schema)
	}
	return result
}
