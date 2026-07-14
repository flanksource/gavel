package native

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

const DefaultImportSource = "grite-import"

var (
	ErrImportConflict = errors.New("native todo import conflict")
	importNamespace   = uuid.MustParse("cb98058a-f57b-5eab-a438-4d8616213a8b")
)

// ImportWorkspace is the normalized durable identity of the workspace being
// imported. ID is optional; when absent ApplyImport derives a stable UUID from
// RepoKey (or RootPath for a non-repository workspace).
type ImportWorkspace struct {
	ID          uuid.UUID
	RepoKey     string
	RootPath    string
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NormalizeImportWorkspace applies the same durable identity normalization as
// ApplyImport while preserving the caller's optional ID semantics.
func NormalizeImportWorkspace(input ImportWorkspace) ImportWorkspace {
	input.RepoKey = normalizeToken(input.RepoKey)
	input.RootPath = normalizeWorkspacePath(input.RootPath)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	return input
}

// DeterministicImportWorkspaceID returns the ID ApplyImport will use for a new
// workspace after repo/path normalization.
func DeterministicImportWorkspaceID(input ImportWorkspace) uuid.UUID {
	input = NormalizeImportWorkspace(input)
	if input.ID != uuid.Nil {
		return input.ID
	}
	identity := input.RepoKey
	if identity == "" {
		if input.RootPath == "" {
			return uuid.Nil
		}
		identity = "path:" + input.RootPath
	}
	return uuid.NewSHA1(importNamespace, []byte("workspace:"+identity))
}

// ImportIssue is the current snapshot of one legacy issue. SourceID must be the
// complete 32-character Grite ID. Short references are resolved by the native
// repository's existing unambiguous alias-prefix lookup and are not persisted
// as separate aliases.
type ImportIssue struct {
	ID                uuid.UUID
	SourceID          string
	Title             string
	Body              string
	Verification      string
	Labels            []string
	Priority          Priority
	Status            IssueStatus
	ExecutionState    ExecutionState
	ActivePromptRunID *uuid.UUID
	SelectedPlanID    *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ImportEvent is one source-owned append-only event. Payload must be JSON; an
// empty payload is normalized to {}. SourceID is globally unique within the
// batch Source, matching todo_issue_events' durable uniqueness boundary.
type ImportEvent struct {
	IssueSourceID string
	SourceID      string
	Order         int
	Kind          string
	Actor         string
	Body          string
	Payload       json.RawMessage
	CreatedAt     time.Time
}

// ImportRelationship is a current materialized relationship. The source event
// remains the audit record; applying this row never creates a synthetic event.
type ImportRelationship struct {
	IssueSourceID       string
	TargetIssueSourceID string
	Relation            RelationshipKind
	CreatedAt           time.Time
}

// ImportPromptRunLink and ImportPlanLink contain already-resolved Captain IDs.
// ApplyImport validates only Gavel ownership/ordinal invariants; the installed
// cross-owner foreign keys validate that the authoritative Captain rows exist.
type ImportPromptRunLink struct {
	IssueSourceID string
	PromptRunID   uuid.UUID
	StepKind      StepKind
	Ordinal       int
	CreatedAt     time.Time
}

type ImportPlanLink struct {
	IssueSourceID string
	PlanID        uuid.UUID
	Ordinal       int
	CreatedAt     time.Time
}

// ImportWarning becomes a deterministic migration_warning event. SourceID may
// be supplied by the outer importer; otherwise it is derived from the warning's
// canonical content, so an exact repeated import cannot duplicate warnings.
type ImportWarning struct {
	IssueSourceID string
	SourceID      string
	Code          string
	Message       string
	Payload       json.RawMessage
	CreatedAt     time.Time
	Order         int
}

// ImportBatch is deliberately source-normalized. Decoding Grite exports,
// resolving Captain provider identities, reading plan files, and producing the
// external rollback artifact belong to the outer migration service.
type ImportBatch struct {
	Source               string
	Fingerprint          string
	ExpectedChecksum     string
	RequireNoPriorImport bool
	Workspace            ImportWorkspace
	Issues               []ImportIssue
	Events               []ImportEvent
	Relationships        []ImportRelationship
	RelationshipDeletes  []ImportRelationship
	PromptRunLinks       []ImportPromptRunLink
	PlanLinks            []ImportPlanLink
	Warnings             []ImportWarning
}

type ImportCounts struct {
	IssuesSeen                  int `json:"issuesSeen"`
	WorkspaceCreated            int `json:"workspaceCreated"`
	WorkspaceUpdated            int `json:"workspaceUpdated"`
	IssuesCreated               int `json:"issuesCreated"`
	IssuesUpdated               int `json:"issuesUpdated"`
	AliasesInserted             int `json:"aliasesInserted"`
	EventsInserted              int `json:"eventsInserted"`
	EventsReplayed              int `json:"eventsReplayed"`
	ProjectionEventsInserted    int `json:"projectionEventsInserted"`
	WarningsInserted            int `json:"warningsInserted"`
	WarningsReplayed            int `json:"warningsReplayed"`
	RelationshipsInserted       int `json:"relationshipsInserted"`
	RelationshipsReplayed       int `json:"relationshipsReplayed"`
	RelationshipsDeleted        int `json:"relationshipsDeleted"`
	RelationshipDeletesReplayed int `json:"relationshipDeletesReplayed"`
	PromptRunLinksInserted      int `json:"promptRunLinksInserted"`
	PromptRunLinksReplayed      int `json:"promptRunLinksReplayed"`
	PlanLinksInserted           int `json:"planLinksInserted"`
	PlanLinksReplayed           int `json:"planLinksReplayed"`
}

type ImportAliasKey struct {
	WorkspaceID uuid.UUID `json:"workspaceId"`
	Alias       string    `json:"alias"`
}

type ImportEventKey struct {
	Source   string `json:"source"`
	SourceID string `json:"sourceId"`
}

type ImportIssuePreimage struct {
	Issue   Issue   `json:"issue"`
	Aliases []Alias `json:"aliases"`
}

// ImportRollback captures the before-images and inserted keys required by the
// outer artifact layer to construct a guarded inverse operation. ApplyImport
// itself intentionally does not write files or implement rollback execution.
type ImportRollback struct {
	WorkspaceCreated       bool                  `json:"workspaceCreated"`
	WorkspaceBefore        *Workspace            `json:"workspaceBefore,omitempty"`
	WorkspacePathsBefore   []WorkspacePath       `json:"workspacePathsBefore,omitempty"`
	CreatedIssueIDs        []uuid.UUID           `json:"createdIssueIds,omitempty"`
	IssuePreimages         []ImportIssuePreimage `json:"issuePreimages,omitempty"`
	CreatedAliases         []ImportAliasKey      `json:"createdAliases,omitempty"`
	InsertedEvents         []ImportEventKey      `json:"insertedEvents,omitempty"`
	InsertedRelationships  []Relationship        `json:"insertedRelationships,omitempty"`
	DeletedRelationships   []Relationship        `json:"deletedRelationships,omitempty"`
	InsertedPromptRunLinks []PromptRunLink       `json:"insertedPromptRunLinks,omitempty"`
	InsertedPlanLinks      []PlanLink            `json:"insertedPlanLinks,omitempty"`
}

type ImportResult struct {
	WorkspaceID uuid.UUID      `json:"workspaceId"`
	Counts      ImportCounts   `json:"counts"`
	Checksum    string         `json:"checksum"`
	Rollback    ImportRollback `json:"rollback"`
}

type normalizedImport struct {
	Source               string
	Fingerprint          string
	ExpectedChecksum     string
	RequireNoPriorImport bool
	Workspace            ImportWorkspace
	Issues               []ImportIssue
	Events               []normalizedImportEvent
	Relationships        []ImportRelationship
	RelationshipDeletes  []ImportRelationship
	PromptRunLinks       []ImportPromptRunLink
	PlanLinks            []ImportPlanLink
}

type normalizedImportEvent struct {
	ImportEvent
	Warning bool
}

// ApplyImport atomically applies one normalized legacy snapshot. It uses a
// source/workspace advisory lock plus the repository's relationship lock, and
// never calls CRUD helpers that would synthesize Gavel events. Source events
// and warnings are the only mutations that advance issue version.
func (r *Repository) ApplyImport(ctx context.Context, batch ImportBatch) (*ImportResult, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("%w: native repository is nil", ErrInvalidInput)
	}
	normalized, err := normalizeImportBatch(batch)
	if err != nil {
		return nil, err
	}

	result := &ImportResult{Counts: ImportCounts{IssuesSeen: len(normalized.Issues)}}
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockImportWorkspace(tx, normalized.Source, normalized.Workspace); err != nil {
			return err
		}
		workspace, created, updated, before, beforePaths, err := upsertImportWorkspace(tx, normalized.Workspace)
		if err != nil {
			return err
		}
		result.WorkspaceID = workspace.ID
		result.Rollback.WorkspaceCreated = created
		result.Rollback.WorkspaceBefore = before
		result.Rollback.WorkspacePathsBefore = beforePaths
		if created {
			result.Counts.WorkspaceCreated = 1
		} else if updated {
			result.Counts.WorkspaceUpdated = 1
		}
		if err := lockWorkspaceRelationships(tx, workspace.ID); err != nil {
			return err
		}
		if err := lockImportWorkspaceIssues(tx, workspace.ID); err != nil {
			return err
		}
		if normalized.RequireNoPriorImport {
			exists, err := hasPriorGriteImportEvidence(tx, workspace.ID)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("%w: workspace %s already contains Grite import events or aliases", ErrImportConflict, workspace.ID)
			}
		}
		if normalized.ExpectedChecksum != "" {
			currentChecksum, err := computeImportChecksum(ctx, tx, workspace.ID)
			if err != nil {
				return err
			}
			if currentChecksum != normalized.ExpectedChecksum {
				return fmt.Errorf("%w: workspace checksum drifted: expected %s, got %s", ErrImportConflict, normalized.ExpectedChecksum, currentChecksum)
			}
		}

		issueIDs, desiredPointers, changedIssues, err := applyImportIssues(tx, workspace.ID, normalized.Issues, result)
		if err != nil {
			return err
		}
		if err := applyImportEvents(tx, normalized.Source, issueIDs, normalized.Events, result); err != nil {
			return err
		}
		if err := applyImportRelationshipDeletes(tx, workspace.ID, issueIDs, normalized.RelationshipDeletes, result); err != nil {
			return err
		}
		if err := applyImportRelationships(tx, workspace.ID, issueIDs, normalized.Relationships, result); err != nil {
			return err
		}
		if err := applyImportPromptRunLinks(tx, issueIDs, normalized.PromptRunLinks, result); err != nil {
			return err
		}
		if err := applyImportPlanLinks(tx, issueIDs, normalized.PlanLinks, result); err != nil {
			return err
		}
		if err := applyImportPointers(tx, issueIDs, desiredPointers, changedIssues, result); err != nil {
			return err
		}
		for _, changed := range changedIssues {
			if changed {
				result.Counts.IssuesUpdated++
			}
		}
		if err := validateImportTarget(ctx, tx, workspace.ID, normalized, issueIDs); err != nil {
			return err
		}

		checksum, err := computeImportChecksum(ctx, tx, workspace.ID)
		if err != nil {
			return err
		}
		result.Checksum = checksum
		sortImportRollback(&result.Rollback)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func validateImportTarget(
	ctx context.Context,
	tx *gorm.DB,
	workspaceID uuid.UUID,
	imported *normalizedImport,
	issueIDs map[string]uuid.UUID,
) error {
	repository, _ := NewRepository(tx)
	desiredAliases := make(map[string]uuid.UUID, len(issueIDs))
	for sourceID, issueID := range issueIDs {
		desiredAliases[sourceID] = issueID
	}
	var aliases []struct {
		Alias   string
		IssueID uuid.UUID
	}
	if err := tx.Raw(`
		SELECT alias, issue_id FROM todo_issue_aliases
		WHERE workspace_id = ? AND kind = 'grite'
		ORDER BY alias`, workspaceID).Scan(&aliases).Error; err != nil {
		return err
	}
	if len(aliases) != len(desiredAliases) {
		return fmt.Errorf("%w: target has %d Grite aliases, expected %d", ErrImportConflict, len(aliases), len(desiredAliases))
	}
	for _, alias := range aliases {
		if desiredAliases[alias.Alias] != alias.IssueID {
			return fmt.Errorf("%w: target alias %s resolves to unexpected issue %s", ErrImportConflict, alias.Alias, alias.IssueID)
		}
	}

	for _, input := range imported.Issues {
		current, err := repository.GetIssue(ctx, issueIDs[input.SourceID])
		if err != nil {
			return err
		}
		if !importIssueMatches(current, input) || !sameUUIDPointer(current.ActivePromptRunID, input.ActivePromptRunID) ||
			!sameUUIDPointer(current.SelectedPlanID, input.SelectedPlanID) {
			return fmt.Errorf("%w: target issue %s does not match normalized import", ErrImportConflict, input.SourceID)
		}
	}

	expectedEvents := make(map[string]uuid.UUID, len(imported.Events))
	expectedEventOrder := make(map[uuid.UUID][]string, len(issueIDs))
	for _, event := range imported.Events {
		issueID := issueIDs[event.IssueSourceID]
		expectedEvents[event.SourceID] = issueID
		if !historicalImportAuditEvent(event.Kind, event.SourceID) {
			expectedEventOrder[issueID] = append(expectedEventOrder[issueID], event.SourceID)
		}
	}
	var actualEvents []struct {
		SourceID string
		IssueID  uuid.UUID
		Kind     string
		Sequence int64
	}
	ids := make([]uuid.UUID, 0, len(issueIDs))
	for _, id := range issueIDs {
		ids = append(ids, id)
	}
	if err := tx.Raw(`
		SELECT source_id, issue_id, kind, sequence FROM todo_issue_events
		WHERE source = ? AND source_id IS NOT NULL AND issue_id IN ?
		ORDER BY issue_id, sequence`, imported.Source, ids).Scan(&actualEvents).Error; err != nil {
		return err
	}
	seenEvents := make(map[string]bool, len(expectedEvents))
	actualEventOrder := make(map[uuid.UUID][]string, len(issueIDs))
	for _, event := range actualEvents {
		expectedIssueID, expected := expectedEvents[event.SourceID]
		if !expected {
			if historicalImportAuditEvent(event.Kind, event.SourceID) {
				continue
			}
			return fmt.Errorf("%w: target has unexpected source event %s", ErrImportConflict, event.SourceID)
		}
		if expectedIssueID != event.IssueID {
			return fmt.Errorf("%w: target source event %s has unexpected owner", ErrImportConflict, event.SourceID)
		}
		seenEvents[event.SourceID] = true
		if !historicalImportAuditEvent(event.Kind, event.SourceID) {
			actualEventOrder[event.IssueID] = append(actualEventOrder[event.IssueID], event.SourceID)
		}
	}
	if len(seenEvents) != len(expectedEvents) {
		return fmt.Errorf("%w: target has %d expected source events, expected %d", ErrImportConflict, len(seenEvents), len(expectedEvents))
	}
	for sourceID, issueID := range issueIDs {
		if !slices.Equal(actualEventOrder[issueID], expectedEventOrder[issueID]) {
			return fmt.Errorf("%w: target source event order for issue %s differs from normalized import", ErrImportConflict, sourceID)
		}
	}

	expectedRelationships := make(map[string]time.Time, len(imported.Relationships))
	for _, relationship := range imported.Relationships {
		issueID, targetID := canonicalRelationship(issueIDs[relationship.IssueSourceID], issueIDs[relationship.TargetIssueSourceID], relationship.Relation)
		key := issueID.String() + "\x00" + targetID.String() + "\x00" + string(relationship.Relation)
		expectedRelationships[key] = relationship.CreatedAt
	}
	var actualRelationships []Relationship
	if err := tx.Raw(`
		SELECT workspace_id, issue_id, target_issue_id, relation, created_at
		FROM todo_issue_relationships
		WHERE workspace_id = ? AND issue_id IN ? AND target_issue_id IN ?
		ORDER BY issue_id, target_issue_id, relation`, workspaceID, ids, ids).Scan(&actualRelationships).Error; err != nil {
		return err
	}
	if len(actualRelationships) != len(expectedRelationships) {
		return fmt.Errorf("%w: target has %d imported-issue relationships, expected %d", ErrImportConflict, len(actualRelationships), len(expectedRelationships))
	}
	for _, relationship := range actualRelationships {
		key := relationship.IssueID.String() + "\x00" + relationship.TargetIssueID.String() + "\x00" + string(relationship.Relation)
		createdAt, ok := expectedRelationships[key]
		if !ok || !createdAt.Equal(relationship.CreatedAt) {
			return fmt.Errorf("%w: target relationship %s differs from normalized import", ErrImportConflict, key)
		}
	}

	expectedPromptLinks := make(map[string]ImportPromptRunLink, len(imported.PromptRunLinks))
	for _, link := range imported.PromptRunLinks {
		key := issueIDs[link.IssueSourceID].String() + "\x00" + link.PromptRunID.String()
		expectedPromptLinks[key] = link
	}
	var actualPromptLinks []PromptRunLink
	if err := tx.Raw(`
		SELECT issue_id, prompt_run_id, step_kind, ordinal, created_at
		FROM todo_issue_prompt_runs WHERE issue_id IN ?
		ORDER BY issue_id, prompt_run_id`, ids).Scan(&actualPromptLinks).Error; err != nil {
		return err
	}
	if len(actualPromptLinks) != len(expectedPromptLinks) {
		return fmt.Errorf("%w: target has %d prompt-run links, expected %d", ErrImportConflict, len(actualPromptLinks), len(expectedPromptLinks))
	}
	for _, actual := range actualPromptLinks {
		key := actual.IssueID.String() + "\x00" + actual.PromptRunID.String()
		expected, ok := expectedPromptLinks[key]
		if !ok || actual.StepKind != expected.StepKind || actual.Ordinal != expected.Ordinal || !actual.CreatedAt.Equal(expected.CreatedAt) {
			return fmt.Errorf("%w: target prompt-run link %s differs from normalized import", ErrImportConflict, key)
		}
	}

	expectedPlanLinks := make(map[string]ImportPlanLink, len(imported.PlanLinks))
	for _, link := range imported.PlanLinks {
		key := issueIDs[link.IssueSourceID].String() + "\x00" + link.PlanID.String()
		expectedPlanLinks[key] = link
	}
	var actualPlanLinks []PlanLink
	if err := tx.Raw(`
		SELECT issue_id, plan_id, ordinal, created_at
		FROM todo_issue_plans WHERE issue_id IN ?
		ORDER BY issue_id, plan_id`, ids).Scan(&actualPlanLinks).Error; err != nil {
		return err
	}
	if len(actualPlanLinks) != len(expectedPlanLinks) {
		return fmt.Errorf("%w: target has %d plan links, expected %d", ErrImportConflict, len(actualPlanLinks), len(expectedPlanLinks))
	}
	for _, actual := range actualPlanLinks {
		key := actual.IssueID.String() + "\x00" + actual.PlanID.String()
		expected, ok := expectedPlanLinks[key]
		if !ok || actual.Ordinal != expected.Ordinal || !actual.CreatedAt.Equal(expected.CreatedAt) {
			return fmt.Errorf("%w: target plan link %s differs from normalized import", ErrImportConflict, key)
		}
	}
	return nil
}

func historicalImportAuditEvent(kind, sourceID string) bool {
	switch kind {
	case "migration_warning":
		return strings.HasPrefix(sourceID, "warning:")
	case "migration_checkpoint":
		return strings.HasPrefix(sourceID, "checkpoint:")
	default:
		return false
	}
}

func lockImportWorkspaceIssues(tx *gorm.DB, workspaceID uuid.UUID) error {
	var ids []uuid.UUID
	return tx.Raw(`
		SELECT id FROM todo_issues
		WHERE workspace_id = ?
		ORDER BY id
		FOR UPDATE`, workspaceID).Scan(&ids).Error
}

func normalizeImportBatch(batch ImportBatch) (*normalizedImport, error) {
	source := normalizeToken(batch.Source)
	if source == "" {
		source = DefaultImportSource
	}
	fingerprint := strings.ToLower(strings.TrimSpace(batch.Fingerprint))
	if fingerprint != "" {
		if _, err := hex.DecodeString(fingerprint); err != nil {
			return nil, fmt.Errorf("%w: import fingerprint must be hexadecimal", ErrInvalidInput)
		}
	}
	expectedChecksum := strings.ToLower(strings.TrimSpace(batch.ExpectedChecksum))
	if expectedChecksum != "" {
		if _, err := hex.DecodeString(expectedChecksum); err != nil {
			return nil, fmt.Errorf("%w: expected checksum must be hexadecimal", ErrInvalidInput)
		}
	}
	workspace := NormalizeImportWorkspace(batch.Workspace)
	if workspace.RepoKey == "" && workspace.RootPath == "" {
		return nil, fmt.Errorf("%w: import workspace repo key or path is required", ErrInvalidInput)
	}
	workspace.CreatedAt = stableImportTime(workspace.CreatedAt, time.Unix(0, 0))
	workspace.UpdatedAt = stableImportTime(workspace.UpdatedAt, workspace.CreatedAt)

	issuesBySource := make(map[string]ImportIssue, len(batch.Issues))
	for _, issue := range batch.Issues {
		sourceID, err := normalizeGriteSourceID(issue.SourceID)
		if err != nil {
			return nil, err
		}
		issue.SourceID = sourceID
		issue.Title = strings.TrimSpace(issue.Title)
		if issue.Title == "" {
			return nil, fmt.Errorf("%w: import issue %s has an empty title", ErrInvalidInput, sourceID)
		}
		issue.Labels = normalizeStrings(issue.Labels)
		if issue.Priority == "" {
			issue.Priority = PriorityMedium
		}
		if !issue.Priority.valid() {
			return nil, fmt.Errorf("%w: import issue %s has priority %q", ErrInvalidInput, sourceID, issue.Priority)
		}
		if issue.Status == "" {
			issue.Status = StatusOpen
		}
		if !issue.Status.valid() {
			return nil, fmt.Errorf("%w: import issue %s has status %q", ErrInvalidInput, sourceID, issue.Status)
		}
		// Execution state is a read-time Captain projection. Legacy artifacts may
		// still carry the retired cached value, but it is not imported.
		issue.CreatedAt = stableImportTime(issue.CreatedAt, workspace.CreatedAt)
		issue.UpdatedAt = stableImportTime(issue.UpdatedAt, issue.CreatedAt)
		if existing, ok := issuesBySource[sourceID]; ok {
			if !equalImportIssue(existing, issue) {
				return nil, fmt.Errorf("%w: duplicate issue source %s has different content", ErrImportConflict, sourceID)
			}
			continue
		}
		issuesBySource[sourceID] = issue
	}
	if len(issuesBySource) == 0 {
		return nil, fmt.Errorf("%w: import contains no issues", ErrInvalidInput)
	}
	issues := make([]ImportIssue, 0, len(issuesBySource))
	for _, issue := range issuesBySource {
		issues = append(issues, issue)
	}
	sort.Slice(issues, func(i, j int) bool { return issues[i].SourceID < issues[j].SourceID })

	eventsBySource := make(map[string]normalizedImportEvent, len(batch.Events)+len(batch.Warnings))
	for _, event := range batch.Events {
		normalized, err := normalizeImportEvent(event, issuesBySource)
		if err != nil {
			return nil, err
		}
		if err := addNormalizedImportEvent(eventsBySource, normalized); err != nil {
			return nil, err
		}
	}
	for _, warning := range batch.Warnings {
		normalized, err := normalizeImportWarning(warning, issuesBySource)
		if err != nil {
			return nil, err
		}
		if err := addNormalizedImportEvent(eventsBySource, normalized); err != nil {
			return nil, err
		}
	}
	if fingerprint != "" {
		for _, issue := range issues {
			payload, err := json.Marshal(map[string]string{"fingerprint": fingerprint})
			if err != nil {
				return nil, err
			}
			checkpoint, err := normalizeImportEvent(ImportEvent{
				IssueSourceID: issue.SourceID,
				SourceID:      "checkpoint:" + fingerprint + ":" + issue.SourceID,
				Order:         math.MaxInt - 1,
				Kind:          "migration_checkpoint",
				Actor:         DefaultImportSource,
				Payload:       payload,
				CreatedAt:     issue.UpdatedAt,
			}, issuesBySource)
			if err != nil {
				return nil, err
			}
			if err := addNormalizedImportEvent(eventsBySource, checkpoint); err != nil {
				return nil, err
			}
		}
	}
	events := make([]normalizedImportEvent, 0, len(eventsBySource))
	for _, event := range eventsBySource {
		events = append(events, event)
	}
	sort.Slice(events, func(i, j int) bool {
		if !events[i].CreatedAt.Equal(events[j].CreatedAt) {
			return events[i].CreatedAt.Before(events[j].CreatedAt)
		}
		if events[i].Order != events[j].Order {
			return events[i].Order < events[j].Order
		}
		return events[i].SourceID < events[j].SourceID
	})

	relationships, err := normalizeImportRelationships(batch.Relationships, issuesBySource)
	if err != nil {
		return nil, err
	}
	relationshipDeletes, err := normalizeImportRelationships(batch.RelationshipDeletes, issuesBySource)
	if err != nil {
		return nil, err
	}
	desiredRelationshipKeys := make(map[string]bool, len(relationships))
	for _, relationship := range relationships {
		desiredRelationshipKeys[importRelationshipKey(relationship)] = true
	}
	for _, relationship := range relationshipDeletes {
		if desiredRelationshipKeys[importRelationshipKey(relationship)] {
			return nil, fmt.Errorf("%w: relationship is both desired and removed", ErrImportConflict)
		}
	}
	promptRunLinks, err := normalizeImportPromptRunLinks(batch.PromptRunLinks, issuesBySource)
	if err != nil {
		return nil, err
	}
	planLinks, err := normalizeImportPlanLinks(batch.PlanLinks, issuesBySource)
	if err != nil {
		return nil, err
	}
	return &normalizedImport{
		Source: source, Fingerprint: fingerprint, ExpectedChecksum: expectedChecksum,
		RequireNoPriorImport: batch.RequireNoPriorImport,
		Workspace:            workspace, Issues: issues, Events: events,
		Relationships: relationships, RelationshipDeletes: relationshipDeletes,
		PromptRunLinks: promptRunLinks, PlanLinks: planLinks,
	}, nil
}

func hasPriorGriteImportEvidence(tx *gorm.DB, workspaceID uuid.UUID) (bool, error) {
	var exists bool
	err := tx.Raw(`
		SELECT EXISTS(
			SELECT 1
			FROM todo_issue_events AS event
			JOIN todo_issues AS issue ON issue.id = event.issue_id
			WHERE issue.workspace_id = ? AND event.source = ?
			UNION ALL
			SELECT 1
			FROM todo_issue_aliases AS alias
			WHERE alias.workspace_id = ? AND alias.kind = 'grite'
		)`, workspaceID, DefaultImportSource, workspaceID).Scan(&exists).Error
	return exists, err
}

func normalizeImportEvent(event ImportEvent, issues map[string]ImportIssue) (normalizedImportEvent, error) {
	issueSourceID, err := normalizeGriteSourceID(event.IssueSourceID)
	if err != nil {
		return normalizedImportEvent{}, err
	}
	issue, ok := issues[issueSourceID]
	if !ok {
		return normalizedImportEvent{}, fmt.Errorf("%w: event %q references unknown issue %s", ErrInvalidInput, event.SourceID, issueSourceID)
	}
	event.IssueSourceID = issueSourceID
	event.SourceID = strings.TrimSpace(event.SourceID)
	if event.SourceID == "" {
		return normalizedImportEvent{}, fmt.Errorf("%w: import event source ID is required", ErrInvalidInput)
	}
	event.Kind = normalizeToken(event.Kind)
	if event.Kind == "" {
		return normalizedImportEvent{}, fmt.Errorf("%w: import event %s kind is required", ErrInvalidInput, event.SourceID)
	}
	event.Actor = strings.TrimSpace(event.Actor)
	event.Payload, err = canonicalImportJSON(event.Payload)
	if err != nil {
		return normalizedImportEvent{}, fmt.Errorf("%w: import event %s payload: %v", ErrInvalidInput, event.SourceID, err)
	}
	event.CreatedAt = stableImportTime(event.CreatedAt, issue.UpdatedAt)
	return normalizedImportEvent{ImportEvent: event}, nil
}

func normalizeImportWarning(warning ImportWarning, issues map[string]ImportIssue) (normalizedImportEvent, error) {
	issueSourceID, err := normalizeGriteSourceID(warning.IssueSourceID)
	if err != nil {
		return normalizedImportEvent{}, err
	}
	issue, ok := issues[issueSourceID]
	if !ok {
		return normalizedImportEvent{}, fmt.Errorf("%w: warning references unknown issue %s", ErrInvalidInput, issueSourceID)
	}
	code := normalizeToken(warning.Code)
	message := strings.TrimSpace(warning.Message)
	if code == "" || message == "" {
		return normalizedImportEvent{}, fmt.Errorf("%w: warning code and message are required", ErrInvalidInput)
	}
	details, err := canonicalImportJSON(warning.Payload)
	if err != nil {
		return normalizedImportEvent{}, fmt.Errorf("%w: warning %s payload: %v", ErrInvalidInput, code, err)
	}
	var decoded any
	if err := json.Unmarshal(details, &decoded); err != nil {
		return normalizedImportEvent{}, err
	}
	payload, err := json.Marshal(map[string]any{"code": code, "message": message, "details": decoded})
	if err != nil {
		return normalizedImportEvent{}, err
	}
	sourceID := strings.TrimSpace(warning.SourceID)
	if sourceID == "" {
		sum := sha256.Sum256(bytes.Join([][]byte{[]byte(issueSourceID), []byte(code), []byte(message), payload}, []byte{0}))
		sourceID = "warning:" + hex.EncodeToString(sum[:])
	}
	return normalizedImportEvent{ImportEvent: ImportEvent{
		IssueSourceID: issueSourceID,
		SourceID:      sourceID,
		Kind:          "migration_warning",
		Actor:         DefaultImportSource,
		Body:          message,
		Payload:       payload,
		CreatedAt:     stableImportTime(warning.CreatedAt, issue.UpdatedAt),
		Order:         warning.Order,
	}, Warning: true}, nil
}

func addNormalizedImportEvent(events map[string]normalizedImportEvent, event normalizedImportEvent) error {
	if existing, ok := events[event.SourceID]; ok {
		if !equalNormalizedImportEvent(existing, event) {
			return fmt.Errorf("%w: source event %q has different content in one batch", ErrEventConflict, event.SourceID)
		}
		return nil
	}
	events[event.SourceID] = event
	return nil
}

func normalizeImportRelationships(inputs []ImportRelationship, issues map[string]ImportIssue) ([]ImportRelationship, error) {
	byKey := map[string]ImportRelationship{}
	for _, relationship := range inputs {
		issueSourceID, err := normalizeGriteSourceID(relationship.IssueSourceID)
		if err != nil {
			return nil, err
		}
		targetSourceID, err := normalizeGriteSourceID(relationship.TargetIssueSourceID)
		if err != nil {
			return nil, err
		}
		if _, ok := issues[issueSourceID]; !ok {
			return nil, fmt.Errorf("%w: relationship references unknown issue %s", ErrInvalidInput, issueSourceID)
		}
		if _, ok := issues[targetSourceID]; !ok {
			return nil, fmt.Errorf("%w: relationship references unknown target %s", ErrInvalidInput, targetSourceID)
		}
		if issueSourceID == targetSourceID {
			return nil, fmt.Errorf("%w: legacy issue %s", ErrSelfRelationship, issueSourceID)
		}
		if !relationship.Relation.valid() {
			return nil, fmt.Errorf("%w: unsupported import relationship %q", ErrInvalidInput, relationship.Relation)
		}
		relationship.IssueSourceID = issueSourceID
		relationship.TargetIssueSourceID = targetSourceID
		relationship.CreatedAt = stableImportTime(relationship.CreatedAt, issues[issueSourceID].UpdatedAt)
		keyIssue, keyTarget := issueSourceID, targetSourceID
		if relationship.Relation == RelationshipRelatedTo && keyIssue > keyTarget {
			keyIssue, keyTarget = keyTarget, keyIssue
		}
		key := keyIssue + "\x00" + keyTarget + "\x00" + string(relationship.Relation)
		if existing, ok := byKey[key]; !ok || relationship.CreatedAt.Before(existing.CreatedAt) {
			byKey[key] = relationship
		}
	}
	out := make([]ImportRelationship, 0, len(byKey))
	for _, relationship := range byKey {
		out = append(out, relationship)
	}
	sort.Slice(out, func(i, j int) bool {
		left := out[i].IssueSourceID + "\x00" + out[i].TargetIssueSourceID + "\x00" + string(out[i].Relation)
		right := out[j].IssueSourceID + "\x00" + out[j].TargetIssueSourceID + "\x00" + string(out[j].Relation)
		return left < right
	})
	return out, nil
}

func importRelationshipKey(relationship ImportRelationship) string {
	issueID, targetID := relationship.IssueSourceID, relationship.TargetIssueSourceID
	if relationship.Relation == RelationshipRelatedTo && issueID > targetID {
		issueID, targetID = targetID, issueID
	}
	return issueID + "\x00" + targetID + "\x00" + string(relationship.Relation)
}

func normalizeImportPromptRunLinks(inputs []ImportPromptRunLink, issues map[string]ImportIssue) ([]ImportPromptRunLink, error) {
	byRun := map[uuid.UUID]ImportPromptRunLink{}
	byOrdinal := map[string]uuid.UUID{}
	for _, link := range inputs {
		sourceID, err := normalizeGriteSourceID(link.IssueSourceID)
		if err != nil {
			return nil, err
		}
		issue, ok := issues[sourceID]
		if !ok {
			return nil, fmt.Errorf("%w: prompt run link references unknown issue %s", ErrInvalidInput, sourceID)
		}
		if link.PromptRunID == uuid.Nil || !link.StepKind.valid() || link.Ordinal < 0 {
			return nil, fmt.Errorf("%w: invalid prompt run link for issue %s", ErrInvalidInput, sourceID)
		}
		link.IssueSourceID = sourceID
		link.CreatedAt = stableImportTime(link.CreatedAt, issue.UpdatedAt)
		if existing, ok := byRun[link.PromptRunID]; ok {
			if existing.IssueSourceID != link.IssueSourceID || existing.StepKind != link.StepKind || existing.Ordinal != link.Ordinal {
				return nil, fmt.Errorf("%w: prompt run %s has multiple import attachments", ErrLinkConflict, link.PromptRunID)
			}
			continue
		}
		ordinalKey := sourceID + "\x00" + string(link.StepKind) + "\x00" + fmt.Sprint(link.Ordinal)
		if existing, ok := byOrdinal[ordinalKey]; ok && existing != link.PromptRunID {
			return nil, fmt.Errorf("%w: issue %s prompt %s ordinal %d is duplicated", ErrLinkConflict, sourceID, link.StepKind, link.Ordinal)
		}
		byRun[link.PromptRunID] = link
		byOrdinal[ordinalKey] = link.PromptRunID
	}
	out := make([]ImportPromptRunLink, 0, len(byRun))
	for _, link := range byRun {
		out = append(out, link)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].IssueSourceID != out[j].IssueSourceID {
			return out[i].IssueSourceID < out[j].IssueSourceID
		}
		if out[i].StepKind != out[j].StepKind {
			return out[i].StepKind < out[j].StepKind
		}
		if out[i].Ordinal != out[j].Ordinal {
			return out[i].Ordinal < out[j].Ordinal
		}
		return out[i].PromptRunID.String() < out[j].PromptRunID.String()
	})
	return out, nil
}

