package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/repomap"
)

// CommitFile is one file changed in a commit: its current path (and previous
// path for a rename), the change kind, and the added/deleted line counts.
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

// CommitFiles returns the per-file change summary for a single commit, parsed
// from its diff and enriched with each file's repomap scope/language. The hash
// is validated before shelling out so untrusted input cannot reach git. Scope
// enrichment is best-effort: a file repomap cannot classify (e.g. a deletion
// whose path no longer resolves) simply carries no chips rather than failing.
func CommitFiles(dir, hash string) ([]CommitFile, error) {
	hash = strings.TrimSpace(hash)
	if !IsValidCommitHash(hash) {
		return nil, fmt.Errorf("invalid commit hash %q", hash)
	}
	// --format= drops the commit message so only the diff body is parsed; -M
	// surfaces renames as rename from/to pairs instead of an add+delete.
	cmd := exec.Command("git", "show", "--format=", "-M", "--patch", hash)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git show %s: %w\nOutput: %s", hash, err, string(out))
	}
	files := parseCommitFiles(string(out))
	enrichCommitFileScopes(dir, files)
	return files, nil
}

// parseCommitFiles splits a unified diff into one CommitFile per file, deriving
// the change kind from the file header (new/deleted/rename) and the +/- counts
// from the hunk body. It is pure so it can be unit-tested without a repository.
func parseCommitFiles(patch string) []CommitFile {
	var files []CommitFile
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
		case strings.HasPrefix(line, "Binary files"):
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
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return "", ""
	}
	return trimDiffPath(fields[2]), trimDiffPath(fields[3])
}

func trimDiffPath(path string) string {
	path = strings.Trim(path, `"`)
	path = strings.TrimPrefix(path, "a/")
	path = strings.TrimPrefix(path, "b/")
	return path
}

// enrichCommitFileScopes annotates each file with its repomap language and
// scopes, in place. Failures (and files repomap has no entry for) are silently
// skipped so the status list still renders for unclassifiable paths. The
// language is dropped from the scope list to avoid a redundant "go · go" chip.
func enrichCommitFileScopes(dir string, files []CommitFile) {
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
