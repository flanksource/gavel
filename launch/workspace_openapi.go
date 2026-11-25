package launch

import (
	"github.com/flanksource/gavel/workspaceaccess"
	"github.com/flanksource/gavel/workspacefiles"
	"strings"
)

func workspaceOpenAPI(paths map[string]any) error {
	denied, err := reflected(workspaceaccess.Denied{})
	if err != nil {
		return err
	}
	for _, route := range []struct {
		path, method, id, summary string
		input, output             any
	}{
		{"/api/workspace/access", "get", "workspace_access", "List effective local workspace access", nil, workspaceaccess.Snapshot{}},
		{"/api/workspace/access", "post", "workspace_access_grant", "Grant reviewed persistent workspace access", workspaceaccess.GrantRequest{}, workspaceaccess.Root{}},
		{"/api/workspace/access/preview", "post", "workspace_access_preview", "Review canonical local folder access", workspaceaccess.PreviewRequest{}, workspaceaccess.PreviewRequest{}},
		{"/api/workspace/access/{id}", "delete", "workspace_access_revoke", "Revoke saved workspace access", nil, nil},
		{"/api/workspace/access/events", "get", "workspace_access_events", "Follow workspace access changes", nil, nil},
		{"/api/workspace/directories", "get", "workspace_directories", "List immediate directories inside enabled local roots", nil, []workspacefiles.DirectoryEntry{}},
		{"/api/workspace/repository", "get", "workspace_repository", "Validate an enabled Git working-tree root", nil, workspacefiles.Repository{}},
		{"/api/workspace/entries", "post", "workspace_entries", "List immediate Markdown children with ordered ignore overrides", workspacefiles.EntriesRequest{}, []workspacefiles.DirectoryEntry{}},
		{"/api/workspace/file", "get", "workspace_file", "Read UTF-8 source and content revision", nil, workspacefiles.File{}},
		{"/api/workspace/file", "put", "workspace_save", "Atomically save with the expected content revision", workspacefiles.SaveRequest{}, workspacefiles.File{}},
		{"/api/workspace/asset", "get", "workspace_asset", "Read a confined image asset", nil, nil},
		{"/api/workspace/blame", "get", "workspace_blame", "Read provenance independently of untracked or unborn document source", nil, workspacefiles.BlameFile{}},
		{"/api/workspace/watch", "post", "workspace_watch", "Acquire an independent open-file lease; default idle timeout 30 seconds", workspacefiles.FileRequest{}, workspacefiles.Lease{}},
		{"/api/workspace/watch/{id}", "post", "workspace_watch_renew", "Renew on user activity only", nil, workspacefiles.Lease{}},
		{"/api/workspace/watch/{id}", "delete", "workspace_watch_release", "Release a viewer lease", nil, nil},
		{"/api/workspace/watch/{id}/stream", "get", "workspace_watch_stream", "SSE file snapshots and expiration; events never renew leases", nil, nil},
	} {
		operation := map[string]any{"operationId": route.id, "summary": route.summary}
		if route.input != nil {
			schema, err := reflected(route.input)
			if err != nil {
				return err
			}
			operation["requestBody"] = map[string]any{"required": true, "content": jsonContent(schema)}
		}
		var schema map[string]any
		if route.output != nil {
			var err error
			schema, err = reflected(route.output)
			if err != nil {
				return err
			}
		}
		operation["responses"] = map[string]any{"200": apiResponse(route.summary, schema), "403": apiResponse("Workspace access denied", denied), "404": apiResponse("File not found", nil), "409": apiResponse("Content revision conflict", nil), "410": apiResponse("Watch lease expired", nil), "422": apiResponse("Invalid path or unavailable database", nil)}
		if route.id == "workspace_watch_stream" || route.id == "workspace_access_events" {
			operation["responses"].(map[string]any)["200"] = map[string]any{"description": route.summary, "content": map[string]any{"text/event-stream": map[string]any{"schema": map[string]any{"type": "string"}}}}
		}
		if route.id == "workspace_asset" {
			operation["responses"].(map[string]any)["200"] = map[string]any{"description": route.summary, "content": map[string]any{"image/*": map[string]any{"schema": map[string]any{"type": "string", "format": "binary"}}}}
		}
		parameters := []any{}
		if route.method == "get" && route.path != "/api/workspace/watch/{id}/stream" && !strings.HasPrefix(route.path, "/api/workspace/access") {
			if route.path == "/api/workspace/file" || route.path == "/api/workspace/asset" || route.path == "/api/workspace/blame" {
				parameters = append(parameters, apiParameter("root", "query", true))
			}
			parameters = append(parameters, apiParameter("path", "query", route.path != "/api/workspace/directories"))
		}
		if route.path == "/api/workspace/watch/{id}" || route.path == "/api/workspace/watch/{id}/stream" || route.path == "/api/workspace/access/{id}" {
			parameters = append(parameters, apiParameter("id", "path", true))
		}
		operation["parameters"] = parameters
		if paths[route.path] == nil {
			paths[route.path] = map[string]any{}
		}
		paths[route.path].(map[string]any)[route.method] = operation
	}
	return nil
}
