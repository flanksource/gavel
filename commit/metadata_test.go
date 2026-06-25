package commit

import (
	"strings"
	"testing"
)

func TestApplyCommitMetadata(t *testing.T) {
	const base = "feat: add widget\n\nbody line"

	tests := []struct {
		name      string
		opts      Options
		env       map[string]string
		wantHas   []string
		wantNotIn []string
	}{
		{
			name:      "disabled returns message unchanged",
			opts:      Options{AddMetadata: false, IssueID: "ISSUE-1", SessionID: "sess-1"},
			wantNotIn: []string{TrailerIssueID, trailerSessionID},
		},
		{
			name:    "options values win over env",
			opts:    Options{AddMetadata: true, IssueID: "ISSUE-1", SessionID: "sess-1"},
			env:     map[string]string{EnvIssueID: "ENV-ISSUE", EnvSessionID: "env-sess"},
			wantHas: []string{TrailerIssueID + ": ISSUE-1", trailerSessionID + ": sess-1"},
		},
		{
			name:    "falls back to env vars",
			opts:    Options{AddMetadata: true},
			env:     map[string]string{EnvIssueID: "ENV-ISSUE", EnvSessionID: "env-sess"},
			wantHas: []string{TrailerIssueID + ": ENV-ISSUE", trailerSessionID + ": env-sess"},
		},
		{
			name:    "claude session id fallback",
			opts:    Options{AddMetadata: true},
			env:     map[string]string{EnvClaudeSessionID: "claude-sess"},
			wantHas: []string{trailerSessionID + ": claude-sess"},
		},
		{
			name:      "no values appends nothing",
			opts:      Options{AddMetadata: true},
			wantNotIn: []string{TrailerIssueID, trailerSessionID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			got := applyCommitMetadata(tt.opts, base)
			for _, want := range tt.wantHas {
				if !strings.Contains(got, want) {
					t.Fatalf("expected message to contain %q, got:\n%s", want, got)
				}
			}
			for _, notWant := range tt.wantNotIn {
				if strings.Contains(got, notWant) {
					t.Fatalf("expected message NOT to contain %q, got:\n%s", notWant, got)
				}
			}
			if !strings.HasPrefix(got, "feat: add widget") {
				t.Fatalf("expected original subject preserved, got:\n%s", got)
			}
		})
	}
}

func TestApplyCommitMetadataIdempotent(t *testing.T) {
	opts := Options{AddMetadata: true, IssueID: "ISSUE-1", SessionID: "sess-1"}
	once := applyCommitMetadata(opts, "feat: thing")
	twice := applyCommitMetadata(opts, once)
	if once != twice {
		t.Fatalf("metadata not idempotent:\nonce:\n%s\ntwice:\n%s", once, twice)
	}
	if got := strings.Count(twice, TrailerIssueID+":"); got != 1 {
		t.Fatalf("expected exactly one issue trailer, got %d:\n%s", got, twice)
	}
}
