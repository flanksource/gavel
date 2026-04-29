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
				{ID: 1, Body: "<!-- sticky-comment:codecov -->\nCoverage report"},
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
				if g.StickyID != w.StickyID {
					t.Errorf("[%d] StickyID = %q, want %q", i, g.StickyID, w.StickyID)
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