func normalizeImportPlanLinks(inputs []ImportPlanLink, issues map[string]ImportIssue) ([]ImportPlanLink, error) {
	byPlan := map[string]ImportPlanLink{}
	byOrdinal := map[string]uuid.UUID{}
	for _, link := range inputs {
		sourceID, err := normalizeGriteSourceID(link.IssueSourceID)
		if err != nil {
			return nil, err
		}
		issue, ok := issues[sourceID]
		if !ok {
			return nil, fmt.Errorf("%w: plan link references unknown issue %s", ErrInvalidInput, sourceID)
		}
		if link.PlanID == uuid.Nil || link.Ordinal < 0 {
			return nil, fmt.Errorf("%w: invalid plan link for issue %s", ErrInvalidInput, sourceID)
		}
		link.IssueSourceID = sourceID
		link.CreatedAt = stableImportTime(link.CreatedAt, issue.UpdatedAt)
		planKey := sourceID + "\x00" + link.PlanID.String()
		if existing, ok := byPlan[planKey]; ok {
			if existing.Ordinal != link.Ordinal {
				return nil, fmt.Errorf("%w: plan %s has multiple ordinals for issue %s", ErrLinkConflict, link.PlanID, sourceID)
			}
			continue
		}
		ordinalKey := sourceID + "\x00" + fmt.Sprint(link.Ordinal)
		if existing, ok := byOrdinal[ordinalKey]; ok && existing != link.PlanID {
			return nil, fmt.Errorf("%w: issue %s plan ordinal %d is duplicated", ErrLinkConflict, sourceID, link.Ordinal)
		}
		byPlan[planKey] = link
		byOrdinal[ordinalKey] = link.PlanID
	}
	out := make([]ImportPlanLink, 0, len(byPlan))
	for _, link := range byPlan {
		out = append(out, link)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].IssueSourceID != out[j].IssueSourceID {
			return out[i].IssueSourceID < out[j].IssueSourceID
		}
		if out[i].Ordinal != out[j].Ordinal {
			return out[i].Ordinal < out[j].Ordinal
		}
		return out[i].PlanID.String() < out[j].PlanID.String()
	})
	return out, nil
}

