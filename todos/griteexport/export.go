// Package griteexport models the JSON files produced by `grite export`.
//
// Grite writes a full issue snapshot together with an incremental event delta.
// Event payloads deliberately remain raw JSON: consumers can project the kinds
// they understand without losing newer payload fields during an import.
package griteexport

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// ID is a Grite identifier. Current exports encode IDs as strings, while some
// Grite command responses encode 128-bit IDs as arrays of 16 bytes.
type ID string

func (id ID) String() string { return string(id) }

func (id ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(id))
}

func (id *ID) UnmarshalJSON(data []byte) error {
	if id == nil {
		return errors.New("decode Grite ID into nil receiver")
	}
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("Grite ID cannot be null")
	}

	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*id = ID(text)
		return nil
	}

	var rawBytes []int
	if err := json.Unmarshal(data, &rawBytes); err != nil {
		return fmt.Errorf("Grite ID must be a string or 16-byte array: %w", err)
	}
	if len(rawBytes) != 16 {
		return fmt.Errorf("Grite byte-array ID has %d bytes, want 16", len(rawBytes))
	}
	decoded := make([]byte, len(rawBytes))
	for i, value := range rawBytes {
		if value < 0 || value > 255 {
			return fmt.Errorf("Grite byte-array ID byte %d is outside 0..255: %d", i, value)
		}
		decoded[i] = byte(value)
	}
	*id = ID(hex.EncodeToString(decoded))
	return nil
}

// Meta describes the export file, not the source watermark. GeneratedTS is the
// time the file was generated and EventCount counts only the exported delta.
type Meta struct {
	SchemaVersion int   `json:"schema_version"`
	GeneratedTS   int64 `json:"generated_ts"`
	EventCount    int   `json:"event_count"`
}

// Issue is one row from the full issue snapshot included in every export.
type Issue struct {
	IssueID      ID       `json:"issue_id"`
	Title        string   `json:"title"`
	State        string   `json:"state"`
	Labels       []string `json:"labels"`
	Assignees    []ID     `json:"assignees"`
	CreatedTS    int64    `json:"created_ts"`
	UpdatedTS    int64    `json:"updated_ts"`
	CommentCount int      `json:"comment_count"`
}

// Kind is the raw, single-key tagged union used by Grite events.
type Kind map[string]json.RawMessage

func (kind *Kind) UnmarshalJSON(data []byte) error {
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		return fmt.Errorf("decode Grite event kind: %w", err)
	}
	if len(decoded) != 1 {
		return fmt.Errorf("Grite event kind has %d keys, want exactly one", len(decoded))
	}
	cloned := make(Kind, 1)
	for name, payload := range decoded {
		cloned[name] = append(json.RawMessage(nil), payload...)
	}
	*kind = cloned
	return nil
}

// NamePayload returns the event kind name and an owned copy of its raw payload.
func (kind Kind) NamePayload() (string, json.RawMessage, error) {
	if len(kind) != 1 {
		return "", nil, fmt.Errorf("Grite event kind has %d keys, want exactly one", len(kind))
	}
	for name, payload := range kind {
		if name == "" {
			return "", nil, errors.New("Grite event kind name is empty")
		}
		if !json.Valid(payload) {
			return "", nil, fmt.Errorf("Grite event kind %q has invalid JSON payload", name)
		}
		return name, append(json.RawMessage(nil), payload...), nil
	}
	panic("unreachable")
}

// Event is one append-only Grite event. Kind remains raw so unknown event kinds
// and payload fields round-trip without being discarded.
type Event struct {
	EventID     ID     `json:"event_id"`
	IssueID     ID     `json:"issue_id"`
	Actor       string `json:"actor"`
	TimestampMS int64  `json:"ts_unix_ms"`
	Parent      *ID    `json:"parent,omitempty"`
	Kind        Kind   `json:"kind"`
}

func (event Event) validate() error {
	if event.EventID == "" {
		return errors.New("Grite event ID is empty")
	}
	if event.IssueID == "" {
		return fmt.Errorf("Grite event %s has an empty issue ID", event.EventID)
	}
	if _, _, err := event.Kind.NamePayload(); err != nil {
		return fmt.Errorf("Grite event %s: %w", event.EventID, err)
	}
	return nil
}

// DependencyPayload is shared by DependencyAdded and DependencyRemoved. Target
// accepts both of Grite's observed ID encodings through ID.UnmarshalJSON.
type DependencyPayload struct {
	DepType string `json:"dep_type"`
	Target  ID     `json:"target"`
}

// File is the on-disk JSON export: a full issue snapshot and an event delta.
type File struct {
	Meta   Meta    `json:"meta"`
	Issues []Issue `json:"issues"`
	Events []Event `json:"events"`
}

// Snapshot is the public name for a decoded export. File remains available to
// make DecodeFile and LoadFile read naturally at call sites.
type Snapshot = File

