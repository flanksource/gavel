package ui

import (
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/todos"
)

// prLinkSchemas names the component schemas the PR link routes reference.
type prLinkSchemas map[string]string

func (s prLinkSchemas) ref(name string) map[string]any {
	return map[string]any{"$ref": "#/components/schemas/" + s[name]}
}

func (s prLinkSchemas) array(name string) map[string]any {
	return map[string]any{"type": "array", "items": s.ref(name)}
}

func addTodoPRLinksOpenAPI(document map[string]any) error {
	components := document["components"].(map[string]any)["schemas"].(map[string]any)
	schemas := prLinkSchemas{}
	for name, value := range map[string]any{
		"link": &todoPRLinkItem{}, "payload": &todoPRLinkPayload{}, "todo": &todos.PRLinkedTODO{}, "pr": &cache.PullRequest{},
	} {
		ref, err := todoOpenAPISchema(value, components)
		if err != nil {
			return err
		}
		schemas[name] = ref
	}
	paths := document["paths"].(map[string]any)
	paths["/api/todos/pr-links"] = todoPRLinksOpenAPIPath(schemas)
	paths["/api/prs/todos"] = map[string]any{"get": map[string]any{
		"operationId": "prs_todos", "summary": "List the workspace's TODOs that link a PR", "tags": []string{"pr"},
		"parameters": []any{queryParam("url", "PR URL", true), prLinkDirParam()},
		"responses": map[string]any{
			"200": jsonResponse("Linked TODOs", schemas.array("todo")),
			"400": map[string]any{"description": "Not a hosted PR URL"},
		},
	}}
	paths["/api/prs/search"] = map[string]any{"get": map[string]any{
		"operationId": "prs_search", "summary": "Search the persisted PR index; an empty q lists the most recently updated PRs", "tags": []string{"pr"},
		"parameters": []any{
			queryParam("q", "Title words (each a prefix) or a PR number", false),
			queryParam("repo", "owner/name, case-insensitive", false),
			queryParam("state", "open, closed, merged or all", false),
			queryParam("limit", "1-200, default 20", false),
		},
		"responses": map[string]any{
			"200": jsonResponse("Matching PRs, most recently updated first", schemas.array("pr")),
			"400": map[string]any{"description": "Invalid state, limit or text"},
			"503": map[string]any{"description": "PR index unavailable"},
		},
	}}
	return nil
}

func todoPRLinksOpenAPIPath(schemas prLinkSchemas) map[string]any {
	id := queryParam("id", "TODO reference (id, short id or alias)", true)
	return map[string]any{
		"get": map[string]any{
			"operationId": "todos_pr_links", "summary": "List a TODO's linked PRs, joined with the PR index for title and state", "tags": []string{"todos"},
			"parameters": []any{prLinkDirParam(), id},
			"responses": map[string]any{
				"200": jsonResponse("Links, the primary (implements) link first", schemas.array("link")),
				"400": map[string]any{"description": "Missing id or unresolvable workspace"},
				"404": map[string]any{"description": "Unknown TODO"},
				"503": map[string]any{"description": "PR index unavailable"},
			},
		},
		"post": map[string]any{
			"operationId": "todos_pr_link", "summary": "Link a hosted PR to a TODO (relation defaults to related; source is manual)", "tags": []string{"todos"},
			"requestBody": jsonRequestBody(schemas["payload"])["requestBody"],
			"responses": map[string]any{
				"201": jsonResponse("The created link", schemas.ref("link")),
				"400": map[string]any{"description": "Not a hosted PR URL, unknown relation, or missing id"},
				"404": map[string]any{"description": "Unknown TODO"},
				"409": map[string]any{"description": "The TODO already links this PR"},
			},
		},
		"delete": map[string]any{
			"operationId": "todos_pr_unlink", "summary": "Unlink a PR; unlinking the primary PR also clears the TODO's pr", "tags": []string{"todos"},
			"parameters": []any{prLinkDirParam(), id, queryParam("url", "PR URL", true)},
			"responses": map[string]any{
				"204": map[string]any{"description": "Unlinked"},
				"400": map[string]any{"description": "Not a hosted PR URL or missing id"},
				"404": map[string]any{"description": "Unknown TODO or PR link"},
			},
		},
	}
}

func queryParam(name, description string, required bool) map[string]any {
	return map[string]any{"name": name, "in": "query", "required": required, "description": description, "schema": map[string]any{"type": "string"}}
}

func prLinkDirParam() map[string]any {
	return queryParam("dir", "Workspace directory; defaults to the server's", false)
}

func jsonResponse(description string, schema map[string]any) map[string]any {
	return map[string]any{"description": description, "content": map[string]any{"application/json": map[string]any{"schema": schema}}}
}
