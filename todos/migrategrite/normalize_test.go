package migrategrite

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/todos/griteexport"
	"github.com/flanksource/gavel/todos/native"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeProjectsIssuesEventsRelationshipsAndHints(t *testing.T) {
	snapshot := decodeSnapshot(t, `{
		"meta":{"schema_version":1,"generated_ts":900,"event_count":9},
		"issues":[
			{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"First","state":"open",
			 "labels":["status:in_progress","priority:high","mode:run","session:captain-1","domain:db","agent:todo"],
			 "created_ts":100,"updated_ts":800,"comment_count":1},
			{"issue_id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","title":"Second","state":"closed",
			 "labels":["status:failed","priority:low"],"created_ts":110,"updated_ts":700,"comment_count":0}
		],
		"events":[
			{"event_id":"01","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","actor":"actor","ts_unix_ms":100,
			 "kind":{"IssueCreated":{"title":"First","body":"old\n\n## Verification\n\ngo test ./..."}}},
			{"event_id":"02","issue_id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","actor":"actor","ts_unix_ms":110,
			 "kind":{"IssueCreated":{"title":"Second","body":"body"}}},
			{"event_id":"03","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","actor":"actor","ts_unix_ms":200,
			 "kind":{"LabelAdded":{"label":"session:captain-1"}}},
			{"event_id":"04","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","actor":"actor","ts_unix_ms":210,
			 "kind":{"LabelAdded":{"label":"mode:run"}}},
			{"event_id":"05","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","actor":"actor","ts_unix_ms":300,
			 "kind":{"DependencyAdded":{"dep_type":"depends_on","target":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}},
			{"event_id":"06","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","actor":"actor","ts_unix_ms":400,
			 "kind":{"DependencyRemoved":{"dep_type":"depends_on","target":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}},
			{"event_id":"07","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","actor":"actor","ts_unix_ms":500,
			 "kind":{"DependencyAdded":{"dep_type":"related_to","target":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}},
			{"event_id":"08","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","actor":"actor","ts_unix_ms":600,
			 "kind":{"CommentAdded":{"body":"**Agent state**\n\n<!-- gavel:state {\"planPath\":\".plans/import.md\"} -->"}}},
			{"event_id":"09","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","actor":"actor","ts_unix_ms":700,
			 "kind":{"IssueUpdated":{"body":""}}}
		]
	}`)

	document, err := Normalize(snapshot)
	require.NoError(t, err)
	require.Len(t, document.Issues, 2)
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", document.Issues[0].SourceID)
	assert.Empty(t, document.Issues[0].Body, "an explicit empty body update must be preserved")
	assert.Empty(t, document.Issues[0].Verification)
	assert.Equal(t, []string{"agent:todo", "domain:db"}, document.Issues[0].Labels)
	assert.Equal(t, native.PriorityHigh, document.Issues[0].Priority)
	assert.Equal(t, native.StatusOpen, document.Issues[0].Status)
	assert.Equal(t, native.ExecutionIdle, document.Issues[0].ExecutionState)
	assert.Equal(t, native.StatusOpen, document.Issues[1].Status, "explicit status label precedes provider-state fallback")
	assert.Equal(t, native.ExecutionFailed, document.Issues[1].ExecutionState)

	require.Len(t, document.Events, 9)
	assert.Equal(t, "issue_created", document.Events[0].Kind)
	assert.Equal(t, "comment", document.Events[7].Kind)
	assert.Equal(t, "", document.Events[8].Body)
	require.Len(t, document.Relationships, 1)
	assert.Equal(t, native.RelationshipRelatedTo, document.Relationships[0].Relation)
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", document.Relationships[0].IssueSourceID)
	assert.Equal(t, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", document.Relationships[0].TargetSourceID)
	require.Len(t, document.RemovedRelationships, 1)
	assert.Equal(t, native.RelationshipDependsOn, document.RemovedRelationships[0].Relation)

	require.Len(t, document.Sessions, 1)
	assert.Equal(t, "captain-1", document.Sessions[0].Identity)
	assert.Equal(t, "run", document.Sessions[0].Mode)
	require.Len(t, document.Plans, 1)
	assert.Equal(t, ".plans/import.md", document.Plans[0].Path)
	assert.Equal(t, "captain-1", document.Plans[0].SessionIdentity)
	assert.True(t, document.Plans[0].Selected)
	assert.False(t, document.Plans[0].Clear)
	assert.NotEmpty(t, document.SourceHash)
	assert.Equal(t, int64(700), document.Watermark.TimestampMS)
}

func TestNormalizePreservesHistoricalPlanHintsAndClearSelectionInSourceOrder(t *testing.T) {
	t.Run("A then B retains both and selects B", func(t *testing.T) {
		snapshot := decodeSnapshot(t, `{
			"meta":{"schema_version":1,"event_count":4},
			"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"Plans","state":"open",
			 "labels":["mode:plan","session:captain-plans"],"created_ts":100,"updated_ts":200,"comment_count":2}],
			"events":[
				{"event_id":"created","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":100,
				 "kind":{"IssueCreated":{"title":"Plans","body":"body"}}},
				{"event_id":"session","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":150,
				 "kind":{"LabelAdded":{"label":"session:captain-plans"}}},
				{"event_id":"plan-a","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":200,
				 "kind":{"CommentAdded":{"body":"<!-- gavel:state {\"planPath\":\"plans/a.md\"} -->"}}},
				{"event_id":"plan-b","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":200,
				 "kind":{"CommentAdded":{"body":"<!-- gavel:state {\"planPath\":\"plans/b.md\"} -->"}}}
			]
		}`)

		document, err := Normalize(snapshot)
		require.NoError(t, err)
		require.Len(t, document.Plans, 2)
		assert.Equal(t, []string{"plans/a.md", "plans/b.md"}, []string{document.Plans[0].Path, document.Plans[1].Path})
		assert.Equal(t, []string{"captain-plans", "captain-plans"}, []string{document.Plans[0].SessionIdentity, document.Plans[1].SessionIdentity})
		assert.False(t, document.Plans[0].Selected)
		assert.True(t, document.Plans[1].Selected)
		assert.Less(t, document.Plans[0].Order, document.Plans[1].Order)
		assert.Equal(t, document.Plans[0].ObservedAt, document.Plans[1].ObservedAt, "equal-ms markers must retain source order")
	})

	t.Run("A then clear retains A and selects nil", func(t *testing.T) {
		snapshot := decodeSnapshot(t, `{
			"meta":{"schema_version":1,"event_count":4},
			"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"Plans","state":"open",
			 "labels":["mode:plan","session:captain-plans"],"created_ts":100,"updated_ts":200,"comment_count":2}],
			"events":[
				{"event_id":"created","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":100,
				 "kind":{"IssueCreated":{"title":"Plans","body":"body"}}},
				{"event_id":"session","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":150,
				 "kind":{"LabelAdded":{"label":"session:captain-plans"}}},
				{"event_id":"plan-a","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":200,
				 "kind":{"CommentAdded":{"body":"<!-- gavel:state {\"planPath\":\"plans/a.md\"} -->"}}},
				{"event_id":"plan-clear","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":200,
				 "kind":{"CommentAdded":{"body":"<!-- gavel:state {} -->"}}}
			]
		}`)

		document, err := Normalize(snapshot)
		require.NoError(t, err)
		require.Len(t, document.Plans, 2)
		assert.Equal(t, "plans/a.md", document.Plans[0].Path)
		assert.False(t, document.Plans[0].Selected)
		assert.False(t, document.Plans[0].Clear)
		assert.Empty(t, document.Plans[1].Path)
		assert.True(t, document.Plans[1].Selected)
		assert.True(t, document.Plans[1].Clear)
		assert.False(t, document.Plans[1].Pathless)
		assert.Less(t, document.Plans[0].Order, document.Plans[1].Order)
	})

	t.Run("active plan status makes an empty marker pathless", func(t *testing.T) {
		snapshot := decodeSnapshot(t, `{
			"meta":{"schema_version":1,"event_count":4},
			"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"Pathless","state":"open",
			 "labels":["mode:plan","plan:new","session:captain-pathless"],"created_ts":100,"updated_ts":200,"comment_count":1}],
			"events":[
				{"event_id":"created","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":100,
				 "kind":{"IssueCreated":{"title":"Pathless","body":"body"}}},
				{"event_id":"session","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":150,
				 "kind":{"LabelAdded":{"label":"session:captain-pathless"}}},
				{"event_id":"plan-status","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":175,
				 "kind":{"LabelAdded":{"label":"plan:new"}}},
				{"event_id":"plan-marker","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":200,
				 "kind":{"CommentAdded":{"body":"<!-- gavel:state {} -->"}}}
			]
		}`)

		document, err := Normalize(snapshot)
		require.NoError(t, err)
		require.Len(t, document.Plans, 1)
		assert.Empty(t, document.Plans[0].Path)
		assert.True(t, document.Plans[0].Selected)
		assert.False(t, document.Plans[0].Clear)
		assert.True(t, document.Plans[0].Pathless)
		assert.Equal(t, "captain-pathless", document.Plans[0].SessionIdentity)
	})
}

func TestNormalizeSessionHintUsesCurrentModeWhenModeLabelLagsSession(t *testing.T) {
	snapshot := decodeSnapshot(t, `{
		"meta":{"schema_version":1,"event_count":5},
		"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"Mode lag","state":"open",
		 "labels":["status:in_progress","mode:run","session:captain-mode-lag"],"created_ts":100,"updated_ts":500}],
		"events":[
			{"event_id":"created","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":100,
			 "kind":{"IssueCreated":{"title":"Mode lag","body":"body"}}},
			{"event_id":"old-mode","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":200,
			 "kind":{"LabelAdded":{"label":"mode:plan"}}},
			{"event_id":"session","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":300,
			 "kind":{"LabelAdded":{"label":"session:captain-mode-lag"}}},
			{"event_id":"remove-old-mode","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":400,
			 "kind":{"LabelRemoved":{"label":"mode:plan"}}},
			{"event_id":"current-mode","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":400,
			 "kind":{"LabelAdded":{"label":"mode:run"}}}
		]
	}`)

	document, err := Normalize(snapshot)
	require.NoError(t, err)
	require.Len(t, document.Sessions, 1)
	assert.Equal(t, "captain-mode-lag", document.Sessions[0].Identity)
	assert.Equal(t, "run", document.Sessions[0].Mode)
	assert.Equal(t, int64(500), document.Sessions[0].ObservedAt.UnixMilli(), "final active labels are authoritative")
	for _, warning := range document.Warnings {
		assert.NotEqual(t, "captain_session_mode_conflict", warning.Code)
	}
}

func TestNormalizeRunStartCommentOverridesStaleModeLabels(t *testing.T) {
	initial := decodeSnapshot(t, `{
		"meta":{"schema_version":1,"event_count":3},
		"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"Interrupted run","state":"open",
		 "labels":["status:in_progress","mode:plan","session:captain-interrupted"],"created_ts":100,"updated_ts":300}],
		"events":[
			{"event_id":"created","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":100,
			 "kind":{"IssueCreated":{"title":"Interrupted run","body":"body"}}},
			{"event_id":"stale-mode","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":200,
			 "kind":{"LabelAdded":{"label":"mode:plan"}}},
			{"event_id":"session","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":300,
			 "kind":{"LabelAdded":{"label":"session:captain-interrupted"}}}
		]
	}`)
	initialDocument, err := Normalize(initial)
	require.NoError(t, err)
	require.Len(t, initialDocument.Sessions, 1)
	assert.Empty(t, initialDocument.Sessions[0].Mode, "a new session must not inherit an older mode label")

	snapshot := decodeSnapshot(t, `{
		"meta":{"schema_version":1,"event_count":4},
		"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"Interrupted run","state":"open",
		 "labels":["status:in_progress","mode:plan","session:captain-interrupted"],"created_ts":100,"updated_ts":500,"comment_count":1}],
		"events":[
			{"event_id":"created","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":100,
			 "kind":{"IssueCreated":{"title":"Interrupted run","body":"body"}}},
			{"event_id":"stale-mode","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":200,
			 "kind":{"LabelAdded":{"label":"mode:plan"}}},
			{"event_id":"session","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":300,
			 "kind":{"LabelAdded":{"label":"session:captain-interrupted"}}},
			{"event_id":"run-start","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":400,
			 "kind":{"CommentAdded":{"body":"**Todo run started**\n\n- **Session ID:** \u0060captain-interrupted\u0060\n- **Mode:** \u0060run\u0060\n- **Resolved Model:** \u0060default\u0060"}}}
		]
	}`)

	document, err := Normalize(snapshot)
	require.NoError(t, err)
	require.Len(t, document.Sessions, 1)
	assert.Equal(t, "captain-interrupted", document.Sessions[0].Identity)
	assert.Equal(t, "run", document.Sessions[0].Mode)
	assert.Equal(t, int64(400), document.Sessions[0].ObservedAt.UnixMilli())
}

func TestNormalizeRunSummaryRetainsPlanningSessionForPlanMarkers(t *testing.T) {
	t.Run("pathful", func(t *testing.T) {
		snapshot := decodeSnapshot(t, `{
			"meta":{"schema_version":1,"event_count":9},
			"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"Ownership","state":"open",
			 "labels":["mode:run","session:captain-run"],"created_ts":100,"updated_ts":900,"comment_count":2}],
			"events":[
				{"event_id":"created","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":100,"kind":{"IssueCreated":{"title":"Ownership","body":"body"}}},
				{"event_id":"plan-mode","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":200,"kind":{"LabelAdded":{"label":"mode:plan"}}},
				{"event_id":"plan-session","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":300,"kind":{"LabelAdded":{"label":"session:captain-plan"}}},
				{"event_id":"plan-marker","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":400,"kind":{"CommentAdded":{"body":"<!-- gavel:state {\"planPath\":\"plans/retained.md\"} -->"}}},
				{"event_id":"remove-plan-session","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":500,"kind":{"LabelRemoved":{"label":"session:captain-plan"}}},
				{"event_id":"remove-plan-mode","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":600,"kind":{"LabelRemoved":{"label":"mode:plan"}}},
				{"event_id":"run-mode","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":700,"kind":{"LabelAdded":{"label":"mode:run"}}},
				{"event_id":"run-session","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":800,"kind":{"LabelAdded":{"label":"session:captain-run"}}},
				{"event_id":"run-summary","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":900,"kind":{"CommentAdded":{"body":"<!-- gavel:state {\"planPath\":\"plans/retained.md\"} -->"}}}
			]
		}`)

		document, err := Normalize(snapshot)
		require.NoError(t, err)
		require.Len(t, document.Plans, 1)
		assert.Equal(t, "plans/retained.md", document.Plans[0].Path)
		assert.Equal(t, "captain-plan", document.Plans[0].SessionIdentity)
		assert.True(t, document.Plans[0].Selected)
	})

	t.Run("pathless", func(t *testing.T) {
		snapshot := decodeSnapshot(t, `{
			"meta":{"schema_version":1,"event_count":10},
			"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"Inline ownership","state":"open",
			 "labels":["mode:run","plan:new","session:captain-run"],"created_ts":100,"updated_ts":900,"comment_count":2}],
			"events":[
				{"event_id":"created","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":100,"kind":{"IssueCreated":{"title":"Inline ownership","body":"body"}}},
				{"event_id":"plan-mode","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":200,"kind":{"LabelAdded":{"label":"mode:plan"}}},
				{"event_id":"plan-session","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":300,"kind":{"LabelAdded":{"label":"session:captain-plan"}}},
				{"event_id":"plan-status","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":350,"kind":{"LabelAdded":{"label":"plan:new"}}},
				{"event_id":"plan-marker","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":400,"kind":{"CommentAdded":{"body":"<!-- gavel:state {} -->"}}},
				{"event_id":"remove-plan-session","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":500,"kind":{"LabelRemoved":{"label":"session:captain-plan"}}},
				{"event_id":"remove-plan-mode","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":600,"kind":{"LabelRemoved":{"label":"mode:plan"}}},
				{"event_id":"run-mode","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":700,"kind":{"LabelAdded":{"label":"mode:run"}}},
				{"event_id":"run-session","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":800,"kind":{"LabelAdded":{"label":"session:captain-run"}}},
				{"event_id":"run-summary","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":900,"kind":{"CommentAdded":{"body":"<!-- gavel:state {} -->"}}}
			]
		}`)

		document, err := Normalize(snapshot)
		require.NoError(t, err)
		require.Len(t, document.Plans, 1)
		assert.Empty(t, document.Plans[0].Path)
		assert.True(t, document.Plans[0].Pathless)
		assert.Equal(t, "captain-plan", document.Plans[0].SessionIdentity)
		assert.True(t, document.Plans[0].Selected)
	})
}

func TestNormalizePlanOutcomeRetainsOwnershipUnlessPlanIsNew(t *testing.T) {
	const issueID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	makeEvent := func(t *testing.T, id string, timestamp int64, name string, payload any) griteexport.Event {
		t.Helper()
		raw, err := json.Marshal(payload)
		require.NoError(t, err)
		return griteexport.Event{
			EventID: griteexport.ID(id), IssueID: griteexport.ID(issueID), TimestampMS: timestamp,
			Kind: griteexport.Kind{name: raw},
		}
	}
	makeSnapshot := func(t *testing.T, outcome string, pathless bool) griteexport.Snapshot {
		t.Helper()
		marker := `<!-- gavel:state {"planPath":"plans/retained.md"} -->`
		if pathless {
			marker = `<!-- gavel:state {} -->`
		}
		events := []griteexport.Event{
			makeEvent(t, "created", 100, "IssueCreated", map[string]any{"title": "Ownership", "body": "body"}),
			makeEvent(t, "plan-mode", 200, "LabelAdded", map[string]any{"label": "mode:plan"}),
			makeEvent(t, "old-session", 300, "LabelAdded", map[string]any{"label": "session:captain-plan-original"}),
			makeEvent(t, "old-plan-new", 350, "LabelAdded", map[string]any{"label": "plan:new"}),
			makeEvent(t, "old-marker", 400, "CommentAdded", map[string]any{"body": marker}),
			makeEvent(t, "old-session-removed", 500, "LabelRemoved", map[string]any{"label": "session:captain-plan-original"}),
			makeEvent(t, "new-session", 600, "LabelAdded", map[string]any{"label": "session:captain-plan-later"}),
		}
		if outcome != "new" {
			events = append(events,
				makeEvent(t, "old-plan-status-removed", 700, "LabelRemoved", map[string]any{"label": "plan:new"}),
				makeEvent(t, "new-plan-status", 800, "LabelAdded", map[string]any{"label": "plan:" + outcome}),
			)
		}
		events = append(events, makeEvent(t, "later-marker", 900, "CommentAdded", map[string]any{"body": marker}))
		return griteexport.Snapshot{
			Meta: griteexport.Meta{SchemaVersion: 1, GeneratedTS: 1_000, EventCount: len(events)},
			Issues: []griteexport.Issue{{
				IssueID: griteexport.ID(issueID), Title: "Ownership", State: "open",
				Labels:    []string{"mode:plan", "plan:" + outcome, "session:captain-plan-later"},
				CreatedTS: 100, UpdatedTS: 900, CommentCount: 2,
			}},
			Events: events,
		}
	}

	for _, pathless := range []bool{false, true} {
		kind := "pathful"
		if pathless {
			kind = "pathless"
		}
		for _, outcome := range []string{"unchanged", "updated", "new"} {
			t.Run(kind+"/"+outcome, func(t *testing.T) {
				document, err := Normalize(makeSnapshot(t, outcome, pathless))
				require.NoError(t, err)
				if outcome == "new" {
					require.Len(t, document.Plans, 2)
					assert.Equal(t, "captain-plan-original", document.Plans[0].SessionIdentity)
					assert.False(t, document.Plans[0].Selected)
					assert.Equal(t, "captain-plan-later", document.Plans[1].SessionIdentity)
					assert.True(t, document.Plans[1].Selected)
					return
				}
				require.Len(t, document.Plans, 1)
				assert.Equal(t, "captain-plan-original", document.Plans[0].SessionIdentity)
				assert.True(t, document.Plans[0].Selected)
				assert.Equal(t, pathless, document.Plans[0].Pathless)
			})
		}
	}
}

func TestNormalizeWarningsAndStableWarningIDs(t *testing.T) {
	snapshot := decodeSnapshot(t, `{
		"meta":{"schema_version":1,"event_count":2},
		"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"Warning","state":"mystery",
		 "labels":["status:future","status:pending","priority:urgent","priority:low"],
		 "assignees":["actor"],"created_ts":1,"updated_ts":2}],
		"events":[
			{"event_id":"1","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":1,
			 "kind":{"IssueCreated":{"title":"Warning","body":"body"}}},
			{"event_id":"2","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":2,
			 "kind":{"DependencyAdded":{"dep_type":"depends_on","target":"ffffffffffffffffffffffffffffffff"}}}
		]
	}`)

	first, err := Normalize(snapshot)
	require.NoError(t, err)
	second, err := Normalize(snapshot)
	require.NoError(t, err)
	require.NotEmpty(t, first.Warnings)
	assert.Equal(t, first, second)

	ids := make(map[string]bool)
	for _, warning := range first.Warnings {
		id := warning.SourceID()
		assert.False(t, ids[id], "warning source IDs must be unique")
		ids[id] = true
	}
	assert.Equal(t, native.PriorityLow, first.Issues[0].Priority)
	assert.Equal(t, native.StatusOpen, first.Issues[0].Status)
	assert.Empty(t, first.Relationships)
}

func TestNormalizeRejectsMetaCountMismatchAndConflictingEvents(t *testing.T) {
	mismatch := decodeSnapshot(t, `{"meta":{"schema_version":1,"event_count":1},"issues":[],"events":[]}`)
	_, err := Normalize(mismatch)
	require.ErrorContains(t, err, "event count")

	base := griteexport.Event{EventID: "same", IssueID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", TimestampMS: 1, Kind: griteexport.Kind{"CommentAdded": json.RawMessage(`{"body":"a"}`)}}
	conflict := base
	conflict.Kind = griteexport.Kind{"CommentAdded": json.RawMessage(`{"body":"b"}`)}
	_, err = Normalize(griteexport.Snapshot{
		Issues: []griteexport.Issue{{IssueID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Title: "x", State: "open"}},
		Events: []griteexport.Event{base, conflict},
	})
	require.ErrorContains(t, err, "conflicting Grite event content")
}

func TestNormalizeSourceHashIgnoresExportGenerationAndOrdering(t *testing.T) {
	first := decodeSnapshot(t, `{
		"meta":{"schema_version":1,"generated_ts":1,"event_count":0},
		"issues":[
			{"issue_id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","title":"b","state":"open","labels":["z","a"]},
			{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"a","state":"open"}
		],"events":[]
	}`)
	second := decodeSnapshot(t, `{
		"meta":{"schema_version":1,"generated_ts":999,"event_count":0},
		"issues":[
			{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"a","state":"open"},
			{"issue_id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","title":"b","state":"open","labels":["a","z"]}
		],"events":[]
	}`)
	left, err := Normalize(first)
	require.NoError(t, err)
	right, err := Normalize(second)
	require.NoError(t, err)
	assert.Equal(t, left.SourceHash, right.SourceHash)
}

func TestNormalizeMatchesUppercaseIssueAndEventIDsBeforeProjection(t *testing.T) {
	snapshot := decodeSnapshot(t, `{
		"meta":{"schema_version":1,"event_count":3},
		"issues":[{"issue_id":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","title":"Uppercase","state":"open",
		 "labels":["mode:plan","session:captain-uppercase"],"created_ts":100,"updated_ts":300,"comment_count":1}],
		"events":[
			{"event_id":"created","issue_id":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","ts_unix_ms":100,
			 "kind":{"IssueCreated":{"title":"Uppercase","body":"projected body"}}},
			{"event_id":"session","issue_id":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","ts_unix_ms":200,
			 "kind":{"LabelAdded":{"label":"session:captain-uppercase"}}},
			{"event_id":"plan","issue_id":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","ts_unix_ms":300,
			 "kind":{"CommentAdded":{"body":"<!-- gavel:state {\"planPath\":\"plans/uppercase.md\"} -->"}}}
		]
	}`)

	document, err := Normalize(snapshot)
	require.NoError(t, err)
	require.Len(t, document.Issues, 1)
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", document.Issues[0].SourceID)
	assert.Equal(t, "projected body", document.Issues[0].Body)
	require.Len(t, document.Sessions, 1)
	assert.Equal(t, "captain-uppercase", document.Sessions[0].Identity)
	require.Len(t, document.Plans, 1)
	assert.Equal(t, "plans/uppercase.md", document.Plans[0].Path)
	assert.True(t, document.Plans[0].Selected)
}

func TestPlanFilePathCannotEscapeWorkspace(t *testing.T) {
	root := t.TempDir()
	path, err := planFilePath(root, ".plans/import.md")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, ".plans", "import.md"), path)
	_, err = planFilePath(root, "../secret")
	require.ErrorContains(t, err, "escapes workspace")
	_, err = planFilePath("", ".plans/import.md")
	require.ErrorContains(t, err, "no root path")
}

