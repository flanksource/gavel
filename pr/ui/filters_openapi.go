package ui

import "github.com/flanksource/gavel/filters"

// addFiltersOpenAPI documents /api/filters, the caller's named filters that list
// endpoints reference with ?exclude=<name>.
func addFiltersOpenAPI(document map[string]any) {
	filter := map[string]any{"type": "object", "required": []string{"name", "kind", "values"}, "properties": map[string]any{
		"name":   map[string]any{"type": "string", "enum": []string{filters.HiddenProjects, filters.HiddenGitHubOrgs}},
		"kind":   map[string]any{"type": "string", "enum": []string{filters.KindWorkspace, filters.KindGitHubOrg}},
		"values": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Sorted. Workspace roots (an absolute directory locally, repo:<id> hosted) or GitHub org logins"},
	}}
	name := []any{map[string]any{"name": "name", "in": "path", "required": true, "schema": filter["properties"].(map[string]any)["name"]}}
	unknown := map[string]any{"description": "Unknown filter name or invalid value"}
	value := map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{
		"type": "object", "required": []string{"value"}, "properties": map[string]any{"value": map[string]any{"type": "string"}},
	}}}}
	edit := func(summary string) map[string]any {
		return map[string]any{"summary": summary, "tags": []string{"filters"}, "parameters": name, "requestBody": value,
			"responses": map[string]any{"200": jsonResponse("The updated filter", filter), "400": unknown}}
	}
	paths := document["paths"].(map[string]any)
	paths["/api/filters"] = map[string]any{"get": map[string]any{"summary": "The caller's named filters", "tags": []string{"filters"},
		"responses": map[string]any{"200": jsonResponse("Every well-known filter, empty when unset", map[string]any{
			"type": "object", "properties": map[string]any{"filters": map[string]any{"type": "array", "items": filter}},
		})}}}
	paths["/api/filters/{name}"] = map[string]any{"get": map[string]any{"summary": "One of the caller's named filters", "tags": []string{"filters"},
		"parameters": name, "responses": map[string]any{"200": jsonResponse("The filter, empty when unset", filter), "400": unknown}}}
	paths["/api/filters/{name}/values"] = map[string]any{
		"post":   edit("Add a value to the filter; adding one it holds changes nothing"),
		"delete": edit("Remove a value from the filter; removing one it lacks changes nothing"),
	}
}
