package git

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/internal/linecounts"
	"github.com/flanksource/repomap"
)

// CommitFile is one file changed in a commit: its current path (and previous
// path for a rename), the change kind, and the added/deleted line counts.
// Binary files and non-SLOC assets have zero line counts.
// Language and Scopes are the repomap classification of the file's path, left
// empty when repomap has no entry for it. The dashboard renders these as the
// "repomap-based git commit status" rows under an expanded commit.
type CommitFile struct {
	Path         string   `json:"path"`
	PreviousPath string   `json:"previousPath,omitempty"`
	Status       string   `json:"status"` // added | modified | deleted | renamed
	Adds         int      `json:"adds"`
	Dels         int      `json:"dels"`
	Binary       bool     `json:"binary,omitempty"`
	Language     string   `json:"language,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
}

// CommitFiles returns the per-file change summary for a single commit or a
// Base..Head range (the same diff CommitDiff renders, untruncated), enriched
// with each file's repomap scope/language unless opts.NoScopes is set (see
// EnrichCommitFileScopes).
func CommitFiles(dir string, opts CommitDiffOptions) ([]CommitFile, error) {
	out, err := commitDiffOutput(dir, opts)
	if err != nil {
		return nil, err
	}
	files := parseCommitFiles(out)
	if !opts.NoScopes {
		EnrichCommitFileScopes(dir, files)
	}
	return files, nil
}

// parseCommitFiles splits a unified diff into one CommitFile per file, deriving
// the change kind from the file header (new/deleted/rename) and the +/- counts
// from the hunk body. It is pure so it can be unit-tested without a repository.
func parseCommitFiles(patch string) []CommitFile {
	files := []CommitFile{}
	var cur *CommitFile

	flush := func() {
		if cur == nil {
			return
		}
		// Only a genuine rename keeps a previous path; a plain edit's a/ and b/
		// paths are identical and would otherwise read as a self-rename.
		if cur.Status != "renamed" {
			cur.PreviousPath = ""
		}
		if cur.Binary || !linecounts.Supported(cur.Path) {
			cur.Adds, cur.Dels = 0, 0
		}
		files = append(files, *cur)
		cur = nil
	}

	for _, line := range strings.Split(patch, "\n") {
		switch {
		case strings.HasPrefix(line, "diff --git "):
			flush()
			oldPath, newPath := diffGitPaths(line)
			cur = &CommitFile{Path: newPath, PreviousPath: oldPath, Status: "modified"}
		case cur == nil:
			continue
		case strings.HasPrefix(line, "new file"):
			cur.Status = "added"
		case strings.HasPrefix(line, "deleted file"):
			cur.Status = "deleted"
		case strings.HasPrefix(line, "rename from "):
			cur.Status = "renamed"
			cur.PreviousPath = strings.TrimSpace(strings.TrimPrefix(line, "rename from "))
		case strings.HasPrefix(line, "rename to "):
			cur.Status = "renamed"
			cur.Path = strings.TrimSpace(strings.TrimPrefix(line, "rename to "))
		case strings.HasPrefix(line, "Binary files"), line == "GIT binary patch":
			cur.Binary = true
		case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
			cur.Adds++
		case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
			cur.Dels++
		}
	}
	flush()
	return files
}

// diffGitPaths pulls the old and new paths from a "diff --git a/x b/y" header,
// stripping the a/ and b/ prefixes and any surrounding quotes git adds for
// paths with special characters.
func diffGitPaths(line string) (oldPath, newPath string) {
	paths := strings.TrimPrefix(line, "diff --git ")
	if strings.HasPrefix(paths, `"`) {
		quoted, err := strconv.QuotedPrefix(paths)
		if err != nil {
			return "", ""
		}
		return trimDiffPath(quoted), trimDiffPath(strings.TrimSpace(paths[len(quoted):]))
	}
	if split := strings.Index(paths, ` "b/`); split >= 0 {
		return trimDiffPath(paths[:split]), trimDiffPath(paths[split+1:])
	}
	oldPath, newPath, ok := strings.Cut(paths, " b/")
	if !ok {
		return "", ""
	}
	return trimDiffPath(oldPath), newPath
}

func trimDiffPath(path string) string {
	if strings.HasPrefix(path, `"`) {
		unquoted, err := strconv.Unquote(path)
		if err != nil {
			return ""
		}
		path = unquoted
	}
	if strings.HasPrefix(path, "a/") || strings.HasPrefix(path, "b/") {
		return path[2:]
	}
	return path
}

// EnrichCommitFileScopes annotates each file of the repository at dir with its
// repomap language and scopes, in place. It is best-effort: a file repomap
// cannot classify (e.g. a deletion whose path no longer resolves) carries no
// chips rather than failing the list. The language is dropped from the scope
// list to avoid a redundant "go · go" chip.
func EnrichCommitFileScopes(dir string, files []CommitFile) {
	for i := range files {
		fm, err := repomap.GetFileMap(filepath.Join(dir, files[i].Path), "")
		if err != nil || fm == nil {
			continue
		}
		files[i].Language = fm.Language
		for _, s := range fm.Scopes {
			scope := strings.TrimSpace(string(s))
			if scope == "" || scope == fm.Language {
				continue
			}
			files[i].Scopes = append(files[i].Scopes, scope)
		}
	}
}
