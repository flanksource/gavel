package ui

func addConnectionsOpenAPI(document map[string]any) {
	paths := document["paths"].(map[string]any)
	paths["/api/connections/storage"] = map[string]any{"get": map[string]any{"summary": "Connection storage layers and KEK availability", "tags": []string{"connections"}, "responses": map[string]any{"200": jsonResponse("Storage destinations and KEK status without key material", map[string]any{"type": "object", "properties": map[string]any{"layers": map[string]any{"type": "array", "items": map[string]any{"type": "object"}}, "defaultLayer": map[string]any{"type": "string"}, "kek": map[string]any{"type": "object", "properties": map[string]any{"state": map[string]any{"type": "string", "enum": []string{"ready", "missing", "unavailable", "not_configured"}}, "detail": map[string]any{"type": "string"}}}}})}}}
	metadata := map[string]any{"type": "object", "description": "Organization connection with credentials redacted"}
	response := map[string]any{"200": jsonResponse("Connection result", metadata), "400": map[string]any{"description": "Invalid connection or failed validation"}, "403": map[string]any{"description": "Organization administrator required"}, "404": map[string]any{"description": "Connection not found in this organization"}, "503": map[string]any{"description": "Database or SecretKeeper unavailable"}}
	for path, methods := range map[string][]string{"/api/connections": {"get", "post"}, "/api/connections/{reference}": {"get", "patch", "delete"}, "/api/connections/{reference}/validate": {"post"}} {
		operations := map[string]any{}
		for _, method := range methods {
			operation := map[string]any{"summary": "Organization-scoped connection " + method, "tags": []string{"connections"}, "responses": response}
			if path != "/api/connections" {
				operation["parameters"] = []any{map[string]any{"name": "reference", "in": "path", "required": true, "schema": map[string]any{"type": "string"}}}
			}
			if method == "patch" || method == "post" && path == "/api/connections" {
				operation["requestBody"] = map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "description": "Connection fields; PATCH merges explicitly supplied fields"}}}}
			}
			operations[method] = operation
		}
		paths[path] = operations
	}
}
