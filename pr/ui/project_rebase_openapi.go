package ui

func projectRebaseOpenAPISchemas() map[string]any {
	operation := objectSchema([]string{"id", "worktree", "branch", "target", "targetSha", "originalHead", "stashSha", "phase", "error", "conflicts", "aborted"}, map[string]any{
		"id": typed("string"), "worktree": typed("string"), "branch": typed("string"), "target": typed("string"), "targetSha": typed("string"), "originalHead": typed("string"), "stashSha": typed("string"), "phase": typed("string"), "error": typed("string"), "conflicts": map[string]any{"type": "array", "items": typed("string")}, "aborted": typed("boolean"),
	})
	return map[string]any{
		"ProjectRebaseOperation": operation,
		"ProjectRebaseRequest": objectSchema([]string{"worktree"}, map[string]any{
			"worktree": typed("string"), "branch": typed("string"), "target": map[string]any{"type": "string", "description": "Full refs/heads branch or refs/remotes/origin/main"}, "expectedHead": typed("string"), "operation": map[string]any{"type": "string", "description": "Required to resume or abort an existing operation; start requires branch, target and expectedHead"},
		}),
		"ProjectRebaseContext": objectSchema([]string{"executable", "branch", "head", "dirtyCount", "targets", "operation"}, map[string]any{
			"executable": typed("string"),
			"branch":     typed("string"), "head": typed("string"), "dirtyCount": typed("integer"),
			"targets":   map[string]any{"type": "array", "items": objectSchema([]string{"ref", "label"}, map[string]any{"ref": typed("string"), "label": typed("string")})},
			"operation": map[string]any{"anyOf": []any{operation, typed("null")}},
		}),
	}
}