func lockImportWorkspace(tx *gorm.DB, source string, workspace ImportWorkspace) error {
	identity := workspace.RepoKey
	if identity == "" {
		identity = "path:" + workspace.RootPath
	}
	return tx.Exec(`SELECT pg_advisory_xact_lock(hashtextextended(?, 0))`, "todo-import:"+source+":"+identity).Error
}

func upsertImportWorkspace(tx *gorm.DB, input ImportWorkspace) (*Workspace, bool, bool, *Workspace, []WorkspacePath, error) {
	var repoID, pathID uuid.UUID
	if input.RepoKey != "" {
		var row struct{ ID uuid.UUID }
		result := tx.Raw(`SELECT id FROM todo_workspaces WHERE repo_key = ?`, input.RepoKey).Scan(&row)
		if result.Error != nil {
			return nil, false, false, nil, nil, result.Error
		}
		if result.RowsAffected > 0 {
			repoID = row.ID
		}
	}
	if input.RootPath != "" {
		var row struct{ WorkspaceID uuid.UUID }
		result := tx.Raw(`SELECT workspace_id FROM todo_workspace_paths WHERE path = ?`, input.RootPath).Scan(&row)
		if result.Error != nil {
			return nil, false, false, nil, nil, result.Error
		}
		if result.RowsAffected > 0 {
			pathID = row.WorkspaceID
		}
	}
	if repoID != uuid.Nil && pathID != uuid.Nil && repoID != pathID {
		return nil, false, false, nil, nil, fmt.Errorf("%w: repo key and path resolve to different workspaces", ErrWorkspaceConflict)
	}
	id := repoID
	if id == uuid.Nil {
		id = pathID
	}
	if id != uuid.Nil {
		if input.ID != uuid.Nil && input.ID != id {
			return nil, false, false, nil, nil, fmt.Errorf("%w: import workspace identity belongs to %s, not %s", ErrWorkspaceConflict, id, input.ID)
		}
		var locked struct{ ID uuid.UUID }
		if err := tx.Raw(`SELECT id FROM todo_workspaces WHERE id = ? FOR UPDATE`, id).Scan(&locked).Error; err != nil {
			return nil, false, false, nil, nil, err
		}
		repository, _ := NewRepository(tx)
		before, err := repository.GetWorkspace(tx.Statement.Context, id)
		if err != nil {
			return nil, false, false, nil, nil, err
		}
		paths, err := repository.ListWorkspacePaths(tx.Statement.Context, id)
		if err != nil {
			return nil, false, false, nil, nil, err
		}
		if input.RepoKey != "" && before.RepoKey != "" && before.RepoKey != input.RepoKey {
			return nil, false, false, nil, nil, fmt.Errorf("%w: workspace %s has repo key %q", ErrWorkspaceConflict, id, before.RepoKey)
		}
		desiredRepoKey := before.RepoKey
		if input.RepoKey != "" {
			desiredRepoKey = input.RepoKey
		}
		desiredDisplayName := before.DisplayName
		if input.DisplayName != "" {
			desiredDisplayName = input.DisplayName
		}
		desiredRootPath := before.RootPath
		if input.RootPath != "" {
			desiredRootPath = input.RootPath
		}
		changed := before.RepoKey != desiredRepoKey || before.DisplayName != desiredDisplayName || before.RootPath != desiredRootPath
		if changed {
			if err := tx.Exec(`
				UPDATE todo_workspaces
				SET repo_key = NULLIF(?, ''), display_name = NULLIF(?, ''), updated_at = ?
				WHERE id = ?`, desiredRepoKey, desiredDisplayName, input.UpdatedAt, id).Error; err != nil {
				return nil, false, false, nil, nil, mapUniqueError(err, ErrWorkspaceConflict, "workspace %s", id)
			}
			if before.RootPath != desiredRootPath {
				if err := setImportedPrimaryPath(tx, id, desiredRootPath, input.UpdatedAt); err != nil {
					return nil, false, false, nil, nil, err
				}
			}
		}
		current, err := repository.GetWorkspace(tx.Statement.Context, id)
		return current, false, changed, before, paths, err
	}

	id = input.ID
	if id == uuid.Nil {
		id = DeterministicImportWorkspaceID(input)
	}
	result := tx.Exec(`
		INSERT INTO todo_workspaces (id, repo_key, display_name, created_at, updated_at)
		VALUES (?, NULLIF(?, ''), NULLIF(?, ''), ?, ?)`,
		id, input.RepoKey, input.DisplayName, input.CreatedAt, input.UpdatedAt)
	if result.Error != nil {
		return nil, false, false, nil, nil, mapUniqueError(result.Error, ErrWorkspaceConflict, "workspace %s", id)
	}
	if input.RootPath != "" {
		if err := tx.Exec(`
			INSERT INTO todo_workspace_paths (workspace_id, path, is_primary, created_at, updated_at)
			VALUES (?, ?, true, ?, ?)`, id, input.RootPath, input.CreatedAt, input.UpdatedAt).Error; err != nil {
			return nil, false, false, nil, nil, mapUniqueError(err, ErrWorkspaceConflict, "workspace path %q", input.RootPath)
		}
	}
	repository, _ := NewRepository(tx)
	workspace, err := repository.GetWorkspace(tx.Statement.Context, id)
	return workspace, true, false, nil, nil, err
}

