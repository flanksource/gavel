package labels

import "strings"

// Split flattens a label list written as slice entries, comma-separated tokens,
// or both, trimming each token and dropping case-insensitive duplicates. It is
// what a CLI flag, an API body, and an agent's structured result all pass
// through, so `--add bug,api` and `--add bug --add API` reach Apply identically.
//
// Tokens keep the case they were written in: the storage layer normalizes what
// it persists, and lowercasing here would make an error message disagree with
// the flag the caller typed.
func Split(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" && !Contains(out, trimmed) {
				out = append(out, trimmed)
			}
		}
	}
	return out
}

// Contains reports case-insensitive membership, ignoring surrounding space.
func Contains(haystack []string, needle string) bool {
	for _, value := range haystack {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(needle)) {
			return true
		}
	}
	return false
}

// Apply merges an add/remove delta onto a label set and returns the result.
//
// It exists because every label write in gavel replaces the whole set — the
// provider has no "append one label" operation — so "add area:ui" has to be
// computed against each TODO's own labels. A bulk action over forty TODOs and a
// triage verdict on one both need exactly this, and a second copy of it would be
// the thing that silently flattens forty different label sets into one.
//
// Removals are applied first, so a label named in both add and remove survives;
// callers that consider that a contradiction reject it before calling here.
// Existing entries keep their order and their stored spelling.
func Apply(existing, add, remove []string) []string {
	add, remove = Split(add), Split(remove)
	next := make([]string, 0, len(existing)+len(add))
	for _, label := range existing {
		if !Contains(remove, label) {
			next = append(next, label)
		}
	}
	for _, label := range add {
		if !Contains(next, label) {
			next = append(next, label)
		}
	}
	return next
}

// Equal reports whether two label sets carry the same labels in the same order,
// which is how a caller tells a delta that changed nothing from one that did.
// An empty and a nil set are equal: both mean "no labels".
func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !strings.EqualFold(strings.TrimSpace(a[i]), strings.TrimSpace(b[i])) {
			return false
		}
	}
	return true
}