// Result is the data payload printed by `grite export --json`.
type Result struct {
	Format     string          `json:"format"`
	OutputPath string          `json:"output_path"`
	WALHead    json.RawMessage `json:"wal_head"`
	EventCount int             `json:"event_count"`
}

// DecodeFile decodes and validates an on-disk Grite export JSON document.
func DecodeFile(data []byte) (File, error) {
	var file File
	if err := json.Unmarshal(data, &file); err != nil {
		return File{}, fmt.Errorf("decode Grite export: %w", err)
	}
	for i, issue := range file.Issues {
		if issue.IssueID == "" {
			return File{}, fmt.Errorf("issues[%d]: Grite issue ID is empty", i)
		}
	}
	for i, event := range file.Events {
		if err := event.validate(); err != nil {
			return File{}, fmt.Errorf("events[%d]: %w", i, err)
		}
	}
	return file, nil
}

// LoadFile reads and decodes an on-disk Grite export JSON document.
func LoadFile(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, fmt.Errorf("read Grite export %s: %w", path, err)
	}
	file, err := DecodeFile(data)
	if err != nil {
		return File{}, fmt.Errorf("decode Grite export %s: %w", path, err)
	}
	return file, nil
}

// Merge returns the stable union of prior and incoming events keyed by
// event_id. Source order is authoritative for events sharing a millisecond;
// prior order is retained and novel delta events are appended in export order.
// Repeated identical events are idempotent; reusing an event ID for different
// content is rejected instead of silently corrupting history.
func Merge(prior, incoming []Event) ([]Event, error) {
	canonical := make(map[ID][]byte, len(prior)+len(incoming))
	merged := make([]Event, 0, len(prior)+len(incoming))
	for _, batch := range [][]Event{prior, incoming} {
		for _, event := range batch {
			if err := event.validate(); err != nil {
				return nil, err
			}
			encoded, err := json.Marshal(event)
			if err != nil {
				return nil, fmt.Errorf("canonicalize Grite event %s: %w", event.EventID, err)
			}
			if existing, ok := canonical[event.EventID]; ok {
				if !bytes.Equal(existing, encoded) {
					return nil, fmt.Errorf("conflicting Grite event content for event_id %s", event.EventID)
				}
				continue
			}
			canonical[event.EventID] = encoded
			merged = append(merged, cloneEvent(event))
		}
	}
	return merged, nil
}

func cloneEvent(event Event) Event {
	cloned := event
	if event.Parent != nil {
		parent := *event.Parent
		cloned.Parent = &parent
	}
	cloned.Kind = make(Kind, len(event.Kind))
	for name, payload := range event.Kind {
		cloned.Kind[name] = append(json.RawMessage(nil), payload...)
	}
	return cloned
}

// Watermark records the greatest source-event timestamp observed and every ID
// observed at that boundary. Event IDs remain the idempotency key; the timestamp
// only bounds the next export request.
type Watermark struct {
	TimestampMS int64 `json:"timestamp_ms"`
	EventIDs    []ID  `json:"event_ids,omitempty"`
}

// WatermarkFor computes a deterministic watermark for events.
func WatermarkFor(events []Event) Watermark {
	var watermark Watermark
	return watermark.Advance(events)
}

// Advance incorporates a new event delta into the watermark.
func (watermark Watermark) Advance(events []Event) Watermark {
	ids := make(map[ID]struct{}, len(watermark.EventIDs))
	for _, id := range watermark.EventIDs {
		if id != "" {
			ids[id] = struct{}{}
		}
	}
	for _, event := range events {
		switch {
		case event.TimestampMS > watermark.TimestampMS:
			watermark.TimestampMS = event.TimestampMS
			ids = map[ID]struct{}{}
			if event.EventID != "" {
				ids[event.EventID] = struct{}{}
			}
		case event.TimestampMS == watermark.TimestampMS && event.EventID != "":
			ids[event.EventID] = struct{}{}
		}
	}
	watermark.EventIDs = make([]ID, 0, len(ids))
	for id := range ids {
		watermark.EventIDs = append(watermark.EventIDs, id)
	}
	sort.Slice(watermark.EventIDs, func(i, j int) bool { return watermark.EventIDs[i] < watermark.EventIDs[j] })
	return watermark
}

// SinceMS returns a one-millisecond-overlap-safe cursor for the next export.
// Replayed boundary events must be deduplicated by event_id with Merge.
func (watermark Watermark) SinceMS() int64 {
	if watermark.TimestampMS <= 1 {
		return 0
	}
	return watermark.TimestampMS - 1
}

// ContainsBoundaryEvent reports whether an event at the watermark timestamp was
// already observed. Events before the watermark are outside this boundary set.
func (watermark Watermark) ContainsBoundaryEvent(event Event) bool {
	if event.TimestampMS != watermark.TimestampMS {
		return false
	}
	for _, id := range watermark.EventIDs {
		if id == event.EventID {
			return true
		}
	}
	return false
}