func setImportedPrimaryPath(tx *gorm.DB, workspaceID uuid.UUID, path string, updatedAt time.Time) error {
	if err := tx.Exec(`
		UPDATE todo_workspace_paths SET is_primary = false, updated_at = ?
		WHERE workspace_id = ? AND is_primary`, updatedAt, workspaceID).Error; err != nil {
		return err
	}
	result := tx.Exec(`
		INSERT INTO todo_workspace_paths (workspace_id, path, is_primary, created_at, updated_at)
		VALUES (?, ?, true, ?, ?)
		ON CONFLICT (workspace_id, path) DO UPDATE
		SET is_primary = true, updated_at = EXCLUDED.updated_at`, workspaceID, path, updatedAt, updatedAt)
	if result.Error != nil {
		return mapUniqueError(result.Error, ErrWorkspaceConflict, "workspace path %q", path)
	}
	return nil
}

type importPointers struct {
	ActivePromptRunID *uuid.UUID
	SelectedPlanID    *uuid.UUID
	UpdatedAt         time.Time
}

func applyImportIssues(
	tx *gorm.DB,
	workspaceID uuid.UUID,
	issues []ImportIssue,
	result *ImportResult,
) (map[string]uuid.UUID, map[string]importPointers, map[string]bool, error) {
	issueIDs := make(map[string]uuid.UUID, len(issues))
	pointers := make(map[string]importPointers, len(issues))
	changedIssues := make(map[string]bool, len(issues))
	repository, _ := NewRepository(tx)
	for _, input := range issues {
		var alias struct {
			IssueID uuid.UUID
			Kind    string
		}
		aliasResult := tx.Raw(`
			SELECT issue_id, COALESCE(kind, '') AS kind FROM todo_issue_aliases
			WHERE workspace_id = ? AND alias = ?`, workspaceID, input.SourceID).Scan(&alias)
		if aliasResult.Error != nil {
			return nil, nil, nil, aliasResult.Error
		}
		id := alias.IssueID
		created := false
		if aliasResult.RowsAffected == 0 {
			id = input.ID
			if id == uuid.Nil {
				id = uuid.NewSHA1(importNamespace, []byte("issue:"+workspaceID.String()+":"+input.SourceID))
			}
			insert := tx.Exec(`
				INSERT INTO todo_issues
					(id, workspace_id, title, body, verification, labels, priority, status,
					 version, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`,
				id, workspaceID, input.Title, input.Body, input.Verification, pq.Array(input.Labels),
				input.Priority, input.Status, input.CreatedAt, input.UpdatedAt)
			if insert.Error != nil {
				return nil, nil, nil, fmt.Errorf("insert imported issue %s: %w", input.SourceID, insert.Error)
			}
			created = true
			result.Counts.IssuesCreated++
			result.Rollback.CreatedIssueIDs = append(result.Rollback.CreatedIssueIDs, id)
		} else {
			if alias.Kind != "" && alias.Kind != "grite" {
				return nil, nil, nil, fmt.Errorf("%w: alias %s is owned by kind %q", ErrAliasConflict, input.SourceID, alias.Kind)
			}
			if input.ID != uuid.Nil && input.ID != id {
				return nil, nil, nil, fmt.Errorf("%w: alias %s belongs to issue %s, not %s", ErrAliasConflict, input.SourceID, id, input.ID)
			}
			if _, err := lockIssueWithoutVersion(tx, id); err != nil {
				return nil, nil, nil, err
			}
			before, err := repository.GetIssue(tx.Statement.Context, id)
			if err != nil {
				return nil, nil, nil, err
			}
			aliases, err := repository.ListAliases(tx.Statement.Context, id)
			if err != nil {
				return nil, nil, nil, err
			}
			result.Rollback.IssuePreimages = append(result.Rollback.IssuePreimages, ImportIssuePreimage{Issue: *before, Aliases: aliases})
			changedIssues[input.SourceID] = !importIssueMatches(before, input)
			if changedIssues[input.SourceID] {
				status := input.Status
				if input.ActivePromptRunID != nil {
					// Once an authoritative Captain run is selected, reconciliation
					// owns durable status. Re-importing an older snapshot must not
					// reset it and manufacture another event.
					status = before.Status
				}
				if err := tx.Exec(`
					UPDATE todo_issues
					SET title = ?, body = ?, verification = ?, labels = ?, priority = ?, status = ?,
					    created_at = ?, updated_at = ?
					WHERE id = ?`, input.Title, input.Body, input.Verification, pq.Array(input.Labels),
					input.Priority, status, input.CreatedAt, input.UpdatedAt, id).Error; err != nil {
					return nil, nil, nil, err
				}
			}
		}
		if !created {
			var locked struct{ ID uuid.UUID }
			if err := tx.Raw(`SELECT id FROM todo_issues WHERE id = ? FOR UPDATE`, id).Scan(&locked).Error; err != nil {
				return nil, nil, nil, err
			}
		}
		aliasInsert := tx.Exec(`
			INSERT INTO todo_issue_aliases (workspace_id, alias, issue_id, kind, created_at)
			VALUES (?, ?, ?, 'grite', ?)
			ON CONFLICT (workspace_id, alias) DO NOTHING`, workspaceID, input.SourceID, id, input.CreatedAt)
		if aliasInsert.Error != nil {
			return nil, nil, nil, aliasInsert.Error
		}
		if aliasInsert.RowsAffected == 1 {
			result.Counts.AliasesInserted++
			result.Rollback.CreatedAliases = append(result.Rollback.CreatedAliases, ImportAliasKey{WorkspaceID: workspaceID, Alias: input.SourceID})
		}
		var owner struct{ IssueID uuid.UUID }
		if err := tx.Raw(`SELECT issue_id FROM todo_issue_aliases WHERE workspace_id = ? AND alias = ?`, workspaceID, input.SourceID).Scan(&owner).Error; err != nil {
			return nil, nil, nil, err
		}
		if owner.IssueID != id {
			return nil, nil, nil, fmt.Errorf("%w: workspace %s alias %q", ErrAliasConflict, workspaceID, input.SourceID)
		}
		if err := tx.Exec(`
			UPDATE todo_issue_aliases SET kind = 'grite'
			WHERE workspace_id = ? AND alias = ? AND kind IS NULL`, workspaceID, input.SourceID).Error; err != nil {
			return nil, nil, nil, err
		}
		issueIDs[input.SourceID] = id
		pointers[input.SourceID] = importPointers{
			ActivePromptRunID: cloneUUIDPointer(input.ActivePromptRunID),
			SelectedPlanID:    cloneUUIDPointer(input.SelectedPlanID),
			UpdatedAt:         input.UpdatedAt,
		}
	}
	return issueIDs, pointers, changedIssues, nil
}

