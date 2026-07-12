// Package migrategrite contains the disposable Grite-to-native TODO import.
//
// It is deliberately separate from provider routing: migration reads a Grite
// export, normalizes it into durable database concepts, and never makes Grite
// part of the native repository's runtime behavior.
package migrategrite

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/griteexport"
	"github.com/flanksource/gavel/todos/native"
)

const (
	GriteEventSource     = "grite"
	MigrationEventSource = "grite-migration"
	agentStateMarker     = "<!-- gavel:state "
	agentStateSuffix     = " -->"
)

// Document is a database-independent, deterministic projection of a Grite
// snapshot. Captain references remain hints until Import resolves them against
// Captain's authoritative tables.
type Document struct {
	SourceHash           string                `json:"sourceHash"`
	Watermark            griteexport.Watermark `json:"watermark"`
	Issues               []Issue               `json:"issues"`
	Events               []Event               `json:"events"`
	Relationships        []Relationship        `json:"relationships"`
	RemovedRelationships []Relationship        `json:"removedRelationships,omitempty"`
	Sessions             []SessionHint         `json:"sessions,omitempty"`
	Plans                []PlanHint            `json:"plans,omitempty"`
	Warnings             []Warning             `json:"warnings,omitempty"`
}

type Issue struct {
	SourceID       string                `json:"sourceId"`
	LegacyStatus   string                `json:"legacyStatus"`
	Title          string                `json:"title"`
	Body           string                `json:"body"`
	Verification   string                `json:"verification,omitempty"`
	Labels         []string              `json:"labels"`
	Priority       native.Priority       `json:"priority"`
	Status         native.IssueStatus    `json:"status"`
	ExecutionState native.ExecutionState `json:"executionState"`
	CreatedAt      time.Time             `json:"createdAt"`
	UpdatedAt      time.Time             `json:"updatedAt"`
}

type Event struct {
	SourceID      string          `json:"sourceId"`
	IssueSourceID string          `json:"issueSourceId"`
	Order         int             `json:"order"`
	Kind          string          `json:"kind"`
	Actor         string          `json:"actor,omitempty"`
	Body          string          `json:"body,omitempty"`
	Payload       json.RawMessage `json:"payload"`
	CreatedAt     time.Time       `json:"createdAt"`
}

type Relationship struct {
	IssueSourceID  string                  `json:"issueSourceId"`
	TargetSourceID string                  `json:"targetSourceId"`
	Relation       native.RelationshipKind `json:"relation"`
	CreatedAt      time.Time               `json:"createdAt"`
}

// SessionHint records the historical mode in effect when a Grite session label
// was attached. Import uses it to avoid guessing a Captain prompt-run step.
type SessionHint struct {
	IssueSourceID string    `json:"issueSourceId"`
	Identity      string    `json:"identity"`
	Mode          string    `json:"mode,omitempty"`
	ObservedAt    time.Time `json:"observedAt"`
}

type PlanHint struct {
	IssueSourceID   string    `json:"issueSourceId"`
	Path            string    `json:"path"`
	SessionIdentity string    `json:"sessionIdentity,omitempty"`
	ObservedAt      time.Time `json:"observedAt"`
	Order           int       `json:"order"`
	Selected        bool      `json:"selected,omitempty"`
	Clear           bool      `json:"clear,omitempty"`
	Pathless        bool      `json:"pathless,omitempty"`
}

// Warning is stable input to a migration_warning event. Its source ID is a
// content hash, so replaying an import cannot create another copy.
type Warning struct {
	Code    string `json:"code"`
	IssueID string `json:"issueId,omitempty"`
	Message string `json:"message"`
}

func (warning Warning) SourceID() string {
	sum := sha256.Sum256([]byte(warning.Code + "\x00" + warning.IssueID + "\x00" + warning.Message))
	return "warning:" + hex.EncodeToString(sum[:])
}

