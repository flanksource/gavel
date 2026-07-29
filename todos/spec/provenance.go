package spec

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/flanksource/captain/pkg/api"
)

// fold merges every layer in order and records, for each leaf of the result,
// the layer that last set it.
//
// Attribution is derived by diffing the accumulator's JSON projection after each
// merge rather than by inspecting the layer's own fields: a layer's value only
// counts as its contribution if it actually survived the merge, and captain's
// merge policy — not this package — decides that.
func fold(layers []Layer) (api.Spec, map[string]string) {
	var acc api.Spec
	before := map[string]string{}
	provenance := map[string]string{}
	for _, layer := range layers {
		acc = acc.Merge(layer.Spec)
		after := flatten(acc)
		for path, value := range after {
			if before[path] != value {
				provenance[path] = layer.Name
			}
		}
		before = after
	}
	return acc, provenance
}

// flatten projects a spec to dotted leaf paths ("model.name", "budget.cost",
// "permissions.tools.allow.0"). It goes through the spec's own JSON encoding, so
// the paths match the wire shape the settings trace endpoint serves and a field
// added to api.Spec appears without touching this file.
//
// A marshalling failure is not silently swallowed: the sentinel path makes the
// trace visibly wrong rather than quietly empty.
func flatten(v any) map[string]string {
	raw, err := json.Marshal(v)
	if err != nil {
		return map[string]string{"": fmt.Sprintf("unencodable spec: %v", err)}
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return map[string]string{"": fmt.Sprintf("unencodable spec: %v", err)}
	}
	out := map[string]string{}
	walk("", decoded, out)
	return out
}

func walk(prefix string, v any, out map[string]string) {
	switch value := v.(type) {
	case map[string]any:
		for key, child := range value {
			walk(join(prefix, key), child, out)
		}
	case []any:
		for i, child := range value {
			walk(join(prefix, strconv.Itoa(i)), child, out)
		}
	case nil:
		// An explicit null carries no value to attribute.
	default:
		out[prefix] = fmt.Sprint(value)
	}
}

func join(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}