func lockIssueWithoutVersion(tx *gorm.DB, id uuid.UUID) (*lockedIssue, error) {
	var issue lockedIssue
	result := tx.Raw(`SELECT id, workspace_id, version FROM todo_issues WHERE id = ? FOR UPDATE`, id).Scan(&issue)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("%w: issue %s", ErrNotFound, id)
	}
	return &issue, nil
}

func applyImportEvents(tx *gorm.DB, source string, issueIDs map[string]uuid.UUID, events []normalizedImportEvent, result *ImportResult) error {
	nextSequences := map[uuid.UUID]int64{}
	for _, input := range events {
		issueID := issueIDs[input.IssueSourceID]
		var existing struct {
			IssueID      uuid.UUID
			Kind         string
			Actor        string
			Body         string
			CreatedAt    time.Time
			PayloadEqual bool
		}
		query := tx.Raw(`
			SELECT issue_id, kind, COALESCE(actor, '') AS actor, COALESCE(body, '') AS body,
			       created_at, payload = CAST(? AS jsonb) AS payload_equal
			FROM todo_issue_events WHERE source = ? AND source_id = ?`,
			string(input.Payload), source, input.SourceID).Scan(&existing)
		if query.Error != nil {
			return query.Error
		}
		if query.RowsAffected > 0 {
			exact := existing.IssueID == issueID && existing.Kind == input.Kind &&
				existing.Actor == input.Actor && existing.Body == input.Body && existing.PayloadEqual &&
				existing.CreatedAt.Equal(input.CreatedAt)
			if !exact {
				return fmt.Errorf("%w: event source %q/%q already has different content", ErrEventConflict, source, input.SourceID)
			}
			result.Counts.EventsReplayed++
			if input.Warning {
				result.Counts.WarningsReplayed++
			}
			continue
		}
		sequence, ok := nextSequences[issueID]
		if !ok {
			if err := tx.Raw(`SELECT COALESCE(MAX(sequence), 0) + 1 FROM todo_issue_events WHERE issue_id = ?`, issueID).Scan(&sequence).Error; err != nil {
				return err
			}
		}
		eventID := uuid.NewSHA1(importNamespace, []byte("event:"+source+":"+input.SourceID))
		insert := tx.Exec(`
			INSERT INTO todo_issue_events
				(id, issue_id, sequence, kind, actor, body, payload, source, source_id, created_at)
			VALUES (?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), CAST(? AS jsonb), ?, ?, ?)`,
			eventID, issueID, sequence, input.Kind, input.Actor, input.Body,
			string(input.Payload), source, input.SourceID, input.CreatedAt)
		if insert.Error != nil {
			return mapUniqueError(insert.Error, ErrEventConflict, "event source %q/%q", source, input.SourceID)
		}
		if err := tx.Exec(`UPDATE todo_issues SET version = version + 1 WHERE id = ?`, issueID).Error; err != nil {
			return err
		}
		nextSequences[issueID] = sequence + 1
		result.Counts.EventsInserted++
		if input.Warning {
			result.Counts.WarningsInserted++
		}
		result.Rollback.InsertedEvents = append(result.Rollback.InsertedEvents, ImportEventKey{Source: source, SourceID: input.SourceID})
	}
	return nil
}

