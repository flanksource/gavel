package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/flanksource/captain/pkg/api"
)

func addTodoNewOpenAPI(document map[string]any) error {
	schemas := document["components"].(map[string]any)["schemas"].(map[string]any)
	request, err := todoOpenAPISchema(&todoNewPayload{}, schemas)
	if err != nil {
		return err
	}
	response, err := todoOpenAPISchema(&todoNewResponse{}, schemas)
	if err != nil {
		return err
	}
	requestSchema := schemas[request].(map[string]any)
	properties := requestSchema["properties"].(map[string]any)
	formProperties := map[string]any{}
	parameters := []any{}
	for _, field := range []string{"dir", "title", "body", "priority", "status", "parent", "labels", "autoSave", "triage"} {
		formProperties[field] = properties[field]
		parameters = append(parameters, map[string]any{"name": field, "in": "query", "required": false, "schema": properties[field]})
	}
	form := map[string]any{"type": "object", "properties": formProperties, "required": []string{"title"}}
	fileProperties := withEntries(map[string]any{}, formProperties)
	fileProperties["attachment"] = map[string]any{"type": "string", "format": "binary", "description": "Screenshot or other attachment saved before triage starts"}
	multipart := map[string]any{"type": "object", "properties": fileProperties, "required": []string{"title"}}
	document["paths"].(map[string]any)["/api/todos/new"] = map[string]any{"post": map[string]any{
		"operationId": "todos_new", "summary": "Create a TODO and optionally triage it after saving", "tags": []string{"todo"},
		"description": "Omitted triage is false. Explicit body values take precedence over query values. Closing the form does not cancel an admitted run.",
		"parameters":  parameters,
		"requestBody": map[string]any{"required": false, "content": map[string]any{
			"application/json":                  map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/" + request}},
			"application/x-www-form-urlencoded": map[string]any{"schema": form},
			"multipart/form-data":               map[string]any{"schema": multipart},
		}},
		"responses": map[string]any{
			"201": map[string]any{"description": "TODO saved, including when triage admission fails. Retry triage on the returned TODO instead of creating it again.", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/" + response}}}},
			"400": map[string]any{"description": "Invalid creation input"},
			"500": map[string]any{"description": "Storage or response failure"},
		},
	}}
	return nil
}

func todoOpenAPISchema(value any, schemas map[string]any) (string, error) {
	raw, err := api.SchemaJSON(value)
	if err != nil {
		return "", err
	}
	var document struct {
		Ref  string         `json:"$ref"`
		Defs map[string]any `json:"$defs"`
	}
	if err := json.Unmarshal(bytes.ReplaceAll(raw, []byte("#/$defs/"), []byte("#/components/schemas/")), &document); err != nil {
		return "", err
	}
	const prefix = "#/components/schemas/"
	if len(document.Ref) <= len(prefix) || document.Ref[:len(prefix)] != prefix {
		return "", fmt.Errorf("creation API schema has no component reference")
	}
	for name, definition := range document.Defs {
		if existing, ok := schemas[name]; ok && !reflect.DeepEqual(existing, definition) {
			return "", fmt.Errorf("conflicting creation API schema %s", name)
		}
		schemas[name] = definition
	}
	return document.Ref[len(prefix):], nil
}
