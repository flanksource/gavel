package ui

import (
	"errors"
	"net/http"

	"github.com/flanksource/gavel/git/branchdelete"
)

type projectBranchDeleteRequest struct {
	Branch             string `json:"branch"`
	ExpectedHead       string `json:"expectedHead"`
	DiscardUncommitted bool   `json:"discardUncommitted,omitempty"`
	DeleteUnmerged     bool   `json:"deleteUnmerged,omitempty"`
}

// handleProjectBranchDeletePlan answers what deleting ?branch= would remove:
// its worktree, the worktree's uncommitted files and the commits not in base.
func (s *Server) handleProjectBranchDeletePlan(w http.ResponseWriter, r *http.Request) {
	opts, ok := projectBranchDeleteOptions(w, r, r.URL.Query().Get("branch"))
	if !ok {
		return
	}
	plan, err := branchdelete.Inspect(r.Context(), opts)
	if err != nil {
		respondBranchDeleteError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, plan)
}

// handleProjectBranchDelete removes the branch and its worktree. Losing
// uncommitted files or unmerged commits needs the matching acknowledgement,
// and expectedHead pins the plan the user confirmed.
func (s *Server) handleProjectBranchDelete(w http.ResponseWriter, r *http.Request) {
	var request projectBranchDeleteRequest
	if err := decodeStrict(r, &request); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	opts, ok := projectBranchDeleteOptions(w, r, request.Branch)
	if !ok {
		return
	}
	opts.ExpectedHead, opts.DiscardUncommitted, opts.DeleteUnmerged = request.ExpectedHead, request.DiscardUncommitted, request.DeleteUnmerged
	result, err := branchdelete.Delete(r.Context(), opts)
	// A refused worktree removal can still have run part-way.
	s.touchGit(opts.Repo)
	if err != nil {
		respondBranchDeleteError(w, err)
		return
	}
	s.notify()
	respondJSON(w, http.StatusOK, result)
}

func projectBranchDeleteOptions(w http.ResponseWriter, r *http.Request, branch string) (branchdelete.Options, bool) {
	ref, ok := resolveProjectBranch(w, r, branch)
	if !ok {
		return branchdelete.Options{}, false
	}
	wt, err := ref.branchWorktree()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return branchdelete.Options{}, false
	}
	return branchdelete.Options{Repo: ref.dir, Branch: ref.branch, Base: ref.base, Worktree: wt}, true
}

func respondBranchDeleteError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, branchdelete.ErrRefused):
		respondError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, branchdelete.ErrNeedsConfirmation), errors.Is(err, branchdelete.ErrStale):
		respondError(w, http.StatusConflict, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, err.Error())
	}
}

func projectBranchDeleteOpenAPIPath() map[string]any {
	return map[string]any{
		"get":  projectLifecycleOp("projects_branch_delete_plan", "Plan deleting a branch and its worktree: uncommitted files and commits not in base", "200", "ProjectBranchDeletePlan", branchQueryParams(false)),
		"post": projectLifecycleOp("projects_branch_delete", "Delete a branch and its linked worktree; losing uncommitted files or unmerged commits needs acknowledgement", "200", "ProjectBranchDeleteResult", jsonRequestBody("ProjectBranchDeleteRequest")),
	}
}

func projectBranchDeleteOpenAPISchemas() map[string]any {
	return map[string]any{
		"ProjectBranchDeletePlan": objectSchema([]string{"branch", "head", "base", "uncommitted", "unmerged"}, map[string]any{
			"branch": typed("string"), "head": typed("string"), "base": typed("string"),
			"worktree":    map[string]any{"type": "string", "description": "The branch's linked worktree, removed with it"},
			"uncommitted": map[string]any{"type": "array", "items": typed("string")},
			"unmerged": map[string]any{"type": "array", "items": objectSchema([]string{"sha", "subject"}, map[string]any{
				"sha": typed("string"), "subject": typed("string"),
			})},
		}),
		"ProjectBranchDeleteRequest": objectSchema([]string{"branch", "expectedHead"}, map[string]any{
			"branch": typed("string"), "expectedHead": typed("string"),
			"discardUncommitted": map[string]any{"type": "boolean", "description": "Acknowledge losing the worktree's uncommitted files"},
			"deleteUnmerged":     map[string]any{"type": "boolean", "description": "Acknowledge losing commits that are not in base"},
		}),
		"ProjectBranchDeleteResult": objectSchema([]string{"worktreeRemoved", "branchDeleted"}, map[string]any{
			"worktreeRemoved": typed("boolean"), "branchDeleted": typed("boolean"),
		}),
	}
}
