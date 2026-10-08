package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/git/gitstate"
)

// todoCommitDiffResponse is a clicky-ui GitDiffPayload: a plain unified diff of
// one commit (or a base..commit range), optionally narrowed to path. Truncated
// is set when the diff exceeded the size cap; Binary when every file in it is
// binary.
type todoCommitDiffResponse struct {
	Diff      string `json:"diff"`
	Truncated bool   `json:"truncated"`
	Binary    bool   `json:"binary"`
	Path      string `json:"path"`
	Commit    string `json:"commit"`
}

// todoCommitFilesResponse carries the per-file change summary of one commit
// (or a base..hash range), each enriched with its repomap scope/language.
type todoCommitFilesResponse struct {
	Hash  string                `json:"hash"`
	Base  string                `json:"base,omitempty"`
	Files []gavelgit.CommitFile `json:"files"`
}

// handleTodoCommitDiff returns the plain unified diff for a commit, or for the
// base..hash range when base is set, optionally narrowed to one file or
// directory, so the dashboard can render it in a diff viewer.
func (s *Server) handleTodoCommitDiff(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dir, opts, ok := s.todoCommitDiffRequest(w, r)
	if !ok {
		return
	}
	result, err := gavelgit.CommitDiff(dir, opts)
	if err != nil {
		writeTodoError(w, todoCommitDiffErrorStatus(err), err)
		return
	}
	json.NewEncoder(w).Encode(todoCommitDiffResponse{ //nolint:errcheck
		Diff:      result.Diff,
		Truncated: result.Truncated,
		Binary:    result.Binary,
		Path:      opts.File,
		Commit:    opts.Head,
	})
}

// handleTodoCommitFiles returns the per-file change summary (path, change kind,
// +/- counts, and repomap scope/language) for a commit or a base..hash range,
// so the dashboard can render a file tree and load each file's diff on demand.
//
// A whole base..hash range is read from the range's cached file list (see
// storedRangeFiles), which lists merge-base(base, hash)..hash: the same files
// whenever base is an ancestor of hash, as a run's setup is of its head. A
// single commit, or a range narrowed to one file, stays live: `git show` of an
// immutable commit is cheap, and the client caches it indefinitely.
func (s *Server) handleTodoCommitFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dir, opts, ok := s.todoCommitDiffRequest(w, r)
	if !ok {
		return
	}
	var files []gavelgit.CommitFile
	var err error
	if opts.Base != "" && opts.File == "" {
		tracker, trackerErr := s.context().GitTracker()
		if trackerErr != nil {
			writeTodoError(w, gitErrorStatus(trackerErr), trackerErr)
			return
		}
		files, _, err = storedRangeFiles(s.requestContext(r), tracker, dir, gitstate.RangeKey{Base: opts.Base, Head: opts.Head})
	} else {
		files, err = gavelgit.CommitFiles(dir, opts)
	}
	if err != nil {
		writeTodoError(w, todoCommitDiffErrorStatus(err), err)
		return
	}
	json.NewEncoder(w).Encode(todoCommitFilesResponse{ //nolint:errcheck
		Hash:  opts.Head,
		Base:  opts.Base,
		Files: files,
	})
}

// todoCommitDiffRequest reads the shared dir/hash/base/file query of the commit
// endpoints, answering 400 (and false) for a missing or malformed value before
// any git command runs. The commit is located within the workspace dir; the
// provider/todo is not needed.
func (s *Server) todoCommitDiffRequest(w http.ResponseWriter, r *http.Request) (string, gavelgit.CommitDiffOptions, bool) {
	query := r.URL.Query()
	opts := gavelgit.CommitDiffOptions{
		Base: strings.TrimSpace(query.Get("base")),
		Head: strings.TrimSpace(query.Get("hash")),
		File: strings.TrimSpace(query.Get("file")),
	}
	if opts.Head == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("hash is required"))
		return "", opts, false
	}
	if err := opts.Validate(); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return "", opts, false
	}
	dir, err := s.resolveTodoDir(query.Get("dir"))
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return "", opts, false
	}
	return dir, opts, true
}

// todoCommitDiffErrorStatus answers 410 Gone for a commit that is no longer in
// the repository (e.g. a run branch deleted and garbage-collected after it
// landed) and 500 for any other git failure.
func todoCommitDiffErrorStatus(err error) int {
	if errors.Is(err, gavelgit.ErrCommitNotFound) {
		return http.StatusGone
	}
	return http.StatusInternalServerError
}
