package git

import (
	"testing"

	"github.com/flanksource/gavel/models"
)

func TestDedupeCommitPrefix(t *testing.T) {
	tests := []struct {
		name        string
		ctype       models.CommitType
		scope       models.ScopeType
		subject     string
		wantType    models.CommitType
		wantScope   models.ScopeType
		wantSubject string
	}{
		{
			name:        "echoed type prefix is stripped",
			ctype:       models.CommitTypeChore,
			subject:     "chore: update dependencies to latest versions",
			wantType:    models.CommitTypeChore,
			wantSubject: "update dependencies to latest versions",
		},
		{
			name:        "echoed type+scope prefix is stripped",
			ctype:       models.CommitTypeFeat,
			scope:       models.ScopeType("api"),
			subject:     "feat(api): add endpoint",
			wantType:    models.CommitTypeFeat,
			wantScope:   models.ScopeType("api"),
			wantSubject: "add endpoint",
		},
		{
			name:        "prefix fills blank type and scope",
			subject:     "fix(auth): handle nil session",
			wantType:    models.CommitTypeFix,
			wantScope:   models.ScopeType("auth"),
			wantSubject: "handle nil session",
		},
		{
			name:        "clean subject is left untouched",
			ctype:       models.CommitTypeChore,
			subject:     "update dependencies to latest versions",
			wantType:    models.CommitTypeChore,
			wantSubject: "update dependencies to latest versions",
		},
		{
			name:        "unrecognised leading token is not treated as a prefix",
			ctype:       models.CommitTypeDocs,
			subject:     "Note: remember to bump version",
			wantType:    models.CommitTypeDocs,
			wantSubject: "Note: remember to bump version",
		},
		{
			name:        "resolved type wins over echoed prefix type",
			ctype:       models.CommitTypeFeat,
			subject:     "fix: correct the thing",
			wantType:    models.CommitTypeFeat,
			wantSubject: "correct the thing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotScope, gotSubject := dedupeCommitPrefix(tt.ctype, tt.scope, tt.subject)
			if gotType != tt.wantType {
				t.Errorf("type = %q, want %q", gotType, tt.wantType)
			}
			if gotScope != tt.wantScope {
				t.Errorf("scope = %q, want %q", gotScope, tt.wantScope)
			}
			if gotSubject != tt.wantSubject {
				t.Errorf("subject = %q, want %q", gotSubject, tt.wantSubject)
			}
		})
	}
}