func TestStatusMappingMatrix(t *testing.T) {
	tests := []struct {
		name      string
		state     string
		label     string
		status    native.IssueStatus
		execution native.ExecutionState
	}{
		{"draft", "open", "status:draft", native.StatusDraft, native.ExecutionIdle},
		{"pending", "open", "status:pending", native.StatusOpen, native.ExecutionIdle},
		{"in progress without linked run", "open", "status:in_progress", native.StatusOpen, native.ExecutionIdle},
		{"review", "open", "status:review", native.StatusOpen, native.ExecutionIdle},
		{"ask without linked request", "open", "status:ask", native.StatusOpen, native.ExecutionIdle},
		{"verified", "open", "status:verified", native.StatusVerified, native.ExecutionIdle},
		{"unverified", "open", "status:unverified", native.StatusOpen, native.ExecutionVerificationFailed},
		{"completed", "open", "status:completed", native.StatusClosed, native.ExecutionIdle},
		{"failed", "open", "status:failed", native.StatusOpen, native.ExecutionFailed},
		{"skipped", "open", "status:skipped", native.StatusClosed, native.ExecutionIdle},
		{"closed fallback", "closed", "", native.StatusClosed, native.ExecutionIdle},
		{"explicit status before provider fallback", "closed", "status:failed", native.StatusOpen, native.ExecutionFailed},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			labels := []string{}
			if test.label != "" {
				labels = append(labels, test.label)
			}
			_, _, status, execution, _, _ := normalizeLabels("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", test.state, labels)
			assert.Equal(t, test.status, status)
			assert.Equal(t, test.execution, execution)
		})
	}
}

func TestNormalizePreservesSourceOrderForEqualMillisecondEvents(t *testing.T) {
	snapshot := decodeSnapshot(t, `{
		"meta":{"schema_version":1,"event_count":2},
		"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"x","state":"open","created_ts":1,"updated_ts":1}],
		"events":[
			{"event_id":"z-created","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":1,"kind":{"IssueCreated":{"title":"x","body":"before"}}},
			{"event_id":"a-updated","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":1,"kind":{"IssueUpdated":{"body":"after"}}}
		]
	}`)
	document, err := Normalize(snapshot)
	require.NoError(t, err)
	assert.Equal(t, "after", document.Issues[0].Body)
	require.Len(t, document.Events, 2)
	assert.Equal(t, "z-created", document.Events[0].SourceID)
	assert.Equal(t, "a-updated", document.Events[1].SourceID)
}

func decodeSnapshot(t *testing.T, raw string) griteexport.Snapshot {
	t.Helper()
	snapshot, err := griteexport.DecodeFile([]byte(raw))
	require.NoError(t, err)
	return snapshot
}
