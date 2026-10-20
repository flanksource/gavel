package ui

import (
	"net/http"
	"os"

	"github.com/flanksource/gavel/git/branchrebase"
)

type projectRebaseRequest struct {
	Worktree     string `json:"worktree"`
	Branch       string `json:"branch"`
	Target       string `json:"target"`
	ExpectedHead string `json:"expectedHead"`
	Operation    string `json:"operation"`
}

func (s *Server) rebaseWorktree(w http.ResponseWriter, r *http.Request, path string) (string, bool) {
	project, err := getVisibleProject(r.PathValue("name"))
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return "", false
	}
	dir, err := projectWorkDir(s.requestContext(r), project, path)
	if err != nil {
		respondError(w, workDirErrorStatus(err), err.Error())
		return "", false
	}
	return dir, true
}

func (s *Server) handleProjectRebaseContext(w http.ResponseWriter, r *http.Request) {
	dir, ok := s.rebaseWorktree(w, r, r.URL.Query().Get("worktree"))
	if !ok {
		return
	}
	result, err := branchrebase.Inspect(r.Context(), dir)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	executable, err := os.Executable()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, struct {
		branchrebase.Context
		Executable string `json:"executable"`
	}{Context: result, Executable: executable})
}

func (s *Server) handleProjectRebase(w http.ResponseWriter, r *http.Request) {
	s.projectRebaseAction(w, r, "start")
}

func (s *Server) handleProjectRebaseAbort(w http.ResponseWriter, r *http.Request) {
	s.projectRebaseAction(w, r, "abort")
}

func (s *Server) projectRebaseAction(w http.ResponseWriter, r *http.Request, action string) {
	var request projectRebaseRequest
	if err := decodeStrict(r, &request); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	dir, ok := s.rebaseWorktree(w, r, request.Worktree)
	if !ok {
		return
	}
	result, err := branchrebase.Run(r.Context(), branchrebase.Options{Action: action, Worktree: dir, Branch: request.Branch, Target: request.Target, ExpectedHead: request.ExpectedHead, Operation: request.Operation})
	s.touchGit(dir)
	if err != nil {
		respondJSON(w, http.StatusConflict, map[string]any{"error": err.Error(), "operation": result})
		return
	}
	respondJSON(w, http.StatusOK, result)
}
