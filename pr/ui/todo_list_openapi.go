package ui

import "github.com/flanksource/gavel/filters"

// addTodoListOpenAPI documents GET /api/todos: one workspace's todos with ?dir=, or
// the flat list of every workspace's todos but the excluded ones with ?exclude=.
func addTodoListOpenAPI(document map[string]any) {
	query := func(name, description string, schema map[string]any) map[string]any {
		return map[string]any{"name": name, "in": "query", "required": false, "description": description, "schema": schema}
	}
	item := map[string]any{"type": "object", "required": []string{"ref", "title", "status", "priority"}, "properties": map[string]any{
		"ref": map[string]any{"type": "string"}, "id": map[string]any{"type": "string"}, "title": map[string]any{"type": "string"},
		"status": map[string]any{"type": "string"}, "priority": map[string]any{"type": "string"}, "workspaceId": map[string]any{"type": "string"},
		"dir":  map[string]any{"type": "string", "description": "The todo's workspace as GET /api/projects names it; set on the ?exclude= list"},
		"diff": map[string]any{"type": "object", "description": "The latest run's prepared diff footprint; omitted until it has been compared"},
	}}
	document["paths"].(map[string]any)["/api/todos"] = map[string]any{"get": map[string]any{
		"operationId": "todos_list",
		"summary":     "List todos",
		"tags":        []string{"todo"},
		"description": "With dir, the todos of that workspace (the server's work dir when neither dir nor exclude is given). Without dir and with exclude, one flat list of the todos of every workspace in the organization except those the excluded filters hold, read from the database only: no workspace is prepared and no git range is compared.",
		"parameters": []any{
			query("dir", "The workspace: an absolute directory locally, repo:<id> hosted", map[string]any{"type": "string"}),
			query("exclude", "A workspace filter whose workspaces to leave out, e.g. "+filters.HiddenProjects+"; repeatable, read only without dir", map[string]any{"type": "array", "items": map[string]any{"type": "string"}}),
			query("status", "Only todos with this status", map[string]any{"type": "string"}),
		},
		"responses": map[string]any{
			"200": jsonResponse("The todos and their counts by status", map[string]any{"type": "object", "required": []string{"counts", "items"}, "properties": map[string]any{
				"dir":    map[string]any{"type": "string", "description": "The listed workspace; omitted on the ?exclude= list"},
				"counts": map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "integer"}},
				"items":  map[string]any{"type": "array", "items": item},
			}}),
			"400": map[string]any{"description": "Invalid status, unknown filter, or a filter that does not hold workspaces"},
			"404": map[string]any{"description": "The workspace is not one of the organization's"},
		},
	}}
}