// Normalize validates and deterministically projects a full Grite export.
func Normalize(snapshot griteexport.Snapshot) (Document, error) {
	if snapshot.Meta.SchemaVersion != 0 && snapshot.Meta.SchemaVersion != 1 {
		return Document{}, fmt.Errorf("unsupported Grite export schema version %d", snapshot.Meta.SchemaVersion)
	}
	if snapshot.Meta.SchemaVersion != 0 && snapshot.Meta.EventCount != len(snapshot.Events) {
		return Document{}, fmt.Errorf("Grite export event count is %d, decoded %d", snapshot.Meta.EventCount, len(snapshot.Events))
	}

	events, err := griteexport.Merge(nil, snapshot.Events)
	if err != nil {
		return Document{}, err
	}
	issueBytes, err := canonicalIssues(snapshot.Issues)
	if err != nil {
		return Document{}, err
	}
	eventBytes, err := json.Marshal(events)
	if err != nil {
		return Document{}, fmt.Errorf("canonicalize Grite events: %w", err)
	}
	sourceBytes, err := json.Marshal(struct {
		SchemaVersion int             `json:"schemaVersion"`
		Issues        json.RawMessage `json:"issues"`
		Events        json.RawMessage `json:"events"`
	}{SchemaVersion: snapshot.Meta.SchemaVersion, Issues: issueBytes, Events: eventBytes})
	if err != nil {
		return Document{}, fmt.Errorf("canonicalize Grite snapshot: %w", err)
	}
	sum := sha256.Sum256(sourceBytes)
	document := Document{
		SourceHash: hex.EncodeToString(sum[:]),
		Watermark:  griteexport.WatermarkFor(events),
	}
	byIssue := make(map[string][]griteexport.Event, len(snapshot.Issues))
	for _, event := range events {
		issueID := strings.ToLower(strings.TrimSpace(event.IssueID.String()))
		byIssue[issueID] = append(byIssue[issueID], event)
	}

	issueIDs := make(map[string]struct{}, len(snapshot.Issues))
	issues := append([]griteexport.Issue(nil), snapshot.Issues...)
	sort.Slice(issues, func(i, j int) bool { return issues[i].IssueID < issues[j].IssueID })
	for _, source := range issues {
		id := strings.ToLower(strings.TrimSpace(source.IssueID.String()))
		if id == "" {
			return Document{}, errors.New("Grite issue has an empty ID")
		}
		if len(id) != 32 {
			return Document{}, fmt.Errorf("Grite issue ID %q must contain 32 hexadecimal characters", id)
		}
		if _, err := hex.DecodeString(id); err != nil {
			return Document{}, fmt.Errorf("Grite issue ID %q is not hexadecimal", id)
		}
		if _, exists := issueIDs[id]; exists {
			return Document{}, fmt.Errorf("duplicate Grite issue ID %s", id)
		}
		issueIDs[id] = struct{}{}

		body, eventTitle, planStates := projectIssueEvents(byIssue[id], source.Labels)
		title := strings.TrimSpace(source.Title)
		if title == "" {
			title = strings.TrimSpace(eventTitle)
		}
		if title == "" {
			return Document{}, fmt.Errorf("Grite issue %s has an empty title", id)
		}
		labels, priority, status, executionState, legacyStatus, warnings := normalizeLabels(id, source.State, source.Labels)
		document.Warnings = append(document.Warnings, warnings...)
		if len(source.Assignees) > 0 {
			document.Warnings = append(document.Warnings, Warning{
				Code: "assignees_not_imported", IssueID: id,
				Message: fmt.Sprintf("%d Grite assignee(s) have no native TODO field", len(source.Assignees)),
			})
		}
		document.Issues = append(document.Issues, Issue{
			SourceID:       id,
			LegacyStatus:   legacyStatus,
			Title:          title,
			Body:           body,
			Verification:   todos.ExtractVerificationFixture(body),
			Labels:         labels,
			Priority:       priority,
			Status:         status,
			ExecutionState: executionState,
			CreatedAt:      time.UnixMilli(source.CreatedTS).UTC(),
			UpdatedAt:      time.UnixMilli(source.UpdatedTS).UTC(),
		})
		hints := sessionHints(id, source, byIssue[id])
		document.Sessions = append(document.Sessions, hints...)
		if len(hints) == 0 && needsCaptainLink(source.Labels) {
			document.Warnings = append(document.Warnings, Warning{
				Code: "captain_session_reference_missing", IssueID: id,
				Message: "Grite execution status has no Captain session reference",
			})
		}
		for _, state := range planStates {
			document.Plans = append(document.Plans, PlanHint{
				IssueSourceID:   id,
				Path:            strings.TrimSpace(state.Path),
				SessionIdentity: state.SessionIdentity,
				ObservedAt:      state.ObservedAt,
				Order:           state.Order,
				Selected:        state.Selected,
				Clear:           state.Clear,
				Pathless:        state.Pathless,
			})
		}
	}

	for order, event := range events {
		issueID := strings.ToLower(strings.TrimSpace(event.IssueID.String()))
		if _, ok := issueIDs[issueID]; !ok {
			return Document{}, fmt.Errorf("Grite event %s references missing issue %s", event.EventID, issueID)
		}
		name, raw, err := event.Kind.NamePayload()
		if err != nil {
			return Document{}, err
		}
		body := eventBody(name, raw)
		payload, err := eventPayload(event, name, raw)
		if err != nil {
			return Document{}, err
		}
		document.Events = append(document.Events, Event{
			SourceID:      event.EventID.String(),
			IssueSourceID: issueID,
			Order:         order,
			Kind:          nativeEventKind(name),
			Actor:         strings.TrimSpace(event.Actor),
			Body:          body,
			Payload:       payload,
			CreatedAt:     time.UnixMilli(event.TimestampMS).UTC(),
		})
	}

	relationships, removedRelationships, warnings, err := projectRelationships(events, issueIDs)
	if err != nil {
		return Document{}, err
	}
	document.Relationships = relationships
	document.RemovedRelationships = removedRelationships
	document.Warnings = append(document.Warnings, warnings...)
	normalizeDocumentOrder(&document)
	return document, nil
}

