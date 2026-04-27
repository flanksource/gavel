package github

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestParseArtifactURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantRepo  string
		wantRunID int64
		wantArtID int64
		wantErr   bool
	}{
		{
			name:      "standard URL",
			url:       "https://github.com/flanksource/gavel/actions/runs/9876543210/artifacts/1122334455",
			wantRepo:  "flanksource/gavel",
			wantRunID: 9876543210,
			wantArtID: 1122334455,
		},
		{
			name:    "invalid URL",
			url:     "https://github.com/flanksource/gavel/pull/42",
			wantErr: true,
		},
		{
			name:    "empty",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, runID, artID, err := ParseArtifactURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if repo != tt.wantRepo {
				t.Errorf("repo = %q, want %q", repo, tt.wantRepo)
			}
			if runID != tt.wantRunID {
				t.Errorf("runID = %d, want %d", runID, tt.wantRunID)
			}
			if artID != tt.wantArtID {
				t.Errorf("artifactID = %d, want %d", artID, tt.wantArtID)
			}
		})
	}
}

func TestFindGavelArtifact(t *testing.T) {
	tests := []struct {
		name      string
		comments  []PRComment
		wantID    int64
		wantURL   string
		wantFound bool
	}{
		{
			name: "gavel sticky comment with artifact link",
			comments: []PRComment{
				{ID: 1, Body: "Some unrelated comment"},
				{
					ID: 2,
					Body: "<!-- sticky-comment:gavel -->\n\n## Gavel summary\n\n" +
						"| Source | Pass | Fail |\n|---|---:|---:|\n| pkg/foo | 10 | 0 |\n\n" +
						"[View full results](https://github.com/flanksource/gavel/actions/runs/999/artifacts/555)",
				},
			},
			wantID:    555,
			wantURL:   "https://github.com/flanksource/gavel/actions/runs/999/artifacts/555",
			wantFound: true,
		},
		{
			name: "custom header still matches",
			comments: []PRComment{
				{
					ID: 3,
					Body: "<!-- sticky-comment:gavel-self-test -->\n\n## Gavel summary\n\n" +
						"[View full results](https://github.com/org/repo/actions/runs/1/artifacts/2)",
				},
			},
			wantID:    2,
			wantURL:   "https://github.com/org/repo/actions/runs/1/artifacts/2",
			wantFound: true,
		},
		{
			name: "most recent comment wins",
			comments: []PRComment{
				{
					ID: 10,
					Body: "<!-- sticky-comment:gavel -->\n\n" +
						"[View full results](https://github.com/a/b/actions/runs/1/artifacts/100)",
				},
				{
					ID: 20,
					Body: "<!-- sticky-comment:gavel -->\n\n" +
						"[View full results](https://github.com/a/b/actions/runs/2/artifacts/200)",
				},
			},
			wantID:    200,
			wantURL:   "https://github.com/a/b/actions/runs/2/artifacts/200",
			wantFound: true,
		},
		{
			name: "no gavel comment",
			comments: []PRComment{
				{ID: 1, Body: "LGTM"},
				{ID: 2, Body: "Please fix the tests"},
			},
			wantFound: false,
		},
		{
			name: "gavel comment without artifact link",
			comments: []PRComment{
				{
					ID:   1,
					Body: "<!-- sticky-comment:gavel -->\n\nGavel exited with code 1.",
				},
			},
			wantFound: false,
		},
		{
			name:      "empty comments",
			comments:  nil,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, url, found := FindGavelArtifact(tt.comments)
			if found != tt.wantFound {
				t.Fatalf("found = %v, want %v", found, tt.wantFound)
			}
			if !found {
				return
			}
			if id != tt.wantID {
				t.Errorf("artifactID = %d, want %d", id, tt.wantID)
			}
			if url != tt.wantURL {
				t.Errorf("url = %q, want %q", url, tt.wantURL)
			}
		})
	}
}

