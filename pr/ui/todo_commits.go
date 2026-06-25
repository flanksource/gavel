package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/commit"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/models"
)

// commitStatTTL bounds how often the todo list recomputes a workspace's git diff
// stats. Within the TTL the per-issue counts are served straight from the cache
// DB; past it, one `git log --numstat` pass refreshes every todo's stats at once.
const commitStatTTL = 60 * time.Second

// commitDiffStats returns the aggregated git diff footprint per todo id for a
// workspace, keyed by Gavel-Issue-Id. It serves the DB cache within commitStatTTL
// and otherwise recomputes from git and refreshes the cache. When no cache DB is
// configured it computes directly every call. A git failure (e.g. a non-repo
// workspace) degrades to no stats rather than failing the list.
func commitDiffStats(ctx context.Context, dir string) map[string]gavelgit.DiffStat {
	repo := repoStatKey(dir)
	store := cache.Shared()
	if cached, syncedAt, err := store.CommitStats(ctx, repo); err != nil {
		logger.Debugf("read commit stats for %s: %v", repo, err)
	} else if !syncedAt.IsZero() && time.Since(syncedAt) < commitStatTTL {
		return cached
	}
	stats, err := gavelgit.TrailerDiffStats(dir, commit.TrailerIssueID)
	if err != nil {
		logger.Debugf("compute commit stats for %s: %v", repo, err)
		return map[string]gavelgit.DiffStat{}
	}
	if err := store.SaveCommitStats(ctx, repo, stats); err != nil {
		logger.Debugf("save commit stats for %s: %v", repo, err)
	}
	return stats
}

// repoStatKey normalizes a workspace directory into the stable cache key the
// commit-stat tables use, matching how the grite cache keys its rows.
func repoStatKey(dir string) string {
	if abs, err := filepath.Abs(dir); err == nil {
		return filepath.Clean(abs)
	}
	return dir
}

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

// todoCommitFilesResponse carries one commit's per-file change summary, each
// enriched with its repomap scope/language for the expanded commit status view.
type todoCommitFilesResponse struct {
	Hash  string                `json:"hash"`
	Files []gavelgit.CommitFile `json:"files"`
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
	dir := s.resolveTodoDir(strings.TrimSpace(r.URL.Query().Get("dir")))
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