type projectedState struct {
	Path            string
	SessionIdentity string
	ObservedAt      time.Time
	Order           int
	Selected        bool
	Clear           bool
	Pathless        bool
}

func projectIssueEvents(events []griteexport.Event, finalLabels []string) (body, title string, states []projectedState) {
	currentSession := ""
	latestSession := ""
	currentMode := ""
	currentPlanStatus := ""
	pathSessions := map[string]string{}
	pathlessSession := ""
	var observed []projectedState
	for order, event := range events {
		name, raw, err := event.Kind.NamePayload()
		if err != nil {
			continue
		}
		switch name {
		case "IssueCreated", "IssueUpdated":
			var payload struct {
				Body  *string `json:"body"`
				Title *string `json:"title"`
			}
			if json.Unmarshal(raw, &payload) == nil {
				if payload.Body != nil {
					body = *payload.Body
				}
				if payload.Title != nil {
					title = *payload.Title
				}
			}
		case "CommentAdded":
			var payload struct {
				Body string `json:"body"`
			}
			if json.Unmarshal(raw, &payload) == nil {
				if planPath, ok := parseAgentStatePath(payload.Body); ok {
					pathless := planPath == "" && knownPathlessPlanStatus(currentPlanStatus)
					sessionIdentity := currentSession
					rebindOwner := strings.EqualFold(strings.TrimSpace(currentMode), "plan") &&
						strings.EqualFold(strings.TrimSpace(currentPlanStatus), "new")
					if pathless {
						if pathlessSession != "" && !rebindOwner {
							sessionIdentity = pathlessSession
						} else if sessionIdentity != "" {
							pathlessSession = sessionIdentity
						}
					} else if planPath != "" {
						pathKey := strings.TrimSpace(planPath)
						if originalSession, seen := pathSessions[pathKey]; seen && !rebindOwner {
							sessionIdentity = originalSession
						} else if sessionIdentity != "" {
							pathSessions[pathKey] = sessionIdentity
						}
					}
					observed = append(observed, projectedState{
						Path:            planPath,
						SessionIdentity: sessionIdentity,
						ObservedAt:      time.UnixMilli(event.TimestampMS).UTC(),
						Order:           order,
						Clear:           planPath == "" && !pathless,
						Pathless:        pathless,
					})
				}
			}
		case "LabelAdded", "LabelRemoved":
			var payload struct {
				Label string `json:"label"`
			}
			if json.Unmarshal(raw, &payload) != nil {
				continue
			}
			prefix, value, ok := strings.Cut(strings.TrimSpace(payload.Label), ":")
			if !ok {
				continue
			}
			value = strings.TrimSpace(value)
			switch prefix {
			case "mode":
				if name == "LabelAdded" {
					currentMode = value
				} else if currentMode == value {
					currentMode = ""
				}
			case "session":
				if name == "LabelAdded" {
					currentSession = value
					latestSession = value
				} else if currentSession == value {
					currentSession = ""
				}
			case "plan":
				if name == "LabelAdded" && knownPathlessPlanStatus(value) {
					currentPlanStatus = value
				} else if name == "LabelRemoved" && currentPlanStatus == value {
					currentPlanStatus = ""
				}
			}
		}
	}

	// Older exports did not always emit the session label before their state
	// comment. Preserve the historical association where it is known, then use
	// the final label (or latest historical label) only as a fallback.
	fallbackSession := labelValue(finalLabels, "session")
	if fallbackSession == "" {
		fallbackSession = latestSession
	}
	seen := make(map[string]int, len(observed))
	selectedKey := ""
	for _, state := range observed {
		if !state.Clear && state.SessionIdentity == "" {
			state.SessionIdentity = fallbackSession
		}
		key := planStateKey(state)
		selectedKey = key
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = len(states)
		states = append(states, state)
	}
	if index, ok := seen[selectedKey]; ok {
		states[index].Selected = true
	}
	return body, title, states
}