func applyImportRelationshipDeletes(tx *gorm.DB, workspaceID uuid.UUID, issueIDs map[string]uuid.UUID, relationships []ImportRelationship, result *ImportResult) error {
	for _, input := range relationships {
		issueID := issueIDs[input.IssueSourceID]
		targetID := issueIDs[input.TargetIssueSourceID]
		storedIssueID, storedTargetID := canonicalRelationship(issueID, targetID, input.Relation)
		var existing Relationship
		query := tx.Raw(`
			SELECT workspace_id, issue_id, target_issue_id, relation, created_at
			FROM todo_issue_relationships
			WHERE workspace_id = ? AND issue_id = ? AND target_issue_id = ? AND relation = ?`,
			workspaceID, storedIssueID, storedTargetID, input.Relation).Scan(&existing)
		if query.Error != nil {
			return query.Error
		}
		if query.RowsAffected == 0 {
			result.Counts.RelationshipDeletesReplayed++
			continue
		}
		deleted := tx.Exec(`
			DELETE FROM todo_issue_relationships
			WHERE workspace_id = ? AND issue_id = ? AND target_issue_id = ? AND relation = ?`,
			workspaceID, storedIssueID, storedTargetID, input.Relation)
		if deleted.Error != nil {
			return deleted.Error
		}
		if deleted.RowsAffected != 1 {
			return fmt.Errorf("%w: imported relationship changed during delete", ErrImportConflict)
		}
		result.Counts.RelationshipsDeleted++
		result.Rollback.DeletedRelationships = append(result.Rollback.DeletedRelationships, existing)
	}
	return nil
}

