package github

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/flanksource/commons/logger"
)

var artifactURLPattern = regexp.MustCompile(
	`github\.com/([^/]+/[^/]+)/actions/runs/(\d+)/artifacts/(\d+)`,
)

// ParseArtifactURL extracts the repo, run ID, and artifact ID from a GitHub
// Actions artifact URL like:
//
//	https://github.com/owner/repo/actions/runs/123/artifacts/456
func ParseArtifactURL(url string) (repo string, runID int64, artifactID int64, err error) {
	m := artifactURLPattern.FindStringSubmatch(url)
	if len(m) < 4 {
		return "", 0, 0, fmt.Errorf("cannot parse artifact URL: %q", url)
	}
	runID, _ = strconv.ParseInt(m[2], 10, 64)
	artifactID, _ = strconv.ParseInt(m[3], 10, 64)
	return m[1], runID, artifactID, nil
}

var artifactLinkPattern = regexp.MustCompile(
	`\[View full results\]\((https://[^)]+/actions/runs/\d+/artifacts/\d+)\)`,
)

// FindGavelArtifact scans PR comments for the gavel sticky comment and
// extracts the artifact ID and URL. It returns the most recent match.
func FindGavelArtifact(comments []PRComment) (artifactID int64, artifactURL string, found bool) {
	for i := len(comments) - 1; i >= 0; i-- {
		body := comments[i].Body
		if !strings.Contains(body, "<!-- sticky-comment:") {
			continue
		}
		m := artifactLinkPattern.FindStringSubmatch(body)
		if len(m) < 2 {
			continue
		}
		url := m[1]
		_, _, id, err := ParseArtifactURL(url)
		if err != nil {
			continue
		}
		return id, url, true
	}
	return 0, "", false
}

// DownloadArtifact downloads a GitHub Actions artifact ZIP and extracts
// gavel-results.json from it. The opts.Repo field must be set to the
// owner/repo that owns the artifact.
func DownloadArtifact(opts Options, artifactID int64) ([]byte, error) {
	token, err := opts.token()
	if err != nil {
		return nil, fmt.Errorf("cannot download artifact: %w", err)
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return nil, fmt.Errorf("cannot resolve repo for artifact: %w", err)
	}

	path := fmt.Sprintf("/repos/%s/actions/artifacts/%d/zip", repo, artifactID)
	result, err := cachedGet(context.Background(), token, path, nil)
	if err != nil {
		return nil, fmt.Errorf("download artifact %d: %w", artifactID, err)
	}

	return extractJSONFromZip(result.Body)
}

func extractJSONFromZip(data []byte) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open artifact zip: %w", err)
	}
	for _, f := range r.File {
		if !strings.HasSuffix(f.Name, ".json") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			logger.Warnf("skip zip entry %s: %v", f.Name, err)
			continue
		}
		defer rc.Close()
		content, err := io.ReadAll(io.LimitReader(rc, 50<<20)) // 50 MB cap
		if err != nil {
			return nil, fmt.Errorf("read zip entry %s: %w", f.Name, err)
		}
		return content, nil
	}
	return nil, fmt.Errorf("no .json file found in artifact zip")
}
