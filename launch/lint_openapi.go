package launch

func lintOpenAPI(paths map[string]any) error {
	paths["/api/launch/linters"] = map[string]any{"get": map[string]any{
		"operationId": "launch_linters", "summary": "List built-in and configured linters in a prepared workspace",
		"parameters": []any{apiParameter("preparationId", "query", true)},
		"responses": map[string]any{"200": apiResponse("Linter names and file patterns", map[string]any{"type": "array", "items": object(map[string]any{
			"name": map[string]any{"type": "string"}, "includes": map[string]any{"type": []string{"array", "null"}, "items": map[string]any{"type": "string"}},
		})}), "404": apiResponse("Unknown caller-scoped preparation", nil)},
	}}
	for _, route := range []struct {
		path, summary string
		input         any
	}{
		{"filter", "Apply configured ignores to findings without changing stored executions", LintFilterRequest{}},
		{"ignore", "Save a scoped ignore rule to the prepared directory's .gavel.yaml", LintIgnoreRequest{}},
	} {
		schema, err := reflected(route.input)
		if err != nil {
			return err
		}
		paths["/api/launch/lint/"+route.path] = map[string]any{"post": map[string]any{
			"operationId": "launch_lint_" + route.path, "summary": route.summary,
			"requestBody": map[string]any{"required": true, "content": jsonContent(schema)},
			"responses":   map[string]any{"200": apiResponse(route.summary, nil), "404": apiResponse("Unknown caller-scoped preparation", nil), "422": apiResponse("Invalid lint request", nil)},
		}}
	}
	return nil
}
