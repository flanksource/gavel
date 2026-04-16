package github

import (
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
