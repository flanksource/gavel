package ui

func projectCheckoutSwitchOpenAPIPaths() map[string]any {
	return map[string]any{
		"/api/projects/{name}/checkout/switch": map[string]any{
			"get":  projectLifecycleOp("projects_checkout_switch_plan", "Plan switching the primary checkout to a local branch and syncing it with its upstream", "200", "ProjectCheckoutSwitchPlan", branchQueryParams(false)),
			"post": projectLifecycleOp("projects_checkout_switch", "Switch the primary checkout to a local branch, optionally fast-forwarding or resetting it to its upstream", "200", "ProjectCheckoutSwitchResult", jsonRequestBody("ProjectCheckoutSwitchRequest")),
		},
	}
}

func projectCheckoutSwitchOpenAPISchemas() map[string]any {
	commit := objectSchema([]string{"sha", "subject"}, map[string]any{"sha": typed("string"), "subject": typed("string")})
	return map[string]any{
		"ProjectCheckoutSwitchPlan": objectSchema([]string{"repo", "current", "head", "branch", "branchHead", "ahead", "behind", "unpreserved", "dirtyCount"}, map[string]any{
			"repo": typed("string"), "current": map[string]any{"type": "string", "description": "Branch the primary checkout is on; empty when detached"},
			"head": typed("string"), "branch": typed("string"), "branchHead": typed("string"),
			"upstream":     map[string]any{"type": "string", "description": "The branch's upstream (e.g. origin/main); absent when it tracks none"},
			"upstreamHead": map[string]any{"type": "string", "description": "Last-fetched upstream tip; absent when the upstream ref is gone"},
			"ahead":        map[string]any{"type": "integer", "description": "Branch commits not in its upstream"},
			"behind":       map[string]any{"type": "integer", "description": "Upstream commits not in the branch"},
			"unpreserved":  map[string]any{"type": "array", "items": commit, "description": "Branch commits no other ref holds and the upstream has not landed; a reset would orphan them"},
			"dirtyCount":   map[string]any{"type": "integer", "description": "Tracked files with uncommitted changes in the primary checkout"},
			"checkedOutIn": map[string]any{"type": "string", "description": "Linked worktree that has the branch checked out"},
		}),
		"ProjectCheckoutSwitchRequest": objectSchema([]string{"branch", "expectedHead", "sync"}, map[string]any{
			"branch":       typed("string"),
			"expectedHead": map[string]any{"type": "string", "description": "The primary checkout's HEAD the plan was read against"},
			"sync":         map[string]any{"type": "string", "enum": []string{"none", "fast-forward", "reset"}},
		}),
		"ProjectCheckoutSwitchResult": objectSchema([]string{"branch", "head", "sync", "previousBranch", "previousHead", "previousBranchHead"}, map[string]any{
			"branch": typed("string"), "head": typed("string"), "sync": typed("string"),
			"previousBranch": typed("string"), "previousHead": typed("string"),
			"previousBranchHead": map[string]any{"type": "string", "description": "Where the branch pointed before the sync, to undo it"},
		}),
	}
}
