package github

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
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
//
// Prefer FindGavelArtifacts (plural) for surfaces that can show multiple
// gavel runs per PR; this single-result helper exists for legacy callers
// that only need the headline artifact.
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

// GavelArtifact identifies a single uploaded artifact whose name starts with
// the "gavel" prefix on a PR's workflow runs.
type GavelArtifact struct {
	// Name is the upload-artifact name (e.g. "gavel-results", "gavel-bench").
	Name string `json:"name"`
	// ID is the GitHub artifact ID; pass to DownloadArtifact / DownloadArtifactFiles.
	ID int64 `json:"id"`
	// RunID is the workflow run that produced the artifact. Used to group
	// artifacts coming from the same job.
	RunID int64 `json:"runId"`
	// URL is the human-facing artifact URL embedded in the PR comment / job summary.
	URL string `json:"url"`
	// SizeBytes is the compressed size reported by the API; useful for surfacing
	// "too large to fetch" hints in the UI.
	SizeBytes int64 `json:"sizeBytes,omitempty"`
	// Expired is true when GitHub has aged out the artifact and the bytes are no
	// longer downloadable.
	Expired bool `json:"expired,omitempty"`
}

// HTMLURL returns the github.com URL for browsing the artifact in the Actions UI.
func (a GavelArtifact) HTMLURL(repo string) string {
	return fmt.Sprintf("https://github.com/%s/actions/runs/%d/artifacts/%d", repo, a.RunID, a.ID)
}

type restArtifact struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	SizeInBytes        int64  `json:"size_in_bytes"`
	ArchiveDownloadURL string `json:"archive_download_url"`
	Expired            bool   `json:"expired"`
	WorkflowRun        struct {
		ID int64 `json:"id"`
	} `json:"workflow_run"`
}

type restArtifactsResponse struct {
	TotalCount int64          `json:"total_count"`
	Artifacts  []restArtifact `json:"artifacts"`
}

// ListRunArtifacts returns every artifact attached to a single workflow run.
// The result is unfiltered — callers (e.g. FindGavelArtifacts) decide which
// names matter.
func ListRunArtifacts(opts Options, runID int64) ([]GavelArtifact, error) {
	token, err := opts.token()
	if err != nil {
		return nil, fmt.Errorf("cannot list artifacts: %w", err)
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return nil, fmt.Errorf("cannot resolve repo for artifacts: %w", err)
	}

	path := fmt.Sprintf("/repos/%s/actions/runs/%d/artifacts?per_page=100", repo, runID)
	resp, err := cachedGet(context.Background(), token, path, nil)
	if err != nil {
		return nil, fmt.Errorf("list artifacts for run %d: %w", runID, err)
	}

	var payload restArtifactsResponse
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		return nil, fmt.Errorf("parse artifacts list for run %d: %w", runID, err)
	}

	out := make([]GavelArtifact, 0, len(payload.Artifacts))
	for _, a := range payload.Artifacts {
		ga := GavelArtifact{
			Name:      a.Name,
			ID:        a.ID,
			RunID:     a.WorkflowRun.ID,
			SizeBytes: a.SizeInBytes,
			Expired:   a.Expired,
		}
		if ga.RunID == 0 {
			ga.RunID = runID
		}
		ga.URL = ga.HTMLURL(repo)
		out = append(out, ga)
	}
	return out, nil
}

// FindGavelArtifacts walks every workflow run on the PR and returns artifacts
// whose name starts with the "gavel" prefix (case-insensitive). The result is
// de-duplicated by artifact ID.
//
// Discovery uses the Actions API rather than the sticky comment, so a PR with
// N parallel test jobs (matrix builds, multiple OSes, separate bench job…)
// surfaces N artifacts instead of just the most recently posted one.
func FindGavelArtifacts(opts Options, pr *PRInfo) ([]GavelArtifact, error) {
	if pr == nil {
		return nil, fmt.Errorf("nil PR")
	}

	seenRun := make(map[int64]bool)
	var runIDs []int64
	for _, check := range pr.StatusCheckRollup {
		runID, err := ExtractRunID(check.DetailsURL)
		if err != nil || seenRun[runID] {
			continue
		}
		seenRun[runID] = true
		runIDs = append(runIDs, runID)
	}

	seenArtifact := make(map[int64]bool)
	var out []GavelArtifact
	for _, runID := range runIDs {
		artifacts, err := ListRunArtifacts(opts, runID)
		if err != nil {
			logger.Warnf("list artifacts for run %d: %v", runID, err)
			continue
		}
		for _, a := range artifacts {
			if !hasGavelPrefix(a.Name) {
				continue
			}
			if seenArtifact[a.ID] {
				continue
			}
			seenArtifact[a.ID] = true
			out = append(out, a)
		}
	}
	return out, nil
}

func hasGavelPrefix(name string) bool {
	return strings.HasPrefix(strings.ToLower(name), "gavel")
}

// DownloadArtifact downloads a GitHub Actions artifact ZIP and returns the
// first .json file found inside. Use DownloadArtifactFiles when an artifact
// can legitimately contain multiple JSON payloads that should be merged.
func DownloadArtifact(opts Options, artifactID int64) ([]byte, error) {
	files, err := DownloadArtifactFiles(opts, artifactID)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if strings.HasSuffix(strings.ToLower(f.Name), ".json") {
			return f.Body, nil
		}
	}
	return nil, fmt.Errorf("no .json file found in artifact zip")
}

// ArtifactFile is one entry pulled from an artifact zip.
type ArtifactFile struct {
	Name string
	Body []byte
}

// DownloadArtifactFiles downloads an artifact ZIP and returns every .json
// entry inside it. Useful when a single uploaded artifact bundles results
// from multiple test packages or jobs that should be merged.
func DownloadArtifactFiles(opts Options, artifactID int64) ([]ArtifactFile, error) {
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

	return extractJSONFilesFromZip(result.Body)
}

func extractJSONFilesFromZip(data []byte) ([]ArtifactFile, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open artifact zip: %w", err)
	}
	var out []ArtifactFile
	for _, f := range r.File {
		if !strings.HasSuffix(strings.ToLower(f.Name), ".json") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			logger.Warnf("skip zip entry %s: %v", f.Name, err)
			continue
		}
		content, err := io.ReadAll(io.LimitReader(rc, 50<<20)) // 50 MB cap per file
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("read zip entry %s: %w", f.Name, err)
		}
		out = append(out, ArtifactFile{Name: f.Name, Body: content})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no .json file found in artifact zip")
	}
	return out, nil
}
