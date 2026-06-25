package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/flanksource/gavel/commit"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/models"
)

// todoCommit is one git commit linked to a todo via its Gavel-Issue-Id trailer.
// URL is the commit's web page on the origin remote, empty for a local-only repo.
type todoCommit struct {
	Hash      string `json:"hash"`
	ShortHash string `json:"shortHash"`
	Subject   string `json:"subject"`
	Author    string `json:"author,omitempty"`
	Date      string `json:"date,omitempty"`
	URL       string `json:"url,omitempty"`
}

type todoCommitsResponse struct {
	// IssueID is the todo's id (the Gavel-Issue-Id trailer value commits are
	// matched against); empty for file todos that carry no id.
	IssueID string       `json:"issueId,omitempty"`
	Commits []todoCommit `json:"commits"`
}

// todoCommitDiffResponse carries one commit's rendered diff (ANSI-colored
// `git show` output). Truncated is set when the diff exceeded the size cap.
type todoCommitDiffResponse struct {
	Hash      string `json:"hash"`
	Diff      string `json:"diff"`
	Truncated bool   `json:"truncated,omitempty"`
}

// handleTodoCommits lists the git commits that reference a todo through their
// Gavel-Issue-Id trailer, each with a link to the commit on the origin remote.
// It resolves the todo to read its id, then scans the workspace's git history.
func (s *Server) handleTodoCommits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	if ref == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	provider, source, err := s.todoProvider(todoSourceFromRequest(r))
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	todo, err := provider.Get(r.Context(), ref)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	commits, err := collectTodoCommits(source.Dir, todo.ID)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(todoCommitsResponse{ //nolint:errcheck
		IssueID: strings.TrimSpace(todo.ID),
		Commits: commits,
	})
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
	dir := s.resolveTodoDir(strings.TrimSpace(r.URL.Query().Get("dir")))
	diff, truncated, err := gavelgit.CommitDiff(dir, hash)
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

// collectTodoCommits finds the commits in dir whose Gavel-Issue-Id trailer
// equals issueID and maps them to todoCommit rows with a web link resolved from
// the origin remote. An empty issueID (e.g. a file-backed todo) returns no
// commits without touching git.
func collectTodoCommits(dir, issueID string) ([]todoCommit, error) {
	issueID = strings.TrimSpace(issueID)
	if issueID == "" {
		return []todoCommit{}, nil
	}
	commits, err := gavelgit.CommitsWithTrailer(dir, commit.TrailerIssueID, issueID)
	if err != nil {
		return nil, err
	}
	// Best-effort: a repo without a parsable origin remote yields no links, just
	// the commit metadata.
	base, _ := gavelgit.RemoteWebURL(dir)

	out := make([]todoCommit, 0, len(commits))
	for _, c := range commits {
		tc := todoCommit{
			Hash:      c.Hash,
			ShortHash: shortCommitHash(c.Hash),
			Subject:   fullCommitSubject(c),
			Author:    c.Author.Name,
		}
		if !c.Author.Date.IsZero() {
			tc.Date = c.Author.Date.Format(time.RFC3339)
		}
		if base != "" {
			tc.URL = base + "/commit/" + c.Hash
		}
		out = append(out, tc)
	}
	return out, nil
}

// fullCommitSubject reassembles the conventional-commit subject line the parser
// split into CommitType/Scope/Subject (e.g. "feat(ui): add panel"), falling back
// to the bare subject when no type was detected.
func fullCommitSubject(c models.Commit) string {
	var b strings.Builder
	if c.CommitType != models.CommitTypeUnknown {
		b.WriteString(string(c.CommitType))
		if c.Scope != models.ScopeTypeUnknown {
			b.WriteString("(" + string(c.Scope) + ")")
		}
		b.WriteString(": ")
	}
	b.WriteString(c.Subject)
	return b.String()
}

func shortCommitHash(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}
