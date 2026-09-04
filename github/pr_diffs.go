package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const maxRemoteDiffBytes = 256 * 1024

type DiffPayload struct {
	Diff      string `json:"diff"`
	Truncated bool   `json:"truncated,omitempty"`
	Binary    bool   `json:"binary,omitempty"`
}

type restPRFile struct {
	Filename         string  `json:"filename"`
	PreviousFilename string  `json:"previous_filename"`
	Status           string  `json:"status"`
	Additions        int     `json:"additions"`
	Deletions        int     `json:"deletions"`
	Patch            *string `json:"patch"`
}

// FetchCommitDiff returns the unified diff for a single commit in opts.Repo.
func FetchCommitDiff(opts Options, sha string) (DiffPayload, error) {
	sha = strings.TrimSpace(sha)
	if sha == "" {
		return DiffPayload{}, fmt.Errorf("commit sha is required")
	}
	token, err := opts.token()
	if err != nil {
		return DiffPayload{}, err
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return DiffPayload{}, err
	}
	repoPath, err := repoAPIPath(repo)
	if err != nil {
		return DiffPayload{}, err
	}

	path := fmt.Sprintf("/repos/%s/commits/%s", repoPath, url.PathEscape(sha))
	result, err := cachedGet(context.Background(), token, path, map[string]string{
		"Accept": "application/vnd.github.diff",
	})
	if err != nil {
		return DiffPayload{}, err
	}
	return truncateRemoteDiff(string(result.Body), false), nil
}

// FetchPRFilePatch returns the text patch for one file in a pull request.
func FetchPRFilePatch(opts Options, number int, filename string) (DiffPayload, error) {
	filename = strings.TrimSpace(filename)
	if number <= 0 {
		return DiffPayload{}, fmt.Errorf("pull request number must be positive")
	}
	if filename == "" {
		return DiffPayload{}, fmt.Errorf("file path is required")
	}
	token, err := opts.token()
	if err != nil {
		return DiffPayload{}, err
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return DiffPayload{}, err
	}
	repoPath, err := repoAPIPath(repo)
	if err != nil {
		return DiffPayload{}, err
	}

	for page := 1; page <= 30; page++ {
		path := fmt.Sprintf("/repos/%s/pulls/%d/files?per_page=100&page=%d", repoPath, number, page)
		result, err := cachedGet(context.Background(), token, path, nil)
		if err != nil {
			return DiffPayload{}, err
		}
		var files []restPRFile
		if err := json.Unmarshal(result.Body, &files); err != nil {
			return DiffPayload{}, fmt.Errorf("parse pull request files: %w", err)
		}
		for _, file := range files {
			if file.Filename != filename {
				continue
			}
			diff := formatPRFilePatch(file)
			return truncateRemoteDiff(diff, diff == ""), nil
		}
		if len(files) < 100 {
			break
		}
	}
	return DiffPayload{}, fmt.Errorf("file %q not found in pull request #%d", filename, number)
}

func repoAPIPath(repo string) (string, error) {
	parts := strings.Split(strings.TrimSpace(repo), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("repo must be owner/name")
	}
	return url.PathEscape(parts[0]) + "/" + url.PathEscape(parts[1]), nil
}

func formatPRFilePatch(file restPRFile) string {
	if file.Patch == nil || *file.Patch == "" {
		return ""
	}
	oldPath := file.Filename
	if file.PreviousFilename != "" {
		oldPath = file.PreviousFilename
	}
	var b strings.Builder
	fmt.Fprintf(&b, "diff --git a/%s b/%s\n", oldPath, file.Filename)
	switch strings.ToLower(file.Status) {
	case "added":
		b.WriteString("new file mode 100644\n")
	case "removed", "deleted":
		b.WriteString("deleted file mode 100644\n")
	case "renamed":
		fmt.Fprintf(&b, "rename from %s\nrename to %s\n", oldPath, file.Filename)
	}
	b.WriteString(*file.Patch)
	if !strings.HasSuffix(*file.Patch, "\n") {
		b.WriteByte('\n')
	}
	return b.String()
}

func truncateRemoteDiff(diff string, binary bool) DiffPayload {
	if len(diff) <= maxRemoteDiffBytes {
		return DiffPayload{Diff: diff, Binary: binary}
	}
	return DiffPayload{
		Diff:      truncateRemoteDiffAtLine(diff, maxRemoteDiffBytes) + "\n\n... diff truncated (showing first 256 KB) ...\n",
		Truncated: true,
		Binary:    binary,
	}
}

func truncateRemoteDiffAtLine(s string, max int) string {
	if len(s) <= max {
		return s
	}
	cut := s[:max]
	if nl := strings.LastIndexByte(cut, '\n'); nl > 0 {
		return cut[:nl]
	}
	return cut
}