func applyImportRelationships(tx *gorm.DB, workspaceID uuid.UUID, issueIDs map[string]uuid.UUID, relationships []ImportRelationship, result *ImportResult) error {
	for _, input := range relationships {
		issueID := issueIDs[input.IssueSourceID]
		targetID := issueIDs[input.TargetIssueSourceID]
		storedIssueID, storedTargetID := canonicalRelationship(issueID, targetID, input.Relation)
		var existing Relationship
		query := tx.Raw(`
			SELECT workspace_id, issue_id, target_issue_id, relation, created_at
			FROM todo_issue_relationships
			WHERE workspace_id = ? AND issue_id = ? AND target_issue_id = ? AND relation = ?`,
			workspaceID, storedIssueID, storedTargetID, input.Relation).Scan(&existing)
		if query.Error != nil {
			return query.Error
		}
		if query.RowsAffected > 0 {
			result.Counts.RelationshipsReplayed++
			continue
		}
		if input.Relation == RelationshipDependsOn {
			cycle, err := dependencyPathExists(tx, workspaceID, targetID, issueID)
			if err != nil {
				return err
			}
			if cycle {
				return fmt.Errorf("%w: adding imported %s -> %s", ErrRelationshipCycle, input.IssueSourceID, input.TargetIssueSourceID)
			}
		}
		created := Relationship{WorkspaceID: workspaceID, IssueID: storedIssueID, TargetIssueID: storedTargetID, Relation: input.Relation, CreatedAt: input.CreatedAt}
		insert := tx.Exec(`
			INSERT INTO todo_issue_relationships (workspace_id, issue_id, target_issue_id, relation, created_at)
			VALUES (?, ?, ?, ?, ?)`, workspaceID, storedIssueID, storedTargetID, input.Relation, input.CreatedAt)
		if insert.Error != nil {
			return mapUniqueError(insert.Error, ErrRelationshipExists, "%s %s %s", issueID, input.Relation, targetID)
		}
		result.Counts.RelationshipsInserted++
		result.Rollback.InsertedRelationships = append(result.Rollback.InsertedRelationships, created)
	}
	return nil
}

func applyImportPromptRunLinks(tx *gorm.DB, issueIDs map[string]uuid.UUID, links []ImportPromptRunLink, result *ImportResult) error {
	for _, input := range links {
		issueID := issueIDs[input.IssueSourceID]
		var existing PromptRunLink
		query := tx.Raw(`
			SELECT issue_id, prompt_run_id, step_kind, ordinal, created_at
			FROM todo_issue_prompt_runs WHERE prompt_run_id = ?`, input.PromptRunID).Scan(&existing)
		if query.Error != nil {
			return query.Error
		}
		if query.RowsAffected > 0 {
			if existing.IssueID != issueID || existing.StepKind != input.StepKind || existing.Ordinal != input.Ordinal {
				return fmt.Errorf("%w: prompt run %s already has a different attachment", ErrLinkConflict, input.PromptRunID)
			}
			result.Counts.PromptRunLinksReplayed++
			continue
		}
		var ordinalOwner struct{ PromptRunID uuid.UUID }
		ordinalQuery := tx.Raw(`
			SELECT prompt_run_id FROM todo_issue_prompt_runs
			WHERE issue_id = ? AND step_kind = ? AND ordinal = ?`, issueID, input.StepKind, input.Ordinal).Scan(&ordinalOwner)
		if ordinalQuery.Error != nil {
			return ordinalQuery.Error
		}
		if ordinalQuery.RowsAffected > 0 && ordinalOwner.PromptRunID != input.PromptRunID {
			return fmt.Errorf("%w: issue %s prompt %s ordinal %d belongs to %s", ErrLinkConflict, issueID, input.StepKind, input.Ordinal, ordinalOwner.PromptRunID)
		}
		insert := tx.Exec(`
			INSERT INTO todo_issue_prompt_runs (issue_id, prompt_run_id, step_kind, ordinal, created_at)
			VALUES (?, ?, ?, ?, ?)`, issueID, input.PromptRunID, input.StepKind, input.Ordinal, input.CreatedAt)
		if insert.Error != nil {
			return mapUniqueError(insert.Error, ErrLinkConflict, "prompt run %s", input.PromptRunID)
		}
		created := PromptRunLink{IssueID: issueID, PromptRunID: input.PromptRunID, StepKind: input.StepKind, Ordinal: input.Ordinal, CreatedAt: input.CreatedAt}
		result.Counts.PromptRunLinksInserted++
		result.Rollback.InsertedPromptRunLinks = append(result.Rollback.InsertedPromptRunLinks, created)
	}
	return nil
}

func applyImportPlanLinks(tx *gorm.DB, issueIDs map[string]uuid.UUID, links []ImportPlanLink, result *ImportResult) error {
	for _, input := range links {
		issueID := issueIDs[input.IssueSourceID]
		var existing PlanLink
		query := tx.Raw(`
			SELECT issue_id, plan_id, ordinal, created_at
			FROM todo_issue_plans WHERE issue_id = ? AND plan_id = ?`, issueID, input.PlanID).Scan(&existing)
		if query.Error != nil {
			return query.Error
		}
		if query.RowsAffected > 0 {
			if existing.Ordinal != input.Ordinal {
				return fmt.Errorf("%w: plan %s already has ordinal %d", ErrLinkConflict, input.PlanID, existing.Ordinal)
			}
			result.Counts.PlanLinksReplayed++
			continue
		}
		var ordinalOwner struct{ PlanID uuid.UUID }
		ordinalQuery := tx.Raw(`SELECT plan_id FROM todo_issue_plans WHERE issue_id = ? AND ordinal = ?`, issueID, input.Ordinal).Scan(&ordinalOwner)
		if ordinalQuery.Error != nil {
			return ordinalQuery.Error
		}
		if ordinalQuery.RowsAffected > 0 && ordinalOwner.PlanID != input.PlanID {
			return fmt.Errorf("%w: issue %s plan ordinal %d belongs to %s", ErrLinkConflict, issueID, input.Ordinal, ordinalOwner.PlanID)
		}
		insert := tx.Exec(`
			INSERT INTO todo_issue_plans (issue_id, plan_id, ordinal, created_at)
			VALUES (?, ?, ?, ?)`, issueID, input.PlanID, input.Ordinal, input.CreatedAt)
		if insert.Error != nil {
			return mapUniqueError(insert.Error, ErrLinkConflict, "plan %s", input.PlanID)
		}
		created := PlanLink{IssueID: issueID, PlanID: input.PlanID, Ordinal: input.Ordinal, CreatedAt: input.CreatedAt}
		result.Counts.PlanLinksInserted++
		result.Rollback.InsertedPlanLinks = append(result.Rollback.InsertedPlanLinks, created)
	}
	return nil
}