func planStateKey(state projectedState) string {
	if state.Clear {
		return "\x00clear"
	}
	if state.Pathless {
		return "\x00pathless\x00" + strings.TrimSpace(state.SessionIdentity)
	}
	return strings.TrimSpace(state.Path) + "\x00" + strings.TrimSpace(state.SessionIdentity)
}

func knownPathlessPlanStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "new", "updated", "unchanged":
		return true
	default:
		return false
	}
}

func normalizeLabels(issueID, providerState string, source []string) ([]string, native.Priority, native.IssueStatus, native.ExecutionState, string, []Warning) {
	var labels, statusValues, priorityValues []string
	var warnings []Warning
	for _, raw := range source {
		label := strings.TrimSpace(raw)
		if label == "" {
			continue
		}
		prefix, value, namespaced := strings.Cut(label, ":")
		if !namespaced {
			labels = append(labels, label)
			continue
		}
		switch prefix {
		case "status":
			statusValues = append(statusValues, value)
			if !knownStatus(value) {
				warnings = append(warnings, Warning{Code: "unknown_status_label", IssueID: issueID, Message: "unknown Grite status label " + label})
			}
		case "priority":
			priorityValues = append(priorityValues, value)
			if !knownPriority(value) {
				warnings = append(warnings, Warning{Code: "unknown_priority_label", IssueID: issueID, Message: "unknown Grite priority label " + label})
			}
		case "session", "plan", "mode":
			// Transport-only labels are represented by native links/events.
		default:
			labels = append(labels, label)
		}
	}
	labels = sortedUnique(labels)
	statusValues = sortedUnique(statusValues)
	priorityValues = sortedUnique(priorityValues)
	if len(statusValues) > 1 {
		warnings = append(warnings, Warning{Code: "conflicting_status_labels", IssueID: issueID, Message: "multiple Grite status labels: " + strings.Join(statusValues, ", ")})
	}
	if len(priorityValues) > 1 {
		warnings = append(warnings, Warning{Code: "conflicting_priority_labels", IssueID: issueID, Message: "multiple Grite priority labels: " + strings.Join(priorityValues, ", ")})
	}
	status, legacyStatus := selectStatus(providerState, statusValues, &warnings, issueID)
	executionState := native.ExecutionIdle
	switch legacyStatus {
	case "failed":
		executionState = native.ExecutionFailed
	case "unverified":
		executionState = native.ExecutionVerificationFailed
	}
	return labels, selectPriority(priorityValues), status, executionState, legacyStatus, warnings
}

func knownStatus(value string) bool {
	switch value {
	case "draft", "pending", "in_progress", "review", "ask", "verified", "unverified", "completed", "failed", "skipped":
		return true
	default:
		return false
	}
}

func knownPriority(value string) bool {
	switch value {
	case string(native.PriorityLow), string(native.PriorityMedium), string(native.PriorityHigh), string(native.PriorityCritical):
		return true
	default:
		return false
	}
}

