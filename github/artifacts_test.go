package github

import (
	"archive/zip"
	"bytes"
	"errors"
	"testing"
)

func TestExtractJSONFromZip(t *testing.T) {
	t.Run("extracts nested JSON result", func(t *testing.T) {
		data := artifactZip(t, map[string]string{
			"report/index.html":         "<html></html>",
			"report/gavel-results.json": `{"tests":[]}`,
		})
		got, err := extractJSONFromZip(data)
		if err != nil {
			t.Fatalf("extract JSON: %v", err)
		}
		if string(got) != `{"tests":[]}` {
			t.Fatalf("content = %q", got)
		}
	})

	t.Run("classifies artifact without JSON as no results", func(t *testing.T) {
		_, err := extractJSONFromZip(artifactZip(t, map[string]string{
			"report/index.html": "<html></html>",
		}))
		if !errors.Is(err, ErrArtifactResultsNotFound) {
			t.Fatalf("error = %v, want ErrArtifactResultsNotFound", err)
		}
	})
}

func artifactZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		entry, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

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

func TestFindGavelArtifacts(t *testing.T) {
	type want struct {
		StickyID    string
		ArtifactID  int64
		ArtifactURL string
		CommentID   int64
	}
	tests := []struct {
		name     string
		comments []PRComment
		want     []want
	}{
		{
			name: "single shard",
			comments: []PRComment{
				{ID: 1, Body: "Some unrelated comment"},
				{
					ID: 2,
					Body: "<!-- sticky-comment:gavel -->\n\n## Gavel summary\n\n" +
						"| Source | Pass | Fail |\n|---|---:|---:|\n| pkg/foo | 10 | 0 |\n\n" +
						"[View full results](https://github.com/flanksource/gavel/actions/runs/999/artifacts/555)",
				},
			},
			want: []want{{StickyID: "gavel", ArtifactID: 555, ArtifactURL: "https://github.com/flanksource/gavel/actions/runs/999/artifacts/555", CommentID: 2}},
		},
		{
			name: "repository-prefixed gavel header",
			comments: []PRComment{
				{
					ID: 42,
					Body: "<!-- sticky-comment:captain-gavel-test -->\n\n## Gavel summary\n\n" +
						"[View full results](https://github.com/flanksource/captain/actions/runs/30736694536/artifacts/8829852604)",
				},
			},
			want: []want{{
				StickyID:    "captain-gavel-test",
				ArtifactID:  8829852604,
				ArtifactURL: "https://github.com/flanksource/captain/actions/runs/30736694536/artifacts/8829852604",
				CommentID:   42,
			}},
		},
		{
			name: "matrix shards (PR 1926 shape)",
			comments: []PRComment{
				{
					ID: 100,
					Body: "<!-- sticky-comment:gavel-test-pg15 -->\n\n## Gavel summary\n\n" +
						"[View full results](https://github.com/flanksource/duty/actions/runs/1/artifacts/100)",
				},
				{
					ID: 101,
					Body: "<!-- sticky-comment:gavel-e2e -->\n\n## Gavel summary\n\n" +
						"[View full results](https://github.com/flanksource/duty/actions/runs/1/artifacts/101)",
				},
				{
					ID: 102,
					Body: "<!-- sticky-comment:gavel-migrate-head-pg15 -->\n\nGavel crashed before producing results\n\n" +
						"[View full results](https://github.com/flanksource/duty/actions/runs/1/artifacts/102)",
				},
			},
			want: []want{
				{StickyID: "gavel-test-pg15", ArtifactID: 100, ArtifactURL: "https://github.com/flanksource/duty/actions/runs/1/artifacts/100", CommentID: 100},
				{StickyID: "gavel-e2e", ArtifactID: 101, ArtifactURL: "https://github.com/flanksource/duty/actions/runs/1/artifacts/101", CommentID: 101},
				{StickyID: "gavel-migrate-head-pg15", ArtifactID: 102, ArtifactURL: "https://github.com/flanksource/duty/actions/runs/1/artifacts/102", CommentID: 102},
			},
		},
		{
			name: "duplicate sticky id keeps latest, preserves first-seen order",
			comments: []PRComment{
				{
					ID: 10,
					Body: "<!-- sticky-comment:gavel-test -->\n\n" +
						"[View full results](https://github.com/a/b/actions/runs/1/artifacts/100)",
				},
				{
					ID: 11,
					Body: "<!-- sticky-comment:gavel-lint -->\n\n" +
						"[View full results](https://github.com/a/b/actions/runs/1/artifacts/110)",
				},
				{
					ID: 20,
					Body: "<!-- sticky-comment:gavel-test -->\n\n" +
						"[View full results](https://github.com/a/b/actions/runs/2/artifacts/200)",
				},
			},
			want: []want{
				{StickyID: "gavel-test", ArtifactID: 200, ArtifactURL: "https://github.com/a/b/actions/runs/2/artifacts/200", CommentID: 20},
				{StickyID: "gavel-lint", ArtifactID: 110, ArtifactURL: "https://github.com/a/b/actions/runs/1/artifacts/110", CommentID: 11},
			},
		},
		{
			name: "gavel comment without artifact link is skipped",
			comments: []PRComment{
				{
					ID:   1,
					Body: "<!-- sticky-comment:gavel -->\n\nGavel exited with code 1.",
				},
				{
					ID: 2,
					Body: "<!-- sticky-comment:gavel-test -->\n\n" +
						"[View full results](https://github.com/a/b/actions/runs/1/artifacts/22)",
				},
			},
			want: []want{{StickyID: "gavel-test", ArtifactID: 22, ArtifactURL: "https://github.com/a/b/actions/runs/1/artifacts/22", CommentID: 2}},
		},
		{
			name: "non-gavel sticky comments are ignored",
			comments: []PRComment{
				{
					ID: 1,
					Body: "<!-- sticky-comment:codecov -->\nCoverage report\n" +
						"[View full results](https://github.com/a/b/actions/runs/1/artifacts/23)",
				},
				{ID: 2, Body: "LGTM"},
			},
			want: nil,
		},
		{
			name:     "empty comments",
			comments: nil,
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindGavelArtifacts(tt.comments)
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d (got=%+v)", len(got), len(tt.want), got)
			}
			for i, g := range got {
				w := tt.want[i]
				_, wantRunID, _, err := ParseArtifactURL(w.ArtifactURL)
				if err != nil {
					t.Fatalf("[%d] parse expected artifact URL: %v", i, err)
				}
				if g.StickyID != w.StickyID {
					t.Errorf("[%d] StickyID = %q, want %q", i, g.StickyID, w.StickyID)
				}
				if g.RunID != wantRunID {
					t.Errorf("[%d] RunID = %d, want %d", i, g.RunID, wantRunID)
				}
				if g.ArtifactID != w.ArtifactID {
					t.Errorf("[%d] ArtifactID = %d, want %d", i, g.ArtifactID, w.ArtifactID)
				}
				if g.ArtifactURL != w.ArtifactURL {
					t.Errorf("[%d] ArtifactURL = %q, want %q", i, g.ArtifactURL, w.ArtifactURL)
				}
				if g.CommentID != w.CommentID {
					t.Errorf("[%d] CommentID = %d, want %d", i, g.CommentID, w.CommentID)
				}
			}
		})
	}
}
