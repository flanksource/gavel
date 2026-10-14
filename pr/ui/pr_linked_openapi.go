package ui

func addPRLinkedOpenAPI(document map[string]any) error {
	schemas := document["components"].(map[string]any)["schemas"].(map[string]any)
	request, err := todoOpenAPISchema(&prMergePayload{}, schemas)
	if err != nil {
		return err
	}
	response, err := todoOpenAPISchema(&linkedPRDetail{}, schemas)
	if err != nil {
		return err
	}
	paths := document["paths"].(map[string]any)
	paths["/api/prs/linked"] = map[string]any{"get": map[string]any{
		"operationId": "prs_linked", "summary": "Fetch a linked GitHub or Azure PR independently of dashboard repository filters", "tags": []string{"pr"},
		"parameters": []any{map[string]any{"name": "url", "in": "query", "required": true, "schema": map[string]any{"type": "string", "format": "uri"}}},
		"responses": map[string]any{
			"200": map[string]any{"description": "Current hosted PR status and verification artifacts", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/" + response}}}},
			"400": map[string]any{"description": "Invalid hosted PR URL"},
			"502": map[string]any{"description": "Host status lookup failed"},
		},
	}}
	paths["/api/prs/merge"] = map[string]any{"post": map[string]any{
		"operationId": "prs_merge", "summary": "Complete a hosted PR using rebase, squash or a merge commit", "tags": []string{"pr"},
		"description": "headSha identifies the reviewed revision; changed heads are rejected. Azure requires headSha and does not support auto. Branch policies are not bypassed.",
		"requestBody": jsonRequestBody(request)["requestBody"],
		"responses": map[string]any{
			"200": map[string]any{"description": "Merge accepted, or GitHub auto-merge enabled"},
			"400": map[string]any{"description": "Invalid PR URL or merge method"},
			"502": map[string]any{"description": "Host rejected the action"},
		},
	}}
	return nil
}