func buildZipFixture(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, body := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("zip create %s: %v", name, err)
		}
		if _, err := f.Write([]byte(body)); err != nil {
			t.Fatalf("zip write %s: %v", name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

// TestExtractJSONFilesFromZipReturnsAllJSON proves the multi-file extractor
// keeps every .json entry — single-file artifacts that bundle multiple test
// shards (e.g. `gavel-results.json`, `gavel-bench.json`) need all of them
// merged downstream, not just the first one alphabetically.
func TestExtractJSONFilesFromZipReturnsAllJSON(t *testing.T) {
	zipData := buildZipFixture(t, map[string]string{
		"gavel-results.json": `{"tests":[]}`,
		"gavel-bench.json":   `{"bench":{}}`,
		"gavel.log":          "log line\n",
		"README.txt":         "ignored",
	})
	files, err := extractJSONFilesFromZip(zipData)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 .json files, got %d (%v)", len(files), names(files))
	}
	got := map[string]string{}
	for _, f := range files {
		got[f.Name] = string(f.Body)
	}
	if got["gavel-results.json"] != `{"tests":[]}` {
		t.Errorf("gavel-results.json body = %q", got["gavel-results.json"])
	}
	if got["gavel-bench.json"] != `{"bench":{}}` {
		t.Errorf("gavel-bench.json body = %q", got["gavel-bench.json"])
	}
}

func TestExtractJSONFilesFromZipNoJSON(t *testing.T) {
	zipData := buildZipFixture(t, map[string]string{
		"gavel.log":  "no json here",
		"README.txt": "still nope",
	})
	if _, err := extractJSONFilesFromZip(zipData); err == nil {
		t.Fatal("expected error when zip has no .json file")
	}
}

func names(files []ArtifactFile) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.Name
	}
	return out
}

// TestHasGavelPrefixAcceptsExpectedNames pins down the set of artifact names
// that should match. Anything starting with "gavel" (case-insensitive) is in;
// names that contain "gavel" elsewhere or use other prefixes are out.
func TestHasGavelPrefixAcceptsExpectedNames(t *testing.T) {
	for _, in := range []string{
		"gavel-results", "gavel-bench", "Gavel-Results",
		"gavel", "gavel-self-test", "gavel-results-linux",
	} {
		if !hasGavelPrefix(in) {
			t.Errorf("expected match for %q", in)
		}
	}
	for _, out := range []string{
		"results", "test-results-gavel", "junit-gavel-results", "",
	} {
		if hasGavelPrefix(out) {
			t.Errorf("unexpected match for %q", out)
		}
	}
}

// TestRestArtifactsResponseUnmarshal sanity-checks the wire shape we parse
// out of /repos/{owner}/{name}/actions/runs/{id}/artifacts so a future GitHub
// payload tweak (snake_case key rename, nested workflow_run shape change)
// fails this test instead of silently returning empty results in production.
func TestRestArtifactsResponseUnmarshal(t *testing.T) {
	const payload = `{
	  "total_count": 2,
	  "artifacts": [
	    {
	      "id": 111,
	      "name": "gavel-results",
	      "size_in_bytes": 4096,
	      "archive_download_url": "https://api.github.com/repos/o/r/actions/artifacts/111/zip",
	      "expired": false,
	      "workflow_run": {"id": 99}
	    },
	    {
	      "id": 222,
	      "name": "gavel-bench",
	      "size_in_bytes": 2048,
	      "expired": true,
	      "workflow_run": {"id": 99}
	    }
	  ]
	}`
	var got restArtifactsResponse
	if err := json.Unmarshal([]byte(payload), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", got.TotalCount)
	}
	if len(got.Artifacts) != 2 {
		t.Fatalf("artifacts = %d, want 2", len(got.Artifacts))
	}
	if got.Artifacts[0].Name != "gavel-results" || got.Artifacts[0].WorkflowRun.ID != 99 {
		t.Errorf("artifact 0 = %+v", got.Artifacts[0])
	}
	if !got.Artifacts[1].Expired {
		t.Errorf("artifact 1 should be Expired")
	}
}

// TestGavelArtifactHTMLURL pins down the artifact URL we hand the UI so
// in-app "open in Actions" links don't silently break if HTMLURL is
// refactored.
func TestGavelArtifactHTMLURL(t *testing.T) {
	a := GavelArtifact{ID: 555, RunID: 999}
	want := "https://github.com/owner/name/actions/runs/999/artifacts/555"
	if got := a.HTMLURL("owner/name"); got != want {
		t.Errorf("HTMLURL = %q, want %q", got, want)
	}
	// Confirm the URL is also recognised by the inverse parser — round-trip
	// guards against a regex/fmt drift between the two helpers.
	if !strings.Contains(want, "/runs/999/artifacts/555") {
		t.Fatalf("test URL fixture is malformed: %q", want)
	}
	repo, runID, artID, err := ParseArtifactURL(want)
	if err != nil {
		t.Fatalf("ParseArtifactURL: %v", err)
	}
	if repo != "owner/name" || runID != 999 || artID != 555 {
		t.Errorf("round-trip mismatch: repo=%q run=%d art=%d", repo, runID, artID)
	}
}
