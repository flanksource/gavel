package native

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *Repository) SetAliases(ctx context.Context, issueID uuid.UUID, expectedVersion int64, aliases []AliasInput, actor string) (*Issue, error) {
	normalized, err := normalizeAliases(aliases)
	if err != nil {
		return nil, err
	}
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		locked, err := lockIssue(tx, issueID, expectedVersion)
		if err != nil {
			return err
		}
		if err := tx.Exec(`DELETE FROM todo_issue_aliases WHERE issue_id = ?`, issueID).Error; err != nil {
			return err
		}
		if err := insertAliases(tx, locked.WorkspaceID, issueID, normalized); err != nil {
			return err
		}
		_, err = recordMutation(tx, locked, EventInput{
			Kind:  "aliases_updated",
			Actor: actor,
			Payload: map[string]any{
				"aliases": normalized,
			},
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return r.GetIssue(ctx, issueID)
}

func (r *Repository) ListAliases(ctx context.Context, issueID uuid.UUID) ([]Alias, error) {
	var aliases []Alias
	result := r.db.WithContext(ctx).Raw(`
		SELECT workspace_id, alias, issue_id, COALESCE(kind, '') AS kind, created_at
		FROM todo_issue_aliases WHERE issue_id = ? ORDER BY alias`, issueID,
	).Scan(&aliases)
	return aliases, result.Error
}

// ListAliasesByKind reads one kind of alias for a whole workspace in a single
// query, keyed by issue. Listing decorates every issue with its external links,
// so the per-issue ListAliases would be an N+1 across the entire backlog.
func (r *Repository) ListAliasesByKind(ctx context.Context, workspaceID uuid.UUID, kind string) (map[uuid.UUID][]Alias, error) {
	var aliases []Alias
	result := r.db.WithContext(ctx).Raw(`
		SELECT workspace_id, alias, issue_id, COALESCE(kind, '') AS kind, created_at
		FROM todo_issue_aliases
		WHERE workspace_id = ? AND lower(COALESCE(kind, '')) = lower(?)
		ORDER BY alias`, workspaceID, kind,
	).Scan(&aliases)
	if result.Error != nil {
		return nil, result.Error
	}
	byIssue := make(map[uuid.UUID][]Alias, len(aliases))
	for _, alias := range aliases {
		byIssue[alias.IssueID] = append(byIssue[alias.IssueID], alias)
	}
	return byIssue, nil
}

func (r *Repository) AppendEvent(ctx context.Context, issueID uuid.UUID, expectedVersion int64, input EventInput) (*Event, error) {
	var event *Event
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		locked, err := lockIssue(tx, issueID, expectedVersion)
		if err != nil {
			return err
		}
		event, err = recordMutation(tx, locked, input)
		return err
	})
	return event, err
}

func (r *Repository) AddComment(ctx context.Context, issueID uuid.UUID, expectedVersion int64, actor, body string) (*Event, error) {
	if strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("%w: comment body is required", ErrInvalidInput)
	}
	return r.AppendEvent(ctx, issueID, expectedVersion, EventInput{
		Kind:  "comment",
		Actor: actor,
		Body:  body,
	})
}

func (r *Repository) ListEvents(ctx context.Context, issueID uuid.UUID) ([]Event, error) {
	type eventRecord struct {
		ID        uuid.UUID
		IssueID   uuid.UUID
		Sequence  int64
		Kind      string
		Actor     sql.NullString
		Body      sql.NullString
		Payload   []byte
		Source    string
		SourceID  sql.NullString
		CreatedAt time.Time
	}
	var records []eventRecord
	result := r.db.WithContext(ctx).Raw(`
		SELECT id, issue_id, sequence, kind, actor, body, payload, source, source_id, created_at
		FROM todo_issue_events WHERE issue_id = ? ORDER BY sequence`, issueID,
	).Scan(&records)
	if result.Error != nil {
		return nil, result.Error
	}
	events := make([]Event, 0, len(records))
	for _, record := range records {
		payload := append(json.RawMessage(nil), record.Payload...)
		if len(payload) == 0 {
			payload = json.RawMessage(`{}`)
		}
		events = append(events, Event{
			ID:        record.ID,
			IssueID:   record.IssueID,
			Sequence:  record.Sequence,
			Kind:      record.Kind,
			Actor:     record.Actor.String,
			Body:      record.Body.String,
			Payload:   payload,
			Source:    record.Source,
			SourceID:  record.SourceID.String,
			CreatedAt: record.CreatedAt,
		})
	}
	return events, nil
}

func insertAliases(tx *gorm.DB, workspaceID, issueID uuid.UUID, aliases []AliasInput) error {
	for _, alias := range aliases {
		result := tx.Exec(`
			INSERT INTO todo_issue_aliases (workspace_id, alias, issue_id, kind, created_at)
			VALUES (?, ?, ?, NULLIF(?, ''), now())`,
			workspaceID, alias.Alias, issueID, alias.Kind,
		)
		if result.Error != nil {
			return mapUniqueError(result.Error, ErrAliasConflict, "workspace %s alias %q", workspaceID, alias.Alias)
		}
	}
	return nil
}