func needsCaptainLink(labels []string) bool {
	for _, label := range labels {
		switch strings.TrimSpace(label) {
		case "status:in_progress", "status:review", "status:ask", "status:failed", "status:unverified":
			return true
		}
	}
	return false
}

func selectPriority(values []string) native.Priority {
	for _, want := range []native.Priority{native.PriorityCritical, native.PriorityHigh, native.PriorityMedium, native.PriorityLow} {
		for _, value := range values {
			if value == string(want) {
				return want
			}
		}
	}
	return native.PriorityMedium
}

func selectStatus(providerState string, values []string, warnings *[]Warning, issueID string) (native.IssueStatus, string) {
	providerState = strings.ToLower(strings.TrimSpace(providerState))
	if providerState != "" && providerState != "open" {
		if providerState != "closed" {
			*warnings = append(*warnings, Warning{Code: "unknown_provider_state", IssueID: issueID, Message: "unknown Grite provider state " + providerState})
		}
	}
	for _, want := range []string{"verified", "completed", "skipped", "draft", "failed", "unverified", "in_progress", "review", "ask", "pending"} {
		for _, value := range values {
			if value != want {
				continue
			}
			switch want {
			case "verified":
				return native.StatusVerified, want
			case "completed", "skipped":
				return native.StatusClosed, want
			case "draft":
				return native.StatusDraft, want
			default:
				return native.StatusOpen, want
			}
		}
	}
	if providerState == "closed" {
		return native.StatusClosed, "closed"
	}
	return native.StatusOpen, "open"
}

func eventBody(name string, raw json.RawMessage) string {
	if name != "CommentAdded" {
		return ""
	}
	var payload struct {
		Body string `json:"body"`
	}
	_ = json.Unmarshal(raw, &payload)
	return payload.Body
}

func eventPayload(event griteexport.Event, name string, raw json.RawMessage) (json.RawMessage, error) {
	payload := struct {
		GriteKind string          `json:"griteKind"`
		Parent    *griteexport.ID `json:"parent,omitempty"`
		Data      json.RawMessage `json:"data"`
	}{GriteKind: name, Parent: event.Parent, Data: raw}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode Grite event %s payload: %w", event.EventID, err)
	}
	return encoded, nil
}

func nativeEventKind(name string) string {
	if name == "CommentAdded" {
		return "comment"
	}
	var out []rune
	for i, r := range name {
		if unicode.IsUpper(r) && i > 0 {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(r))
	}
	return string(out)
}

type relationshipState struct {
	Relationship
	Present     bool
	ObservedAdd bool
}

