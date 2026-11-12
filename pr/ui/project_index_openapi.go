package ui

func worktreeOnly() map[string]any {
	return map[string]any{"parameters": append(nameParam()["parameters"].([]any), worktreeParam())}
}

// projectIndexOpenAPIPaths documents the git index routes of the project page:
// staging whole entries or single lines, discarding changes, and opening a
// file in a host editor.
func projectIndexOpenAPIPaths() map[string]any {
	return map[string]any{
		"/api/projects/{name}/stage": map[string]any{
			"post": projectLifecycleOp("projects_stage", "Stage (git add) or unstage (git restore --staged) current changes", "200", "ProjectOK",
				jsonRequestBody("ProjectStageRequest"), worktreeOnly()),
		},
		"/api/projects/{name}/stage/lines": map[string]any{
			"post": projectLifecycleOp("projects_stage_lines", "Stage or unstage individual changed lines of one file", "200", "ProjectOK",
				jsonRequestBody("ProjectStageLinesRequest"), worktreeOnly()),
		},
		"/api/projects/{name}/discard": map[string]any{
			"post": projectLifecycleOp("projects_discard", "Discard changes: restore tracked paths to HEAD and delete untracked ones", "200", "ProjectOK",
				jsonRequestBody("ProjectDiscardRequest"), worktreeOnly()),
		},
		"/api/projects/{name}/open": map[string]any{
			"post": projectLifecycleOp("projects_open", "Open a project file in an editor installed on the gavel host", "200", "ProjectOK",
				jsonRequestBody("ProjectOpenRequest"), worktreeOnly()),
		},
		"/api/editors": map[string]any{
			"get": map[string]any{
				"operationId": "editors_list",
				"summary":     "List the editors detected on the gavel host",
				"tags":        []string{"project"},
				"responses": map[string]any{"200": map[string]any{
					"description": "OK",
					"content": map[string]any{"application/json": map[string]any{
						"schema": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/ProjectEditor"}},
					}},
				}},
			},
		},
	}
}

func projectIndexOpenAPISchemas() map[string]any {
	paths := map[string]any{"type": "array", "minItems": 1, "items": map[string]any{"type": "string"},
		"description": "Project-relative paths of current changes; a folder names every change under it"}
	return map[string]any{
		"ProjectOK": map[string]any{
			"type": "object", "required": []string{"ok"},
			"properties": map[string]any{"ok": map[string]any{"type": "boolean", "enum": []bool{true}}},
		},
		"ProjectStageRequest": map[string]any{
			"type": "object", "required": []string{"paths"}, "additionalProperties": false,
			"properties": map[string]any{"paths": paths, "unstage": map[string]any{"type": "boolean", "default": false}},
		},
		"ProjectStageLinesRequest": map[string]any{
			"type": "object", "required": []string{"path", "lines"}, "additionalProperties": false,
			"properties": map[string]any{
				"path":    map[string]any{"type": "string"},
				"unstage": map[string]any{"type": "boolean", "default": false, "description": "Unstage the lines from the staged (HEAD→index) diff instead of staging them from the unstaged (index→worktree) diff"},
				"lines": map[string]any{"type": "array", "minItems": 1, "items": map[string]any{
					"type": "object", "required": []string{"side", "line"}, "additionalProperties": false,
					"properties": map[string]any{
						"side": map[string]any{"type": "string", "enum": []string{"old", "new"}, "description": "old: the old-file number of a '-' line; new: the new-file number of a '+' line"},
						"line": map[string]any{"type": "integer", "minimum": 1},
					},
				}},
			},
		},
		"ProjectDiscardRequest": map[string]any{
			"type": "object", "required": []string{"paths"}, "additionalProperties": false,
			"properties": map[string]any{"paths": paths},
		},
		"ProjectOpenRequest": map[string]any{
			"type": "object", "required": []string{"path", "editor"}, "additionalProperties": false,
			"properties": map[string]any{
				"path":   map[string]any{"type": "string"},
				"editor": map[string]any{"type": "string", "description": "An id from GET /api/editors"},
				"line":   map[string]any{"type": "integer", "minimum": 1},
			},
		},
		"ProjectEditor": map[string]any{
			"type": "object", "required": []string{"id", "label"},
			"properties": map[string]any{"id": map[string]any{"type": "string"}, "label": map[string]any{"type": "string"}},
		},
	}
}
