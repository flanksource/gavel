package ui

import (
	"encoding/json"
	captaindb "github.com/flanksource/captain/pkg/database"
	prlifecycle "github.com/flanksource/gavel/pr/lifecycle"
	"github.com/flanksource/gavel/workflows"
	"github.com/invopop/jsonschema"
	"reflect"
	"strings"
)

func addLifecycleOpenAPI(document map[string]any) error {
	schemas := document["components"].(map[string]any)["schemas"].(map[string]any)
	types := map[string]any{"LifecycleView": &workflows.View{}, "LifecycleEdit": &workflows.Edit{}, "PRLifecyclePreview": &prlifecycle.Preview{}, "PRLifecycleRequest": &prLifecycleRequest{}, "PRLifecyclePrepared": &prLifecyclePrepared{}, "PRLifecycleRun": &captaindb.PromptRun{}}
	names := map[string]string{}
	for name, value := range types {
		reflector := jsonschema.Reflector{FieldNameTag: "json", RequiredFromJSONSchemaTags: true, Namer: func(t reflect.Type) string {
			if t.Name() == "" {
				return ""
			}
			return strings.NewReplacer("/", "_", ".", "_", "-", "_").Replace(t.PkgPath() + "_" + t.Name())
		}}
		raw, err := json.Marshal(reflector.Reflect(value))
		if err != nil {
			return err
		}
		ref, err := mergeOpenAPISchema(raw, schemas)
		if err != nil {
			return err
		}
		names[name] = ref
	}
	parameters := func(keys ...string) []any {
		result := []any{}
		for _, key := range keys {
			result = append(result, map[string]any{"name": key, "in": "query", "schema": typed("string")})
		}
		return result
	}
	response := func(name string) map[string]any {
		return map[string]any{"description": name, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/" + names[name]}}}}
	}
	operation := func(id, summary, result string, params []any) map[string]any {
		return map[string]any{"operationId": id, "summary": summary, "tags": []string{"lifecycles"}, "parameters": params, "responses": map[string]any{"200": response(result), "400": map[string]any{"description": "Invalid scope or lifecycle"}, "409": map[string]any{"description": "Preview changed or another PR run is active"}}}
	}
	paths := document["paths"].(map[string]any)
	settingsParams := parameters("scope", "project", "kind")
	read := operation("lifecycles_read", "Read one scope's authored and resolved lifecycle", "LifecycleView", settingsParams)
	preview := operation("lifecycles_preview", "Validate a lifecycle without saving", "LifecycleView", settingsParams)
	preview["requestBody"] = jsonRequestBody(names["LifecycleEdit"])["requestBody"]
	save := operation("lifecycles_save", "Save a validated lifecycle with revision protection", "LifecycleView", settingsParams)
	save["requestBody"] = jsonRequestBody(names["LifecycleEdit"])["requestBody"]
	paths["/api/settings/lifecycles"] = map[string]any{"get": read, "post": preview, "put": save}
	prParams := parameters("project", "repo", "number")
	context := operation("pr_lifecycle_context", "Offer matching steps using current PR context and native history", "PRLifecyclePreview", prParams)
	action := operation("pr_lifecycle_action", "Preview, start or cancel a manual PR lifecycle action", "PRLifecyclePrepared", prParams)
	action["requestBody"] = jsonRequestBody(names["PRLifecycleRequest"])["requestBody"]
	action["responses"].(map[string]any)["202"] = response("PRLifecycleRun")
	paths["/api/pr/lifecycle"] = map[string]any{"get": context, "post": action}
	return nil
}
