package ui

func addWorkspaceOpenAPI(document map[string]any) error {
	schemas := document["components"].(map[string]any)["schemas"].(map[string]any)
	name, err := todoOpenAPISchema(&workspaceSnapshot{}, schemas)
	if err != nil {
		return err
	}
	document["paths"].(map[string]any)["/api/todos/changes"] = map[string]any{"get": map[string]any{
		"operationId": "todos_changes",
		"summary":     "Follow committed workspace generations",
		"tags":        []string{"todo"},
		"description": "SSE snapshots of authorized workspace identities and separate TODO, label and configuration generations. Subscribe before reading data. A changed epoch requires resynchronization. An error reports unavailable change tracking; clients do not fall back to polling. Also available as a logical subscription through /api/events.",
		"responses": map[string]any{
			"200": map[string]any{"description": "Initial snapshot and committed changes; unchanged in-memory snapshots may be sent as heartbeats", "content": map[string]any{"text/event-stream": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/" + name}}}},
			"503": map[string]any{"description": "Workspace services have not started"},
		},
	}}
	return nil
}
