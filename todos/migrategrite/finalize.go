package migrategrite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/flanksource/gavel/todos/griteexport"
)

// ValidateFullSnapshotHistory rejects incremental exports used as standalone
// imports. Grite issue snapshot rows do not contain Markdown bodies, so every
// issue needs its creation history before body/verification projection is safe.
func ValidateFullSnapshotHistory(snapshot griteexport.Snapshot) error {
	created := make(map[griteexport.ID]bool, len(snapshot.Issues))
	for _, event := range snapshot.Events {
		name, _, err := event.Kind.NamePayload()
		if err != nil {
			return err
		}
		if name == "IssueCreated" {
			created[event.IssueID] = true
		}
	}
	var missing []string
	for _, issue := range snapshot.Issues {
		if !created[issue.IssueID] {
			missing = append(missing, issue.IssueID.String())
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("Grite import is not a full-history snapshot; missing IssueCreated event(s) for: %v", missing)
	}
	return nil
}

// MergeSnapshots combines the initial history with an overlapping final delta.
// Grite always returns a full issue snapshot, so the delta's issues are the
// authoritative final rows while events are unioned by immutable event ID.
func MergeSnapshots(initial, delta griteexport.Snapshot) (griteexport.Snapshot, error) {
	if err := requireWatermarkBoundary(griteexport.WatermarkFor(initial.Events), delta.Events, "final delta"); err != nil {
		return griteexport.Snapshot{}, err
	}
	events, err := griteexport.Merge(initial.Events, delta.Events)
	if err != nil {
		return griteexport.Snapshot{}, err
	}
	merged := delta
	merged.Events = events
	merged.Meta.EventCount = len(events)
	if merged.Meta.SchemaVersion == 0 {
		merged.Meta.SchemaVersion = initial.Meta.SchemaVersion
	}
	return merged, nil
}

// ValidateFrozenProbe proves a second overlapping export contains no event or
// issue-snapshot change beyond the candidate final snapshot. Generation time
// and delta-local event counts are intentionally ignored.
func ValidateFrozenProbe(candidate, probe griteexport.Snapshot) error {
	if err := requireWatermarkBoundary(griteexport.WatermarkFor(candidate.Events), probe.Events, "frozen probe"); err != nil {
		return err
	}
	merged, err := griteexport.Merge(candidate.Events, probe.Events)
	if err != nil {
		return err
	}
	if len(merged) != len(candidate.Events) {
		return fmt.Errorf("Grite source is not frozen: probe contains %d novel event(s)", len(merged)-len(candidate.Events))
	}
	left, err := canonicalIssues(candidate.Issues)
	if err != nil {
		return err
	}
	right, err := canonicalIssues(probe.Issues)
	if err != nil {
		return err
	}
	if !bytes.Equal(left, right) {
		return fmt.Errorf("Grite source is not frozen: issue snapshot changed during final validation")
	}
	return nil
}

func requireWatermarkBoundary(watermark griteexport.Watermark, events []griteexport.Event, source string) error {
	if len(watermark.EventIDs) == 0 {
		return nil
	}
	seen := make(map[griteexport.ID]bool, len(events))
	for _, event := range events {
		if event.TimestampMS == watermark.TimestampMS {
			seen[event.EventID] = true
		}
	}
	var missing []string
	for _, id := range watermark.EventIDs {
		if !seen[id] {
			missing = append(missing, id.String())
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("%s does not overlap watermark %d; missing boundary event(s): %v", source, watermark.TimestampMS, missing)
	}
	return nil
}

func canonicalIssues(issues []griteexport.Issue) ([]byte, error) {
	cloned := append([]griteexport.Issue(nil), issues...)
	for i := range cloned {
		cloned[i].Labels = append([]string(nil), cloned[i].Labels...)
		cloned[i].Assignees = append([]griteexport.ID(nil), cloned[i].Assignees...)
		if cloned[i].Labels == nil {
			cloned[i].Labels = []string{}
		}
		if cloned[i].Assignees == nil {
			cloned[i].Assignees = []griteexport.ID{}
		}
		sort.Strings(cloned[i].Labels)
		sort.Slice(cloned[i].Assignees, func(a, b int) bool { return cloned[i].Assignees[a] < cloned[i].Assignees[b] })
	}
	sort.Slice(cloned, func(i, j int) bool { return cloned[i].IssueID < cloned[j].IssueID })
	encoded, err := json.Marshal(cloned)
	if err != nil {
		return nil, fmt.Errorf("canonicalize Grite issue snapshot: %w", err)
	}
	return encoded, nil
}
