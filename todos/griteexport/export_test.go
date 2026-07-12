package griteexport

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeFilePreservesExportDataAndIDEncodings(t *testing.T) {
	raw := []byte(`{
		"meta":{"schema_version":1,"generated_ts":1234,"event_count":1},
		"issues":[{
			"issue_id":[226,163,184,194,208,247,201,169,139,64,13,199,142,138,148,165],
			"title":"Import Grite","state":"open","labels":["status:pending","db"],
			"assignees":["1604eaddc513891131b1d19dea2df875"],
			"created_ts":100,"updated_ts":200,"comment_count":2
		}],
		"events":[{
			"event_id":"event-1","issue_id":"e2a3b8c2d0f7c9a98b400dc78e8a94a5",
			"actor":"actor-1","ts_unix_ms":200,"parent":"parent-1",
			"kind":{"DependencyAdded":{"dep_type":"depends_on","target":[116,43,187,204,20,9,4,215,34,237,29,149,228,111,89,171]}}
		}]
	}`)

	file, err := DecodeFile(raw)
	if err != nil {
		t.Fatalf("DecodeFile failed: %v", err)
	}
	if file.Meta.SchemaVersion != 1 || file.Meta.GeneratedTS != 1234 || file.Meta.EventCount != 1 {
		t.Fatalf("meta not preserved: %+v", file.Meta)
	}
	if got := file.Issues[0].IssueID.String(); got != "e2a3b8c2d0f7c9a98b400dc78e8a94a5" {
		t.Fatalf("byte-array issue ID = %q", got)
	}
	if got := file.Issues[0].Assignees[0].String(); got != "1604eaddc513891131b1d19dea2df875" {
		t.Fatalf("assignee = %q", got)
	}
	if file.Events[0].Parent == nil || file.Events[0].Parent.String() != "parent-1" {
		t.Fatalf("parent not preserved: %+v", file.Events[0].Parent)
	}
	name, payload, err := file.Events[0].Kind.NamePayload()
	if err != nil || name != "DependencyAdded" {
		t.Fatalf("kind = %q, %v", name, err)
	}
	var dependency DependencyPayload
	if err := json.Unmarshal(payload, &dependency); err != nil {
		t.Fatalf("decode dependency payload: %v", err)
	}
	if dependency.DepType != "depends_on" || dependency.Target.String() != "742bbbcc140904d722ed1d95e46f59ab" {
		t.Fatalf("dependency not preserved: %+v", dependency)
	}

	path := filepath.Join(t.TempDir(), "export.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}
	if loaded.Events[0].EventID != file.Events[0].EventID {
		t.Fatalf("loaded event ID = %q", loaded.Events[0].EventID)
	}
}

func TestDecodeFileRejectsNonSingletonKindAndInvalidByteID(t *testing.T) {
	for name, raw := range map[string]string{
		"empty kind": `{"issues":[],"events":[{"event_id":"e","issue_id":"i","kind":{}}]}`,
		"multi kind": `{"issues":[],"events":[{"event_id":"e","issue_id":"i","kind":{"A":{},"B":{}}}]}`,
		"short ID":   `{"issues":[{"issue_id":[1,2]}],"events":[]}`,
		"null ID":    `{"issues":[{"issue_id":null}],"events":[]}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeFile([]byte(raw)); err == nil {
				t.Fatal("expected decode error")
			}
		})
	}
}

func TestMergeIsDeterministicAndRejectsConflictingIDs(t *testing.T) {
	event := func(id string, ts int64, body string) Event {
		payload, _ := json.Marshal(map[string]string{"body": body})
		return Event{
			EventID: ID(id), IssueID: "issue", TimestampMS: ts,
			Kind: Kind{"CommentAdded": payload},
		}
	}

	merged, err := Merge(
		[]Event{event("z", 200, "last"), event("a", 100, "first")},
		[]Event{event("a", 100, "first"), event("b", 200, "middle")},
	)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}
	if len(merged) != 3 || merged[0].EventID != "z" || merged[1].EventID != "a" || merged[2].EventID != "b" {
		t.Fatalf("unexpected stable source order: %+v", merged)
	}
	if _, err := Merge([]Event{event("same", 1, "left")}, []Event{event("same", 1, "right")}); err == nil {
		t.Fatal("expected conflicting duplicate error")
	}
}

func TestWatermarkUsesBoundaryIDsAndOneMillisecondOverlap(t *testing.T) {
	events := []Event{
		{EventID: "old", TimestampMS: 99},
		{EventID: "z", TimestampMS: 100},
		{EventID: "a", TimestampMS: 100},
	}
	watermark := WatermarkFor(events)
	if watermark.TimestampMS != 100 || watermark.SinceMS() != 99 {
		t.Fatalf("watermark = %+v, since=%d", watermark, watermark.SinceMS())
	}
	if len(watermark.EventIDs) != 2 || watermark.EventIDs[0] != "a" || watermark.EventIDs[1] != "z" {
		t.Fatalf("boundary IDs = %#v", watermark.EventIDs)
	}
	if !watermark.ContainsBoundaryEvent(events[1]) || watermark.ContainsBoundaryEvent(events[0]) {
		t.Fatalf("boundary membership is wrong: %+v", watermark)
	}

	watermark = watermark.Advance([]Event{{EventID: "new", TimestampMS: 101}})
	if watermark.TimestampMS != 101 || watermark.SinceMS() != 100 || len(watermark.EventIDs) != 1 || watermark.EventIDs[0] != "new" {
		t.Fatalf("advanced watermark = %+v", watermark)
	}
	if got := (Watermark{TimestampMS: 1}).SinceMS(); got != 0 {
		t.Fatalf("small watermark since = %d, want 0", got)
	}
}
