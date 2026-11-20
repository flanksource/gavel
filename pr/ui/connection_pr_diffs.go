// Reads a git connection's pull request's commits, changed files and patches from git, through a
// cache of its repository: the commits since the merge base of its head and base, and the patch
// from that merge base.
package ui

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
)

// connectionPatchLimit is the longest patch a diff response carries before it is truncated.
const connectionPatchLimit = 1 << 20

// SetGitCache reads the commits and patches of pull requests from git connections through cache.
func (s *Server) SetGitCache(cache *git.Cache) { s.gitCache = cache }

// connectionChanges fills info's commits, changed files and line counts: the commits on its head
// since the merge base with its base, and the patch from that merge base.
func connectionChanges(ctx context.Context, cache *git.Cache, remote *git.Repository, info *model.PRInfo) error {
	mergeBase, err := cache.MergeBase(ctx, remote, git.MergeBaseOptions{Base: info.BaseRefName, Head: info.HeadRefName})
	if err != nil {
		return fmt.Errorf("merge base of %s and %s: %w", info.BaseRefName, info.HeadRefName, err)
	}
	commits, err := cache.Log(ctx, remote, git.LogOptions{Base: mergeBase, Head: info.HeadRefName})
	if err != nil {
		return fmt.Errorf("commits of %s: %w", info.HeadRefName, err)
	}
	for _, commit := range commits {
		info.PRCommits = append(info.PRCommits, model.PRCommitInfo{OID: commit.SHA, MessageHeadline: commit.Subject, MessageBody: commit.Body,
			CommittedDate: commit.Date.Format(time.RFC3339), AuthorName: commit.Author.Name, ChangedFiles: len(commit.Files)})
	}
	patch, err := cache.Diff(ctx, git.DiffOptions{Repository: remote, Base: mergeBase, Head: info.HeadRefName, Renames: true})
	if err != nil {
		return fmt.Errorf("patch of %s: %w", info.HeadRefName, err)
	}
	info.PRFiles = patchFiles(patch)
	info.ChangedFiles = len(info.PRFiles)
	for _, file := range info.PRFiles {
		info.Additions += file.Additions
		info.Deletions += file.Deletions
	}
	return nil
}

// patchFiles are the files a unified patch changes, with their added and deleted lines.
func patchFiles(patch string) []model.PRFileInfo {
	var files []model.PRFileInfo
	inHunk := false
	for _, line := range strings.Split(patch, "\n") {
		switch {
		case strings.HasPrefix(line, "diff --git "):
			path := line[strings.LastIndex(line, " b/")+3:]
			files = append(files, model.PRFileInfo{Path: path, ChangeType: "MODIFIED"})
			inHunk = false
		case len(files) == 0:
		case strings.HasPrefix(line, "@@"):
			inHunk = true
		case !inHunk && strings.HasPrefix(line, "new file mode"):
			files[len(files)-1].ChangeType = "ADDED"
		case !inHunk && strings.HasPrefix(line, "deleted file mode"):
			files[len(files)-1].ChangeType = "DELETED"
		case !inHunk && strings.HasPrefix(line, "rename from "):
			files[len(files)-1].ChangeType = "RENAMED"
		case inHunk && strings.HasPrefix(line, "+"):
			files[len(files)-1].Additions++
		case inHunk && strings.HasPrefix(line, "-"):
			files[len(files)-1].Deletions++
		}
	}
	return files
}

// connectionPRInRepo is a listed pull request a git connection brings in repo, whose remote
// reaches the repository's commits.
func (s *Server) connectionPRInRepo(repo string) (github.PRListItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, pr := range s.prs {
		if pr.Provider != "" && pr.Repo == repo {
			return pr, true
		}
	}
	return github.PRListItem{}, false
}

// writeConnectionDiff answers a diff request with the patch diff reads through pr's remote and the
// server's git cache.
func (s *Server) writeConnectionDiff(w http.ResponseWriter, r *http.Request, pr github.PRListItem, response prDiffResponse,
	diff func(ctx context.Context, cache *git.Cache, remote *git.Repository) (string, error)) {
	if s.prRemote == nil || s.gitCache == nil {
		writeJSONError(w, http.StatusBadGateway, fmt.Errorf("this dashboard reads no %s pull request's patches: it has no connection or git cache to read them through", pr.Provider))
		return
	}
	remote, err := s.prRemote(r.Context(), pr)
	if err == nil {
		response.Diff, err = diff(r.Context(), s.gitCache, remote)
	}
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, fmt.Errorf("patch of %s: %w", pr.Repo, err))
		return
	}
	if len(response.Diff) > connectionPatchLimit {
		response.Diff, response.Truncated = response.Diff[:connectionPatchLimit], true
	}
	response.Binary = strings.Contains(response.Diff, "\nBinary files ") || strings.HasPrefix(response.Diff, "Binary files ")
	writeJSON(w, response)
}

// prFilePatch is the patch of one file of pr since the merge base of its head (Source) and base
// (Target) branches.
func prFilePatch(pr github.PRListItem, path string) func(context.Context, *git.Cache, *git.Repository) (string, error) {
	return func(ctx context.Context, cache *git.Cache, remote *git.Repository) (string, error) {
		mergeBase, err := cache.MergeBase(ctx, remote, git.MergeBaseOptions{Base: pr.Target, Head: pr.Source})
		if err != nil {
			return "", err
		}
		return cache.Diff(ctx, git.DiffOptions{Repository: remote, Base: mergeBase, Head: pr.Source, Paths: []string{path}, Renames: true})
	}
}

// commitPatch is the patch sha makes on its first parent.
func commitPatch(sha string) func(context.Context, *git.Cache, *git.Repository) (string, error) {
	return func(ctx context.Context, cache *git.Cache, remote *git.Repository) (string, error) {
		commits, err := cache.History(ctx, remote, sha, git.HistoryOptions{Limit: 1})
		if err != nil {
			return "", err
		}
		if len(commits) == 0 || len(commits[0].Parents) == 0 {
			return "", fmt.Errorf("commit %s has no parent to compare it with", sha)
		}
		return cache.Diff(ctx, git.DiffOptions{Repository: remote, Base: commits[0].Parents[0], Head: sha, Renames: true})
	}
}
