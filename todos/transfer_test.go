package todos

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

func TestTransferMovesTodoBetweenFileWorkspaces(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	source := NewFileProvider(srcDir, "")
	target := NewFileProvider(dstDir, "")

	original, err := source.Create(context.Background(), CreateRequest{
		Title:    "Wire transfer endpoint",
		Body:     "Move todos between project workspaces.",
		Priority: types.PriorityHigh,
		Status:   types.StatusInProgress,
	})
	if err != nil {
		t.Fatalf("seed Create failed: %v", err)
	}

	moved, err := Transfer(context.Background(), source, target, original.FilePath)
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

	if moved.Title != "Wire transfer endpoint" || moved.Priority != types.PriorityHigh || moved.Status != types.StatusInProgress {
		t.Fatalf("transferred todo lost fields: %+v", moved)
	}
	if !strings.HasPrefix(moved.FilePath, dstDir) {
		t.Fatalf("transferred todo not written to target dir %q: %s", dstDir, moved.FilePath)
	}
	if !strings.Contains(moved.MarkdownBody, "between project workspaces") {
		t.Fatalf("transferred body not preserved: %q", moved.MarkdownBody)
	}

	// Source is emptied, target holds exactly the moved todo.
	if _, err := os.Stat(original.FilePath); !os.IsNotExist(err) {
		t.Fatalf("expected source todo removed, stat err=%v", err)
	}
	targetItems, err := target.List(context.Background(), DiscoveryFilters{})
	if err != nil {
		t.Fatalf("target List failed: %v", err)
	}
	if len(targetItems) != 1 || targetItems[0].Title != "Wire transfer endpoint" {
		t.Fatalf("unexpected target contents: %+v", targetItems)
	}
}

func TestTransferGriteIssueIntoFileWorkspace(t *testing.T) {
	var closed bool
	runner := func(_ context.Context, _, _ string, args ...string) ([]byte, error) {
		joined := strings.Join(args, " ")
		switch {
		case strings.HasPrefix(joined, "issue show"):
			return []byte(`{
				"ok": true,
				"data": {
					"issue": {
						"issue_id": "abc123def456",
						"title": "Cross-provider move",
						"state": "open",
						"labels": ["status:pending", "priority:low"]
					},
					"events": [{
						"event_id": "e1",
						"kind": {"IssueCreated": {"title": "Cross-provider move", "body": "Grite issue body to carry over."}}
					}]
				}
			}`), nil
		case strings.HasPrefix(joined, "issue close"):
			closed = true
			return []byte(`{"ok": true, "data": {}}`), nil
		default:
			t.Fatalf("unexpected grite args: %v", args)
			return nil, nil
		}
	}
	source := &GriteProvider{WorkDir: t.TempDir(), Binary: "grite", Runner: runner}
	target := NewFileProvider(t.TempDir(), "")

	moved, err := Transfer(context.Background(), source, target, "abc123def456")
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

	if moved.Provider != ProviderFiles {
		t.Fatalf("expected target file provider, got %q", moved.Provider)
	}
	if moved.Title != "Cross-provider move" || moved.Priority != types.PriorityLow {
		t.Fatalf("transferred todo lost fields: %+v", moved)
	}
	if !strings.Contains(moved.MarkdownBody, "carry over") {
		t.Fatalf("grite body not carried over: %q", moved.MarkdownBody)
	}
	if !closed {
		t.Fatal("expected source Grite issue to be closed after transfer")
	}
}
