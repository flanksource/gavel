package ui

// worktreeParam is the optional ?worktree= that scopes a project request to
// one of its linked worktrees.
func worktreeParam() map[string]any {
	return map[string]any{
		"name": "worktree", "in": "query", "required": false,
		"description": "Absolute path of one of the project's git worktrees; defaults to the project directory",
		"schema":      map[string]any{"type": "string"},
	}
}

func branchQueryParams(withFile bool) map[string]any {
	params := append(nameParam()["parameters"].([]any), map[string]any{
		"name": "branch", "in": "query", "required": true, "schema": map[string]any{"type": "string"},
	})
	if withFile {
		params = append(params, map[string]any{
			"name": "file", "in": "query", "required": false, "schema": map[string]any{"type": "string"},
		})
	}
	return map[string]any{"parameters": params}
}

func jsonRequestBody(schema string) map[string]any {
	return map[string]any{"requestBody": map[string]any{
		"required": true,
		"content": map[string]any{"application/json": map[string]any{
			"schema": map[string]any{"$ref": "#/components/schemas/" + schema},
		}},
	}}
}

func projectGitOpenAPIPaths() map[string]any {
	summary := projectLifecycleOp("projects_git_summary", "Get every project's unmerged line counts, worktree and branch counts", "200", "ProjectGitSummaries")
	delete(summary, "parameters")
	return map[string]any{
		"/api/projects/git-summary": map[string]any{"get": summary},
		"/api/projects/{name}/git": map[string]any{
			"get": projectLifecycleOp("projects_git", "Get a project's worktrees and unmerged branches", "200", "ProjectGit"),
		},
		"/api/projects/{name}/branch/files": map[string]any{
			"get": projectLifecycleOp("projects_branch_files", "List the files a branch changed since it forked from base", "200", "ProjectBranchFiles", branchQueryParams(false)),
		},
		"/api/projects/{name}/branch/diff": map[string]any{
			"get": projectLifecycleOp("projects_branch_diff", "Get a branch's diff since it forked from base", "200", "ProjectBranchDiff", branchQueryParams(true)),
		},
		"/api/projects/{name}/branch/merge": map[string]any{
			"post": projectLifecycleOp("projects_branch_merge", "Merge a branch into the base branch, then remove its worktree and branch", "200", "ProjectBranchMergeResponse", jsonRequestBody("ProjectBranchMergeRequest")),
		},
		"/api/projects/{name}/branch/pr": map[string]any{
			"post": projectLifecycleOp("projects_branch_pr", "Open a pull request carrying a branch's commits", "200", "ProjectBranchPRResponse", jsonRequestBody("ProjectBranchPRRequest")),
		},
	}
}

func objectSchema(required []string, properties map[string]any) map[string]any {
	return map[string]any{"type": "object", "required": required, "properties": properties}
}

func typed(kind string) map[string]any { return map[string]any{"type": kind} }

func projectGitOpenAPISchemas() map[string]any {
	diffStat := objectSchema([]string{"commits", "files", "adds", "dels"}, map[string]any{
		"commits": typed("integer"), "files": typed("integer"), "adds": typed("integer"), "dels": typed("integer"),
	})
	changes := objectSchema([]string{"staged", "unstaged", "both", "untracked", "conflict", "adds", "dels"}, map[string]any{
		"staged": typed("integer"), "unstaged": typed("integer"), "both": typed("integer"), "untracked": typed("integer"),
		"conflict": typed("integer"), "adds": typed("integer"), "dels": typed("integer"),
	})
	dateTime := map[string]any{"type": "string", "format": "date-time"}
	worktree := objectSchema([]string{"path", "branch", "head", "primary", "detached", "prunable", "changes", "ahead", "lastCommitAt"}, map[string]any{
		"path": typed("string"), "branch": typed("string"), "head": typed("string"), "primary": typed("boolean"),
		"detached": typed("boolean"), "prunable": typed("boolean"), "changes": changes, "ahead": typed("integer"),
		"lastCommitAt": dateTime, "touchedAt": dateTime,
	})
	branch := objectSchema([]string{"name", "head", "ahead", "behind", "worktree", "diff", "lastCommitAt"}, map[string]any{
		"name": typed("string"), "head": typed("string"), "ahead": typed("integer"), "behind": typed("integer"),
		"worktree": typed("string"), "diff": diffStat, "lastCommitAt": dateTime,
	})
	summary := objectSchema([]string{"name", "base", "adds", "dels", "worktrees", "branches"}, map[string]any{
		"name": typed("string"), "base": typed("string"), "adds": typed("integer"), "dels": typed("integer"),
		"worktrees": map[string]any{"type": "integer", "description": "Linked worktrees, the primary checkout excluded"},
		"branches":  typed("integer"), "error": typed("string"),
	})
	return map[string]any{
		"ProjectGitSummaries": map[string]any{"type": "array", "items": summary},
		"ProjectGit": objectSchema([]string{"base", "currentBranch", "baseCheckedOut", "worktrees", "branches"}, map[string]any{
			"base": typed("string"), "currentBranch": typed("string"), "baseCheckedOut": typed("boolean"),
			"worktrees": map[string]any{"type": "array", "items": worktree},
			"branches":  map[string]any{"type": "array", "items": branch},
		}),
		"ProjectBranchFiles": objectSchema([]string{"hash", "files"}, map[string]any{
			"hash": typed("string"), "base": typed("string"), "files": map[string]any{"type": "array", "items": typed("object")},
		}),
		"ProjectBranchDiff": objectSchema([]string{"diff", "truncated", "binary", "path", "commit"}, map[string]any{
			"diff": typed("string"), "truncated": typed("boolean"), "binary": typed("boolean"), "path": typed("string"), "commit": typed("string"),
		}),
		"ProjectBranchMergeRequest": objectSchema([]string{"branch", "mode"}, map[string]any{
			"branch": typed("string"), "mode": map[string]any{"type": "string", "enum": []string{"squash", "incremental"}},
			"message": map[string]any{"type": "string", "description": "Squash commit message; generated from the branch's commits when empty"},
		}),
		"ProjectBranchMergeResponse": objectSchema([]string{"targetBranch", "landedSha", "mode", "commits", "worktreeRemoved", "branchDeleted"}, map[string]any{
			"targetBranch": typed("string"), "landedSha": typed("string"), "mode": typed("string"), "commits": typed("integer"),
			"worktreeRemoved": typed("boolean"), "branchDeleted": typed("boolean"),
		}),
		"ProjectBranchPRRequest": objectSchema([]string{"branch"}, map[string]any{
			"branch": typed("string"), "draft": typed("boolean"),
		}),
		"ProjectBranchPRResponse": objectSchema([]string{"number", "url", "topicBranch"}, map[string]any{
			"number": typed("integer"), "url": typed("string"), "topicBranch": typed("string"),
		}),
	}
}
