package todos

import (
	"context"
	"encoding/json"
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

func TestGriteProviderEditUpdatesTitleAndBody(t *testing.T) {
	var calls []string
	runner := func(ctx context.Context, workDir, binary string, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return []byte(`{"ok": true, "data": {}}`), nil
	}

	provider := &GriteProvider{WorkDir: "/repo", Binary: "grite", Runner: runner}
	todo := &types.TODO{ID: "abc123", Provider: ProviderGrite, ProviderState: "open"}
	title := "Revised title"
	body := "Revised body"
	if err := provider.Edit(context.Background(), todo, EditRequest{Title: &title, Body: &body}); err != nil {
		t.Fatalf("Edit failed: %v", err)
	}
	want := []string{"issue update abc123 --title Revised title --body Revised body --json"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls mismatch\nwant: %#v\n got: %#v", want, calls)
	}
	if todo.Title != "Revised title" || todo.MarkdownBody != "Revised body" {
		t.Fatalf("in-memory todo not updated: %+v", todo)
	}
}

func TestGriteProviderEditTitleOnlyOmitsBodyFlag(t *testing.T) {
	var calls []string
	runner := func(ctx context.Context, workDir, binary string, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return []byte(`{"ok": true, "data": {}}`), nil
	}

	provider := &GriteProvider{WorkDir: "/repo", Binary: "grite", Runner: runner}
	todo := &types.TODO{ID: "abc123", Provider: ProviderGrite, ProviderState: "open", MarkdownBody: "unchanged"}
	title := "Only the title"
	if err := provider.Edit(context.Background(), todo, EditRequest{Title: &title}); err != nil {
		t.Fatalf("Edit failed: %v", err)
	}
	want := []string{"issue update abc123 --title Only the title --json"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls mismatch\nwant: %#v\n got: %#v", want, calls)
	}
	if todo.MarkdownBody != "unchanged" {
		t.Fatalf("body should be untouched, got %q", todo.MarkdownBody)
	}
}

func TestGriteProviderEditRejectsEmptyEdit(t *testing.T) {
	provider := &GriteProvider{WorkDir: "/repo", Binary: "grite", Runner: func(context.Context, string, string, ...string) ([]byte, error) {
		t.Fatal("runner should not be called for an empty edit")
		return nil, nil
	}}
	todo := &types.TODO{ID: "abc123", Provider: ProviderGrite}
	if err := provider.Edit(context.Background(), todo, EditRequest{}); err == nil {
		t.Fatal("expected error for empty edit")
	}
}

func TestGriteProviderCommentPostsBody(t *testing.T) {
	var calls []string
	runner := func(ctx context.Context, workDir, binary string, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return []byte(`{"ok": true, "data": {}}`), nil
	}

	provider := &GriteProvider{WorkDir: "/repo", Binary: "grite", Runner: runner}
	todo := &types.TODO{ID: "abc123", Provider: ProviderGrite, ProviderState: "open"}
	if err := provider.Comment(context.Background(), todo, "looks good to me"); err != nil {
		t.Fatalf("Comment failed: %v", err)
	}
	want := []string{"issue comment abc123 --body looks good to me --json"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls mismatch\nwant: %#v\n got: %#v", want, calls)
	}

	if err := provider.Comment(context.Background(), todo, "   "); err == nil {
		t.Fatal("expected error for blank comment")
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

func TestGriteProviderUpdateStateSwapsPriorityLabel(t *testing.T) {
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
		Labels:        []string{"priority:high", "status:pending"},
	}
	low := types.PriorityLow
	if err := provider.UpdateState(context.Background(), todo, StateUpdate{Priority: &low}); err != nil {
		t.Fatalf("UpdateState failed: %v", err)
	}

	want := []string{
		"issue label remove abc123 --label priority:high --json",
		"issue label add abc123 --label priority:low --json",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls mismatch\nwant: %#v\n got: %#v", want, calls)
	}
	if todo.Priority != types.PriorityLow {
		t.Fatalf("expected priority low, got %q", todo.Priority)
	}
	if !hasLabel(todo.Labels, "priority:low") || hasLabel(todo.Labels, "priority:high") {
		t.Fatalf("labels not updated: %#v", todo.Labels)
	}
	// status:pending untouched when only priority changes.
	if !hasLabel(todo.Labels, "status:pending") {
		t.Fatalf("status label should be preserved: %#v", todo.Labels)
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

func TestGriteProviderSupportsDraftAndVerifiedStatusLabels(t *testing.T) {
	if got := statusFromGriteIssue("open", []string{"status:draft"}); got != types.StatusDraft {
		t.Fatalf("draft label mapped to %q, want draft", got)
	}
	if got := statusFromGriteIssue("open", []string{"status:verified"}); got != types.StatusVerified {
		t.Fatalf("verified label mapped to %q, want verified", got)
	}

	labels := griteCreateLabels(types.PriorityLow, types.StatusVerified)
	want := []string{"priority:low", "status:verified"}
	if !reflect.DeepEqual(labels, want) {
		t.Fatalf("create labels mismatch\nwant: %#v\n got: %#v", want, labels)
	}
}

func TestProviderEventsCaptureLabelDetail(t *testing.T) {
	kind := func(name, payload string) map[string]json.RawMessage {
		return map[string]json.RawMessage{name: json.RawMessage(payload)}
	}
	events := []griteEvent{
		{EventID: "2dfd815d4bd2", Actor: "agent", TimestampMS: 1782148373153, Kind: kind("LabelRemoved", `{"label": "status:pending"}`)},
		{EventID: "5ee3e0c5d06e", Actor: "agent", TimestampMS: 1782148373193, Kind: kind("LabelAdded", `{"label": "status:in_progress"}`)},
	}

	got := providerEventsFromGriteEvents(events)
	if len(got) != 1 {
		t.Fatalf("expected the remove/add pair to merge into 1 event, got %#v", got)
	}
	if got[0].Kind != "LabelChanged" || got[0].OldLabel != "status:pending" || got[0].NewLabel != "status:in_progress" {
		t.Fatalf("LabelChanged not merged: %#v", got[0])
	}

	details := types.TODO{ProviderEvents: got}.PrettyDetailed().String()
	for _, want := range []string{"LabelChanged", "status:pending", "status:in_progress"} {
		if !strings.Contains(details, want) {
			t.Fatalf("expected detailed output to contain %q, got:\n%s", want, details)
		}
	}
}

func TestProviderEventsMergeLabelChangeKeepsUnrelated(t *testing.T) {
	kind := func(name, payload string) map[string]json.RawMessage {
		return map[string]json.RawMessage{name: json.RawMessage(payload)}
	}
	events := []griteEvent{
		// priority remove/add in the same namespace -> merge.
		{EventID: "f5b668b8aaaa", Actor: "agent", TimestampMS: 1782148373100, Kind: kind("LabelRemoved", `{"label": "priority:medium"}`)},
		{EventID: "3d81c76dbbbb", Actor: "agent", TimestampMS: 1782148373110, Kind: kind("LabelAdded", `{"label": "priority:high"}`)},
		// remove + add across different namespaces -> stay separate.
		{EventID: "cccccccccccc", Actor: "agent", TimestampMS: 1782148373200, Kind: kind("LabelRemoved", `{"label": "status:pending"}`)},
		{EventID: "dddddddddddd", Actor: "agent", TimestampMS: 1782148373210, Kind: kind("LabelAdded", `{"label": "area:ui"}`)},
		// lone add with no preceding remove -> stays LabelAdded.
		{EventID: "eeeeeeeeeeee", Actor: "agent", TimestampMS: 1782148373300, Kind: kind("LabelAdded", `{"label": "priority:low"}`)},
	}

	got := providerEventsFromGriteEvents(events)
	if len(got) != 4 {
		t.Fatalf("expected 4 events (one merged), got %d: %#v", len(got), got)
	}
	if got[0].Kind != "LabelChanged" || got[0].OldLabel != "priority:medium" || got[0].NewLabel != "priority:high" {
		t.Fatalf("priority change not merged: %#v", got[0])
	}
	if got[1].Kind != "LabelRemoved" || got[1].Label != "status:pending" {
		t.Fatalf("cross-namespace remove should stay separate: %#v", got[1])
	}
	if got[2].Kind != "LabelAdded" || got[2].Label != "area:ui" {
		t.Fatalf("cross-namespace add should stay separate: %#v", got[2])
	}
	if got[3].Kind != "LabelAdded" || got[3].Label != "priority:low" {
		t.Fatalf("lone add should stay LabelAdded: %#v", got[3])
	}
}

func TestGriteProviderUpdateStateSwapsSessionLabel(t *testing.T) {
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
		Labels:        []string{"session:old-session", "status:in_progress"},
	}
	session := "new-session-id"
	if err := provider.UpdateState(context.Background(), todo, StateUpdate{SessionID: &session}); err != nil {
		t.Fatalf("UpdateState failed: %v", err)
	}

	want := []string{
		"issue label remove abc123 --label session:old-session --json",
		"issue label add abc123 --label session:new-session-id --json",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls mismatch\nwant: %#v\n got: %#v", want, calls)
	}
	if !hasLabel(todo.Labels, "session:new-session-id") || hasLabel(todo.Labels, "session:old-session") {
		t.Fatalf("session label not swapped: %#v", todo.Labels)
	}
	// status:in_progress untouched when only the session id changes.
	if !hasLabel(todo.Labels, "status:in_progress") {
		t.Fatalf("status label should be preserved: %#v", todo.Labels)
	}
}

func TestGriteProviderUpdateStateIgnoresEmptySessionID(t *testing.T) {
	var calls []string
	runner := func(ctx context.Context, workDir, binary string, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return []byte(`{"ok": true, "data": {}}`), nil
	}

	provider := &GriteProvider{WorkDir: "/repo", Binary: "grite", Runner: runner}
	todo := &types.TODO{ID: "abc123", Provider: ProviderGrite, ProviderState: "open"}
	empty := ""
	if err := provider.UpdateState(context.Background(), todo, StateUpdate{SessionID: &empty}); err != nil {
		t.Fatalf("UpdateState failed: %v", err)
	}
	if len(calls) != 0 {
		t.Fatalf("empty session id should not touch labels, got calls: %#v", calls)
	}
}

func TestGriteProviderSaveAttemptOmitsTranscript(t *testing.T) {
	var commentBody string
	runner := func(ctx context.Context, workDir, binary string, args ...string) ([]byte, error) {
		if len(args) >= 2 && args[0] == "issue" && args[1] == "comment" {
			for i, a := range args {
				if a == "--body" && i+1 < len(args) {
					commentBody = args[i+1]
				}
			}
		}
		return []byte(`{"ok": true, "data": {}}`), nil
	}

	provider := &GriteProvider{WorkDir: "/repo", Binary: "grite", Runner: runner}
	todo := &types.TODO{
		ID:              "abc123",
		Provider:        ProviderGrite,
		Labels:          []string{"session:sess-xyz", "status:in_progress"},
		TODOFrontmatter: types.TODOFrontmatter{Attempts: 1},
	}
	result := &ExecutionResult{
		Success:      true,
		ExecutorName: "cmux-claude",
		Transcript: &ExecutionTranscript{Entries: []TranscriptEntry{
			{Type: EntryText, Role: "executor", Content: "SECRET_TRANSCRIPT_CONTENT"},
		}},
	}
	if err := provider.SaveAttempt(context.Background(), todo, result); err != nil {
		t.Fatalf("SaveAttempt failed: %v", err)
	}
	if commentBody == "" {
		t.Fatal("expected an attempt comment to be posted")
	}
	if strings.Contains(commentBody, "Transcript") || strings.Contains(commentBody, "SECRET_TRANSCRIPT_CONTENT") {
		t.Fatalf("attempt comment must not embed the transcript:\n%s", commentBody)
	}
	if !strings.Contains(commentBody, "sess-xyz") {
		t.Fatalf("attempt comment should record the session id for re-parsing:\n%s", commentBody)
	}
}

func TestSessionIDFromLabels(t *testing.T) {
	if got := sessionIDFromLabels([]string{"status:pending", "session:s1", "priority:high"}); got != "s1" {
		t.Fatalf("sessionIDFromLabels = %q, want s1", got)
	}
	if got := sessionIDFromLabels([]string{"status:pending"}); got != "" {
		t.Fatalf("sessionIDFromLabels = %q, want empty", got)
	}
}

func TestTodoFromGriteIssuePopulatesSession(t *testing.T) {
	todo := todoFromGriteIssue(griteIssue{
		IssueID: "abc123",
		Title:   "Fix it",
		State:   "open",
		Labels:  []string{"session:sess-9", "status:in_progress"},
	}, "/repo")
	if todo.LLM == nil || todo.LLM.SessionId != "sess-9" {
		t.Fatalf("expected session id carried onto todo, got %+v", todo.LLM)
	}
}

func TestDecodeGriteReturnsCommandError(t *testing.T) {
	_, err := decodeGrite[griteIssueListData]([]byte(`{"ok": false, "error": {"code": "bad", "message": "nope"}}`))
	if err == nil || !strings.Contains(err.Error(), "nope") {
		t.Fatalf("expected grite error, got %v", err)
	}
}
