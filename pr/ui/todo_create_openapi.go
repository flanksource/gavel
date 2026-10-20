package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
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
	for _, field := range []string{"dir", "title", "body", "priority", "status", "parent", "labels", "autoSave", "step"} {
		formProperties[field] = properties[field]
		parameters = append(parameters, map[string]any{"name": field, "in": "query", "required": false, "schema": properties[field]})
	}
	form := map[string]any{"type": "object", "properties": formProperties, "required": []string{"title"}}
	fileProperties := withEntries(map[string]any{}, formProperties)
	fileProperties["attachment"] = map[string]any{"type": "string", "format": "binary", "description": "Screenshot or other attachment saved before the step starts"}
	multipart := map[string]any{"type": "object", "properties": fileProperties, "required": []string{"title"}}
	document["paths"].(map[string]any)["/api/todos/new"] = map[string]any{"post": map[string]any{
		"operationId": "todos_new", "summary": "Create a TODO and optionally start a lifecycle step on it after saving", "tags": []string{"todo"},
		"description": "Omitted step starts nothing. A named step must apply to the new TODO's status (see GET /api/todos/new/steps). Explicit body values take precedence over query values. Closing the form does not cancel an admitted run.",
		"parameters":  parameters,
		"requestBody": map[string]any{"required": false, "content": map[string]any{
			"application/json":                  map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/" + request}},
			"application/x-www-form-urlencoded": map[string]any{"schema": form},
			"multipart/form-data":               map[string]any{"schema": multipart},
		}},
		"responses": map[string]any{
			"201": map[string]any{"description": "TODO saved, including when step admission fails. Retry the step on the returned TODO instead of creating it again.", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/" + response}}}},
			"400": map[string]any{"description": "Invalid creation input, or a step that does not apply to the new TODO; nothing is saved"},
			"500": map[string]any{"description": "Storage or response failure"},
		},
	}}
	return addTodoNewStepsOpenAPI(document, schemas)
}

func addTodoNewStepsOpenAPI(document map[string]any, schemas map[string]any) error {
	response, err := todoOpenAPISchema(&todoCreateStepsResponse{}, schemas)
	if err != nil {
		return err
	}
	statuses := []string{}
	for _, status := range types.AssignableStatuses() {
		statuses = append(statuses, string(status))
	}
	document["paths"].(map[string]any)["/api/todos/new/steps"] = map[string]any{"get": map[string]any{
		"operationId": "todos_new_steps", "summary": "List the lifecycle steps a new TODO can start", "tags": []string{"todo"},
		"description": "Evaluates the workspace lifecycle for a TODO not yet created: the given status, medium priority, no labels, empty body, no plan, no verification and no runs. Steps are in definition order, auxiliary steps included.",
		"parameters": []any{
			map[string]any{"name": "dir", "in": "query", "required": false, "schema": map[string]any{"type": "string"}},
			map[string]any{"name": "status", "in": "query", "required": false, "schema": map[string]any{"type": "string", "enum": statuses, "default": string(types.StatusDraft)}},
		},
		"responses": map[string]any{
			"200": map[string]any{"description": "Applicable steps", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/" + response}}}},
			"400": map[string]any{"description": "Status cannot be assigned, or the workspace cannot be resolved"},
			"500": map[string]any{"description": "Lifecycle could not be loaded or evaluated"},
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
