package types

import (
	"fmt"
	"strings"
)

// RelationKind is a directed link between two TODOs.
type RelationKind string

const (
	// RelationDependsOn means this TODO is blocked until the target reaches a
	// satisfied status (verified or completed).
	RelationDependsOn RelationKind = "depends_on"
	// RelationRelatedTo is a symmetric, non-blocking association — the relation
	// to use when two TODOs overlap, duplicate, or should be read together.
	RelationRelatedTo RelationKind = "related_to"
	// RelationBlocks is the read-only reverse of depends_on: it appears when
	// listing links but can never be written directly.
	RelationBlocks RelationKind = "blocks"
)

// LinkableRelations returns the relations a caller may create or delete.
func LinkableRelations() []RelationKind {
	return []RelationKind{RelationDependsOn, RelationRelatedTo}
}

// ParseRelationKind normalizes CLI and API spellings ("depends-on",
// "depends_on") to a writable relation. An empty value means related_to, the
// non-blocking default. The derived blocks relation is rejected: write the
// depends_on edge from the blocked TODO instead.
func ParseRelationKind(value string) (RelationKind, error) {
	normalized := RelationKind(strings.ReplaceAll(strings.ToLower(strings.TrimSpace(value)), "-", "_"))
	if normalized == "" {
		return RelationRelatedTo, nil
	}
	for _, linkable := range LinkableRelations() {
		if normalized == linkable {
			return normalized, nil
		}
	}
	if normalized == RelationBlocks {
		return "", fmt.Errorf("relation %q is derived from depends_on and cannot be written; add depends_on from the blocked TODO instead", value)
	}
	return "", fmt.Errorf("unknown relation %q; valid relations: %s, %s", value, RelationDependsOn, RelationRelatedTo)
}
