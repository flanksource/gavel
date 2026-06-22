package todos

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

func TestGriteProviderListMapsStatusesAndLabels(t *testing.T) {
	runner := func(ctx context.Context, workDir, binary string, args ...string) ([]byte, error) {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "--state open"):
			return []byte(`{
				"ok": true,
				"data": {
					"issues": [{
						"issue_id": "open123456789",
						"title": "Fix open issue",
						"state": "open",
						"labels": ["status:failed", "priority:high"],
						"updated_ts": 1782104417615
					}]
				}
			}`), nil
		case strings.Contains(joined, "--state closed"):
			return []byte(`{
				"ok": true,
				"data": {
					"issues": [{
						"issue_id": "closed123456",
						"title": "Closed issue",
						"state": "closed",
						"labels": []
					}]
				}
			}`), nil
		default:
			t.Fatalf("unexpected args: %v", args)
		}
		return nil, nil
	}

	provider := &GriteProvider{WorkDir: "/repo", Binary: "grite", Runner: runner}
	got, err := provider.List(context.Background(), DiscoveryFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 TODOs, got %d", len(got))
	}

	open := got[0]
	if open.ID != "open123456789" || open.Status != types.StatusFailed || open.Priority != types.PriorityHigh {
		t.Fatalf("unexpected open issue mapping: %+v", open)
	}
	row := open.PrettyRow(nil)
	if row["ID"].String() != "open1234" {
		t.Fatalf("expected short id column, got %#v", row["ID"].String())
	}
	closed := got[1]
	if closed.ID != "closed123456" || closed.Status != types.StatusCompleted || closed.Priority != types.PriorityMedium {
		t.Fatalf("unexpected closed issue mapping: %+v", closed)
	}
}

