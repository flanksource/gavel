package migrategrite

import (
	"encoding/json"
	"testing"

	"github.com/flanksource/gavel/todos/griteexport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeSnapshotsOverlapsAndUsesFinalIssues(t *testing.T) {
	issueID := griteexport.ID("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	event := func(id string, timestamp int64) griteexport.Event {
		return griteexport.Event{EventID: griteexport.ID(id), IssueID: issueID, TimestampMS: timestamp,
			Kind: griteexport.Kind{"CommentAdded": json.RawMessage(`{"body":"` + id + `"}`)}}
	}
	initial := griteexport.Snapshot{
		Meta:   griteexport.Meta{SchemaVersion: 1, EventCount: 2},
		Issues: []griteexport.Issue{{IssueID: issueID, Title: "before", State: "open"}},
		Events: []griteexport.Event{event("one", 1), event("boundary", 2)},
	}
	delta := griteexport.Snapshot{
		Meta:   griteexport.Meta{SchemaVersion: 1, EventCount: 2},
		Issues: []griteexport.Issue{{IssueID: issueID, Title: "after", State: "closed"}},
		Events: []griteexport.Event{event("boundary", 2), event("three", 3)},
	}

	merged, err := MergeSnapshots(initial, delta)
	require.NoError(t, err)
	assert.Len(t, merged.Events, 3)
	assert.Equal(t, 3, merged.Meta.EventCount)
	assert.Equal(t, "after", merged.Issues[0].Title)
	assert.Equal(t, int64(3), griteexport.WatermarkFor(merged.Events).TimestampMS)
}

func TestValidateFullSnapshotHistoryRejectsIncrementalBodyLoss(t *testing.T) {
	incremental := decodeSnapshot(t, `{
		"meta":{"schema_version":1,"event_count":0},
		"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"body is not in this row","state":"open"}],
		"events":[]
	}`)
	require.ErrorContains(t, ValidateFullSnapshotHistory(incremental), "missing IssueCreated")

	full := incremental
	full.Events = []griteexport.Event{{
		EventID: "created", IssueID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", TimestampMS: 1,
		Kind: griteexport.Kind{"IssueCreated": json.RawMessage(`{"title":"x","body":"body"}`)},
	}}
	require.NoError(t, ValidateFullSnapshotHistory(full))
}

func TestValidateFrozenProbeDetectsNovelEventsAndSnapshotChanges(t *testing.T) {
	base := decodeSnapshot(t, `{
		"meta":{"schema_version":1,"event_count":1},
		"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"same","state":"open","labels":["b","a"]}],
		"events":[{"event_id":"one","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":1,"kind":{"CommentAdded":{"body":"x"}}}]
	}`)
	reordered := base
	reordered.Issues = append([]griteexport.Issue(nil), base.Issues...)
	reordered.Issues[0].Labels = []string{"a", "b"}
	require.NoError(t, ValidateFrozenProbe(base, reordered))

	missingBoundary := base
	missingBoundary.Events = nil
	require.ErrorContains(t, ValidateFrozenProbe(base, missingBoundary), "missing boundary event")

	novel := base
	novel.Events = append(append([]griteexport.Event(nil), base.Events...), griteexport.Event{
		EventID: "two", IssueID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", TimestampMS: 2,
		Kind: griteexport.Kind{"CommentAdded": json.RawMessage(`{"body":"y"}`)},
	})
	require.ErrorContains(t, ValidateFrozenProbe(base, novel), "novel event")

	changed := base
	changed.Issues = append([]griteexport.Issue(nil), base.Issues...)
	changed.Issues[0].Title = "changed"
	require.ErrorContains(t, ValidateFrozenProbe(base, changed), "snapshot changed")
}