func recordMutation(tx *gorm.DB, issue *lockedIssue, input EventInput) (*Event, error) {
	kind := normalizeToken(input.Kind)
	if kind == "" {
		return nil, fmt.Errorf("%w: event kind is required", ErrInvalidInput)
	}
	var sequence int64
	if err := tx.Raw(`
		SELECT COALESCE(MAX(sequence), 0) + 1
		FROM todo_issue_events WHERE issue_id = ?`, issue.ID,
	).Scan(&sequence).Error; err != nil {
		return nil, err
	}
	nextVersion := issue.Version + 1
	result := tx.Exec(`
		UPDATE todo_issues SET version = ?, updated_at = now()
		WHERE id = ? AND version = ?`, nextVersion, issue.ID, issue.Version,
	)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected != 1 {
		return nil, fmt.Errorf("%w: issue %s expected version %d", ErrVersionConflict, issue.ID, issue.Version)
	}
	event, err := insertEvent(tx, issue.ID, sequence, input)
	if err != nil {
		return nil, err
	}
	issue.Version = nextVersion
	return event, nil
}

func insertEvent(tx *gorm.DB, issueID uuid.UUID, sequence int64, input EventInput) (*Event, error) {
	kind := normalizeToken(input.Kind)
	if kind == "" {
		return nil, fmt.Errorf("%w: event kind is required", ErrInvalidInput)
	}
	payload, err := marshalPayload(input.Payload)
	if err != nil {
		return nil, err
	}
	source := normalizeToken(input.Source)
	if source == "" {
		source = "gavel"
	}

	type eventRecord struct {
		ID        uuid.UUID
		IssueID   uuid.UUID
		Sequence  int64
		Kind      string
		Actor     sql.NullString
		Body      sql.NullString
		Payload   []byte
		Source    string
		SourceID  sql.NullString
		CreatedAt time.Time
	}
	var record eventRecord
	result := tx.Raw(`
		INSERT INTO todo_issue_events
			(id, issue_id, sequence, kind, actor, body, payload, source, source_id, created_at)
		VALUES (?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), CAST(? AS jsonb), ?, NULLIF(?, ''), now())
		RETURNING id, issue_id, sequence, kind, actor, body, payload, source, source_id, created_at`,
		uuid.New(), issueID, sequence, kind, strings.TrimSpace(input.Actor), input.Body,
		string(payload), source, strings.TrimSpace(input.SourceID),
	).Scan(&record)
	if result.Error != nil {
		return nil, mapUniqueError(result.Error, ErrEventConflict, "event source %q/%q", source, input.SourceID)
	}
	return &Event{
		ID:        record.ID,
		IssueID:   record.IssueID,
		Sequence:  record.Sequence,
		Kind:      record.Kind,
		Actor:     record.Actor.String,
		Body:      record.Body.String,
		Payload:   append(json.RawMessage(nil), record.Payload...),
		Source:    record.Source,
		SourceID:  record.SourceID.String,
		CreatedAt: record.CreatedAt,
	}, nil
}

func marshalPayload(value any) ([]byte, error) {
	if value == nil {
		return []byte(`{}`), nil
	}
	switch payload := value.(type) {
	case json.RawMessage:
		if !json.Valid(payload) {
			return nil, fmt.Errorf("%w: event payload is not valid JSON", ErrInvalidInput)
		}
		return payload, nil
	case []byte:
		if !json.Valid(payload) {
			return nil, fmt.Errorf("%w: event payload is not valid JSON", ErrInvalidInput)
		}
		return payload, nil
	default:
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("encode native todo event payload: %w", err)
		}
		return encoded, nil
	}
}

func normalizeAliases(inputs []AliasInput) ([]AliasInput, error) {
	byAlias := make(map[string]AliasInput, len(inputs))
	for _, input := range inputs {
		alias := normalizeToken(input.Alias)
		if alias == "" {
			return nil, fmt.Errorf("%w: alias cannot be empty", ErrInvalidInput)
		}
		kind := normalizeToken(input.Kind)
		if existing, ok := byAlias[alias]; ok && existing.Kind != kind {
			return nil, fmt.Errorf("%w: alias %q has conflicting kinds", ErrInvalidInput, alias)
		}
		byAlias[alias] = AliasInput{Alias: alias, Kind: kind}
	}
	aliases := make([]AliasInput, 0, len(byAlias))
	for _, alias := range byAlias {
		aliases = append(aliases, alias)
	}
	sort.Slice(aliases, func(i, j int) bool { return aliases[i].Alias < aliases[j].Alias })
	return aliases, nil
}