func projectRelationships(events []griteexport.Event, issueIDs map[string]struct{}) ([]Relationship, []Relationship, []Warning, error) {
	states := map[string]relationshipState{}
	var warnings []Warning
	for _, event := range events {
		name, raw, err := event.Kind.NamePayload()
		if err != nil {
			return nil, nil, nil, err
		}
		if name != "DependencyAdded" && name != "DependencyRemoved" {
			continue
		}
		var payload griteexport.DependencyPayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, nil, nil, fmt.Errorf("decode dependency event %s: %w", event.EventID, err)
		}
		issueID := strings.ToLower(strings.TrimSpace(event.IssueID.String()))
		targetID := strings.ToLower(strings.TrimSpace(payload.Target.String()))
		if _, ok := issueIDs[targetID]; !ok {
			warnings = append(warnings, Warning{Code: "missing_dependency_target", IssueID: issueID, Message: "Grite dependency target " + targetID + " is not in the snapshot"})
			continue
		}
		relation := native.RelationshipKind(strings.ToLower(strings.TrimSpace(payload.DepType)))
		if relation == native.RelationshipBlocks {
			issueID, targetID = targetID, issueID
			relation = native.RelationshipDependsOn
		}
		if relation != native.RelationshipDependsOn && relation != native.RelationshipRelatedTo {
			warnings = append(warnings, Warning{Code: "unknown_dependency_type", IssueID: issueID, Message: "unknown Grite dependency type " + payload.DepType})
			continue
		}
		if issueID == targetID {
			warnings = append(warnings, Warning{Code: "self_dependency", IssueID: issueID, Message: "Grite dependency targets the same issue"})
			continue
		}
		if relation == native.RelationshipRelatedTo && issueID > targetID {
			issueID, targetID = targetID, issueID
		}
		key := issueID + "\x00" + targetID + "\x00" + string(relation)
		state, exists := states[key]
		if !exists {
			state.Relationship = Relationship{
				IssueSourceID: issueID, TargetSourceID: targetID, Relation: relation,
				CreatedAt: time.UnixMilli(event.TimestampMS).UTC(),
			}
		}
		if name == "DependencyAdded" {
			// Keep the first observed add as the durable lifecycle timestamp.
			// A remove/re-add in the frozen full history should replay the same
			// native relationship row instead of forcing timestamp churn.
			if !state.ObservedAdd {
				state.CreatedAt = time.UnixMilli(event.TimestampMS).UTC()
				state.ObservedAdd = true
			}
			state.Present = true
		} else {
			state.Present = false
		}
		states[key] = state
	}
	var out, removed []Relationship
	for _, state := range states {
		if state.Present {
			out = append(out, state.Relationship)
		} else {
			removed = append(removed, state.Relationship)
		}
	}
	sortRelationships := func(values []Relationship) {
		sort.Slice(values, func(i, j int) bool {
			left := values[i].IssueSourceID + "\x00" + values[i].TargetSourceID + "\x00" + string(values[i].Relation)
			right := values[j].IssueSourceID + "\x00" + values[j].TargetSourceID + "\x00" + string(values[j].Relation)
			return left < right
		})
	}
	sortRelationships(out)
	sortRelationships(removed)
	return out, removed, warnings, nil
}

func sessionHints(issueID string, issue griteexport.Issue, events []griteexport.Event) []SessionHint {
	mode := ""
	active := map[string]bool{}
	byIdentity := map[string]SessionHint{}
	authoritative := map[string]bool{}
	update := func(identity, observedMode string, at time.Time, force bool) {
		identity = strings.TrimSpace(identity)
		if identity == "" {
			return
		}
		if authoritative[identity] && !force {
			return
		}
		byIdentity[identity] = SessionHint{IssueSourceID: issueID, Identity: identity, Mode: observedMode, ObservedAt: at}
		if force {
			authoritative[identity] = true
		}
	}
	for _, event := range events {
		name, raw, err := event.Kind.NamePayload()
		if err != nil {
			continue
		}
		if name == "CommentAdded" {
			var payload struct {
				Body string `json:"body"`
			}
			if json.Unmarshal(raw, &payload) == nil {
				if identity, runMode, ok := parseRunStartComment(payload.Body); ok {
					update(identity, runMode, time.UnixMilli(event.TimestampMS).UTC(), true)
				}
			}
			continue
		}
		if name != "LabelAdded" && name != "LabelRemoved" {
			continue
		}
		var payload struct {
			Label string `json:"label"`
		}
		if json.Unmarshal(raw, &payload) != nil {
			continue
		}
		prefix, value, ok := strings.Cut(strings.TrimSpace(payload.Label), ":")
		if !ok {
			continue
		}
		switch prefix {
		case "mode":
			if name == "LabelAdded" {
				mode = value
				// A mode label often arrives just after the session label it
				// qualifies. Update every still-active session to the nearest
				// explicit mode instead of retaining a stale plan/run pairing.
				for identity := range active {
					update(identity, mode, time.UnixMilli(event.TimestampMS).UTC(), false)
				}
			} else if mode == value {
				mode = ""
			}
		case "session":
			if name == "LabelAdded" {
				identity := strings.TrimSpace(value)
				if identity != "" {
					active[identity] = true
					// A mode label that predates this session may belong to the
					// previous run. Leave the new session unclassified until a later
					// mode event or authoritative run-start comment identifies it.
					update(identity, "", time.UnixMilli(event.TimestampMS).UTC(), false)
				}
			} else {
				delete(active, strings.TrimSpace(value))
			}
		}
	}
	finalMode := labelValue(issue.Labels, "mode")
	for _, label := range issue.Labels {
		if prefix, identity, ok := strings.Cut(strings.TrimSpace(label), ":"); ok && prefix == "session" {
			identity = strings.TrimSpace(identity)
			if authoritative[identity] {
				continue
			}
			observedMode := finalMode
			if observed, exists := byIdentity[identity]; exists {
				// Preserve an explicitly observed unknown mode. Falling back to
				// the final label here would reintroduce the stale-label race.
				observedMode = observed.Mode
			}
			update(identity, observedMode, time.UnixMilli(issue.UpdatedTS).UTC(), false)
		}
	}
	hints := make([]SessionHint, 0, len(byIdentity))
	for _, hint := range byIdentity {
		hints = append(hints, hint)
	}
	sort.Slice(hints, func(i, j int) bool {
		if !hints[i].ObservedAt.Equal(hints[j].ObservedAt) {
			return hints[i].ObservedAt.Before(hints[j].ObservedAt)
		}
		if hints[i].Identity != hints[j].Identity {
			return hints[i].Identity < hints[j].Identity
		}
		return hints[i].Mode < hints[j].Mode
	})
	return hints
}

