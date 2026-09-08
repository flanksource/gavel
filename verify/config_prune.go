package verify

import (
	"bytes"
	"encoding/json"
)

// pruneEmptyConfigNodes drops every key that carries no configuration from an
// encoded config document.
//
// It exists because `omitempty` cannot express this. encoding/json applies
// omitempty to strings, numbers, pointers, maps and slices but never to a
// struct value, so a GavelConfig the user never touched still serializes
// `checks: {}`, `fixtures: {}`, `commit: {lint: {}, tidy: {}, ...}`. Those keys
// configure nothing, and every SaveGavelConfig call scattered more of them into
// the repo's .gavel.yaml. They are also a liability: once a field is renamed or
// removed, the dead key it left behind starts warning as an unknown field in a
// file the user never edited by hand.
//
// The walk is bottom-up, so one pass reaches a fixed point: an object is tested
// for emptiness only after its own children have been pruned, which collapses
// `commit: {lint: {}}` entirely rather than leaving a hollow parent.
func pruneEmptyConfigNodes(encoded []byte) ([]byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	// Numbers stay json.Number so a re-marshal reproduces the literal the config
	// carried; decoding into float64 would rewrite large integers in exponent
	// form and quietly lose precision past 2^53.
	decoder.UseNumber()
	var tree any
	if err := decoder.Decode(&tree); err != nil {
		return nil, err
	}
	pruned, _ := pruneNode(tree)
	return json.Marshal(pruned)
}

// pruneNode returns node with its empty descendants removed, and reports
// whether what remains is itself empty.
//
// Only null, the empty object and the empty list count as empty. false, 0 and
// "" are values a user can mean — commit.lint.secrets: false is the documented
// way to turn the secrets linter off — so they survive.
func pruneNode(node any) (any, bool) {
	switch typed := node.(type) {
	case map[string]any:
		for key, value := range typed {
			child, empty := pruneNode(value)
			if empty {
				delete(typed, key)
				continue
			}
			typed[key] = child
		}
		return typed, len(typed) == 0
	case []any:
		// List entries are cleaned but never dropped: position is meaningful in
		// hook and ignore lists, so removing an entry would shift the rest.
		for i, value := range typed {
			child, _ := pruneNode(value)
			typed[i] = child
		}
		return typed, len(typed) == 0
	case nil:
		return nil, true
	default:
		return node, false
	}
}
