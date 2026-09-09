package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/github"
)

type prDiffResponse struct {
	Repo      string `json:"repo"`
	Number    int    `json:"number,omitempty"`
	Commit    string `json:"commit,omitempty"`
	Path      string `json:"path,omitempty"`
	Diff      string `json:"diff"`
	Truncated bool   `json:"truncated,omitempty"`
	Binary    bool   `json:"binary,omitempty"`
}

func (s *Server) handlePRCommitDiff(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	repo := strings.TrimSpace(r.URL.Query().Get("repo"))
	sha := strings.TrimSpace(r.URL.Query().Get("sha"))
	if err := validatePRDiffRepo(repo); err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	if sha == "" {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("sha is required"))
		return
	}
	if !gavelgit.IsValidCommitHash(sha) {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("invalid commit sha %q", sha))
		return
	}
	opts := s.ghOpts
	opts.Repo = repo
	payload, err := github.FetchCommitDiff(opts, sha)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err)
		return
	}
	json.NewEncoder(w).Encode(prDiffResponse{ //nolint:errcheck
		Repo:      repo,
		Commit:    sha,
		Diff:      payload.Diff,
		Truncated: payload.Truncated,
		Binary:    payload.Binary,
	})
}

func (s *Server) handlePRFileDiff(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	repo := strings.TrimSpace(r.URL.Query().Get("repo"))
	numberText := strings.TrimSpace(r.URL.Query().Get("number"))
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	if err := validatePRDiffRepo(repo); err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	number, err := strconv.Atoi(numberText)
	if numberText == "" || err != nil || number <= 0 {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("invalid number"))
		return
	}
	if path == "" {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("path is required"))
		return
	}
	opts := s.ghOpts
	opts.Repo = repo
	payload, err := github.FetchPRFilePatch(opts, number, path)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err)
		return
	}
	json.NewEncoder(w).Encode(prDiffResponse{ //nolint:errcheck
		Repo:      repo,
		Number:    number,
		Path:      path,
		Diff:      payload.Diff,
		Truncated: payload.Truncated,
		Binary:    payload.Binary,
	})
}

func validatePRDiffRepo(repo string) error {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("repo must be owner/name")
	}
	if strings.Contains(repo, "..") {
		return fmt.Errorf("invalid repo")
	}
	return nil
}