func parseRunStartComment(body string) (identity, mode string, ok bool) {
	if !strings.Contains(body, "**Todo run started**") {
		return "", "", false
	}
	for _, line := range strings.Split(body, "\n") {
		switch {
		case strings.Contains(line, "**Session ID:**"):
			identity = backtickFieldValue(line, "**Session ID:**")
		case strings.Contains(line, "**Mode:**"):
			mode = backtickFieldValue(line, "**Mode:**")
		}
	}
	identity = strings.TrimSpace(identity)
	mode = strings.ToLower(strings.TrimSpace(mode))
	return identity, mode, identity != "" && identity != "unknown" && mode != ""
}

func backtickFieldValue(line, label string) string {
	index := strings.Index(line, label)
	if index < 0 {
		return ""
	}
	rest := line[index+len(label):]
	start := strings.Index(rest, "`")
	if start < 0 {
		return ""
	}
	rest = rest[start+1:]
	end := strings.Index(rest, "`")
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(rest[:end])
}

func labelValue(labels []string, namespace string) string {
	for _, label := range labels {
		if prefix, value, ok := strings.Cut(strings.TrimSpace(label), ":"); ok && prefix == namespace {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func parseAgentStatePath(body string) (string, bool) {
	index := strings.LastIndex(body, agentStateMarker)
	if index < 0 {
		return "", false
	}
	rest := body[index+len(agentStateMarker):]
	end := strings.Index(rest, agentStateSuffix)
	if end < 0 {
		return "", false
	}
	var state struct {
		PlanPath *string `json:"planPath"`
	}
	if err := json.Unmarshal([]byte(rest[:end]), &state); err != nil {
		return "", false
	}
	if state.PlanPath == nil {
		return "", true
	}
	return strings.TrimSpace(*state.PlanPath), true
}

func sortedUnique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func normalizeDocumentOrder(document *Document) {
	sort.Slice(document.Sessions, func(i, j int) bool {
		left := document.Sessions[i].IssueSourceID + "\x00" + document.Sessions[i].Identity + "\x00" + document.Sessions[i].Mode
		right := document.Sessions[j].IssueSourceID + "\x00" + document.Sessions[j].Identity + "\x00" + document.Sessions[j].Mode
		return left < right
	})
	sort.SliceStable(document.Plans, func(i, j int) bool {
		if document.Plans[i].IssueSourceID != document.Plans[j].IssueSourceID {
			return document.Plans[i].IssueSourceID < document.Plans[j].IssueSourceID
		}
		if document.Plans[i].Order != document.Plans[j].Order {
			return document.Plans[i].Order < document.Plans[j].Order
		}
		return document.Plans[i].ObservedAt.Before(document.Plans[j].ObservedAt)
	})
	sort.Slice(document.Warnings, func(i, j int) bool {
		if document.Warnings[i].IssueID != document.Warnings[j].IssueID {
			return document.Warnings[i].IssueID < document.Warnings[j].IssueID
		}
		if document.Warnings[i].Code != document.Warnings[j].Code {
			return document.Warnings[i].Code < document.Warnings[j].Code
		}
		return document.Warnings[i].Message < document.Warnings[j].Message
	})
}
