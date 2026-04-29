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

var stickyIDPattern = regexp.MustCompile(`<!-- sticky-comment:(gavel[^\s>]*) -->`)

// GavelArtifact identifies one gavel sticky comment on a PR — typically
// one per matrix shard (e.g. gavel-test-pg15, gavel-e2e). A single PR can
// have many of these.
type GavelArtifact struct {
	StickyID    string
	ArtifactID  int64
	ArtifactURL string
	CommentID   int64
}

// FindGavelArtifacts scans PR comments and returns one GavelArtifact per
// distinct sticky id. When the same id appears more than once (gavel
// rewrites the sticky comment on every push), the most recent occurrence
// wins. Order is determined by first appearance in the comment list so
// the UI renders shards in a stable, source-controlled order.
func FindGavelArtifacts(comments []PRComment) []GavelArtifact {
	type slot struct {
		idx int
		art GavelArtifact
	}
	byID := make(map[string]*slot)
	order := make([]string, 0)
	for _, c := range comments {
		body := c.Body
		sm := stickyIDPattern.FindStringSubmatch(body)
		if len(sm) < 2 {
			continue
		}
		stickyID := sm[1]
		am := artifactLinkPattern.FindStringSubmatch(body)
		if len(am) < 2 {
			continue
		}
		url := am[1]
		_, _, id, err := ParseArtifactURL(url)
		if err != nil {
			continue
		}
		art := GavelArtifact{
			StickyID:    stickyID,
			ArtifactID:  id,
			ArtifactURL: url,
			CommentID:   c.ID,
		}
		if existing, ok := byID[stickyID]; ok {
			existing.art = art
			continue
		}
		byID[stickyID] = &slot{idx: len(order), art: art}
		order = append(order, stickyID)
	}
	out := make([]GavelArtifact, len(order))
	for i, id := range order {
		out[i] = byID[id].art
	}
	return out
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