func TestGriteProviderGetParsesIssueBody(t *testing.T) {
	runner := func(ctx context.Context, workDir, binary string, args ...string) ([]byte, error) {
		if strings.Join(args, " ") != "issue show abc123 --json" {
			t.Fatalf("unexpected args: %v", args)
		}
		return []byte(`{
			"ok": true,
			"data": {
				"issue": {
					"issue_id": "abc123",
					"title": "Run verification",
					"state": "open",
					"labels": ["status:pending"]
				},
				"events": [
					{
						"event_id": "created123456789",
						"actor": "agent",
						"ts_unix_ms": 1782107374959,
						"kind": {
							"IssueCreated": {
								"title": "Run verification",
								"body": "## Verification\n\n` + "```bash\\necho ok\\n```" + `\n"
							}
						}
					},
					{
						"event_id": "comment123456789",
						"actor": "reviewer",
						"ts_unix_ms": 1782107384557,
						"kind": {
							"CommentAdded": {
								"body": "Please include history"
							}
						}
					}
				]
			}
		}`), nil
	}

	provider := &GriteProvider{WorkDir: "/repo", Binary: "grite", Runner: runner}
	todo, err := provider.Get(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if todo.ID != "abc123" || todo.Provider != ProviderGrite || todo.Status != types.StatusPending {
		t.Fatalf("unexpected TODO identity: %+v", todo)
	}
	if len(todo.Verification) != 1 {
		t.Fatalf("expected one verification section, got %d", len(todo.Verification))
	}
	if !strings.Contains(todo.MarkdownBody, "echo ok") {
		t.Fatalf("markdown body was not preserved: %q", todo.MarkdownBody)
	}
	if len(todo.ProviderEvents) != 2 {
		t.Fatalf("expected provider event history, got %#v", todo.ProviderEvents)
	}
	if todo.ProviderEvents[0].ShortID != "created1" || todo.ProviderEvents[0].Kind != "IssueCreated" {
		t.Fatalf("unexpected created event: %#v", todo.ProviderEvents[0])
	}
	if todo.ProviderEvents[1].Kind != "CommentAdded" || todo.ProviderEvents[1].Body != "Please include history" {
		t.Fatalf("unexpected comment event: %#v", todo.ProviderEvents[1])
	}
	details := todo.PrettyDetailed().String()
	for _, want := range []string{"Issue Body", "## Verification", "Event History", "CommentAdded", "Please include history"} {
		if !strings.Contains(details, want) {
			t.Fatalf("expected detailed output to contain %q, got:\n%s", want, details)
		}
	}
}

func TestGriteProviderCreateUsesLabelsAndFetchesDetail(t *testing.T) {
	var calls []string
	runner := func(ctx context.Context, workDir, binary string, args ...string) ([]byte, error) {
		joined := strings.Join(args, " ")
		calls = append(calls, joined)
		switch joined {
		case "issue create --title Fix workspace --body Body text --label priority:high --label status:in_progress --json":
			return []byte(`{"ok": true, "data": {"issue_id": "new123456789"}}`), nil
		case "issue show new123456789 --json":
			return []byte(`{
				"ok": true,
				"data": {
					"issue": {
						"issue_id": "new123456789",
						"title": "Fix workspace",
						"state": "open",
						"labels": ["priority:high", "status:in_progress"]
					},
					"events": [{
						"event_id": "created123",
						"actor": "agent",
						"kind": {"IssueCreated": {"title": "Fix workspace", "body": "Body text"}}
					}]
				}
			}`), nil
		default:
			t.Fatalf("unexpected args: %v", args)
		}
		return nil, nil
	}

	provider := &GriteProvider{WorkDir: "/repo", Binary: "grite", Runner: runner}
	todo, err := provider.Create(context.Background(), CreateRequest{
		Title:    "Fix workspace",
		Body:     "Body text",
		Priority: types.PriorityHigh,
		Status:   types.StatusInProgress,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if todo.ID != "new123456789" || todo.Status != types.StatusInProgress || todo.Priority != types.PriorityHigh {
		t.Fatalf("unexpected created TODO: %+v", todo)
	}
	want := []string{
		"issue create --title Fix workspace --body Body text --label priority:high --label status:in_progress --json",
		"issue show new123456789 --json",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls mismatch\nwant: %#v\n got: %#v", want, calls)
	}
}

func TestGriteProviderDeleteClosesIssue(t *testing.T) {
	var calls []string
	runner := func(ctx context.Context, workDir, binary string, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return []byte(`{"ok": true, "data": {}}`), nil
	}

	provider := &GriteProvider{WorkDir: "/repo", Binary: "grite", Runner: runner}
	todo := &types.TODO{ID: "abc123", Provider: ProviderGrite, ProviderState: "open"}
	if err := provider.Delete(context.Background(), todo); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if !reflect.DeepEqual(calls, []string{"issue close abc123 --json"}) {
		t.Fatalf("calls = %#v", calls)
	}
	if todo.ProviderState != "closed" || todo.Status != types.StatusCompleted {
		t.Fatalf("todo state not updated: %+v", todo)
	}
}

func TestGriteProviderUpdateStateUsesStatusLabels(t *testing.T) {
	var calls []string
	runner := func(ctx context.Context, workDir, binary string, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return []byte(`{"ok": true, "data": {}}`), nil
	}

	provider := &GriteProvider{WorkDir: "/repo", Binary: "grite", Runner: runner}
	todo := &types.TODO{
		ID:            "abc123",
		Provider:      ProviderGrite,
		ProviderState: "open",
		Labels:        []string{"status:pending", "priority:high"},
	}
	failed := types.StatusFailed
	if err := provider.UpdateState(context.Background(), todo, StateUpdate{Status: &failed}); err != nil {
		t.Fatalf("UpdateState failed: %v", err)
	}

	want := []string{
		"issue label remove abc123 --label status:pending --json",
		"issue label add abc123 --label status:failed --json",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls mismatch\nwant: %#v\n got: %#v", want, calls)
	}
	if !hasLabel(todo.Labels, "status:failed") || hasLabel(todo.Labels, "status:pending") {
		t.Fatalf("labels not updated: %#v", todo.Labels)
	}
}

func TestGriteProviderUpdateStateClosesCompletedIssue(t *testing.T) {
	var calls []string
	runner := func(ctx context.Context, workDir, binary string, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return []byte(`{"ok": true, "data": {}}`), nil
	}

	provider := &GriteProvider{WorkDir: "/repo", Binary: "grite", Runner: runner}
	todo := &types.TODO{
		ID:            "abc123",
		Provider:      ProviderGrite,
		ProviderState: "open",
		Labels:        []string{"status:failed"},
	}
	completed := types.StatusCompleted
	if err := provider.UpdateState(context.Background(), todo, StateUpdate{Status: &completed}); err != nil {
		t.Fatalf("UpdateState failed: %v", err)
	}

	want := []string{
		"issue close abc123 --json",
		"issue label remove abc123 --label status:failed --json",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls mismatch\nwant: %#v\n got: %#v", want, calls)
	}
	if todo.ProviderState != "closed" {
		t.Fatalf("expected closed state, got %q", todo.ProviderState)
	}
}

func TestDecodeGriteReturnsCommandError(t *testing.T) {
	_, err := decodeGrite[griteIssueListData]([]byte(`{"ok": false, "error": {"code": "bad", "message": "nope"}}`))
	if err == nil || !strings.Contains(err.Error(), "nope") {
		t.Fatalf("expected grite error, got %v", err)
	}
}
