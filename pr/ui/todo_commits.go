package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	gavelgit "github.com/flanksource/gavel/git"
)

// todoCommitDiffResponse carries one commit's rendered diff (ANSI-colored
// `git show` output). Truncated is set when the diff exceeded the size cap.
type todoCommitDiffResponse struct {
	Hash      string `json:"hash"`
	Diff      string `json:"diff"`
	Truncated bool   `json:"truncated,omitempty"`
}

// todoCommitFilesResponse carries one commit's per-file change summary, each
// enriched with its repomap scope/language for the expanded commit status view.
type todoCommitFilesResponse struct {
	Hash  string                `json:"hash"`
	Files []gavelgit.CommitFile `json:"files"`
}

// handleTodoCommitDiff returns the ANSI-colored diff for a single commit so the
// dashboard can expand a commit row to show its changes. The commit is located
// by hash within the workspace dir; the provider/todo is not needed.
func (s *Server) handleTodoCommitDiff(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	hash := strings.TrimSpace(r.URL.Query().Get("hash"))
	if hash == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("hash is required"))
		return
	}
	if !gavelgit.IsValidCommitHash(hash) {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid commit hash %q", hash))
		return
	}
	dir, err := s.resolveTodoDir(r.URL.Query().Get("dir"))
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	// An optional file narrows the diff to a single path (the per-file hover
	// card); empty shows the whole commit.
	file := strings.TrimSpace(r.URL.Query().Get("file"))
	diff, truncated, err := gavelgit.CommitDiff(dir, hash, file)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(todoCommitDiffResponse{ //nolint:errcheck
		Hash:      hash,
		Diff:      diff,
		Truncated: truncated,
	})
}

// handleTodoCommitFiles returns the per-file change summary for a single commit
// (path, change kind, +/- counts, and repomap scope/language), so the dashboard
// can render a commit's "repomap-based status" rows and load each file's diff on
// demand. The commit is located by hash within the workspace dir.
func (s *Server) handleTodoCommitFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	hash := strings.TrimSpace(r.URL.Query().Get("hash"))
	if hash == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("hash is required"))
		return
	}
	if !gavelgit.IsValidCommitHash(hash) {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid commit hash %q", hash))
		return
	}
	dir, err := s.resolveTodoDir(r.URL.Query().Get("dir"))
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	files, err := gavelgit.CommitFiles(dir, hash)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(todoCommitFilesResponse{ //nolint:errcheck
		Hash:  hash,
		Files: files,
	})
}