func applyImportPointers(
	tx *gorm.DB,
	issueIDs map[string]uuid.UUID,
	pointers map[string]importPointers,
	changed map[string]bool,
	result *ImportResult,
) error {
	for sourceID, desired := range pointers {
		issueID := issueIDs[sourceID]
		if desired.ActivePromptRunID != nil {
			var linked bool
			if err := tx.Raw(`SELECT EXISTS(SELECT 1 FROM todo_issue_prompt_runs WHERE issue_id = ? AND prompt_run_id = ?)`, issueID, *desired.ActivePromptRunID).Scan(&linked).Error; err != nil {
				return err
			}
			if !linked {
				return fmt.Errorf("%w: active prompt run %s is not linked to imported issue %s", ErrLinkConflict, *desired.ActivePromptRunID, sourceID)
			}
		}
		if desired.SelectedPlanID != nil {
			var linked bool
			if err := tx.Raw(`SELECT EXISTS(SELECT 1 FROM todo_issue_plans WHERE issue_id = ? AND plan_id = ?)`, issueID, *desired.SelectedPlanID).Scan(&linked).Error; err != nil {
				return err
			}
			if !linked {
				return fmt.Errorf("%w: selected plan %s is not linked to imported issue %s", ErrLinkConflict, *desired.SelectedPlanID, sourceID)
			}
		}
		var current struct {
			ActivePromptRunID *uuid.UUID
			SelectedPlanID    *uuid.UUID
		}
		if err := tx.Raw(`SELECT active_prompt_run_id, selected_plan_id FROM todo_issues WHERE id = ?`, issueID).Scan(&current).Error; err != nil {
			return err
		}
		if !sameUUIDPointer(current.ActivePromptRunID, desired.ActivePromptRunID) || !sameUUIDPointer(current.SelectedPlanID, desired.SelectedPlanID) {
			if err := tx.Exec(`UPDATE todo_issues SET active_prompt_run_id = ?, selected_plan_id = ? WHERE id = ?`, desired.ActivePromptRunID, desired.SelectedPlanID, issueID).Error; err != nil {
				return err
			}
			changed[sourceID] = true
		}
		if desired.ActivePromptRunID != nil {
			var beforeSequence int64
			if err := tx.Raw(`SELECT COALESCE(MAX(sequence), 0) FROM todo_issue_events WHERE issue_id = ?`, issueID).Scan(&beforeSequence).Error; err != nil {
				return err
			}
			if err := projectPromptRun(tx, *desired.ActivePromptRunID); err != nil {
				return err
			}
			var projected []ImportEventKey
			if err := tx.Raw(`
				SELECT source, source_id FROM todo_issue_events
				WHERE issue_id = ? AND sequence > ? AND source_id IS NOT NULL
				ORDER BY sequence`, issueID, beforeSequence).Scan(&projected).Error; err != nil {
				return err
			}
			result.Counts.ProjectionEventsInserted += len(projected)
			result.Rollback.InsertedEvents = append(result.Rollback.InsertedEvents, projected...)
			// Projection correctly timestamps its mutation, but the one-off import's
			// current snapshot timestamp must remain the legacy source timestamp.
			// The projection event retains its own authoritative creation time.
			if err := tx.Exec(`UPDATE todo_issues SET updated_at = ? WHERE id = ?`, desired.UpdatedAt, issueID).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

type checksumEvent struct {
	Sequence  int64           `json:"sequence"`
	Kind      string          `json:"kind"`
	Actor     string          `json:"actor,omitempty"`
	Body      string          `json:"body,omitempty"`
	Payload   json.RawMessage `json:"payload"`
	Source    string          `json:"source"`
	SourceID  string          `json:"sourceId,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
}

type checksumIssue struct {
	Issue      Issue           `json:"issue"`
	Aliases    []Alias         `json:"aliases"`
	Events     []checksumEvent `json:"events"`
	PromptRuns []PromptRunLink `json:"promptRuns"`
	Plans      []PlanLink      `json:"plans"`
}

type checksumWorkspace struct {
	Workspace     Workspace       `json:"workspace"`
	Paths         []WorkspacePath `json:"paths"`
	Issues        []checksumIssue `json:"issues"`
	Relationships []Relationship  `json:"relationships"`
}

func computeImportChecksum(ctx context.Context, tx *gorm.DB, workspaceID uuid.UUID) (string, error) {
	repository, _ := NewRepository(tx)
	workspace, err := repository.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return "", err
	}
	paths, err := repository.ListWorkspacePaths(ctx, workspaceID)
	if err != nil {
		return "", err
	}
	issues, err := repository.ListIssues(ctx, workspaceID)
	if err != nil {
		return "", err
	}
	sort.Slice(issues, func(i, j int) bool { return issues[i].ID.String() < issues[j].ID.String() })
	snapshot := checksumWorkspace{Workspace: *workspace, Paths: paths, Issues: make([]checksumIssue, 0, len(issues)), Relationships: []Relationship{}}
	for _, issue := range issues {
		aliases, err := repository.ListAliases(ctx, issue.ID)
		if err != nil {
			return "", err
		}
		if aliases == nil {
			aliases = []Alias{}
		}
		events, err := repository.ListEvents(ctx, issue.ID)
		if err != nil {
			return "", err
		}
		checksumEvents := make([]checksumEvent, 0, len(events))
		for _, event := range events {
			payload, err := canonicalImportJSON(event.Payload)
			if err != nil {
				return "", err
			}
			checksumEvents = append(checksumEvents, checksumEvent{
				Sequence: event.Sequence, Kind: event.Kind, Actor: event.Actor, Body: event.Body,
				Payload: payload, Source: event.Source, SourceID: event.SourceID, CreatedAt: event.CreatedAt,
			})
		}
		promptRuns, err := repository.ListPromptRuns(ctx, issue.ID)
		if err != nil {
			return "", err
		}
		if promptRuns == nil {
			promptRuns = []PromptRunLink{}
		}
		plans, err := repository.ListPlans(ctx, issue.ID)
		if err != nil {
			return "", err
		}
		if plans == nil {
			plans = []PlanLink{}
		}
		snapshot.Issues = append(snapshot.Issues, checksumIssue{Issue: issue, Aliases: aliases, Events: checksumEvents, PromptRuns: promptRuns, Plans: plans})
	}
	if err := tx.Raw(`
		SELECT workspace_id, issue_id, target_issue_id, relation, created_at
		FROM todo_issue_relationships WHERE workspace_id = ?
		ORDER BY issue_id, target_issue_id, relation`, workspaceID).Scan(&snapshot.Relationships).Error; err != nil {
		return "", err
	}
	if snapshot.Relationships == nil {
		snapshot.Relationships = []Relationship{}
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func sortImportRollback(rollback *ImportRollback) {
	sort.Slice(rollback.CreatedIssueIDs, func(i, j int) bool {
		return rollback.CreatedIssueIDs[i].String() < rollback.CreatedIssueIDs[j].String()
	})
	sort.Slice(rollback.IssuePreimages, func(i, j int) bool {
		return rollback.IssuePreimages[i].Issue.ID.String() < rollback.IssuePreimages[j].Issue.ID.String()
	})
	sort.Slice(rollback.CreatedAliases, func(i, j int) bool { return rollback.CreatedAliases[i].Alias < rollback.CreatedAliases[j].Alias })
	sort.Slice(rollback.InsertedEvents, func(i, j int) bool {
		if rollback.InsertedEvents[i].Source != rollback.InsertedEvents[j].Source {
			return rollback.InsertedEvents[i].Source < rollback.InsertedEvents[j].Source
		}
		return rollback.InsertedEvents[i].SourceID < rollback.InsertedEvents[j].SourceID
	})
	sortRelationships := func(values []Relationship) {
		sort.Slice(values, func(i, j int) bool {
			left := values[i].IssueID.String() + "\x00" + values[i].TargetIssueID.String() + "\x00" + string(values[i].Relation)
			right := values[j].IssueID.String() + "\x00" + values[j].TargetIssueID.String() + "\x00" + string(values[j].Relation)
			return left < right
		})
	}
	sortRelationships(rollback.InsertedRelationships)
	sortRelationships(rollback.DeletedRelationships)
}

func normalizeGriteSourceID(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "")
	if len(value) != 32 {
		return "", fmt.Errorf("%w: full Grite issue ID %q must contain 32 hexadecimal characters", ErrInvalidInput, value)
	}
	if _, err := hex.DecodeString(value); err != nil {
		return "", fmt.Errorf("%w: full Grite issue ID %q is not hexadecimal", ErrInvalidInput, value)
	}
	return value, nil
}

func stableImportTime(value, fallback time.Time) time.Time {
	if value.IsZero() {
		value = fallback
	}
	if value.IsZero() {
		value = time.Unix(0, 0)
	}
	return value.UTC().Truncate(time.Microsecond)
}

func canonicalImportJSON(raw json.RawMessage) (json.RawMessage, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return json.RawMessage(`{}`), nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err == nil {
		return nil, errors.New("multiple JSON values")
	} else if !errors.Is(err, io.EOF) {
		return nil, err
	}
	encoded, err := json.Marshal(value)
	return json.RawMessage(encoded), err
}

func equalImportIssue(left, right ImportIssue) bool {
	return left.ID == right.ID && left.SourceID == right.SourceID && left.Title == right.Title &&
		left.Body == right.Body && left.Verification == right.Verification && slices.Equal(left.Labels, right.Labels) &&
		left.Priority == right.Priority && left.Status == right.Status && left.ExecutionState == right.ExecutionState &&
		sameUUIDPointer(left.ActivePromptRunID, right.ActivePromptRunID) && sameUUIDPointer(left.SelectedPlanID, right.SelectedPlanID) &&
		left.CreatedAt.Equal(right.CreatedAt) && left.UpdatedAt.Equal(right.UpdatedAt)
}

func equalNormalizedImportEvent(left, right normalizedImportEvent) bool {
	return left.Warning == right.Warning && left.IssueSourceID == right.IssueSourceID && left.SourceID == right.SourceID &&
		left.Order == right.Order && left.Kind == right.Kind && left.Actor == right.Actor && left.Body == right.Body && bytes.Equal(left.Payload, right.Payload) &&
		left.CreatedAt.Equal(right.CreatedAt)
}

func importIssueMatches(current *Issue, imported ImportIssue) bool {
	statusMatches := current.Status == imported.Status
	if imported.ActivePromptRunID != nil {
		statusMatches = true
	}
	return current.Title == imported.Title && current.Body == imported.Body && current.Verification == imported.Verification &&
		slices.Equal(current.Labels, imported.Labels) && current.Priority == imported.Priority && statusMatches &&
		current.CreatedAt.Equal(imported.CreatedAt) && current.UpdatedAt.Equal(imported.UpdatedAt)
}

func cloneUUIDPointer(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
