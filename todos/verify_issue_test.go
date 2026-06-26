package todos

import (
	"reflect"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

func TestBuildIssueContextSplitsChecksAndCriteria(t *testing.T) {
	todo := &types.TODO{
		ID:           "ISSUE-1",
		MarkdownBody: "Implement NDJSON export",
		TODOFrontmatter: types.TODOFrontmatter{
			Title: "NDJSON export",
			LLM:   &types.LLM{SessionId: "sess-42"},
		},
		AcceptanceCriteria: []types.AcceptanceCriterion{
			{CheckID: "tests-added", Text: "New logic includes tests"},
			{Text: "Streams rows for payloads over 10k"},
		},
		ProviderEvents: []types.ProviderEvent{
			{Kind: "CommentAdded", Actor: "alice", Body: "please paginate"},
			{Kind: "LabelAdded", Label: "priority:high"},
		},
	}

	ic := BuildIssueContext(todo, []string{"abc123"})

	if ic.Title != "NDJSON export" || ic.SessionID != "sess-42" || ic.ID != "ISSUE-1" {
		t.Errorf("unexpected header fields: %+v", ic)
	}
	if !reflect.DeepEqual(ic.CheckIDs, []string{"tests-added"}) {
		t.Errorf("CheckIDs = %v, want [tests-added]", ic.CheckIDs)
	}
	if !reflect.DeepEqual(ic.Criteria, []string{"Streams rows for payloads over 10k"}) {
		t.Errorf("Criteria = %v", ic.Criteria)
	}
	if len(ic.Comments) != 1 || ic.Comments[0].Author != "alice" || ic.Comments[0].Body != "please paginate" {
		t.Errorf("Comments = %+v, want one CommentAdded", ic.Comments)
	}
	if !reflect.DeepEqual(ic.CommitSHAs, []string{"abc123"}) {
		t.Errorf("CommitSHAs = %v", ic.CommitSHAs)
	}
}

func TestResolveIssueCommitsPrefersKnown(t *testing.T) {
	known := []string{"sha1", "sha2"}
	got, err := ResolveIssueCommits("/nonexistent", "ISSUE-1", known)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, known) {
		t.Errorf("ResolveIssueCommits with known = %v, want %v", got, known)
	}
}
